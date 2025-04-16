CREATE TABLE IF NOT EXISTS telemetry (
    id SERIAL PRIMARY KEY,
    device_id TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL
);

-- Create an index on timestamp for faster queries.
CREATE INDEX IF NOT EXISTS idx_telemetry_timestamp ON telemetry (timestamp);
