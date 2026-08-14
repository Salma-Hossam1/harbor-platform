
# Cloud Native Artifact Platform

A Cloud Native Artifact Platform for managing, processing, securing, and observing container artifacts throughout their lifecycle.

The platform integrates **Harbor, Kubernetes, Kafka, vulnerability scanning, image signing, metadata processing, and observability** to provide an automated artifact supply-chain workflow.

---

## 📌 Overview

Modern cloud-native environments rely heavily on container images. Managing these images securely requires more than simply storing them in a container registry.

This project provides a platform that:

- Stores container images using Harbor
- Detects image push events
- Processes registry events asynchronously through Kafka
- Stores artifact metadata in PostgreSQL
- Scans images for vulnerabilities using Trivy
- Signs approved container images using Cosign
- Provides metadata APIs
- Runs platform services on Kubernetes
- Collects metrics, logs, and traces for observability

The main goal is to demonstrate how different cloud-native tools can be integrated into a secure and observable container artifact lifecycle.

---

## 🏗️ Architecture

The platform is divided into several major layers:

                         ┌─────────────────────┐
                         │      Developer      │
                         └──────────┬──────────┘
                                    │
                                    │ docker push
                                    ▼
                    ┌──────────────────────────────┐
                    │            Harbor            │
                    │     Container Registry       │
                    │                              │
                    │  • Image Storage             │
                    │  • Webhooks                  │
                    │  • Trivy Vulnerability Scan  |
                    │  • SBOM                      │
                    └──────────────┬───────────────┘
                                   │
                                   │ PUSH_ARTIFACT
                                   │ event
                                   ▼
                    ┌──────────────────────────────┐
                    │        Event Gateway         │
                    │             Go               │
                    │                              │
                    │  Harbor Webhook Receiver     │
                    └──────────────┬───────────────┘
                                   │
                                   │ publish
                                   ▼
                    ┌──────────────────────────────┐
                    │            Kafka             │
                    │         harbor-events        │
                    │                              │
                    │        Event Backbone        │
                    └──────────────┬───────────────┘
                                   │
                    ┌──────────────┴──────────────┐
                    │                             │
                 consume                       consume
                    │                             │
                    ▼                             ▼
          ┌───────────────────┐       ┌────────────────────┐
          │  Metadata Worker  │       │    Image Signer    │
          │        Go         │       │         Go         │
          │                   │       │                    │
          │ Event Processing  │       │  Artifact Signing  │
          └─────────┬─────────┘       └──────────┬─────────┘
                    │                            │
                    │                            │
                    ▼                            ▼
          ┌───────────────────┐       ┌────────────────────┐
          │    PostgreSQL     │       │       Cosign       │
          │                   │       │                    │
          │ Metadata Events   │       │  OCI Image Signing │
          └───────────────────┘       └──────────┬─────────┘
                                                 │
                                                 │ signature
                                                 ▼
                                      ┌────────────────────┐
                                      │       Harbor       │
                                      │                    │
                                      │ Signed Artifacts   │
                                      └────────────────────┘


              ┌─────────────────────────────────────────────┐
              │              Platform Infrastructure        │
              │                                             │
              │  ┌───────────────┐   ┌───────────────────┐ │
              │  │    Redis      │   │      MinIO        │ │
              │  │               │   │                   │ │
              │  │ Cache /       │   │ Object Storage /  │ │
              │  │ Supporting    │   │ Harbor Storage    │ │
              │  │ Data          │   │                   │ │
              │  └───────────────┘   └───────────────────┘ │
              │                                             │
              └─────────────────────────────────────────────┘


              ┌─────────────────────────────────────────────┐
              │                Security Layer               │
              │                                             │
              │  ┌────────────────┐   ┌──────────────────┐ │
              │  │ Harbor + Trivy │   │ Cosign           │ │
              │  │                │   │                  │ │
              │  │ Vulnerability  │   │ Image Signing &  │ │
              │  │ Scanning       │   │ Verification     │ │
              │  └────────────────┘   └──────────────────┘ │
              │                                             │
              └─────────────────────────────────────────────┘


              ┌─────────────────────────────────────────────┐
              │                Observability                │
              │                                             │
              │  ┌────────────┐      ┌────────────┐         │
              │  │ Prometheus │ ───► │  Grafana   │         │
              │  └────────────┘      └────────────┘         │
              │                                             │
              │  ┌────────────┐      ┌────────────┐         │
              │  │   Loki     │ ◄─── │  Promtail  │         │
              │  └────────────┘      └────────────┘         │
              │                                             │
              │  ┌────────────┐      ┌────────────┐         │
              │  │   Tempo    │ ◄─── │    OTEL    │         │
              │  │            │      │  Collector │         │
              │  └────────────┘      └────────────┘         │
              │                                             │
              └─────────────────────────────────────────────┘
