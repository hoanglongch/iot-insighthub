package telemetry

import (
	"encoding/json"
	"log"

	"iot-insighthub/pkg/api"
	"iot-insighthub/pkg/kinesis"
)

// ProcessRecord converts a raw Kinesis record into TelemetryData and processes it.
func ProcessRecord(record *kinesis.Record) error {
	// In production, you would unmarshal record.Data (assumed to be JSON) into TelemetryData.
	var data api.TelemetryData
	if err := json.Unmarshal(record.Data, &data); err != nil {
		log.Printf("Error parsing record data: %v", err)
		return err
	}

	// Process telemetry data (business logic such as anomaly detection, enrichment, etc.)
	log.Printf("Processed telemetry from device %s: %+v", data.DeviceID, data)

	// Simulate checkpointing (acknowledging record processing)
	Checkpoint(record)

	return nil
}

// Checkpoint simulates recording that this record was processed successfully.
func Checkpoint(record *kinesis.Record) {
	log.Printf("Checkpointing record: %s", *record.SequenceNumber)
}
