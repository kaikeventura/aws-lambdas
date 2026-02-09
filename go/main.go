package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ===================================================================
// Configuration
// ===================================================================

var (
	tableName string
	jwtSecret string
	svc       *dynamodb.DynamoDB
)

func init() {
	tableName = os.Getenv("TABLE_NAME")
	jwtSecret = os.Getenv("JWT_SECRET")

	if tableName == "" || jwtSecret == "" {
		log.Println("Warning: TABLE_NAME or JWT_SECRET not set in environment")
	} else {
		sess := session.Must(session.NewSession())
		svc = dynamodb.New(sess)
	}
}

// ===================================================================
// Structs
// ===================================================================

type User struct {
	Email     string `json:"email" dynamodbav:"email"`
	Password  string `json:"password" dynamodbav:"password"`
	Name      string `json:"name" dynamodbav:"name"`
	CreatedAt string `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt string `json:"updatedAt" dynamodbav:"updatedAt"`
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ===================================================================
// Helper Functions
// ===================================================================

func jsonResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(jsonBody),
	}, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ===================================================================
// Route Handlers
// ===================================================================

func signup(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req SignupRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return jsonResponse(400, map[string]string{"error": "Invalid JSON in request body."})
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		return jsonResponse(400, map[string]string{"error": "Missing required fields (email, password, name)."})
	}

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {S: aws.String(req.Email)},
		},
	})
	if err != nil {
		log.Printf("DynamoDB error on GetItem: %v", err)
		return jsonResponse(500, map[string]string{"error": "Internal server error."})
	}
	if result.Item != nil {
		return jsonResponse(409, map[string]string{"error": "User with this email already exists."})
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return jsonResponse(500, map[string]string{"error": "Internal server error."})
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	newUser := User{
		Email:     req.Email,
		Password:  hashedPassword,
		Name:      req.Name,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}

	av, err := dynamodbattribute.MarshalMap(newUser)
	if err != nil {
		log.Printf("Error marshalling user: %v", err)
		return jsonResponse(500, map[string]string{"error": "Internal server error."})
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		log.Printf("DynamoDB error on PutItem: %v", err)
		return jsonResponse(500, map[string]string{"error": "Internal server error."})
	}

	return jsonResponse(201, map[string]string{"message": "User created successfully."})
}

func signin(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req SigninRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return jsonResponse(400, map[string]string{"error": "Invalid JSON in request body."})
	}

	if req.Email == "" || req.Password == "" {
		return jsonResponse(400, map[string]string{"error": "Missing required fields (email, password)."})
	}

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {S: aws.String(req.Email)},
		},
	})
	if err != nil {
		log.Printf("DynamoDB error on GetItem: %v", err)
		return jsonResponse(500, map[string]string{"error": "Internal server error."})
	}

	if result.Item == nil {
		return jsonResponse(401, map[string]string{"error": "Invalid credentials."})
	}

	var user User
	if err := dynamodbattribute.UnmarshalMap(result.Item, &user); err != nil {
		log.Printf("Error unmarshalling user: %v", err)
		return jsonResponse(500, map[string]string{"error": "Internal server error."})
	}

	if !checkPassword(req.Password, user.Password) {
		return jsonResponse(401, map[string]string{"error": "Invalid credentials."})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return jsonResponse(500, map[string]string{"error": "Internal server error."})
	}

	return jsonResponse(200, map[string]string{"token": tokenString})
}

func authenticate(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"]
	}

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return jsonResponse(401, map[string]string{"error": "Authorization header is missing or malformed."})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return jsonResponse(401, map[string]string{"error": "Token has expired."})
		}
		return jsonResponse(401, map[string]string{"error": "Token is invalid."})
	}

	if !token.Valid {
		return jsonResponse(401, map[string]string{"error": "Token is invalid."})
	}

	return jsonResponse(200, map[string]string{"message": "Token is valid."})
}

// ===================================================================
// Main Lambda Handler
// ===================================================================

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received event: %v", request)

	switch request.HTTPMethod {
	case "POST":
		switch request.Path {
		case "/signup":
			return signup(request)
		case "/signin":
			return signin(request)
		case "/authentication":
			return authenticate(request)
		}
	}

	return jsonResponse(404, map[string]string{"error": "Not Found"})
}

func main() {
	lambda.Start(lambdaHandler)
}
