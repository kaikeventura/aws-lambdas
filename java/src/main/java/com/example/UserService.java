package com.example;

import io.smallrye.jwt.build.Jwt;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.mindrot.jbcrypt.BCrypt;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;
import org.jboss.logging.Logger;
import org.jose4j.jwt.consumer.JwtConsumer;
import org.jose4j.jwt.consumer.JwtConsumerBuilder;
import org.jose4j.keys.HmacKey;
import org.jose4j.jws.AlgorithmIdentifiers;
import org.jose4j.jwa.AlgorithmConstraints;

import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Optional;

@ApplicationScoped
public class UserService {

    private static final Logger LOG = Logger.getLogger(UserService.class);

    @Inject
    DynamoDbTable<User> userTable;

    @ConfigProperty(name = "mp.jwt.verify.issuer")
    String jwtIssuer;

    @ConfigProperty(name = "JWT_SECRET")
    String jwtSecret;

    public Optional<User> createUser(User user) {
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

    public Optional<String> authenticateUser(String email, String password) {
        User user = userTable.getItem(r -> r.key(k -> k.partitionValue(email)));
        if (user == null || !BCrypt.checkpw(password, user.getPassword())) {
            return Optional.empty();
        }

        return Optional.of(
            Jwt.issuer(jwtIssuer)
                .subject(email)
                .expiresIn(3600) // 1 hour
                .signWithSecret(jwtSecret)
        );
    }

    public void validateToken(String token) throws Exception {
        // Log para debug
        LOG.info("Validating token with issuer: " + jwtIssuer);
        
        // Simplificando a validação para isolar o problema
        // Se a validação completa falhar, vamos tentar uma validação mais permissiva temporariamente
        // para ver se o problema é a chave ou outra claim.
        
        JwtConsumer consumer = new JwtConsumerBuilder()
                .setRequireExpirationTime()
                .setAllowedClockSkewInSeconds(30)
                .setVerificationKey(new HmacKey(jwtSecret.getBytes(StandardCharsets.UTF_8)))
                .setJwsAlgorithmConstraints(
                        AlgorithmConstraints.ConstraintType.PERMIT, 
                        AlgorithmIdentifiers.HMAC_SHA256
                )
                .setExpectedIssuer(jwtIssuer)
                .build();
        
        consumer.processToClaims(token);
    }
}
