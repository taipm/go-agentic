# 🚀 Deployment Guide - go-agentic

**Status**: Production Ready
**Version**: 1.0
**Last Updated**: 2025-12-22

---

## 🎯 Overview

This guide covers deploying go-agentic in production environments. Deployments range from simple local setups to Kubernetes clusters with full observability.

**Deployment Options**:
- ✅ Local development
- ✅ Docker containers
- ✅ Kubernetes clusters
- ✅ Cloud platforms (AWS, GCP, Azure)
- ✅ Traditional VMs

---

## 📋 Pre-Deployment Checklist

### Before You Deploy

- [ ] Go 1.21+ installed (for building from source)
- [ ] OpenAI API key obtained and tested
- [ ] Configuration files prepared (crew.yaml, agents/*.yaml)
- [ ] Network/firewall rules configured
- [ ] Monitoring/logging system available
- [ ] Backup and recovery strategy planned
- [ ] Load balancer configured (if needed)
- [ ] SSL/TLS certificates ready (for HTTPS)
- [ ] Health check endpoints configured
- [ ] Resource limits defined (CPU, memory)

---

## 🏠 Local Development Deployment

### Quick Start (5 minutes)

```bash
# 1. Clone repository
git clone https://github.com/taipm/go-agentic.git
cd go-agentic

# 2. Set environment variables
export OPENAI_API_KEY="sk-..."

# 3. Run server
go run ./cmd/main.go --server --port 8081

# 4. Test
curl http://localhost:8081/health

# 5. Open web UI (if available)
open http://localhost:8081
```

### Development Configuration

```yaml
# config/crew.yaml (development)
logging:
  level: "debug"

performance:
  max_concurrent_requests: 10

timeouts:
  default_tool_timeout: 30
```

### Local Health Check

```bash
# Run every 5 seconds
watch -n 5 'curl -s http://localhost:8081/health | jq'

# Should show: {"status": "ok", ...}
```

---

## 🐳 Docker Deployment

### Build Docker Image

```dockerfile
# Dockerfile
FROM golang:1.21 AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o go-agentic ./cmd

# Runtime stage
FROM alpine:3.18

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/go-agentic .

# Copy config
COPY config/ ./config/

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run
ENTRYPOINT ["./go-agentic"]
CMD ["--server", "--port", "8081"]
```

### Build and Run

```bash
# Build image
docker build -t go-agentic:latest .

# Run container
docker run \
  -d \
  -p 8081:8081 \
  -e OPENAI_API_KEY="sk-..." \
  -v $(pwd)/config:/app/config:ro \
  --name go-agentic \
  go-agentic:latest

# Check logs
docker logs -f go-agentic

# Test
curl http://localhost:8081/health
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  go-agentic:
    image: go-agentic:latest
    ports:
      - "8081:8081"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - GO_AGENTIC_LOG_LEVEL=info
    volumes:
      - ./config:/app/config:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 10s
      timeout: 5s
      retries: 3
    restart: unless-stopped
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
```

**Start with docker-compose**:

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f go-agentic

# Stop
docker-compose down
```

### Multi-Stage Build (Optimized)

```dockerfile
# Dockerfile.optimized
FROM golang:1.21 AS builder

WORKDIR /build
COPY . .

RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o go-agentic ./cmd

FROM scratch

COPY --from=builder /build/go-agentic /
COPY --from=builder /etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/

EXPOSE 8081
ENTRYPOINT ["/go-agentic", "--server"]
```

**Size comparison**:
- Standard: ~250MB
- Optimized: ~20MB

---

## ☸️ Kubernetes Deployment

### Kubernetes Manifest

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-agentic
  labels:
    app: go-agentic
spec:
  replicas: 3  # High availability

  selector:
    matchLabels:
      app: go-agentic

  template:
    metadata:
      labels:
        app: go-agentic
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8081"
        prometheus.io/path: "/metrics"

    spec:
      # Graceful shutdown
      terminationGracePeriodSeconds: 40

      containers:
      - name: go-agentic
        image: go-agentic:latest
        imagePullPolicy: IfNotPresent

        ports:
        - containerPort: 8081
          name: http
          protocol: TCP

        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: openai-secret
              key: api-key

        - name: GO_AGENTIC_LOG_LEVEL
          value: "info"

        # Resource limits
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi

        # Readiness probe (for load balancing)
        readinessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        # Liveness probe (for restart)
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20
          timeoutSeconds: 5
          failureThreshold: 3

        # Volume mounts
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true

      # Volumes
      volumes:
      - name: config
        configMap:
          name: go-agentic-config

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: go-agentic
  labels:
    app: go-agentic

spec:
  type: LoadBalancer  # Or ClusterIP for internal-only

  selector:
    app: go-agentic

  ports:
  - name: http
    port: 80
    targetPort: 8081
    protocol: TCP

---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: go-agentic-config

data:
  crew.yaml: |
    entry_point: orchestrator
    agents:
      - orchestrator
      - clarifier
      - executor
    max_handoffs: 5
    max_rounds: 10

  orchestrator.yaml: |
    id: orchestrator
    name: "Smart Router"
    role: "Route requests to appropriate agents"
    model: "gpt-4o"
    tools: []
    is_terminal: false

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: openai-secret
type: Opaque
stringData:
  api-key: "sk-..."  # Will be injected from environment
```

### Deploy to Kubernetes

```bash
# Create secret
kubectl create secret generic openai-secret \
  --from-literal=api-key=$OPENAI_API_KEY

# Apply manifests
kubectl apply -f deployment.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

# Check deployment
kubectl get deployments
kubectl get pods

# Check logs
kubectl logs -f deployment/go-agentic

# Port forward for testing
kubectl port-forward svc/go-agentic 8081:80

# Test
curl http://localhost:8081/health
```

### Kubernetes Best Practices

**1. Resource Quotas** (limit namespace usage)
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: go-agentic-quota
spec:
  hard:
    requests.cpu: "2"
    requests.memory: "2Gi"
    limits.cpu: "4"
    limits.memory: "4Gi"
```

**2. Network Policies** (restrict traffic)
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: go-agentic-netpol
spec:
  podSelector:
    matchLabels:
      app: go-agentic
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: ingress
    ports:
    - protocol: TCP
      port: 8081
```

**3. Pod Disruption Budget** (for rolling updates)
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: go-agentic-pdb
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: go-agentic
```

---

## 🌩️ Cloud Platform Deployments

### AWS ECS (Elastic Container Service)

```yaml
# task-definition.json
{
  "family": "go-agentic",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "go-agentic",
      "image": "YOUR_ECR_URI/go-agentic:latest",
      "portMappings": [
        {
          "containerPort": 8081,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "GO_AGENTIC_LOG_LEVEL",
          "value": "info"
        }
      ],
      "secrets": [
        {
          "name": "OPENAI_API_KEY",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:openai-key"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/go-agentic",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8081/health || exit 1"],
        "interval": 10,
        "timeout": 5,
        "retries": 3
      }
    }
  ]
}
```

### Google Cloud Run

```yaml
# cloud-run-deploy.sh
#!/bin/bash

# Build and push to GCR
gcloud builds submit --tag gcr.io/$PROJECT_ID/go-agentic

# Deploy to Cloud Run
gcloud run deploy go-agentic \
  --image gcr.io/$PROJECT_ID/go-agentic:latest \
  --platform managed \
  --region us-central1 \
  --memory 512Mi \
  --cpu 1 \
  --set-env-vars OPENAI_API_KEY=$OPENAI_API_KEY \
  --allow-unauthenticated \
  --timeout 3600

# Get service URL
gcloud run services describe go-agentic --region us-central1
```

### Azure Container Instances

```bash
# Deploy to ACI
az container create \
  --resource-group myresourcegroup \
  --name go-agentic \
  --image go-agentic:latest \
  --port 8081 \
  --cpu 1 \
  --memory 0.5 \
  --environment-variables \
    OPENAI_API_KEY=$OPENAI_API_KEY \
    GO_AGENTIC_LOG_LEVEL=info \
  --dns-name-label go-agentic

# Get public IP
az container show \
  --resource-group myresourcegroup \
  --name go-agentic \
  --query ipAddress.fqdn
```

---

## 🔐 Security Configuration

### SSL/TLS with nginx

```nginx
# nginx.conf
upstream go_agentic {
    server go-agentic:8081;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;

    # SSL certificates
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
    ssl_protocols TLSv1.2 TLSv1.3;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;

    # Proxy
    location / {
        proxy_pass http://go_agentic;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # For SSE streaming
        proxy_buffering off;
        proxy_cache off;
    }
}
```

### Secrets Management

**Option 1: Environment Variables** (for local/Docker)
```bash
export OPENAI_API_KEY="sk-..."
go-agentic --server
```

**Option 2: Kubernetes Secrets**
```bash
kubectl create secret generic openai-secret \
  --from-literal=api-key=$OPENAI_API_KEY
```

**Option 3: AWS Secrets Manager**
```bash
aws secretsmanager create-secret \
  --name go-agentic/openai-key \
  --secret-string "sk-..."
```

**Option 4: HashiCorp Vault**
```bash
vault write secret/go-agentic/openai \
  api_key="sk-..."
```

### Firewall Rules

```bash
# Allow only necessary ports
# Allow HTTPS (443)
sudo ufw allow 443/tcp

# Allow from specific IPs only
sudo ufw allow from 10.0.0.0/8 to any port 8081

# Deny all else
sudo ufw default deny incoming
```

---

## 📊 Monitoring and Observability

### Prometheus Integration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'go-agentic'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
    params:
      format: ['prometheus']
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "go-agentic Monitoring",
    "panels": [
      {
        "title": "Requests Per Second",
        "targets": [{
          "expr": "rate(crew_requests_total[1m])"
        }]
      },
      {
        "title": "Average Response Time",
        "targets": [{
          "expr": "crew_average_request_duration_seconds"
        }]
      },
      {
        "title": "Memory Usage",
        "targets": [{
          "expr": "crew_memory_usage_bytes"
        }]
      }
    ]
  }
}
```

### Log Aggregation (ELK Stack)

```yaml
# filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/go-agentic/*.log

output.elasticsearch:
  hosts: ["elasticsearch:9200"]

setup.kibana:
  host: "kibana:5601"
```

---

## 🔄 Upgrades and Rollbacks

### Rolling Update Strategy

```bash
# Kubernetes rolling update
kubectl set image deployment/go-agentic \
  go-agentic=go-agentic:v2.0

# Monitor progress
kubectl rollout status deployment/go-agentic

# Rollback if needed
kubectl rollout undo deployment/go-agentic
```

### Blue-Green Deployment

```bash
# Deploy new version (green) alongside current (blue)
kubectl apply -f deployment-green.yaml

# Test green environment
curl http://go-agentic-green:8081/health

# Switch traffic
kubectl patch service go-agentic \
  -p '{"spec":{"selector":{"version":"green"}}}'

# Remove blue
kubectl delete deployment go-agentic-blue
```

### Canary Deployment

```bash
# Deploy new version to 10% of pods
kubectl patch deployment go-agentic \
  -p '{"spec":{"replicas":9}}'

kubectl apply -f deployment-canary.yaml

# Monitor metrics
# If good: scale up green
# If bad: rollback
```

---

## ✅ Deployment Checklist

- [ ] Binary builds successfully
- [ ] Docker image builds and runs
- [ ] Configuration files are valid
- [ ] Environment variables set correctly
- [ ] Health check endpoint works
- [ ] Metrics endpoint works
- [ ] API responds to requests
- [ ] Graceful shutdown works
- [ ] Load testing shows expected performance
- [ ] Monitoring/alerting configured
- [ ] Backup strategy in place
- [ ] Rollback procedure tested
- [ ] Security policies applied
- [ ] SSL/TLS configured (if needed)
- [ ] Documentation updated

---

## 🎓 Best Practices

1. **Use configuration management** (ansible, terraform, helm)
2. **Automate deployments** (CI/CD pipelines)
3. **Monitor continuously** (Prometheus, Grafana)
4. **Log aggregation** (ELK, Splunk, Datadog)
5. **Health checks** (readiness, liveness probes)
6. **Graceful shutdown** (respect SIGTERM signal)
7. **Resource limits** (prevent resource exhaustion)
8. **Security scanning** (container images, dependencies)
9. **Testing in staging** (before production)
10. **Runbooks** (for common operations)

---

## 🔗 Related Documentation

- [Architecture Overview](ARCHITECTURE_OVERVIEW.md) - System design
- [Configuration Guide](CONFIGURATION_GUIDE.md) - Configuration reference
- [Troubleshooting Guide](TROUBLESHOOTING_GUIDE.md) - Common issues
- [Metrics Guide](METRICS_GUIDE.md) - Monitoring setup

---

**Version**: 1.0
**Last Updated**: 2025-12-22
**Status**: Production Ready ✅
