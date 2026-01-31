resource "aws_api_gateway_rest_api" "this" {
  name        = "${var.project}-${var.env}-api"
  description = "API Gateway for Lambda functions"
}

# --- Resource: /python ---
resource "aws_api_gateway_resource" "python" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_rest_api.this.root_resource_id
  path_part   = "python"
}

resource "aws_api_gateway_method" "python_any" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.python.id
  http_method   = "ANY"
  authorization = "NONE"
  api_key_required = true # Requer API Key
}

resource "aws_api_gateway_integration" "python_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.python.id
  http_method             = aws_api_gateway_method.python_any.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.python_lambda_invoke_arn
}

# Permissão para o API Gateway invocar a Lambda Python
resource "aws_lambda_permission" "apigw_python" {
  statement_id  = "AllowAPIGatewayInvokePython"
  action        = "lambda:InvokeFunction"
  function_name = var.python_lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.this.execution_arn}/*/*"
}

# --- Resource: /java ---
resource "aws_api_gateway_resource" "java" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_rest_api.this.root_resource_id
  path_part   = "java"
}

resource "aws_api_gateway_method" "java_any" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.java.id
  http_method   = "ANY"
  authorization = "NONE"
  api_key_required = true # Requer API Key
}

resource "aws_api_gateway_integration" "java_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.java.id
  http_method             = aws_api_gateway_method.java_any.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.java_lambda_invoke_arn
}

# Permissão para o API Gateway invocar a Lambda Java
resource "aws_lambda_permission" "apigw_java" {
  statement_id  = "AllowAPIGatewayInvokeJava"
  action        = "lambda:InvokeFunction"
  function_name = var.java_lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.this.execution_arn}/*/*"
}

# --- Deployment & Stage ---
resource "aws_api_gateway_deployment" "this" {
  rest_api_id = aws_api_gateway_rest_api.this.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.python.id,
      aws_api_gateway_method.python_any.id,
      aws_api_gateway_integration.python_lambda.id,
      aws_api_gateway_resource.java.id,
      aws_api_gateway_method.java_any.id,
      aws_api_gateway_integration.java_lambda.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "dev" {
  deployment_id = aws_api_gateway_deployment.this.id
  rest_api_id   = aws_api_gateway_rest_api.this.id
  stage_name    = var.env
}

# --- API Key & Usage Plan ---
resource "aws_api_gateway_api_key" "mykey" {
  name = "${var.project}-${var.env}-key"
}

resource "aws_api_gateway_usage_plan" "myplan" {
  name = "${var.project}-${var.env}-plan"

  api_stages {
    api_id = aws_api_gateway_rest_api.this.id
    stage  = aws_api_gateway_stage.dev.stage_name
  }
}

resource "aws_api_gateway_usage_plan_key" "main" {
  key_id        = aws_api_gateway_api_key.mykey.id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.myplan.id
}
