terraform {
  backend "s3" {
    bucket       = "aws-lambdas-dev-tfstate-2025"
    key          = "envs/dev/terraform.tfstate"
    region       = "us-east-1"
    profile      = "aws-lambdas-dev"
    encrypt      = true
    use_lockfile = true
  }
}
