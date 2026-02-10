import { GenericContainer, Wait } from "testcontainers";
import { DynamoDBClient, CreateTableCommand } from "@aws-sdk/client-dynamodb";
import { handler } from "./index.mjs";

describe("Lambda Integration Tests", () => {
  let container;
  let dynamoClient;
  const tableName = "Users";
  const jwtSecret = "testsecret";

  beforeAll(async () => {
    // Start LocalStack container
    container = await new GenericContainer("localstack/localstack:3.2.0")
      .withExposedPorts(4566)
      .withEnvironment({ SERVICES: "dynamodb" })
      .withWaitStrategy(Wait.forLogMessage("Ready."))
      .start();

    const endpoint = `http://${container.getHost()}:${container.getMappedPort(4566)}`;

    // Configure Environment Variables
    process.env.TABLE_NAME = tableName;
    process.env.JWT_SECRET = jwtSecret;
    // Override AWS SDK endpoint via env var is tricky without re-importing,
    // but since we instantiate client in global scope of index.mjs, we might need to mock or rely on AWS_ENDPOINT_URL if SDK supports it.
    // AWS SDK v3 supports AWS_ENDPOINT_URL env var.
    process.env.AWS_ENDPOINT_URL = endpoint;
    process.env.AWS_REGION = "us-east-1";
    process.env.AWS_ACCESS_KEY_ID = "test";
    process.env.AWS_SECRET_ACCESS_KEY = "test";

    // Initialize DynamoDB Client for setup
    dynamoClient = new DynamoDBClient({
      endpoint,
      region: "us-east-1",
      credentials: { accessKeyId: "test", secretAccessKey: "test" },
    });

    // Create Table
    await dynamoClient.send(
      new CreateTableCommand({
        TableName: tableName,
        KeySchema: [{ AttributeName: "email", KeyType: "HASH" }],
        AttributeDefinitions: [{ AttributeName: "email", AttributeType: "S" }],
        ProvisionedThroughput: { ReadCapacityUnits: 5, WriteCapacityUnits: 5 },
      })
    );
  }, 60000);

  afterAll(async () => {
    if (container) await container.stop();
  });

  test("Signup - Success", async () => {
    const event = {
      httpMethod: "POST",
      path: "/signup",
      body: JSON.stringify({
        email: "test@example.com",
        password: "password123",
        name: "Test User",
      }),
    };

    const response = await handler(event);
    expect(response.statusCode).toBe(201);
    const body = JSON.parse(response.body);
    expect(body.message).toBe("User created successfully.");
  });

  test("Signup - Duplicate", async () => {
    const event = {
      httpMethod: "POST",
      path: "/signup",
      body: JSON.stringify({
        email: "test@example.com",
        password: "password123",
        name: "Test User",
      }),
    };

    const response = await handler(event);
    expect(response.statusCode).toBe(409);
  });

  test("Signin - Success", async () => {
    const event = {
      httpMethod: "POST",
      path: "/signin",
      body: JSON.stringify({
        email: "test@example.com",
        password: "password123",
      }),
    };

    const response = await handler(event);
    expect(response.statusCode).toBe(200);
    const body = JSON.parse(response.body);
    expect(body.token).toBeDefined();
  });

  test("Signin - Invalid Credentials", async () => {
    const event = {
      httpMethod: "POST",
      path: "/signin",
      body: JSON.stringify({
        email: "test@example.com",
        password: "wrongpassword",
      }),
    };

    const response = await handler(event);
    expect(response.statusCode).toBe(401);
  });

  test("Authentication - Success", async () => {
    // Get token first
    const signinEvent = {
      httpMethod: "POST",
      path: "/signin",
      body: JSON.stringify({
        email: "test@example.com",
        password: "password123",
      }),
    };
    const signinResponse = await handler(signinEvent);
    const token = JSON.parse(signinResponse.body).token;

    const event = {
      httpMethod: "POST",
      path: "/authentication",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    };

    const response = await handler(event);
    expect(response.statusCode).toBe(200);
  });

  test("Authentication - Invalid Token", async () => {
    const event = {
      httpMethod: "POST",
      path: "/authentication",
      headers: {
        Authorization: "Bearer invalidtoken",
      },
    };

    const response = await handler(event);
    expect(response.statusCode).toBe(401);
  });
});
