using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using Amazon.Lambda.APIGatewayEvents;
using Amazon.Lambda.Core;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;
using BCrypt.Net;
using Microsoft.IdentityModel.Tokens;

// Assembly attribute to enable the Lambda function's JSON input to be converted into a .NET class.
[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.SystemTextJson.DefaultLambdaJsonSerializer))]

namespace AwsLambdasDotNet;

public class Function
{
    private static readonly string TableName = Environment.GetEnvironmentVariable("TABLE_NAME") ?? "Users";
    private static readonly string JwtSecret = Environment.GetEnvironmentVariable("JWT_SECRET") ?? "secret";
    private readonly IAmazonDynamoDB _ddbClient;

    public Function()
    {
        // Initialize with default client, which picks up env vars (AWS_ENDPOINT_URL, etc.)
        // or allow tests to set env vars before this constructor is called.
        // For better testability, we could add a constructor that accepts IAmazonDynamoDB.
        var config = new AmazonDynamoDBConfig();
        var serviceUrl = Environment.GetEnvironmentVariable("AWS_ENDPOINT_URL");
        if (!string.IsNullOrEmpty(serviceUrl))
        {
            config.ServiceURL = serviceUrl;
        }

        _ddbClient = new AmazonDynamoDBClient(config);
    }

    // Constructor for testing
    public Function(IAmazonDynamoDB ddbClient)
    {
        _ddbClient = ddbClient;
    }

    public async Task<APIGatewayProxyResponse> FunctionHandler(APIGatewayProxyRequest request, ILambdaContext context)
    {
        context.Logger.LogInformation($"Received event: {JsonSerializer.Serialize(request)}");

        var method = request.HttpMethod;
        var path = request.Path;

        if (method == "POST" && path == "/signup")
        {
            return await Signup(request.Body);
        }
        if (method == "POST" && path == "/signin")
        {
            return await Signin(request.Body);
        }
        if (method == "POST" && path == "/authentication")
        {
            return Authenticate(request.Headers);
        }

        return JsonResponse(404, new { error = "Not Found" });
    }

    private async Task<APIGatewayProxyResponse> Signup(string body)
    {
        SignupRequest? req;
        try
        {
            req = JsonSerializer.Deserialize<SignupRequest>(body);
        }
        catch
        {
            return JsonResponse(400, new { error = "Invalid JSON in request body." });
        }

        if (req == null || string.IsNullOrEmpty(req.Email) || string.IsNullOrEmpty(req.Password) || string.IsNullOrEmpty(req.Name))
        {
            return JsonResponse(400, new { error = "Missing required fields (email, password, name)." });
        }

        // Check if user exists
        var getItemRequest = new GetItemRequest
        {
            TableName = TableName,
            Key = new Dictionary<string, AttributeValue> { { "email", new AttributeValue { S = req.Email } } }
        };

        try
        {
            var response = await _ddbClient.GetItemAsync(getItemRequest);
            if (response.Item != null && response.Item.Count > 0)
            {
                return JsonResponse(409, new { error = "User with this email already exists." });
            }
        }
        catch (Exception e)
        {
            Console.WriteLine(e);
            return JsonResponse(500, new { error = "Internal server error." });
        }

        // Create user
        var hashedPassword = BCrypt.Net.BCrypt.HashPassword(req.Password);
        var timestamp = DateTime.UtcNow.ToString("o");

        var putItemRequest = new PutItemRequest
        {
            TableName = TableName,
            Item = new Dictionary<string, AttributeValue>
            {
                { "email", new AttributeValue { S = req.Email } },
                { "password", new AttributeValue { S = hashedPassword } },
                { "name", new AttributeValue { S = req.Name } },
                { "createdAt", new AttributeValue { S = timestamp } },
                { "updatedAt", new AttributeValue { S = timestamp } }
            }
        };

        try
        {
            await _ddbClient.PutItemAsync(putItemRequest);
            return JsonResponse(201, new { message = "User created successfully." });
        }
        catch (Exception e)
        {
            Console.WriteLine(e);
            return JsonResponse(500, new { error = "Internal server error." });
        }
    }

