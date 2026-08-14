Infrastructure/network problem
connection timeout
different network namespaces
Docker bridge
Minikube bridge
routing
IP forwarding
iptables

Your application was healthy.

Your Kubernetes Service was healthy.

Your EndpointSlice was healthy.

Your Pod was healthy.

Your Ingress was healthy.

The missing piece was:

Allowing traffic from the Docker Compose network (172.29.0.0/16) to the Minikube network (192.168.49.0/24) through the host.

That's the root cause.
//////////////////////////////////////////////////////////////////////////////////

Absolutely. Here is a **copy-paste-ready documentation section** you can put directly into your project notes, for example:

`docs/harbor-k8s-webhook-networking.md`

````markdown
# Harbor → Kubernetes Webhook Networking

## Problem

The project uses two different environments on the same machine:

1. Harbor is running with Docker Compose.
2. The application platform is running inside Minikube/Kubernetes.

The important network ranges were:

| Environment | Network |
|---|---|
| Harbor Docker Compose | `172.29.0.0/16` |
| Minikube | `192.168.49.0/24` |

The Harbor Jobservice container had:

```text
IP:      172.29.0.24
Gateway: 172.29.0.1
````

Minikube had:

```text
IP:      192.168.49.2
Gateway: 192.168.49.1
```

The Kubernetes Ingress was reachable through:

```text
192.168.49.240:80
```

### What was failing?

Harbor's Jobservice was configured to send webhook requests to:

```text
http://192.168.49.240/webhook
```

From the host machine, this worked.

From inside the Harbor Jobservice container, however, the request timed out:

```text
curl: (28) Connection timed out
```

Harbor therefore reported:

```text
context deadline exceeded
Client.Timeout exceeded while awaiting headers
```

This was initially confusing because the Event Gateway itself was healthy.

---

# Diagnosis

## Kubernetes was healthy

The Event Gateway Pod was running:

```bash
kubectl get pods -n artifact-platform -l app=event-gateway -o wide
```

The Service had endpoints:

```bash
kubectl describe svc event-gateway -n artifact-platform
```

Example:

```text
Endpoints: 10.244.1.96:8080
```

The EndpointSlice also contained the Pod:

```bash
kubectl get endpointslice -n artifact-platform \
  -l kubernetes.io/service-name=event-gateway
```

Therefore:

```text
Pod → Service
```

was working.

---

## Kubernetes internal request worked

A temporary curl Pod inside Kubernetes successfully called:

```bash
kubectl run curl-test \
  -n artifact-platform \
  --rm -it \
  --image=curlimages/curl \
  --restart=Never \
  -- \
  curl -i -X POST \
  -H "Content-Type: application/json" \
  -d '{"event":"image_push","repository":"demo/backend"}' \
  http://event-gateway:8080/webhook
```

Response:

```text
HTTP/1.1 200 OK
```

Therefore the Event Gateway application was working.

---

## Host → Ingress worked

The host machine could successfully call:

```bash
curl -i \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"event":"image_push","repository":"demo/backend"}' \
  http://192.168.49.240/webhook
```

Response:

```text
HTTP/1.1 200 OK
```

Therefore NGINX Ingress and the Event Gateway were working.

---

## Harbor → Ingress failed

The important test was from the Harbor Jobservice container:

```bash
docker exec harbor-jobservice sh
```

Then:

```bash
curl -v --max-time 10 http://192.168.49.240/webhook
```

This timed out.

The Jobservice network was:

```text
172.29.0.0/16
```

while Minikube was:

```text
192.168.49.0/24
```

The host had both networks, but traffic from the Docker container to the Minikube network was not being forwarded correctly.

Therefore the problem was:

> A network routing/forwarding problem between the Docker Compose network and the Minikube network.

It was NOT:

* an Event Gateway application problem
* a Worker image problem
* a Kafka problem
* a PostgreSQL problem
* a Kubernetes Service problem
* an Ingress configuration problem

---

# Network Architecture

Before the fix:

```text
                    HOST
                     |
       +-------------+-------------+
       |                           |
       |                           |
172.29.0.0/16                192.168.49.0/24
Docker Compose                   Minikube
       |                           |
       |                           |
Harbor Jobservice              Kubernetes
172.29.0.24                       |
       |                           |
       X  <-- traffic could       |
       |      not cross           |
       |                           |
       +---------------------------+
