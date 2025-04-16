package secureapi

import (
	"context"
	"os"
	"testing"
	"time"

	"iot-insighthub/pkg/api"
)

func TestStoreTelemetryData_Success(t *testing.T) {
	// Set TEST_DB_DSN to point to your test database.
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("Skipping test; TEST_DB_DSN environment variable not set")
	}

	// For testing, override the DSN inside initDB.
	// (In production, you would modify your code to read DSN from env vars.)
	// Here we assume that the DSN in the code works with the test DB if TEST_DB_DSN is set.

	// Reset the global db connection.
	db = nil

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := api.TelemetryData{
		DeviceID: "test-device",
		Value:    55.5,
		Time:     time.Now().Unix(),
	}
	err := StoreTelemetryData(ctx, data)
	if err != nil {
		t.Errorf("expected success storing telemetry data, got error: %v", err)
	}
}

func TestStoreTelemetryData_DBInitFailure(t *testing.T) {
	// To simulate a DB init failure, temporarily override the connection string inside initDB.
	// In this test, we force initDB to fail by setting an invalid DSN.
	originalInitDB := initDB
	defer func() { initDB = originalInitDB }()

	initDB = func() error {
		return os.ErrInvalid
	}

	// Reset the db so that StoreTelemetryData calls initDB.
	db = nil

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := api.TelemetryData{
		DeviceID: "test-device",
		Value:    55.5,
		Time:     time.Now().Unix(),
	}
	err := StoreTelemetryData(ctx, data)
	if err == nil {
		t.Error("expected error when DB initialization fails, but got nil")
	}
}
