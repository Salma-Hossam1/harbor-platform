# Metadata Worker

## Run

```bash
go run ./cmd/worker
```

Environment variables:

```text
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=harbor-events
```

adding port in docker-compose -> 2112 is only for getting http://localhost:2112/metrics on localhost , but it doesn't need it in order to communicate with kafka , so it is not important to expose it in docker compose , but to test metrics in local host we added it (delete it later) , also prom can connect to it without exposing it to 2112 on local host 
### Why does the Metadata Worker have an HTTP server?

The Metadata Worker is a background service that consumes Kafka messages and writes them to PostgreSQL. It does **not** require HTTP to perform its primary job.

However, Prometheus collects metrics by **scraping an HTTP endpoint**. Since the worker has no existing HTTP API, we start a **small, dedicated HTTP server** whose only responsibility is exposing `GET /metrics`.

This server is completely independent of the Kafka consumer and exists only for observability.

In Docker Compose, we publish port `2112` (for example, `2112:2112`) so the metrics endpoint is reachable from the host during development:

```text
http://localhost:2112/metrics
```

In production, publishing this port is often unnecessary. Prometheus typically runs on the same network and scrapes the worker directly using its service name (for example, `http://metadata-worker:2112/metrics`).

**Key idea:** The Go application creates the HTTP server. Docker Compose only exposes it outside the container when needed.
