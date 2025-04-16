package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/go-playground/validator/v10"
	"iot-insighthub/pkg/api"
	"iot-insighthub/pkg/auth"
	"iot-insighthub/pkg/secureapi"
)

// generateTestTokenForHandler creates a valid JWT token for testing the API handler.
func generateTestTokenForHandler() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	return token.SignedString([]byte("testsecret"))
}

// OverrideStoreTelemetryData is used to bypass actual DB calls during testing.
func overrideStoreTelemetryData(ctx context.Context, data api.TelemetryData) error {
	// Simply return nil to simulate a successful insert.
	return nil
}

func TestTelemetryHandler_ValidRequest(t *testing.T) {
	// Ensure the validator is initialized.
	validate = validator.New()

	// Set environment variable for JWT secret.
	os.Setenv("JWT_SECRET", "testsecret")

	// Override the database call in secureapi for testing.
	originalStoreFunc := secureapi.StoreTelemetryData
	secureapi.StoreTelemetryData = overrideStoreTelemetryData
	defer func() { secureapi.StoreTelemetryData = originalStoreFunc }()

	// Create a valid telemetry payload.
	payload := api.TelemetryData{
		DeviceID: "device123",
		Value:    45.6,
		Time:     time.Now().Unix(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	// Generate a valid JWT.
	token, err := generateTestTokenForHandler()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	// Wrap the telemetryHandler with the AuthMiddleware.
	handler := auth.AuthMiddleware(http.HandlerFunc(telemetryHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected status 202 Accepted for valid request, got %d", rr.Code)
	}
}

func TestTelemetryHandler_InvalidPayload(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	// Generate a valid JWT.
	token, err := generateTestTokenForHandler()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Use an invalid JSON body.
	req := httptest.NewRequest("POST", "/ingest", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler := auth.AuthMiddleware(http.HandlerFunc(telemetryHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request for invalid JSON, got %d", rr.Code)
	}
}
