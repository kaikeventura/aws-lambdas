# -*- coding: utf-8 -*-
"""
AWS Lambda function for user authentication via API Gateway.

Handles user signup, signin, and token authentication using
DynamoDB for persistence and JWT for session management.
"""
import json
import logging
import os
from datetime import datetime, timedelta, timezone

import boto3
import bcrypt
import jwt
from botocore.exceptions import ClientError

# ===================================================================
# Configuration
# ===================================================================
logging.basicConfig(level=logging.INFO)
LOGGER = logging.getLogger(__name__)

try:
    TABLE_NAME = os.environ["TABLE_NAME"]
    JWT_SECRET = os.environ["JWT_SECRET"]
    DDB_RESOURCE = boto3.resource("dynamodb")
    USER_TABLE = DDB_RESOURCE.Table(TABLE_NAME)
except KeyError as e:
    LOGGER.error("Missing environment variable: %s", e)
    # This will cause the Lambda container to fail on init, which is the desired behavior.
    raise

# ===================================================================
# Helper Functions
# ===================================================================

def _json_response(status_code, body):
    """Constructs a valid API Gateway proxy response."""
    return {
        "statusCode": status_code,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps(body),
    }

def _hash_password(password):
    """Hashes a password using bcrypt."""
    return bcrypt.hashpw(password.encode("utf-8"), bcrypt.gensalt())

def _check_password(password, hashed):
    """Checks a password against a bcrypt hash."""
    return bcrypt.checkpw(password.encode("utf-8"), hashed.encode("utf-8"))

# ===================================================================
# Route Handlers
# ===================================================================

def _signup(body):
    """
    Handles new user registration.
    POST /signup
    """
    try:
        email = body["email"]
        password = body["password"]
        name = body["name"]
    except KeyError:
        return _json_response(400, {"error": "Missing required fields (email, password, name)."})

    # Check if user already exists
    try:
        response = USER_TABLE.get_item(Key={"email": email})
        if "Item" in response:
            return _json_response(409, {"error": "User with this email already exists."})
    except ClientError as e:
        LOGGER.error("DynamoDB error on get_item: %s", e)
        return _json_response(500, {"error": "Internal server error."})

    # Create and save user
    hashed_password = _hash_password(password).decode("utf-8")
    timestamp = datetime.utcnow().isoformat()
    new_user = {
        "email": email,
        "password": hashed_password,
        "name": name,
        "createdAt": timestamp,
        "updatedAt": timestamp,
    }

    try:
        USER_TABLE.put_item(Item=new_user)
        return _json_response(201, {"message": "User created successfully."})
    except ClientError as e:
        LOGGER.error("DynamoDB error on put_item: %s", e)
        return _json_response(500, {"error": "Internal server error."})


def _signin(body):
    """
    Handles user authentication and JWT generation.
    POST /signin
    """
    try:
        email = body["email"]
        password = body["password"]
    except KeyError:
        return _json_response(400, {"error": "Missing required fields (email, password)."})

    # Fetch user
    try:
        response = USER_TABLE.get_item(Key={"email": email})
        user = response.get("Item")
        if not user:
            # Use a generic error message to avoid user enumeration
            return _json_response(401, {"error": "Invalid credentials."})
    except ClientError as e:
        LOGGER.error("DynamoDB error on get_item: %s", e)
        return _json_response(500, {"error": "Internal server error."})

    # Verify password
    if not _check_password(password, user["password"]):
        return _json_response(401, {"error": "Invalid credentials."})

    # Generate JWT
    payload = {
        "sub": user["email"],
        "exp": datetime.now(timezone.utc) + timedelta(hours=1),
    }
    token = jwt.encode(payload, JWT_SECRET, algorithm="HS256")

    return _json_response(200, {"token": token})


def _authenticate(headers):
    """
    Validates a JWT from the Authorization header.
    POST /authentication
    """
    auth_header = headers.get("Authorization")
    if not auth_header or not auth_header.startswith("Bearer "):
        return _json_response(401, {"error": "Authorization header is missing or malformed."})

    token = auth_header.split(" ")[1]

    try:
        # PyJWT automatically validates signature and expiration
        jwt.decode(token, JWT_SECRET, algorithms=["HS256"])
        return _json_response(200, {"message": "Token is valid."})
    except jwt.ExpiredSignatureError:
        return _json_response(401, {"error": "Token has expired."})
    except jwt.InvalidTokenError:
        return _json_response(401, {"error": "Token is invalid."})


# ===================================================================
# Main Lambda Handler
# ===================================================================

def lambda_handler(event, context):
    """
    Main entry point for the Lambda function.
    Routes requests based on HTTP method and path.
    """
    LOGGER.info("Received event: %s", event)
    
    http_method = event.get("httpMethod")
    path = event.get("path")

    try:
        body = json.loads(event.get("body") or "{}")
    except json.JSONDecodeError:
        return _json_response(400, {"error": "Invalid JSON in request body."})

    # Simple router
    routes = {
        ("POST", "/signup"): lambda: _signup(body),
        ("POST", "/signin"): lambda: _signin(body),
        ("POST", "/authentication"): lambda: _authenticate(event.get("headers", {})),
    }

    handler = routes.get((http_method, path))

    if handler:
        return handler()
    
    return _json_response(404, {"error": "Not Found"})

