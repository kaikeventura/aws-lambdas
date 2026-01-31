package com.example;

import jakarta.inject.Inject;
import jakarta.ws.rs.*;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;

@Path("/api")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class UserResource {

    @Inject
    UserService userService;

    @POST
    @Path("/signup")
    public Response signup(User user) {
        return userService.createUser(user)
                .map(u -> Response.status(Response.Status.CREATED).entity(u).build())
                .orElseGet(() -> Response.status(Response.Status.CONFLICT)
                        .entity("{\"error\":\"User with this email already exists.\"}")
                        .build());
    }

    @POST
    @Path("/signin")
    public Response signin(SigninRequest request) {
        return userService.authenticateUser(request.email(), request.password())
                .map(token -> Response.ok(new AuthResponse(token)).build())
                .orElseGet(() -> Response.status(Response.Status.UNAUTHORIZED)
                        .entity("{\"error\":\"Invalid credentials.\"}")
                        .build());
    }

    @POST
    @Path("/authentication")
    public Response authentication(@HeaderParam("Authorization") String authorizationHeader) {
        if (authorizationHeader == null || !authorizationHeader.startsWith("Bearer ")) {
            return Response.status(Response.Status.UNAUTHORIZED)
                    .entity("{\"error\":\"Authorization header is missing or malformed.\"}")
                    .build();
        }
        String token = authorizationHeader.substring(7);

        if (userService.validateToken(token)) {
            return Response.ok("{\"message\":\"Token is valid.\"}").build();
        } else {
            return Response.status(Response.Status.UNAUTHORIZED)
                    .entity("{\"error\":\"Token is invalid.\"}")
                    .build();
        }
    }
}