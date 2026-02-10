using System.Collections.Generic;
using System.Net;
using System.Text.Json;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;
using Amazon.Lambda.APIGatewayEvents;
using Amazon.Lambda.TestUtilities;
using Amazon.Runtime;
using AwsLambdasDotNet;
using DotNet.Testcontainers.Builders;
using DotNet.Testcontainers.Containers;
using Xunit;

namespace AwsLambdasDotNet.Tests;

public class FunctionTests : IAsyncLifetime
{
    private IContainer _localStackContainer = null!;
    private AmazonDynamoDBClient _ddbClient = null!;
    private Function _function = null!;
    private const string TableName = "Users";
    // Secret must be at least 16 characters (128 bits) for HS256
    private const string JwtSecret = "testsecret_must_be_longer_than_16_chars";

    public async Task InitializeAsync()
    {
        _localStackContainer = new ContainerBuilder()
            .WithImage("localstack/localstack:3.2.0")
            .WithPortBinding(4566, true)
            .WithEnvironment("SERVICES", "dynamodb")
            .WithWaitStrategy(Wait.ForUnixContainer().UntilPortIsAvailable(4566))
            .Build();

        await _localStackContainer.StartAsync();

        var endpoint = $"http://{_localStackContainer.Hostname}:{_localStackContainer.GetMappedPublicPort(4566)}";

        // Configure Environment Variables
        Environment.SetEnvironmentVariable("TABLE_NAME", TableName);
        Environment.SetEnvironmentVariable("JWT_SECRET", JwtSecret);

        // Initialize DynamoDB Client pointing to LocalStack with fake credentials
        var config = new AmazonDynamoDBConfig { ServiceURL = endpoint };
        var credentials = new BasicAWSCredentials("test", "test");
        _ddbClient = new AmazonDynamoDBClient(credentials, config);

        // Create Table
        var createTableRequest = new CreateTableRequest
        {
            TableName = TableName,
            KeySchema = new List<KeySchemaElement> { new KeySchemaElement { AttributeName = "email", KeyType = KeyType.HASH } },
            AttributeDefinitions = new List<AttributeDefinition> { new AttributeDefinition { AttributeName = "email", AttributeType = ScalarAttributeType.S } },
            ProvisionedThroughput = new ProvisionedThroughput { ReadCapacityUnits = 5, WriteCapacityUnits = 5 }
        };
        await _ddbClient.CreateTableAsync(createTableRequest);

        // Inject the client into the Function
        _function = new Function(_ddbClient);
    }

    public async Task DisposeAsync()
    {
        if (_localStackContainer != null)
        {
            await _localStackContainer.StopAsync();
        }
    }

    [Fact]
    public async Task Signup_Success()
    {
        var request = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/signup",
            Body = JsonSerializer.Serialize(new { email = "test@example.com", password = "password123", name = "Test User" })
        };
        var context = new TestLambdaContext();

        var response = await _function.FunctionHandler(request, context);
        Assert.Equal((int)HttpStatusCode.Created, response.StatusCode);
        Assert.Contains("User created successfully.", response.Body);
    }

    [Fact]
    public async Task Signup_Duplicate()
    {
        // Create user first
        var request1 = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/signup",
            Body = JsonSerializer.Serialize(new { email = "test@example.com", password = "password123", name = "Test User" })
        };
        var context1 = new TestLambdaContext();
        await _function.FunctionHandler(request1, context1);

        // Try to create again
        var request = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/signup",
            Body = JsonSerializer.Serialize(new { email = "test@example.com", password = "password123", name = "Test User" })
        };
        var context = new TestLambdaContext();

        var response = await _function.FunctionHandler(request, context);
        Assert.Equal((int)HttpStatusCode.Conflict, response.StatusCode);
        Assert.Contains("User with this email already exists.", response.Body);
    }

    [Fact]
    public async Task Signin_Success()
    {
        // Create user first
        var request1 = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/signup",
            Body = JsonSerializer.Serialize(new { email = "signin@example.com", password = "password123", name = "Signin User" })
        };
        var context1 = new TestLambdaContext();
        await _function.FunctionHandler(request1, context1);

        var request = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/signin",
            Body = JsonSerializer.Serialize(new { email = "signin@example.com", password = "password123" })
        };
        var context = new TestLambdaContext();

        var response = await _function.FunctionHandler(request, context);
        Assert.Equal((int)HttpStatusCode.OK, response.StatusCode);
        Assert.Contains("token", response.Body);
    }

    [Fact]
    public async Task Signin_InvalidCredentials()
    {
        var request = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/signin",
            Body = JsonSerializer.Serialize(new { email = "test@example.com", password = "wrongpassword" })
        };
        var context = new TestLambdaContext();

        var response = await _function.FunctionHandler(request, context);
        Assert.Equal((int)HttpStatusCode.Unauthorized, response.StatusCode);
        Assert.Contains("Invalid credentials.", response.Body);
    }

    [Fact]
    public async Task Authentication_Success()
    {
        // Create user and signin to get token
        var request1 = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/signup",
            Body = JsonSerializer.Serialize(new { email = "auth@example.com", password = "password123", name = "Auth User" })
        };
        var context1 = new TestLambdaContext();
        await _function.FunctionHandler(request1, context1);

        var request2 = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/signin",
            Body = JsonSerializer.Serialize(new { email = "auth@example.com", password = "password123" })
        };
        var context2 = new TestLambdaContext();
        var signinResponse = await _function.FunctionHandler(request2, context2);
        var token = JsonSerializer.Deserialize<Dictionary<string, string>>(signinResponse.Body)["token"];

        // Authenticate with token
        var request = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/authentication",
            Headers = new Dictionary<string, string> { { "Authorization", $"Bearer {token}" } }
        };
        var context = new TestLambdaContext();

        var response = await _function.FunctionHandler(request, context);
        Assert.Equal((int)HttpStatusCode.OK, response.StatusCode);
        Assert.Contains("Token is valid.", response.Body);
    }

    [Fact]
    public async Task Authentication_InvalidToken()
    {
        var request = new APIGatewayProxyRequest
        {
            HttpMethod = "POST",
            Path = "/authentication",
            Headers = new Dictionary<string, string> { { "Authorization", "Bearer invalidtoken" } }
        };
        var context = new TestLambdaContext();

        var response = await _function.FunctionHandler(request, context);
        Assert.Equal((int)HttpStatusCode.Unauthorized, response.StatusCode);
        Assert.Contains("Token is invalid.", response.Body);
    }
}
