use aws_config::BehaviorVersion;
use aws_sdk_dynamodb::{types::AttributeValue, Client};
use bcrypt::{hash, verify, DEFAULT_COST};
use chrono::{Duration, Utc};
use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
use lambda_runtime::{Error, LambdaEvent};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::env;

#[derive(Deserialize, Serialize, Debug)]
pub struct User {
    pub email: String,
    pub password: String,
    pub name: String,
    #[serde(rename = "createdAt")]
    pub created_at: String,
    #[serde(rename = "updatedAt")]
    pub updated_at: String,
}

#[derive(Deserialize)]
pub struct SignupRequest {
    pub email: String,
    pub password: String,
    pub name: String,
}

#[derive(Deserialize)]
pub struct SigninRequest {
    pub email: String,
    pub password: String,
}

#[derive(Serialize, Deserialize)]
pub struct Claims {
    pub sub: String,
    pub exp: usize,
}

pub async fn get_dynamodb_client() -> (Client, String) {
    let config = aws_config::load_defaults(BehaviorVersion::latest()).await;
    let client = Client::new(&config);
    let table_name = env::var("TABLE_NAME").expect("TABLE_NAME must be set");
    (client, table_name)
}

pub fn json_response(status_code: i32, body: Value) -> Result<Value, Error> {
    Ok(json!({
        "statusCode": status_code,
        "headers": { "Content-Type": "application/json" },
        "body": body.to_string()
    }))
}

pub async fn signup(body: &str, client: &Client, table_name: &str) -> Result<Value, Error> {
    let req: SignupRequest = match serde_json::from_str(body) {
        Ok(r) => r,
        Err(_) => return json_response(400, json!({ "error": "Invalid JSON or missing fields" })),
    };

    let result = client
        .get_item()
        .table_name(table_name)
        .key("email", AttributeValue::S(req.email.clone()))
        .send()
        .await;

    match result {
        Ok(output) => {
            if output.item.is_some() {
                return json_response(409, json!({ "error": "User with this email already exists." }));
            }
        }
        Err(e) => {
            eprintln!("DynamoDB error: {:?}", e);
            return json_response(500, json!({ "error": "Internal server error." }));
        }
    }

    let hashed_password = match hash(&req.password, DEFAULT_COST) {
        Ok(h) => h,
        Err(_) => return json_response(500, json!({ "error": "Internal server error." })),
    };

    let now = Utc::now().to_rfc3339();
    let new_user = User {
        email: req.email.clone(),
        password: hashed_password,
        name: req.name,
        created_at: now.clone(),
        updated_at: now,
    };

    let put_result = client
        .put_item()
        .table_name(table_name)
        .item("email", AttributeValue::S(new_user.email))
        .item("password", AttributeValue::S(new_user.password))
        .item("name", AttributeValue::S(new_user.name))
        .item("createdAt", AttributeValue::S(new_user.created_at))
        .item("updatedAt", AttributeValue::S(new_user.updated_at))
        .send()
        .await;

    match put_result {
        Ok(_) => json_response(201, json!({ "message": "User created successfully." })),
        Err(e) => {
            eprintln!("DynamoDB error: {:?}", e);
            json_response(500, json!({ "error": "Internal server error." }))
        }
    }
}

pub async fn signin(body: &str, client: &Client, table_name: &str, jwt_secret: &str) -> Result<Value, Error> {
    let req: SigninRequest = match serde_json::from_str(body) {
        Ok(r) => r,
        Err(_) => return json_response(400, json!({ "error": "Invalid JSON or missing fields" })),
    };

    let result = client
        .get_item()
        .table_name(table_name)
        .key("email", AttributeValue::S(req.email.clone()))
        .send()
        .await;

    let item = match result {
        Ok(output) => match output.item {
            Some(i) => i,
            None => return json_response(401, json!({ "error": "Invalid credentials." })),
        },
        Err(e) => {
            eprintln!("DynamoDB error: {:?}", e);
            return json_response(500, json!({ "error": "Internal server error." }));
        }
    };

    let stored_password = item.get("password").and_then(|av| av.as_s().ok());

    if let Some(pwd) = stored_password {
        match verify(&req.password, pwd) {
            Ok(valid) => {
                if !valid {
                    return json_response(401, json!({ "error": "Invalid credentials." }));
                }
            }
            Err(_) => return json_response(500, json!({ "error": "Internal server error." })),
        }
    } else {
        return json_response(500, json!({ "error": "Internal server error." }));
    }

    let expiration = Utc::now()
        .checked_add_signed(Duration::hours(1))
        .expect("valid timestamp")
        .timestamp() as usize;

    let claims = Claims {
        sub: req.email,
        exp: expiration,
    };

    let token = match encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(jwt_secret.as_bytes()),
    ) {
        Ok(t) => t,
        Err(_) => return json_response(500, json!({ "error": "Internal server error." })),
    };

    json_response(200, json!({ "token": token }))
}

pub fn authenticate(headers: &Value, jwt_secret: &str) -> Result<Value, Error> {
    let auth_header = headers
        .get("Authorization")
        .or_else(|| headers.get("authorization"))
        .and_then(|v| v.as_str());

    let token = match auth_header {
        Some(h) if h.starts_with("Bearer ") => &h[7..],
        _ => return json_response(401, json!({ "error": "Authorization header is missing or malformed." })),
    };

    let validation = Validation::default();
    match decode::<Claims>(
        token,
        &DecodingKey::from_secret(jwt_secret.as_bytes()),
        &validation,
    ) {
        Ok(_) => json_response(200, json!({ "message": "Token is valid." })),
        Err(_) => json_response(401, json!({ "error": "Token is invalid or expired." })),
    }
}

pub async fn function_handler(event: LambdaEvent<Value>) -> Result<Value, Error> {
    let (client, table_name) = get_dynamodb_client().await;
    let jwt_secret = env::var("JWT_SECRET").expect("JWT_SECRET must be set");

    let http_method = event.payload.get("httpMethod").and_then(|v| v.as_str()).unwrap_or("");
    let path = event.payload.get("path").and_then(|v| v.as_str()).unwrap_or("");
    let body = event.payload.get("body").and_then(|v| v.as_str()).unwrap_or("{}");

    let default_headers = json!({});
    let headers = event.payload.get("headers").unwrap_or(&default_headers);

    match (http_method, path) {
        ("POST", "/signup") => signup(body, &client, &table_name).await,
        ("POST", "/signin") => signin(body, &client, &table_name, &jwt_secret).await,
        ("POST", "/authentication") => authenticate(headers, &jwt_secret),
        _ => json_response(404, json!({ "error": "Not Found" })),
    }
}