```

Harbor could not successfully reach:

```text
192.168.49.240:80
```

---

# Solution

The solution was to allow the Linux host to forward traffic between:

```text
172.29.0.0/16
```

and:

```text
192.168.49.0/24
```

Two things were required:

1. Enable IPv4 forwarding.
2. Allow the traffic through Docker's `DOCKER-USER` chain.

---

# 1. Enable IPv4 forwarding

First, IPv4 forwarding was enabled:

```bash
sudo sysctl -w net.ipv4.ip_forward=1
```

Verify:

```bash
sysctl net.ipv4.ip_forward
```

Expected:

```text
net.ipv4.ip_forward = 1
```

---

# 2. Make IPv4 forwarding permanent

A sysctl configuration file was created:

```bash
sudo tee /etc/sysctl.d/99-minikube-docker.conf <<'EOF'
net.ipv4.ip_forward=1
EOF
```

Then the configuration was applied:

```bash
sudo sysctl --system
```

This makes IPv4 forwarding survive a reboot.

---

# 3. Allow Docker → Minikube traffic

The following rule allows traffic originating from the Harbor Docker network to reach the Minikube network:

```bash
sudo iptables -I DOCKER-USER 1 \
  -s 172.29.0.0/16 \
  -d 192.168.49.0/24 \
  -j ACCEPT
```

Meaning:

```text
source:
172.29.0.0/16

destination:
192.168.49.0/24

action:
ACCEPT
```

---

# 4. Allow return traffic

The return traffic was allowed with:

```bash
sudo iptables -I DOCKER-USER 2 \
  -s 192.168.49.0/24 \
  -d 172.29.0.0/16 \
  -m conntrack \
  --ctstate ESTABLISHED,RELATED \
  -j ACCEPT
```

This allows responses belonging to connections initiated from the Harbor network.

---

# 5. Make iptables rules permanent

The persistence package was installed:

```bash
sudo apt update
sudo apt install -y iptables-persistent
```

Then the rules were saved:

```bash
sudo netfilter-persistent save
```

Alternatively:

```bash
sudo iptables-save | sudo tee /etc/iptables/rules.v4 > /dev/null
```

Verify:

```bash
sudo iptables -S DOCKER-USER
```

The rules should include:

```text
-A DOCKER-USER -s 172.29.0.0/16 -d 192.168.49.0/24 -j ACCEPT

-A DOCKER-USER -s 192.168.49.0/24 -d 172.29.0.0/16 \
    -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
```

---

# 6. Verify the final connection

The final test should be performed from the Harbor Jobservice container:

```bash
docker exec harbor-jobservice sh -c '
curl -v --max-time 10 \
-X POST \
-H "Content-Type: application/json" \
-d "{\"event\":\"image_push\",\"repository\":\"demo/backend\"}" \
http://192.168.49.240/webhook
'
```

Expected:

```text
HTTP/1.1 200 OK
```

Then verify the Event Gateway:

```bash
kubectl logs -n artifact-platform deployment/event-gateway --since=1m
```

Expected:

```text
{"event":"image_push","repository":"demo/backend"}
```

---

# Final Architecture

After the fix:

```text
Harbor Jobservice
172.29.0.24
       |
       | Docker network
       |
172.29.0.1
       |
       | Linux IP forwarding
       | + iptables
       |
192.168.49.1
       |
       |
192.168.49.240:80
       |
       v
NGINX Ingress
       |
       v
event-gateway Service
       |
       v
Event Gateway Pod
       |
       v
Kafka
       |
       v
Metadata Worker
       |
       v
PostgreSQL
```

---

# Why the application images did not need to change

The Metadata Worker image was working correctly.

The failure happened before the event reached the application:

```text
Harbor Jobservice
       |
       X
       |
NGINX Ingress
       |
Event Gateway
```

The Worker was downstream:

```text
Event Gateway
       |
       v
Kafka
       |
       v
