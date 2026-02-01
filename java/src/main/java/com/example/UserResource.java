package com.example;

import jakarta.inject.Inject;
import jakarta.ws.rs.*;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;
import org.jboss.logging.Logger;

import java.util.Map;

@Path("/api")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class UserResource {

    private static final Logger LOG = Logger.getLogger(UserResource.class);

    @Inject
    UserService userService;

    @POST
    @Path("/signup")
    public Response signup(User user) {
        return userService.createUser(user)
                .map(u -> Response.status(Response.Status.CREATED).entity(u).build())
                .orElseGet(() -> Response.status(Response.Status.CONFLICT)
                        .entity(Map.of("error", "User with this email already exists."))
                        .build());
    }

    @POST
    @Path("/signin")
    public Response signin(SigninRequest request) {
        return userService.authenticateUser(request.email(), request.password())
                .map(token -> Response.ok(new AuthResponse(token)).build())
                .orElseGet(() -> Response.status(Response.Status.UNAUTHORIZED)
                        .entity(Map.of("error", "Invalid credentials."))
                        .build());
    }

    @POST
    @Path("/authentication")
    public Response authentication(@HeaderParam("Authorization") String authorizationHeader) {
        if (authorizationHeader == null || !authorizationHeader.startsWith("Bearer ")) {
            return Response.status(Response.Status.UNAUTHORIZED)
                    .entity(Map.of("error", "Authorization header is missing or malformed."))
                    .build();
        }
        String token = authorizationHeader.substring(7);
        try {
            userService.validateToken(token);
            return Response.ok(Map.of("message", "Token is valid.")).build();
        } catch (Exception e) {
            // Imprime no console para garantir visibilidade nos testes
            System.err.println("Token validation failed: " + e.getMessage());
            e.printStackTrace();

            return Response.status(Response.Status.UNAUTHORIZED)
                    .entity(Map.of("error", "Token validation failed: " + e.getMessage()))
                    .build();
        }
    }
}
