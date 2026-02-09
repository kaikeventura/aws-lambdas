package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

var (
	localStackContainer *localstack.LocalStackContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	localStackContainer, err = localstack.RunContainer(ctx,
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "localstack/localstack:3.2.0",
				Env: map[string]string{
					"SERVICES": "dynamodb",
				},
			},
		}),
	)
	if err != nil {
		fmt.Printf("failed to start localstack container: %s\n", err)
		os.Exit(1)
	}

	os.Setenv("TABLE_NAME", "Users")
	os.Setenv("JWT_SECRET", "testsecret")

	endpoint, err := localStackContainer.PortEndpoint(ctx, "4566/tcp", "")
	if err != nil {
		fmt.Printf("failed to get localstack endpoint: %s\n", err)
		os.Exit(1)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
		DisableSSL:       aws.Bool(true),
	}))
	svc = dynamodb.New(sess)
	tableName = "Users"
	jwtSecret = "testsecret"

	_, err = svc.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("email"), KeyType: aws.String("HASH")},
		},
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("email"), AttributeType: aws.String("S")},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})
	if err != nil {
		fmt.Printf("failed to create table: %s\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := localStackContainer.Terminate(ctx); err != nil {
		fmt.Printf("failed to terminate container: %s\n", err)
	}

	os.Exit(code)
}

func TestSignup(t *testing.T) {
	body := `{"email": "test@example.com", "password": "password123", "name": "Test User"}`
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/signup",
		Body:       body,
	}

	resp, err := lambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var respBody map[string]string
	json.Unmarshal([]byte(resp.Body), &respBody)
	assert.Equal(t, "User created successfully.", respBody["message"])

	resp, err = lambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestSignin(t *testing.T) {
	signupBody := `{"email": "signin@example.com", "password": "password123", "name": "Signin User"}`
	signupReq := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/signup",
		Body:       signupBody,
	}
	lambdaHandler(context.Background(), signupReq)

	signinBody := `{"email": "signin@example.com", "password": "password123"}`
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/signin",
		Body:       signinBody,
	}

	resp, err := lambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var respBody map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &respBody)
	assert.NotEmpty(t, respBody["token"])

	invalidBody := `{"email": "signin@example.com", "password": "wrongpassword"}`
	req.Body = invalidBody
	resp, err = lambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	nonExistentBody := `{"email": "nobody@example.com", "password": "password123"}`
	req.Body = nonExistentBody
	resp, err = lambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestAuthentication(t *testing.T) {
	signupBody := `{"email": "auth@example.com", "password": "password123", "name": "Auth User"}`
	signupReq := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/signup",
		Body:       signupBody,
	}
	lambdaHandler(context.Background(), signupReq)

	signinBody := `{"email": "auth@example.com", "password": "password123"}`
	signinReq := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/signin",
		Body:       signinBody,
	}
	resp, _ := lambdaHandler(context.Background(), signinReq)

	var respBody map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &respBody)
	token := respBody["token"].(string)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/authentication",
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
	}

	resp, err := lambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	req.Headers["Authorization"] = "Bearer invalidtoken"
	resp, err = lambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	req.Headers = map[string]string{}
	resp, err = lambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}
