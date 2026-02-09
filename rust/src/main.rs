use aws_lambdas_rust::function_handler;
use lambda_runtime::{run, service_fn, Error};

#[tokio::main]
async fn main() -> Result<(), Error> {
    simple_logger::init_with_level(log::Level::Info)?;
    run(service_fn(function_handler)).await
}
