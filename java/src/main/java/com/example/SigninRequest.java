package com.example;

// Using records for concise, immutable data carriers
public record SigninRequest(String email, String password) {}
