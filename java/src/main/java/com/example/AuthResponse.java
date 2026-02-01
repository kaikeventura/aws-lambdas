package com.example;

import io.quarkus.runtime.annotations.RegisterForReflection;

@RegisterForReflection
public record AuthResponse(String token) {
}