package com.example;

import io.quarkus.test.common.QuarkusTestResource;
import io.quarkus.test.junit.QuarkusTest;
import io.restassured.http.ContentType;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.testcontainers.containers.localstack.LocalStackContainer;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.*;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.notNullValue;

@QuarkusTest
@QuarkusTestResource(DynamoDBResource.class)
public class UserResourceTest {

    @BeforeAll
    public static void setup() {
        DynamoDbClient client = DynamoDbClient.builder()
                .endpointOverride(DynamoDBResource.localStack.getEndpointOverride(LocalStackContainer.Service.DYNAMODB))
                .region(Region.of(DynamoDBResource.localStack.getRegion()))
                .credentialsProvider(StaticCredentialsProvider.create(
                        AwsBasicCredentials.create(DynamoDBResource.localStack.getAccessKey(), DynamoDBResource.localStack.getSecretKey())
                ))
                .build();

        try {
            client.createTable(CreateTableRequest.builder()
                    .tableName("users")
                    .keySchema(KeySchemaElement.builder()
                            .attributeName("email")
                            .keyType(KeyType.HASH)
                            .build())
                    .attributeDefinitions(AttributeDefinition.builder()
                            .attributeName("email")
                            .attributeType(ScalarAttributeType.S)
                            .build())
                    .provisionedThroughput(ProvisionedThroughput.builder()
                            .readCapacityUnits(5L)
                            .writeCapacityUnits(5L)
                            .build())
                    .build());
        } catch (ResourceInUseException e) {
            // Table already exists
        }
    }

    @Test
    public void testSignup() {
        User user = new User();
        user.setEmail("test@example.com");
        user.setPassword("password123");
        user.setName("Test User");

        given()
                .contentType(ContentType.JSON)
                .body(user)
                .when()
                .post("/api/signup")
                .then()
                .statusCode(201)
                .body("email", notNullValue());
    }

    @Test
    public void testSignin() {
        // First create a user
        User user = new User();
        user.setEmail("signin@example.com");
        user.setPassword("password123");
        user.setName("Signin User");

        given()
                .contentType(ContentType.JSON)
                .body(user)
                .post("/api/signup");

        // Then try to sign in
        SigninRequest request = new SigninRequest("signin@example.com", "password123");

        given()
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .post("/api/signin")
                .then()
                .statusCode(200)
                .body("token", notNullValue());
    }

    @Test
    public void testAuthentication() {
        // First create a user
        User user = new User();
        user.setEmail("auth@example.com");
        user.setPassword("password123");
        user.setName("Auth User");

        given()
                .contentType(ContentType.JSON)
                .body(user)
                .post("/api/signup");

        // Sign in to get token
        SigninRequest request = new SigninRequest("auth@example.com", "password123");

        String token = given()
                .contentType(ContentType.JSON)
                .body(request)
                .post("/api/signin")
                .jsonPath()
                .getString("token");

        // Test authentication endpoint
        given()
                .contentType(ContentType.JSON) // Adicionado explicitamente
                .header("Authorization", "Bearer " + token)
                .when()
                .post("/api/authentication")
                .then()
                .statusCode(200)
                .body("message", notNullValue());
    }
}
