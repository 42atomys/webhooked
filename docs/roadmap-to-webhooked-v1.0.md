---
description: The roadmap to stable version
icon: rocket-launch
---

# Roadmap to Webhooked v1.0

## Key Objectives

{% hint style="info" %}
All objectives comes from feedback, suggestions and personal knowledge. You can discuss the v1.0 directly [on GitHub](https://github.com/42atomys/webhooked/discussions/199).
{% endhint %}

### :clock1: Response Headers - <mark style="color:orange;">TO DO</mark>

Allow editing headers in the response of webhooks with formatting feature.

### :clock1: Prometheus Metrics - <mark style="color:orange;">TO DO</mark>

### :clock1: Allow customize rate limit error response - <mark style="color:orange;">TO DO</mark>

```yaml
# Example
  errorResponse:
    statusCode: 429
    body: |
      {
        "error": {
          "code": "RATE_LIMIT_EXCEEDED",
          "message": "Too many requests. Please retry after {{ .ResetTime }}",
          "limit": {{ .Limit }},
          "window": {{ .Window }},
          "retry_after": {{ .RetryAfter }}
        }
      }
```

### :clock1: Send back ratelimit information on headers - <mark style="color:orange;">TO DO</mark>

Rate limit information in response headers:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1609459200
```

### :clock1: Per wbhook throttling - <mark style="color:orange;">TO DO</mark>

Different limits based on paths:

```yaml
# Per-webhook rate limits
webhooks:
  - name: free-tier
    throttling:
      maxRequests: 100
      window: 3600
  
  - name: premium-tier
    throttling:
      maxRequests: 10000
      window: 3600
```

### :clock1: Advanced Throttling - <mark style="color:orange;">TO DO</mark>

Different limits for different times:

```yaml
throttling:
  # Time-based rules
  timeRules:
    - schedule: "0 9-17 * * MON-FRI"  # Business hours
      maxRequests: 1000
      window: 60
    - schedule: "0 0-23 * * SAT,SUN"  # Weekends
      maxRequests: 500
      window: 60
```

### :clock1: Allow custom metrics labels - <mark style="color:orange;">TO DO</mark>

```yaml
# Custom labels
metrics:
  labels:
    environment: production
    region: us-east-1
    team: platform
```

### :clock1: \[STORAGE]\[Redis] Allow formatting and valuable of key - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[Redis] Allow usage of redis clusters (nodes) - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[Redis] Allow setting the command used - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[Redis] Allow passing TTL - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[Redis] Retry mechanism  - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[Postgres] Allow formatting and valuable of query - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[Postgres] Allow connection pooling - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[Postgres] Retry mechanism  - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[RabbitMQ] Allow formatting and valuable of queue name - <mark style="color:orange;">TO DO</mark>

### :clock1: \[STORAGE]\[RabbitMQ] Allow passing args / headers - <mark style="color:orange;">TO DO</mark>

### :white\_check\_mark: Function Library - <mark style="color:green;">**DONE**</mark>

Use an external library (selected sprout) to provide a complete list of tempalte functions

{% hint style="success" %}
These features are implemented on v0.9, documentation can be found here :

[formatting.md](configuration/formatting.md "mention")
{% endhint %}
