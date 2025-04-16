package secureapi

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"iot-insighthub/pkg/api"
)

var db *sql.DB

// initDB initializes the database connection to TimescaleDB.
func initDB() error {
	// DSN should come from a secure configuration (environment variable or secrets manager).
	dsn := "postgres://user:password@localhost:5432/telemetrydb?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	// Configure connection pooling.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db.Ping()
}

var InitDBFunc = initDB

// StoreTelemetryData persists telemetry data into the database with retry and exponential backoff.
func StoreTelemetryData(ctx context.Context, data api.TelemetryData) error {
    // Ensure the database is initialized.
	if db == nil {
        if err := InitDBFunc(); err != nil {
            return fmt.Errorf("failed to initialize database: %w", err)
        }
    }

	// Prepare the SQL insert statement.
	stmt := `INSERT INTO telemetry (device_id, value, timestamp) VALUES ($1, $2, to_timestamp($3))`

	// Retry logic: attempt up to 3 times with exponential backoff.
	attempts := 3
	var err error
	for i := 0; i < attempts; i++ {
		_, err = db.ExecContext(ctx, stmt, data.DeviceID, data.Value, data.Time)
		if err == nil {
			return nil
		}
		// Log error and wait before retrying.
		log.Printf("Error inserting telemetry data (attempt %d): %v", i+1, err)
		time.Sleep(time.Duration(1<<i) * time.Second) // Backoff intervals: 1, 2, 4 seconds.
	}
	return fmt.Errorf("failed to store telemetry data after %d attempts: %w", attempts, err)
}

/* 
additional steps havent implemented yet:
1. In your database connection setup (in pkg/secureapi/service.go), ensure that the DSN points to your TimescaleDB instance which has been migrated. During deployment, you might run the migration tool as a separate CI/CD step so that every time you deploy, your schema is checked and updated if needed.
*/