## 🔄 Artifact Lifecycle

The intended artifact lifecycle is:

```text
Developer
    │
    │ docker push
    ▼
Harbor
    │
    ├──► Trivy vulnerability scanning
    │
    └──► PUSH_ARTIFACT webhook
              │
              ▼
        Event Gateway
              │
              ▼
           Kafka
              │
        ┌─────┴─────┐
        ▼           ▼
 Metadata       Image Signer
 Worker             │
        │           ▼
        ▼         Cosign
   PostgreSQL        │
                    ▼
                  Harbor
```

The event-driven architecture allows components to remain loosely coupled.

Instead of Harbor directly calling every downstream service, Harbor sends an event to the Event Gateway, which publishes it to Kafka.

Consumers can then independently process the event.

---

# 🧩 Main Components

## Harbor

Harbor acts as the primary container registry.

Responsibilities include:

* Container image storage
* Image/project management
* Registry API
* Webhooks
* Vulnerability scanning integration
* Artifact metadata
* OCI artifact storage

Harbor is the entry point for container artifacts entering the platform.

---

## Event Gateway

**Technology:** Go

The Event Gateway receives Harbor webhook events and publishes them to Kafka.

### Responsibilities

* Receive Harbor webhook events
* Validate incoming requests
* Publish events to Kafka
* Expose health and metrics endpoints

### Endpoints

```text
GET  /health
POST /webhook
GET  /metrics
```

---

## Kafka

Kafka provides the event streaming layer.

The main topic used by the platform is:

```text
harbor-events
```

Kafka decouples Harbor events from downstream processing.

For example:

```text
Harbor
   │
   ▼
Event Gateway
   │
   ▼
Kafka
   │
   ├──► Metadata Worker
   │
   └──► Image Signer
```

This allows additional consumers to be added without changing the Event Gateway.

---

## Metadata Worker

**Technology:** Go

The Metadata Worker consumes Harbor events from Kafka and stores relevant metadata in PostgreSQL.

### Responsibilities

* Consume Kafka events
* Parse Harbor artifact events
* Extract artifact metadata
* Persist metadata
* Expose Prometheus metrics

---

## Metadata API

**Technology:** Go

The Metadata API provides access to stored artifact metadata.

### Endpoints

```text
GET /health
GET /events
GET /metrics
```

The API queries PostgreSQL and exposes the stored event information to clients.

---

# 🔐 Security

Security is an important part of the platform.

The project integrates multiple security mechanisms.

## Trivy

Trivy is integrated with Harbor as the vulnerability scanner.

It scans container images for known vulnerabilities.

The intended workflow is:

```text
Image Push
    │
    ▼
Harbor
    │
    ▼
Trivy Scan
    │
    ├── Vulnerabilities found
    │        │
    │        ▼
    │     Reject / block deployment
    │
    └── Acceptable result
             │
             ▼
          Continue
```

The vulnerability scanner is responsible for **image vulnerability detection**.

---

## Cosign

Cosign is used to cryptographically sign container images.

The Image Signer service consumes Harbor artifact events from Kafka.

```text
Harbor
   │
   │ PUSH_ARTIFACT
   ▼
Event Gateway
   │
   ▼
Kafka
   │
   ▼
Image Signer
   │
   ▼
Cosign
   │
   ▼
Signed Image
```

Image signing provides a mechanism for establishing artifact authenticity and integrity.

---

# ☸️ Kubernetes

The platform is designed to run on Kubernetes.

The Kubernetes environment currently uses:

* Minikube
* Kubernetes
* Ingress-NGINX
* MetalLB

Platform services are deployed into Kubernetes namespaces.

Example:

```text
artifact-platform
monitoring
metallb-system
ingress-nginx
```

The Kubernetes deployment includes workloads such as:

* Event Gateway
* Metadata Worker
* Metadata API
* Kafka
* PostgreSQL
* Redis
* MinIO
* Exporters
* Monitoring components

---

# 📊 Observability

The platform includes a complete observability stack.

## Metrics

**Prometheus** collects application and infrastructure metrics.

Services expose Prometheus-compatible endpoints:

```text
/metrics
```

Grafana is used to visualize metrics.

---

## Logs

The logging pipeline uses:

```text
Kubernetes Containers
        │
        ▼
     Promtail
        │
        ▼
       Loki
        │
        ▼
     Grafana
```

Promtail collects container logs and sends them to Loki.

---

## Tracing

Distributed tracing uses:

