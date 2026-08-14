# ---------- Build Stage ----------
FROM golang:1.25 AS builder

WORKDIR /app

# Copy module files first to improve build cache
# COPY go.mod go.sum ./
COPY go.mod ./

RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -o event-gateway ./cmd/server


# ---------- Runtime Stage ----------
FROM alpine:3.22

# Create a non-root user
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy only the compiled binary
COPY --from=builder /app/event-gateway /app/event-gateway

# Labels gor OCI image (metadata)
LABEL org.opencontainers.image.title="event-gateway"
LABEL org.opencontainers.image.description="Receives Harbor webhooks and publishes platform events"
LABEL org.opencontainers.image.version="0.1.0"


# Switch to the non-root user
USER app

# Application listens on this port
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Start the application
ENTRYPOINT ["./event-gateway"]
