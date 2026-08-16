# ---------- Go Build Stage ----------
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o image-verifier ./cmd/server


# ---------- Cosign Stage ----------
FROM ghcr.io/sigstore/cosign/cosign:v3.0.2 AS cosign


# ---------- Runtime Stage ----------
FROM alpine:3.22

RUN apk add --no-cache ca-certificates su-exec && \
    mkdir -p /usr/local/share/ca-certificates

RUN addgroup -S app && \
    adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/image-verifier ./image-verifier

COPY --from=cosign /ko-app/cosign /usr/local/bin/cosign

RUN chmod +x /usr/local/bin/cosign

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080
EXPOSE 2112

ENTRYPOINT ["/entrypoint.sh"]