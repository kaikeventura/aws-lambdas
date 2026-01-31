package com.example;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.inject.Produces;
import jakarta.inject.Inject;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbEnhancedClient;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;
import software.amazon.awssdk.enhanced.dynamodb.TableSchema;

@ApplicationScoped
public class DynamoDbClientProducer {

    @Inject
    DynamoDbEnhancedClient enhancedClient;

    @ConfigProperty(name = "quarkus.dynamodb.table-name")
    String tableName;

    @Produces
    @ApplicationScoped
    public DynamoDbTable<User> userTable() {
        return enhancedClient.table(tableName, TableSchema.fromBean(User.class));
    }
}
