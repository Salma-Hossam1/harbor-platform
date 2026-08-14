# Image Signer

The Image Signer is a Kafka consumer responsible for receiving Harbor events.

## Current milestone

The current implementation only consumes messages from Kafka and logs them.

It does not perform image signing yet.

## Configuration

Environment variables:

- `KAFKA_BROKERS` - Kafka broker address. Default: `kafka:9092`
- `KAFKA_TOPIC` - Kafka topic. Default: `harbor-events`
- `KAFKA_GROUP_ID` - Kafka consumer group. Default: `image-signer`

## Architecture

```text
Harbor
   |
   | webhook
   v
Event Gateway
   |
   v
Kafka
   |
   +----------------------+
   |                      |
   v                      v
Metadata Worker       Image Signer
   |                      |
   v                      v
PostgreSQL             logging