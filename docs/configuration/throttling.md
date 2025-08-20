# Throttling

## Introduction

Protect your webhook endpoints from abuse and ensure fair resource usage with Webhooked's comprehensive rate limiting capabilities.

{% hint style="info" %}
Rate limiting is applied per client IP, keep it in mind
{% endhint %}

## Rate Limiting Algorithms

### Sliding Window

```yaml
throttling:
  enabled: true
  maxRequests: 1000
  window: 3600        # 1 hour sliding window
```

### Token Bucket Algorithm

The default algorithm used by Webhooked:

```yaml
throttling:
  enabled: true
  maxRequests: 100    # Bucket capacity
  window: 60          # Refill rate: 100 tokens per 60 seconds
  burst: 20           # Allow bursts up to 20 requests
```

**How it works:**

1. Bucket starts with `maxRequests` tokens
2. Each request consumes one token
3. Tokens refill at rate of `maxRequests/window`
4. Burst allows temporary exceeding of rate

### Request Queuing

#### Basic Queue Configuration

```yaml
throttling:
  enabled: true
  maxRequests: 100
  window: 60
  queueCapacity: 50         # Queue up to 50 requests
  queueTimeout: 5           # Wait max 5 seconds
  queueTimeoutCode: 503     # Return 503 on timeout
```

#### Queue Behavior

When rate limit is exceeded:

1. Request enters queue if capacity available
2. Waits for token to become available
3. Processes when token available or timeout
4. Returns configured status code on timeout

## Configuration

```yaml
specs:
  - throttling:
      enabled: true
      maxRequests: 1000     # Sustained rate
      window: 60            # Per minute
      burst: 50             # Allow burst of 50 requests
      burstWindow: 10       # Within 10 seconds
```

| Property           | Description             | Default | Example |
| ------------------ | ----------------------- | ------- | ------- |
| `enabled`          | Enable rate limiting    | `false` | `true`  |
| `maxRequests`      | Max requests per window | -       | `1000`  |
| `window`           | Time window (seconds)   | -       | `60`    |
| `burst`            | Burst capacity          | -       | `50`    |
| `burstWindow`      | Burst window (seconds)  | -       | `10`    |
| `queueCapacity`    | Request queue size      | `0`     | `100`   |
| `queueTimeout`     | Queue timeout (seconds) | `0`     | `5`     |
| `queueTimeoutCode` | Status code on timeout  | `503`   | `503`   |

## Common Configurations

### API Rate Limiting

Standard API rate limiting:

```yaml
specs:
  - throttling:
      enabled: true
      maxRequests: 10000    # 10K requests
      window: 3600          # Per hour
      burst: 100            # Allow small bursts
      burstWindow: 60       # Per minute
```

### Webhook Protection

Protect against webhook storms:

```yaml
specs:
  - throttling:
      enabled: true
      maxRequests: 100      # Conservative limit
      window: 60            # Per minute
      burst: 20             # Small burst allowance
      burstWindow: 10       # Quick burst window
```

### High-Traffic Configuration

For high-volume webhooks:

```yaml
specs:
  - throttling:
      enabled: true
      maxRequests: 50000    # 50K requests
      window: 60            # Per minute
      burst: 1000           # Large burst capacity
      burstWindow: 5        # 5 second burst
      queueCapacity: 500    # Queue excess requests
      queueTimeout: 10      # Wait up to 10 seconds
```