```text
Application
    │
    ▼
 OpenTelemetry
    │
    ▼
OTEL Collector
    │
    ▼
   Tempo
    │
    ▼
 Grafana
```

This allows requests and operations to be traced across services.

---

# 🐳 Docker Compose

The project was initially developed using Docker Compose.

The Compose environment provides the core platform components and allows individual services to be developed and tested independently before deployment to Kubernetes.

The Compose phase includes:

* Harbor
* Kafka
* PostgreSQL
* Redis
* MinIO
* Event Gateway
* Metadata Worker
* Metadata API
* Observability components

---

# 📁 Repository Structure

The repository is organized around the platform's major components.

```text
harbor-platform/
│
├── harbor/
│
├── services/
│   ├── event-gateway/
│   ├── metadata-worker/
│   ├── metadata-api/
│   └── image-signer/
│
├── docker-compose/
│
├── k8s/
│   ├── artifact-platform/
│   ├── monitoring/
│   └── ...
│
├── secrets/
│
├── scripts/
│
└── README.md
```

Each major service is maintained as an independent component with its own source code and deployment configuration.

---

# 🛠️ Technologies

| Category              | Technology               |
| --------------------- | ------------------------ |
| Container Registry    | Harbor                   |
| Container Runtime     | Docker                   |
| Orchestration         | Kubernetes               |
| Local Kubernetes      | Minikube                 |
| Event Streaming       | Apache Kafka             |
| Backend Services      | Go                       |
| Database              | PostgreSQL               |
| Cache                 | Redis                    |
| Object Storage        | MinIO                    |
| Vulnerability Scanner | Trivy                    |
| Image Signing         | Cosign                   |
| Metrics               | Prometheus               |
| Visualization         | Grafana                  |
| Logging               | Loki                     |
| Log Collection        | Promtail                 |
| Tracing               | OpenTelemetry + Tempo    |
| Ingress               | NGINX Ingress Controller |
| Load Balancing        | MetalLB                  |

---

# 🚀 Development Phases

The project is being developed incrementally.

## Phase 1 — Infrastructure

Core infrastructure:

* PostgreSQL
* Redis
* MinIO
* Kafka

## Phase 2 — Platform Services

Implemented services:

* Event Gateway
* Metadata Worker
* Metadata API

## Phase 3 — Harbor Integration

Integration with Harbor:

* Harbor webhooks
* Artifact events
* Kafka event pipeline
* Metadata processing

## Phase 4 — Security

Security pipeline:

* Trivy vulnerability scanning
* Image signing with Cosign
* Artifact verification

## Phase 5 — Kubernetes

Migration and deployment of platform components to Kubernetes.

## Phase 6 — Observability

Observability stack:

* Prometheus
* Grafana
* Loki
* Promtail
* Tempo
* OpenTelemetry Collector

---

# 🧪 Testing

The platform can be tested by pushing an image to Harbor:

```bash
docker tag alpine:latest localhost:8088/demo/test:v1

docker push localhost:8088/demo/test:v1
```

The push should generate a Harbor artifact event.

The expected event flow is:

```text
Docker Push
     │
     ▼
   Harbor
     │
     ▼
Event Gateway
     │
     ▼
   Kafka
     │
     ├───────────────┐
     ▼               ▼
Metadata Worker   Image Signer
     │               │
     ▼               ▼
PostgreSQL         Cosign
```

---

# 🔭 Future Improvements

Potential future improvements include:

* Automated deployment admission based on vulnerability results
* Signature verification before deployment
* Kubernetes admission control
* SBOM generation and storage
* Artifact provenance
* Policy enforcement
* Improved failure/rejection notifications
* CI/CD integration
* Production-grade TLS
* Secret management using Kubernetes Secrets or Vault
* High availability for stateful components
* Advanced Grafana dashboards
* Alerting and incident notification

---

# 🎯 Project Goals

The main goals of this project are to demonstrate:

1. **Cloud-native architecture**
2. **Event-driven communication**
3. **Container artifact management**
4. **Container security**
5. **Image vulnerability scanning**
6. **Cryptographic image signing**
7. **Kubernetes deployment**
8. **Observability**
9. **Loose coupling through Kafka**
10. **Secure artifact lifecycle management**

---

## 📌 Project Status

The platform is under active development.

The Docker Compose environment has been used to develop and validate the core event-driven platform, while Kubernetes is being used for the cloud-native deployment and observability phases.

Security and artifact-signing workflows are currently being integrated and tested.

---

## 👩‍💻 Project

**Cloud Native Artifact Platform**

Built as a cloud-native / DevOps engineering project focusing on:

**Container Registries · Kubernetes · Event-Driven Architecture · Security · Observability · Platform Engineering**

````

