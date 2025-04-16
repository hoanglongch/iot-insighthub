package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"iot-insighthub/pkg/api"
	"iot-insighthub/pkg/auth"
	"iot-insighthub/pkg/secureapi"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// telemetryHandler processes incoming telemetry data.
func telemetryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse JSON payload.
	var data api.TelemetryData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Validate payload fields.
	if err := validate.Struct(data); err != nil {
		http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Set a context with timeout for database operations.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Persist telemetry data with fault tolerance.
	if err := secureapi.StoreTelemetryData(ctx, data); err != nil {
		log.Printf("Error storing telemetry data: %v", err)
		http.Error(w, "failed to store telemetry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Telemetry data accepted"))
}

// docsHandler serves static Swagger documentation (e.g. swagger.json).
func docsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./docs/swagger.json")
}

func main() {
	// Initialize the validator instance.
	validate = validator.New()

	mux := http.NewServeMux()
	// Apply JWT-based authentication and rate limiting on the /ingest endpoint.
	mux.Handle("/ingest", auth.AuthMiddleware(http.HandlerFunc(telemetryHandler)))
	// Serve API documentation at /docs.
	mux.Handle("/docs", http.HandlerFunc(docsHandler))

	log.Println("Secure API running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
