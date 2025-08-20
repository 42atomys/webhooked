---
description: >-
  This guide will help you get Webhooked up and running in minutes. We'll cover
  installation, basic configuration, and your first webhook.
---

# Getting Started

{% hint style="info" %}
This page evolves with each version based on modifications and community feedback. Having trouble following this guide?

[Open an issue](https://github.com/go-sprout/sprout/issues/new/choose) to get help, and contribute to improving this guide for future users. :purple\_heart:
{% endhint %}

## Prerequisites

* **For Docker**: Docker and Docker Compose installed
* **For Binary**: Linux/macOS/Windows system
* **For Source**: Go 1.24+ installed
* **Optional**: Redis, PostgreSQL, RabbitMQ, etc... for storage backends

## Quick Start (5 minutes)

### Option 1: Docker (Recommended)

```bash
# Create a configuration file
cat > webhooked.yaml << 'EOF'
apiVersion: v1alpha2
kind: Configuration
metadata:
  name: my-webhooks
specs:
  - webhooks:
    - name: hello-webhook
      entrypointUrl: /webhook/hello
      security:
        type: noop  # No auth for testing
      response:
        statusCode: 200
        formatting:
          templateString: '{"message": "Hello from Webhooked!", "received": {{ .Payload }}}'
EOF

# Run Webhooked
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/webhooked.yaml:/config/webhooked.yaml \
  atomys/webhooked:latest \
  --config /config/webhooked.yaml

# Test your webhook
curl -X POST http://localhost:8080/webhooks/v1alpha2/webhook/hello \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

#### Option 2: Binary Installation

```bash
# Download latest release (Linux example)
curl -L https://github.com/42atomys/webhooked/releases/latest/download/webhooked-linux-amd64.tar.gz | tar xz

# Create configuration
./webhooked --init --config webhooked.yaml

# Start the server
./webhooked --config webhooked.yaml --port 8080
```

#### Option 3: Install from Source

```bash
# Clone and build
git clone https://github.com/42atomys/webhooked.git
cd webhooked
go build -o webhooked ./cmd/webhooked

# Create configuration
./webhooked --init --config webhooked.yaml

# Run
./webhooked --config webhooked.yaml
```

### Your First Real Webhook

Let's create a more realistic webhook that stores data in Redis:

#### 1. Create Configuration

```yaml
# webhooked.yaml
apiVersion: v1alpha2
kind: Configuration
metadata:
  name: production-webhooks
specs:
  - metricsEnabled: true
    webhooks:
      - name: user-events
        entrypointUrl: /users
        security:
          type: custom
          specs:
            condition: |
              {{ eq (.Request.Header.Peek "X-API-Key" | toString) (env "API_TOKEN") }}
        storage:
          - type: redis
            specs:
              host: redis
              port: 6379
              database: 0
              key: "webhooks:users"
            formatting:
              templateString: |
                {
                  "timestamp": "{{ now | date \"2006-01-02T15:04:05Z07:00\" }}",
                  "event": {{ .Payload }},
                  "headers": {
                    "user-agent": "{{ .Request.Header.Peek \"User-Agent\" | toString }}",
                    "content-type": "{{ .Request.Header.Peek \"Content-Type\" | toString }}"
                  }
                }
        response:
          statusCode: 201
          headers:
            X-Request-ID: "{{ uuidv4 }}"
          formatting:
            templateString: |
              {
                "status": "accepted",
                "timestamp": "{{ now | unixEpoch }}"
              }
```

#### 2. Start Services with Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  webhooked:
    image: atomys/webhooked:latest
    ports:
      - "8080:8080"
    volumes:
      - ./webhooked.yaml:/webhooked.yaml
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

```bash
# Start everything
docker-compose up -d

# Check logs
docker-compose logs -f webhooked
```

#### 3. Test the Webhook

```bash
# Send a test webhook
curl -X POST http://localhost:8080/webhooks/v1alpha2/users \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secret-key-123" \
  -d '{
    "user_id": "12345",
    "action": "login",
    "ip": "192.168.1.1"
  }'

# Response:
# {
#   "status": "accepted",
#   "timestamp": "1704067200"
# }

# Check Redis for stored data
docker exec -it $(docker ps -qf "name=redis") redis-cli
> GET "webhooks:users"
```

### Essential Configuration Concepts

#### Webhook URL Structure

All webhook endpoints follow this pattern:

```
/webhooks/{apiVersion}/{entrypointUrl}
```

Example:

* Configuration: `entrypointUrl: /my/webhook`
* Actual URL: `/webhooks/v1alpha2/my/webhook`

#### Security Providers

Webhooked supports multiple security providers:

{% hint style="info" %}
See the dedicated page [security-layer.md](../configuration/security-layer.md "mention")
{% endhint %}

#### Storage Backends

Configure one or more storage backends:

{% hint style="info" %}
See the dedicated page [storage-layer.md](../configuration/storage-layer.md "mention")
{% endhint %}

### Environment Variables

Webhooked supports environment variable substitution:

```yaml
security:
  type: github
  specs:
    secret:
      valueFrom:
        envRef: GITHUB_SECRET

storage:
  - type: redis
    specs:
      host:
        valueFrom:
          envRef: REDIS_HOST
      password:
        valueFrom:
          envRef: REDIS_PASSWORD
```

Set environment variables:

```bash
export GITHUB_SECRET="my-secret"
export REDIS_HOST="redis.example.com"
export REDIS_PASSWORD="redis-password"

./webhooked serve --config webhooked.yaml
```

### Health Checks

Webhooked provides built-in health endpoints:

```bash
# Liveness probe
curl http://localhost:8080/health
# {"status":"healthy","version":"v1.0.0"}

# Readiness probe
curl http://localhost:8080/ready
# {"status":"ready","version":"v1.0.0"}
```

### Monitoring

When metrics are enabled, Prometheus metrics are available:

```yaml
specs:
  - metricsEnabled: true
```

```bash
# Access metrics
curl http://localhost:8080/metrics

# Example metrics:
# webhooked_http_requests_total{webhook="user-events",status="200"} 42
# webhooked_http_request_duration_seconds{webhook="user-events",quantile="0.99"} 0.05
```

### Common Patterns

#### 1. Multi-Environment Configuration

```yaml
# Use environment variables for environment-specific values
specs:
  - webhooks:
    - name: payment-webhook
      entrypointUrl: /payments
      security:
        type: custom
        specs:
          condition: |
            {{ eq (.Request.Header.Peek "X-Token" | toString) (env "WEBHOOK_TOKEN") }}
```

#### 2. Data Transformation

```yaml
# Transform data before storage
formatting:
  templateString: |
    {{ with $event := fromJSON .Payload }}
    {
      "event_id": "{{ uuidv4 }}",
      "timestamp": "{{ now }}",
      "user": "{{ $event.user_id }}",
      "action": "{{ $event.action | upper }}",
      "metadata": {{ $event | toJSON }}
    }
    {{ end }}
```



***

_Congratulations! You're now ready to process webhooks at scale with Webhooked._
