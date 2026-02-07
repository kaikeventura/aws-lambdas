resource "null_resource" "build_lambda" {
  count = var.enable_build ? 1 : 0

  provisioner "local-exec" {
    working_dir = "${abspath(path.module)}/../../../../java"
    command     = var.architecture == "arm64" ? "./build-arm64.sh" : "./build-x86.sh"
  }

  triggers = {
    always_run = timestamp()
  }
}

resource "aws_iam_role" "lambda_exec_role" {
  name = "lambda-java-exec-role"

  assume_role_policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Action    = "sts:AssumeRole",
        Effect    = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "dynamodb_access" {
  name   = "dynamodb-access-policy-java"
  role   = aws_iam_role.lambda_exec_role.id
  policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Action   = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ],
        Effect   = "Allow",
        Resource = var.table_arn
      }
    ]
  })
}

resource "aws_lambda_function" "java_lambda" {
  function_name    = "MyJavaLambda"
  handler          = "not.used"
  runtime          = "provided.al2023"
  architectures    = [var.architecture]
  memory_size      = 256
  timeout          = 15

  role             = aws_iam_role.lambda_exec_role.arn

  filename         = "${abspath(path.module)}/../../../../java/function.zip"

  source_code_hash = fileexists("${abspath(path.module)}/../../../../java/function.zip") ? filebase64sha256("${abspath(path.module)}/../../../../java/function.zip") : null

  environment {
    variables = {
      TABLE_NAME                  = var.table_name
      JWT_SECRET                  = var.jwt_secret
      DISABLE_SIGNAL_HANDLERS     = "true"
      MP_JWT_VERIFY_ISSUER        = "https://my-app.com"
      QUARKUS_DYNAMODB_TABLE_NAME = var.table_name
    }
  }

  depends_on = [null_resource.build_lambda]
}
