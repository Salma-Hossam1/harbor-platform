# ---------- Go Build Stage ----------
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o image-signer ./cmd/worker


# ---------- Cosign Stage ----------
FROM ghcr.io/sigstore/cosign/cosign:v3.0.2 AS cosign


# ---------- Runtime Stage ----------
FROM alpine:3.22

RUN addgroup -S app && \
    adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/image-signer ./image-signer

COPY --from=cosign /ko-app/cosign /usr/local/bin/cosign

RUN chmod +x /usr/local/bin/cosign

USER app

ENTRYPOINT ["./image-signer"]
