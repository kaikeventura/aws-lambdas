package com.example;

import io.quarkus.test.common.QuarkusTestResourceLifecycleManager;
import org.testcontainers.containers.localstack.LocalStackContainer;
import org.testcontainers.utility.DockerImageName;

import java.util.Map;

public class DynamoDBResource implements QuarkusTestResourceLifecycleManager {

    public static LocalStackContainer localStack;

    static {
        // Tenta configurar o cliente Docker manualmente se necessário, mas geralmente a propriedade no arquivo deve bastar.
        // System.setProperty("docker.client.strategy", "org.testcontainers.dockerclient.UnixSocketClientProviderStrategy");
        
        localStack = new LocalStackContainer(DockerImageName.parse("localstack/localstack:3.2.0"))
            .withServices(LocalStackContainer.Service.DYNAMODB);
    }

    @Override
    public Map<String, String> start() {
        localStack.start();
        return Map.of(
                "quarkus.dynamodb.endpoint-override", localStack.getEndpointOverride(LocalStackContainer.Service.DYNAMODB).toString(),
                "quarkus.dynamodb.aws.region", localStack.getRegion(),
                "quarkus.dynamodb.aws.credentials.type", "static",
                "quarkus.dynamodb.aws.credentials.static-provider.access-key-id", localStack.getAccessKey(),
                "quarkus.dynamodb.aws.credentials.static-provider.secret-access-key", localStack.getSecretKey()
        );
    }

    @Override
    public void stop() {
        if (localStack != null) {
            localStack.stop();
        }
    }
}