    private async Task<APIGatewayProxyResponse> Signin(string body)
    {
        SigninRequest? req;
        try
        {
            req = JsonSerializer.Deserialize<SigninRequest>(body);
        }
        catch
        {
            return JsonResponse(400, new { error = "Invalid JSON in request body." });
        }

        if (req == null || string.IsNullOrEmpty(req.Email) || string.IsNullOrEmpty(req.Password))
        {
            return JsonResponse(400, new { error = "Missing required fields (email, password)." });
        }

        var getItemRequest = new GetItemRequest
        {
            TableName = TableName,
            Key = new Dictionary<string, AttributeValue> { { "email", new AttributeValue { S = req.Email } } }
        };

        Dictionary<string, AttributeValue>? userItem = null;
        try
        {
            var response = await _ddbClient.GetItemAsync(getItemRequest);
            if (response.Item == null || response.Item.Count == 0)
            {
                return JsonResponse(401, new { error = "Invalid credentials." });
            }
            userItem = response.Item;
        }
        catch (Exception e)
        {
            Console.WriteLine(e);
            return JsonResponse(500, new { error = "Internal server error." });
        }

        if (!BCrypt.Net.BCrypt.Verify(req.Password, userItem["password"].S))
        {
            return JsonResponse(401, new { error = "Invalid credentials." });
        }

        var tokenHandler = new JwtSecurityTokenHandler();
        var key = Encoding.ASCII.GetBytes(JwtSecret);
        var tokenDescriptor = new SecurityTokenDescriptor
        {
            Subject = new ClaimsIdentity(new[] { new Claim("sub", req.Email) }),
            Expires = DateTime.UtcNow.AddHours(1),
            SigningCredentials = new SigningCredentials(new SymmetricSecurityKey(key), SecurityAlgorithms.HmacSha256Signature)
        };
        var token = tokenHandler.CreateToken(tokenDescriptor);
        var tokenString = tokenHandler.WriteToken(token);

        return JsonResponse(200, new { token = tokenString });
    }

    private APIGatewayProxyResponse Authenticate(IDictionary<string, string> headers)
    {
        string? authHeader = null;
        if (headers != null)
        {
            if (headers.ContainsKey("Authorization")) authHeader = headers["Authorization"];
            else if (headers.ContainsKey("authorization")) authHeader = headers["authorization"];
        }

        if (string.IsNullOrEmpty(authHeader) || !authHeader.StartsWith("Bearer "))
        {
            return JsonResponse(401, new { error = "Authorization header is missing or malformed." });
        }

        var token = authHeader.Substring("Bearer ".Length).Trim();
        var tokenHandler = new JwtSecurityTokenHandler();
        var key = Encoding.ASCII.GetBytes(JwtSecret);

        try
        {
            tokenHandler.ValidateToken(token, new TokenValidationParameters
            {
                ValidateIssuerSigningKey = true,
                IssuerSigningKey = new SymmetricSecurityKey(key),
                ValidateIssuer = false,
                ValidateAudience = false,
                ClockSkew = TimeSpan.Zero
            }, out SecurityToken validatedToken);

            return JsonResponse(200, new { message = "Token is valid." });
        }
        catch (SecurityTokenExpiredException)
        {
            return JsonResponse(401, new { error = "Token has expired." });
        }
        catch (Exception)
        {
            return JsonResponse(401, new { error = "Token is invalid." });
        }
    }

    private APIGatewayProxyResponse JsonResponse(int statusCode, object body)
    {
        return new APIGatewayProxyResponse
        {
            StatusCode = statusCode,
            Headers = new Dictionary<string, string> { { "Content-Type", "application/json" } },
            Body = JsonSerializer.Serialize(body)
        };
    }
}

public class SignupRequest
{
    [JsonPropertyName("email")]
    public string? Email { get; set; }
    [JsonPropertyName("password")]
    public string? Password { get; set; }
    [JsonPropertyName("name")]
    public string? Name { get; set; }
}

public class SigninRequest
{
    [JsonPropertyName("email")]
    public string? Email { get; set; }
    [JsonPropertyName("password")]
    public string? Password { get; set; }
}
