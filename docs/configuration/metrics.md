---
description: >-
  Metrics provide visibility into your webhook processing. Use them to ensure
  reliability and performance.
hidden: true
---

# Metrics

## Enabling Metrics

```yaml
specs:
  - metricsEnabled: true
    webhooks:
      - name: my-webhook
        # ...
```

Metrics are available at: `/metrics`

### Available Metrics

#### HTTP Metrics

```prometheus
# Total requests
webhooked_http_requests_total{
  webhook="github-events",
  method="POST",
  status="200"
} 1234

# Request duration histogram
webhooked_http_request_duration_seconds_bucket{
  webhook="github-events",
  le="0.005"
} 100
webhooked_http_request_duration_seconds_bucket{
  webhook="github-events",
  le="0.01"
} 150

# Request size
webhooked_http_request_size_bytes{
  webhook="github-events"
} 2048

# Response size
webhooked_http_response_size_bytes{
  webhook="github-events"
} 256
```

#### Storage Metrics

```prometheus
# Storage operations
webhooked_storage_operations_total{
  webhook="payment-webhook",
  storage="redis",
  operation="store",
  status="success"
} 5678

# Storage duration
webhooked_storage_duration_seconds{
  webhook="payment-webhook",
  storage="postgres",
  operation="store",
  quantile="0.5"
} 0.02

# Storage errors
webhooked_storage_errors_total{
  webhook="payment-webhook",
  storage="redis",
  error="connection_timeout"
} 12
```

#### Security Metrics

```prometheus
# Security validations
webhooked_security_validations_total{
  webhook="api-webhook",
  provider="github",
  result="allowed"
} 9012

webhooked_security_validations_total{
  webhook="api-webhook",
  provider="github",
  result="denied"
} 34

# Security rejections by reason
webhooked_security_rejections_total{
  webhook="api-webhook",
  reason="invalid_signature"
} 25

webhooked_security_rejections_total{
  webhook="api-webhook",
  reason="missing_header"
} 9
```

#### Rate Limiting Metrics

```prometheus
# Rate limit status
webhooked_ratelimit_requests_total{
  status="allowed"
} 10000

webhooked_ratelimit_requests_total{
  status="limited"
} 523

# Current tokens
webhooked_ratelimit_tokens_remaining{
  client="192.168.1.100"
} 45

# Queue metrics
webhooked_ratelimit_queue_size{} 12
webhooked_ratelimit_queue_timeouts_total{} 3
```

#### System Metrics

```prometheus
# Go runtime metrics
go_memstats_alloc_bytes{} 12345678
go_memstats_sys_bytes{} 23456789
go_goroutines{} 42

# Process metrics
process_cpu_seconds_total{} 123.45
process_resident_memory_bytes{} 56789012
process_virtual_memory_bytes{} 89012345
process_open_fds{} 15
```

### Metric Labels

#### Standard Labels

| Label       | Description       | Example Values          |
| ----------- | ----------------- | ----------------------- |
| `webhook`   | Webhook name      | `"github-push"`         |
| `method`    | HTTP method       | `"POST"`, `"GET"`       |
| `status`    | HTTP status code  | `"200"`, `"404"`        |
| `storage`   | Storage backend   | `"redis"`, `"postgres"` |
| `operation` | Operation type    | `"store"`, `"retrieve"` |
| `provider`  | Security provider | `"github"`, `"custom"`  |
