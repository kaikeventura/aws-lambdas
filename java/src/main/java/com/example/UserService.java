package com.example;

import io.smallrye.common.annotation.RunOnVirtualThread;
import io.smallrye.jwt.build.Jwt;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.mindrot.jbcrypt.BCrypt;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;

import java.time.Instant;
import java.util.Optional;

@ApplicationScoped
public class UserService {

    @Inject
    DynamoDbTable<User> userTable;

    @ConfigProperty(name = "mp.jwt.verify.issuer")
    String jwtIssuer;

    @RunOnVirtualThread
    public Optional<User> createUser(User user) {
        // Check if user already exists
        if (userTable.getItem(r -> r.key(k -> k.partitionValue(user.getEmail()))) != null) {
            return Optional.empty();
        }

        String hashedPassword = BCrypt.hashpw(user.getPassword(), BCrypt.gensalt());
        user.setPassword(hashedPassword);

        String now = Instant.now().toString();
        user.setCreatedAt(now);
        user.setUpdatedAt(now);

        userTable.putItem(user);
        return Optional.of(user);
    }

    @RunOnVirtualThread
    public Optional<String> authenticateUser(String email, String password) {
        User user = userTable.getItem(r -> r.key(k -> k.partitionValue(email)));
        if (user == null || !BCrypt.checkpw(password, user.getPassword())) {
            return Optional.empty();
        }

        return Optional.of(
            Jwt.issuer(jwtIssuer)
                .subject(email)
                .expiresIn(3600) // 1 hour
                .sign()
        );
    }

    public boolean validateToken(String token) {
        // This is a simplified validation. In a real app, you'd use quarkus-smallrye-jwt for this.
        // For this exercise, we will just check if the token can be decoded.
        // The programmatic validation API is more complex.
        try {
            // SmallRye's JWT build focuses on creation, not simple programmatic validation.
            // A true manual validation requires setting up the parser with the key.
            io.smallrye.jwt.util.JwtUtil.parseClaims(token);
            return true;
        } catch (Exception e) {
            return false;
        }
    }
}
