import { DynamoDBClient } from "@aws-sdk/client-dynamodb";
import { DynamoDBDocumentClient, GetCommand, PutCommand } from "@aws-sdk/lib-dynamodb";
import bcrypt from "bcryptjs";
import jwt from "jsonwebtoken";

// ===================================================================
// Configuration
// ===================================================================

let docClient;

const getDocClient = () => {
  if (!docClient) {
    const client = new DynamoDBClient({});
    docClient = DynamoDBDocumentClient.from(client);
  }
  return docClient;
};

const getTableName = () => process.env.TABLE_NAME;
const getJwtSecret = () => process.env.JWT_SECRET;

// ===================================================================
// Helper Functions
// ===================================================================

const jsonResponse = (statusCode, body) => {
  return {
    statusCode,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  };
};

const hashPassword = async (password) => {
  return await bcrypt.hash(password, 10);
};

const checkPassword = async (password, hash) => {
  return await bcrypt.compare(password, hash);
};

// ===================================================================
// Route Handlers
// ===================================================================

const signup = async (body) => {
  const { email, password, name } = body;

  if (!email || !password || !name) {
    return jsonResponse(400, { error: "Missing required fields (email, password, name)." });
  }

  const client = getDocClient();
  const tableName = getTableName();

  // Check if user already exists
  try {
    const command = new GetCommand({
      TableName: tableName,
      Key: { email },
    });
    const response = await client.send(command);
    if (response.Item) {
      return jsonResponse(409, { error: "User with this email already exists." });
    }
  } catch (error) {
    console.error("DynamoDB error on GetItem:", error);
    return jsonResponse(500, { error: "Internal server error." });
  }

  // Create and save user
  try {
    const hashedPassword = await hashPassword(password);
    const timestamp = new Date().toISOString();
    const newUser = {
      email,
      password: hashedPassword,
      name,
      createdAt: timestamp,
      updatedAt: timestamp,
    };

    const command = new PutCommand({
      TableName: tableName,
      Item: newUser,
    });
    await client.send(command);
    return jsonResponse(201, { message: "User created successfully." });
  } catch (error) {
    console.error("DynamoDB error on PutItem:", error);
    return jsonResponse(500, { error: "Internal server error." });
  }
};

const signin = async (body) => {
  const { email, password } = body;

  if (!email || !password) {
    return jsonResponse(400, { error: "Missing required fields (email, password)." });
  }

  const client = getDocClient();
  const tableName = getTableName();

  // Fetch user
  let user;
  try {
    const command = new GetCommand({
      TableName: tableName,
      Key: { email },
    });
    const response = await client.send(command);
    user = response.Item;
    if (!user) {
      return jsonResponse(401, { error: "Invalid credentials." });
    }
  } catch (error) {
    console.error("DynamoDB error on GetItem:", error);
    return jsonResponse(500, { error: "Internal server error." });
  }

  // Verify password
  if (!(await checkPassword(password, user.password))) {
    return jsonResponse(401, { error: "Invalid credentials." });
  }

  // Generate JWT
  const token = jwt.sign({ sub: user.email }, getJwtSecret(), { expiresIn: "1h" });

  return jsonResponse(200, { token });
};

const authenticate = async (headers) => {
  let authHeader = headers.Authorization || headers.authorization;

  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    return jsonResponse(401, { error: "Authorization header is missing or malformed." });
  }

  const token = authHeader.split(" ")[1];

  try {
    jwt.verify(token, getJwtSecret());
    return jsonResponse(200, { message: "Token is valid." });
  } catch (error) {
    if (error.name === "TokenExpiredError") {
      return jsonResponse(401, { error: "Token has expired." });
    }
    return jsonResponse(401, { error: "Token is invalid." });
  }
};

// ===================================================================
// Main Lambda Handler
// ===================================================================

export const handler = async (event) => {
  console.log("Received event:", JSON.stringify(event, null, 2));

  const httpMethod = event.httpMethod;
  const path = event.path;
  let body = {};

  try {
    if (event.body) {
      body = JSON.parse(event.body);
    }
  } catch (error) {
    return jsonResponse(400, { error: "Invalid JSON in request body." });
  }

  if (httpMethod === "POST" && path === "/signup") {
    return await signup(body);
  } else if (httpMethod === "POST" && path === "/signin") {
    return await signin(body);
  } else if (httpMethod === "POST" && path === "/authentication") {
    return await authenticate(event.headers || {});
  }

  return jsonResponse(404, { error: "Not Found" });
};
