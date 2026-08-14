# ---------- Build Stage ----------
FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o admission-webhook ./cmd/server


# ---------- Runtime Stage ----------
FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/admission-webhook .

USER app

EXPOSE 8080
EXPOSE 2112

ENTRYPOINT ["./admission-webhook"]