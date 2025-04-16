package api

// TelemetryData defines the structure for incoming telemetry data.
// The struct tags include JSON mappings and validation tags.
type TelemetryData struct {
	DeviceID string  `json:"device_id" validate:"required"`
	Value    float64 `json:"value" validate:"required"`
	Time     int64   `json:"time" validate:"required"`
}
