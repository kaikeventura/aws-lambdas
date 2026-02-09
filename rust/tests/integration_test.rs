use aws_config::{BehaviorVersion, Region};
use aws_credential_types::{Credentials, provider::SharedCredentialsProvider};
use aws_lambdas_rust::{authenticate, signin, signup};
use aws_sdk_dynamodb::{types::{AttributeDefinition, KeySchemaElement, KeyType, ProvisionedThroughput, ScalarAttributeType}, Client};
use serde_json::{json, Value};
use testcontainers::clients;
use testcontainers_modules::localstack::LocalStack;

#[tokio::test]
async fn test_integration() {
    let docker = clients::Cli::default();
    let node = docker.run(LocalStack::default());
    let host_port = node.get_host_port_ipv4(4566);
    let endpoint_url = format!("http://127.0.0.1:{}", host_port);

    let credentials = Credentials::new("test", "test", None, None, "static");
    let credentials_provider = SharedCredentialsProvider::new(credentials);

    let config = aws_config::load_defaults(BehaviorVersion::latest())
        .await
        .into_builder()
        .region(Region::new("us-east-1"))
        .endpoint_url(&endpoint_url)
        .credentials_provider(credentials_provider)
        .build();

    let client = Client::new(&config);
    let table_name = "Users";
    let jwt_secret = "testsecret";

    client
        .create_table()
        .table_name(table_name)
        .key_schema(
            KeySchemaElement::builder()
                .attribute_name("email")
                .key_type(KeyType::Hash)
                .build()
                .unwrap(),
        )
        .attribute_definitions(
            AttributeDefinition::builder()
                .attribute_name("email")
                .attribute_type(ScalarAttributeType::S)
                .build()
                .unwrap(),
        )
        .provisioned_throughput(
            ProvisionedThroughput::builder()
                .read_capacity_units(5)
                .write_capacity_units(5)
                .build()
                .unwrap(),
        )
        .send()
        .await
        .expect("Failed to create table");

    let signup_body = json!({
        "email": "test@example.com",
        "password": "password123",
        "name": "Test User"
    }).to_string();

    let resp = signup(&signup_body, &client, table_name).await.unwrap();
    let status = resp["statusCode"].as_i64().unwrap();
    assert_eq!(status, 201);

    let resp = signup(&signup_body, &client, table_name).await.unwrap();
    let status = resp["statusCode"].as_i64().unwrap();
    assert_eq!(status, 409);

    let signin_body = json!({
        "email": "test@example.com",
        "password": "password123"
    }).to_string();

    let resp = signin(&signin_body, &client, table_name, jwt_secret).await.unwrap();
    let status = resp["statusCode"].as_i64().unwrap();
    assert_eq!(status, 200);

    let body_str = resp["body"].as_str().unwrap();
    let body_json: Value = serde_json::from_str(body_str).unwrap();
    let token = body_json["token"].as_str().unwrap();

    let headers = json!({
        "Authorization": format!("Bearer {}", token)
    });

    let resp = authenticate(&headers, jwt_secret).unwrap();
    let status = resp["statusCode"].as_i64().unwrap();
    assert_eq!(status, 200);

    let headers = json!({
        "Authorization": "Bearer invalidtoken"
    });
    let resp = authenticate(&headers, jwt_secret).unwrap();
    let status = resp["statusCode"].as_i64().unwrap();
    assert_eq!(status, 401);
}