Metadata Worker
```

Therefore changing the Worker image could not solve the problem.

The correct fix was networking.

---

# Why this is needed only in the current hybrid setup

The problem exists because the current architecture mixes:

```text
Docker Compose
+
Minikube/Kubernetes
```

on the same machine.

Harbor lives in:

```text
172.29.0.0/16
```

while the Kubernetes application lives in:

```text
192.168.49.0/24
```

The host must therefore route traffic between the two networks.

---

# What happens if Harbor is moved to Kubernetes?

If Harbor is eventually deployed entirely inside Kubernetes, this specific Docker Compose → Minikube networking problem goes away.

For example:

```text
Kubernetes
┌─────────────────────────────────────┐
│                                     │
│  Harbor Jobservice                  │
│        |                            │
│        v                            │
│  Kubernetes Service                 │
│        |                            │
│        v                            │
│  Event Gateway                      │
│                                     │
└─────────────────────────────────────┘
```

The components are then communicating through Kubernetes networking rather than crossing:

```text
Docker Compose network
        ↓
Linux host
        ↓
Minikube network
```

The iptables rules created specifically for this hybrid setup would no longer be required.

However, the Kubernetes networking configuration itself would still be required.

---

# How to remove the configuration

If Harbor is later moved completely into Kubernetes and these rules are no longer needed, remove them.

## 1. Remove the iptables rules

First inspect them:

```bash
sudo iptables -S DOCKER-USER
```

Then delete the rules:

```bash
sudo iptables -D DOCKER-USER \
  -s 172.29.0.0/16 \
  -d 192.168.49.0/24 \
  -j ACCEPT
```

And:

```bash
sudo iptables -D DOCKER-USER \
  -s 192.168.49.0/24 \
  -d 172.29.0.0/16 \
  -m conntrack \
  --ctstate ESTABLISHED,RELATED \
  -j ACCEPT
```

Save the updated rules:

```bash
sudo netfilter-persistent save
```

---

# 2. Remove the permanent sysctl configuration

Delete:

```bash
sudo rm /etc/sysctl.d/99-minikube-docker.conf
```

Then reload sysctl configuration:

```bash
sudo sysctl --system
```

If the machine still needs IPv4 forwarding for other applications, do NOT disable it.

If this was the only reason for enabling forwarding and you are sure nothing else needs it, you can disable it:

```bash
sudo sysctl -w net.ipv4.ip_forward=0
```

Verify:

```bash
sysctl net.ipv4.ip_forward
```

---

# Important caution when removing the rules

Do not blindly flush:

```bash
sudo iptables -F
```

and do not flush the entire `DOCKER-USER` chain:

```bash
sudo iptables -F DOCKER-USER
```

because Docker or other applications may have rules there.

Remove only the rules that were created for this project.

---

# Useful troubleshooting commands

If the same problem happens again, these commands help identify it quickly.

### Check Docker Jobservice network

```bash
docker inspect harbor-jobservice \
  --format '{{range .NetworkSettings.Networks}}{{.IPAddress}} {{.Gateway}}{{end}}'
```

Expected:

```text
172.29.0.24 172.29.0.1
```

### Check Docker network

```bash
docker network inspect <network-name> \
  | grep -E '"Subnet"|"Gateway"'
```

### Check Minikube IP

```bash
minikube ip
```

### Check host route

```bash
ip route | grep 192.168.49
```

### Check IPv4 forwarding

```bash
sysctl net.ipv4.ip_forward
```

### Check Docker forwarding rules

```bash
sudo iptables -S DOCKER-USER
```

### Test from Harbor

```bash
docker exec harbor-jobservice \
  curl -v --max-time 10 \
  http://192.168.49.240/webhook
```

If you receive:

```text
HTTP/1.1 405 Method Not Allowed
Allow: POST
```

that means **network connectivity is working**.

The request reached the Event Gateway; you simply used GET instead of POST.

A successful POST should return:

```text
HTTP/1.1 200 OK
```

---

# Summary

### Root cause

Harbor Jobservice was inside the Docker Compose network:

```text
172.29.0.0/16
```

while the Kubernetes Ingress was inside the Minikube network:

```text
192.168.49.0/24
```

Traffic between those networks was not being forwarded correctly.

### Solution

Enable:

```text
net.ipv4.ip_forward=1
```

and allow:

```text
172.29.0.0/16
        ↓
192.168.49.0/24
```

through the Docker `DOCKER-USER` chain, with the corresponding established/related return traffic.

### Result

```text
Harbor
  ↓
Docker network
  ↓
Host forwarding
  ↓
Minikube
  ↓
NGINX Ingress
  ↓
