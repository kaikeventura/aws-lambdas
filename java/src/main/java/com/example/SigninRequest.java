package com.example;

import io.quarkus.runtime.annotations.RegisterForReflection;

@RegisterForReflection
public record SigninRequest(String email, String password) {
}