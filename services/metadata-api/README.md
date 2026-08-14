# Metadata API

## Run

```bash
go run ./cmd/server
```

Environment variables:

```text
PORT=8081

DB_HOST=postgres
DB_PORT=5432
DB_NAME=metadata
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
```

Endpoints:

```text
GET /health
GET /events
```