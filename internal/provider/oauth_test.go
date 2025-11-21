// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestConfig_Client_OAuth2Success tests successful OAuth2 token acquisition
func TestConfig_Client_OAuth2Success(t *testing.T) {
	// Create mock OAuth2 server
	tokenCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/oauth/token" {
			tokenCalled = true
			t.Logf("Token endpoint called with method: %s", r.Method)

			// Verify it's a POST request
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			// Mock successful token response
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-access-token-12345",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}
		t.Logf("Unexpected path called: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := &Config{
		Endpoint:     server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	client, err := config.Client(context.Background())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if !tokenCalled {
		t.Error("Token endpoint was never called")
	}

	t.Log("✓ OAuth2 client created successfully")
	t.Log("✓ Token endpoint was called")
	t.Log("✓ HTTP client with OAuth2 transport is ready")
}

// TestConfig_Client_InvalidCredentials tests OAuth2 with invalid credentials
func TestConfig_Client_InvalidCredentials(t *testing.T) {
	// Create mock OAuth2 server that rejects credentials
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_client",
				"error_description": "Client authentication failed",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := &Config{
		Endpoint:     server.URL,
		ClientID:     "bad-client",
		ClientSecret: "bad-secret",
	}

	_, err := config.Client(context.Background())
	if err == nil {
		t.Fatal("Expected error for invalid credentials, got nil")
	}

	t.Logf("✓ Error correctly returned for invalid credentials: %v", err)
}

// TestConfig_Client_UnreachableServer tests behavior with unreachable server
func TestConfig_Client_UnreachableServer(t *testing.T) {
	config := &Config{
		Endpoint:     "http://localhost:9999", // Non-existent server
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}

	_, err := config.Client(context.Background())
	if err == nil {
		t.Fatal("Expected error for unreachable server, got nil")
	}

	t.Logf("✓ Error correctly returned for unreachable server: %v", err)
}
