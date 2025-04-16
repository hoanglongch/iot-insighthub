## Ingestion API
**Endpoint:** `/ingest`

**Method:** POST

**Authentication:** JWT Bearer token (see documentation for token generation)

**Request Body:**
```json
{
  "device_id": "string - Unique device ID",
  "value": "number - Sensor reading",
  "time": "integer - Unix epoch timestamp"
}
```

## Success Response:

- Code: 202 Accepted

- Body: "Telemetry data accepted"

## Error Responses:

- 400 Bad Request: Invalid payload or missing required fields.

- 401 Unauthorized: Missing or invalid token.

- 500 Internal Server Error: Failure in data persistence.