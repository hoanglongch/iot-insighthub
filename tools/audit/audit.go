package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// AuditReport is the structure for baseline metrics and health data.
type AuditReport struct {
	TelemetryIngestedEvents float64 `json:"telemetry_ingested_events"`
	APIResponseTimeMS       float64 `json:"api_response_time_ms"`
	APIStatusCode           int     `json:"api_status_code"`
	Timestamp               string  `json:"timestamp"`
}

// extractMetricValue is a helper to parse a Prometheus metric value from text.
func extractMetricValue(metricsText, metricName string) float64 {
	lines := strings.Split(metricsText, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, metricName) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				var value float64
				fmt.Sscanf(parts[1], "%f", &value)
				return value
			}
		}
	}
	// Return zero if metric not found.
	return 0
}

func main() {
	// --- Step 1: Audit Telemetry Metrics ---
	telemetryMetricsURL := "http://localhost:9090/metrics"
	resp, err := http.Get(telemetryMetricsURL)
	if err != nil {
		fmt.Println("Error fetching telemetry metrics:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	metricsText := string(body)

	// Extract the ingested events counter (ensure your ingestor exposes a metric named "ingested_events_total")
	ingestedEvents := extractMetricValue(metricsText, "ingested_events_total")
	fmt.Printf("Ingested events: %f\n", ingestedEvents)

	// --- Step 2: Test the Secure API Endpoint ---
	// Prepare a dummy payload (ensure your secure API at /ingest is running on port 8080)
	apiURL := "http://localhost:8080/ingest"
	dummyPayload := strings.NewReader(`{"device_id": "test-device", "value": 42, "time": 1234567890}`)

	startTime := time.Now()
	apiResp, err := http.Post(apiURL, "application/json", dummyPayload)
	elapsed := time.Since(startTime).Milliseconds()

	var apiStatusCode int
	if err != nil {
		fmt.Println("Error calling API:", err)
		apiStatusCode = 0
	} else {
		apiStatusCode = apiResp.StatusCode
		apiResp.Body.Close()
	}
	fmt.Printf("API Response Time: %d ms, Status Code: %d\n", elapsed, apiStatusCode)

	// --- Step 3: Generate an Audit Report ---
	report := AuditReport{
		TelemetryIngestedEvents: ingestedEvents,
		APIResponseTimeMS:       float64(elapsed),
		APIStatusCode:           apiStatusCode,
		Timestamp:               time.Now().Format(time.RFC3339),
	}

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling audit report:", err)
		return
	}

	fmt.Println("\n--- Audit Report ---")
	fmt.Println(string(reportJSON))
}