Event Gateway
```

### Future

If Harbor is deployed entirely inside Kubernetes, this **specific hybrid-network workaround can be removed** because Docker Compose will no longer need to reach the Minikube network.

```
```


###############################################################################

The problem

Your Harbor is running with Docker Compose on the Linux host:

Linux host
│
├── Docker Compose
│     └── Harbor
│          └── :8088
│
└── Minikube
      └── Kubernetes Pods

From your Linux host, this works:

localhost:8088

because localhost means:

this Linux machine

But when the Image Verifier runs inside Kubernetes:

Image Verifier Pod
       │
       │ localhost:8088
       ▼
      ??? 

localhost now means:

the Image Verifier Pod itself

It does NOT mean your Linux host.

So this would be wrong:

HARBOR_REGISTRY=localhost:8088
The solution

Minikube provides:

host.minikube.internal

which allows workloads inside Minikube to reach the host machine.

Therefore:

HARBOR_REGISTRY=host.minikube.internal:8088

means:

Image Verifier Pod
       │
       │ host.minikube.internal:8088
       ▼
Minikube → Linux Host
       │
       ▼
Docker Compose Harbor :8088

And we already proved this works with:

minikube ssh -- curl -s \
  http://host.minikube.internal:8088/api/v2.0/systeminfo

which returned Harbor's system information.

Why this is a good solution for our project

We are intentionally keeping:

Harbor → Docker Compose
Kubernetes platform → Minikube

as separate environments.

We don't need to move Harbor into Kubernetes just to make the admission architecture work.

The connection is simply:

                    Linux Host
              ┌────────────────────┐
              │                    │
              │  Harbor Compose    │
              │  :8088             │
              │                    │
              └─────────▲──────────┘
                        │
             host.minikube.internal
                        │
              ┌─────────┴──────────┐
              │     Minikube       │
              │                    │
              │ Image Verifier     │
              │        ▲           │
              │        │           │
              │ Admission Webhook  │
              │        ▲           │
              │        │           │
              │ Kubernetes API     │
              └────────────────────┘
Documentation wording

You can document it as:

Problem: Harbor is deployed using Docker Compose on the host machine, while the security services are deployed inside Minikube. Therefore, localhost:8088 from inside a Kubernetes Pod refers to the Pod itself rather than the host running Harbor.

Solution: Minikube's host.minikube.internal hostname is used to provide Pod-to-host connectivity. Therefore, the Image Verifier accesses Harbor through host.minikube.internal:8088.

Validation: Connectivity was verified from the Minikube environment using Harbor's /api/v2.0/systeminfo endpoint.

This is a real deployment/environment networking issue, and documenting it is worthwhile.


####################################################################################################
# Explain signer and verifier 
Signing cycle → happens when the image is published.
Verification cycle → happens later when someone tries to deploy the image.

The important thing is: the image itself is not changed by signing. A signature is stored in Harbor as a separate OCI artifact.

1. Signing cycle

Imagine the developer builds:

node-trivy-test:2.0

and pushes it to Harbor.

Step 1 — Push image
Developer
   │
   │ docker push
   ▼
Harbor
   │
   ▼
Image
node-trivy-test@sha256:ABC...

Harbor now has the actual image.

Step 2 — Trivy scans it

Harbor's Trivy scanner checks:

Image
  │
  ▼
Trivy
  │
  ├── vulnerabilities
  └── SBOM

Harbor stores the scan/SBOM information associated with that image.

So conceptually:

Harbor
 └── Image @sha256:ABC
       ├── vulnerabilities
       └── SBOM
Step 3 — Image Signer signs the image

This is the part we haven't implemented yet.

Our future Image Signer receives:

image =
node-trivy-test@sha256:ABC...

It has the private Cosign key:

cosign.key

The signer tells Cosign:

Sign this exact digest using my private key.

Conceptually:

                 private key
                     │
                     ▼
Image digest ──► Cosign
                     │
                     ▼
                 Signature

The private key never goes to Harbor.

Step 4 — Signature is pushed to Harbor

This is the part that was confusing before.

Cosign does not modify:

node-trivy-test@sha256:ABC

Instead, it creates a separate OCI artifact:

Harbor
 ├── Image @sha256:ABC
 │
 └── Cosign signature
       │
       └── associated with @sha256:ABC

And that's exactly why you previously saw:

type: signature.cosign

in Harbor.

So the final state is:

                    Harbor
                      │
          ┌───────────┴───────────┐
          │                       │
          ▼                       ▼
    Container Image         Cosign Signature
    @sha256:ABC              associated with
                             @sha256:ABC
2. Verification cycle

Now imagine later someone wants to deploy:

node-trivy-test@sha256:ABC

They create a Kubernetes Deployment/Pod.

Developer
   │
   │ kubectl apply
   ▼
Kubernetes API Server

The API Server doesn't simply create the Pod.

Because we configured an admission webhook:

Kubernetes API Server
          │
          ▼
   Admission Webhook
Step 1 — Webhook extracts the image

The webhook sees:

node-trivy-test@sha256:ABC

It says:

I need to know whether this exact image is trusted.

So it calls:

POST /verify

on Image Verifier.

Admission Webhook
       │
       │ "verify this image"
       ▼
Image Verifier
3. What Image Verifier does

The verifier has:

cosign.pub

This is the public key corresponding to the private key used by the signer.

But the verifier doesn't have the signature yet.

So it asks Harbor indirectly through Cosign:

Image Verifier
      │
      │ Cosign verify
      ▼
    Harbor
      │
      │ give me signature
      ▼
Signature

Now the verifier has:

Image digest
      +
Signature
      +
Public key

and Cosign checks:

Was this signature created using the private key corresponding to my trusted public key?

and also:

Does this signature belong to this exact image digest?

4. Why the digest is so important

Suppose the developer signed:

image@sha256:ABC

Someone cannot simply replace the image with:

image@sha256:XYZ

and keep using the old signature.

The signature is tied to:

ABC

not merely:

node-trivy-test

or:

2.0

That's why we chose:

repository@sha256:digest

instead of:

repository:2.0
5. The complete project cycle

Now put everything together:

                 PUBLISHING PHASE
                 ────────────────

Developer
    │
    │ docker push
    ▼
  Harbor
    │
    ├───────────────┐
    │               │
    ▼               ▼
  Trivy           Image
    │             @sha256:ABC
    │
    ├── SBOM
    └── vulnerabilities

                  │
                  │ image digest
                  ▼
             Image Signer
                  │
                  │ private key
                  ▼
                Cosign
                  │
                  │ signature
                  ▼
                Harbor
                  │
                  └── signature.cosign
                       associated with
                       @sha256:ABC

Then later:

                 DEPLOYMENT PHASE
                 ────────────────

Developer
    │
    │ kubectl apply
    ▼
Kubernetes API Server
    │
    │ AdmissionReview
    ▼
Admission Webhook
    │
    │ image @sha256:ABC
    ▼
Image Verifier
    │
    │ Cosign verify
    │
    ├──────────────► Harbor
    │                  │
    │                  │ signature
    │◄─────────────────┘
    │
    │ + cosign.pub
    ▼
Cryptographic verification
    │
    ├───────────────┐
    │               │
   VALID          INVALID
    │               │
    ▼               ▼
  ALLOW           REJECT
    │               │
    ▼               ▼
 Kubernetes       Pod NOT
 creates Pod      created
6. So where is the Image Signer?

That's our next missing piece.

Right now we have:

                    DONE
                     │
                     ▼
Harbor ──► Trivy/SBOM
   │
   └──► Cosign signature ◄── manually created for testing
                         │
                         ▼
                   Image Verifier
                         │
                         ▼
                 Admission Webhook
                         │
                         ▼
                    Kubernetes

We manually created the Cosign signature earlier because we needed something to test the verifier with.

That is not the final lifecycle.

The final lifecycle should automate this:

Image pushed to Harbor
        │
        ▼
Harbor event/webhook
        │
        ▼
Image Signer
        │
        │ private key
        ▼
Cosign sign
        │
        ▼
Signature pushed to Harbor

Then the deployment side is:

kubectl apply
     │
     ▼
API Server
     │
     ▼
Admission Webhook
     │
     ▼
Image Verifier
     │
     ▼
Cosign + Harbor + public key
     │
     ▼
ALLOW / REJECT

This is the cycle we should implement next: Image Signer.

And notice the beautiful separation:

Component	Responsibility
Harbor	Stores images, SBOMs, signatures
Trivy	Finds vulnerabilities / generates SBOM
Image Signer	Creates Cosign signatures
Image Verifier	Cryptographically verifies signatures
Admission Webhook	Tells Kubernetes ALLOW/REJECT
Kubernetes API Server	Actually enforces the decision

So yes: until we implement Image Signer, you were missing the first half of the story. The manual Cosign command we used was only simulating that missing Image Signer.