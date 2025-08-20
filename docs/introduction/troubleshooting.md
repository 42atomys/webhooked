# Troubleshooting

## Common Issues

### Webhook Not Found (404)

**Symptom:**

```
HTTP/1.1 404 Not Found
```

**Causes & Solutions:**

1.  **Incorrect URL path**

    ```bash
    # Wrong: Missing prefix
    curl http://localhost:8080/my-webhook

    # Correct: Include full path
    curl http://localhost:8080/webhooks/v1alpha2/my-webhook
    ```
2.  **Configuration mismatch**

    ```yaml
    # Configuration
    entrypointUrl: /github/events

    # URL should be:
    /webhooks/v1alpha2/github/events
    ```

### Authentication Failed (401)

**Symptom:**

```json
{
  "error": "security validation failed"
}
```

**Solutions:**

1.  **Check secret/token**

    ```bash
    # Verify environment variable is set
    echo $WEBHOOK_SECRET

    # Ensure it matches configuration
    export WEBHOOK_SECRET="correct-secret"
    ```
2.  **Check headers**

    ```bash
    # Missing required header
    curl -H "X-API-Key: key" ...
    ```

### Rate Limit Exceeded (429)

**Symptom:**

```json
{
  "error": "rate limit exceeded",
  "client_ip": "192.168.1.100"
}
```

**Solutions:**

1.  **Increase limits**

    ```yaml
    throttling:
      maxRequests: 10000  # Increase
      window: 60
    ```
2.  **Add burst capacity**

    ```yaml
    throttling:
      burst: 100
      burstWindow: 10
    ```

### Storage Connection Failed

**Symptom:**

```
Error: storage failed: connection refused
```

**Solutions:**

1.  **Check connectivity**

    ```bash
    # Test Redis
    redis-cli -h localhost ping
    ```
2.  **Verify credentials**

    ```yaml
    storage:
      - type: redis
        specs:
          host: localhost  # Correct host
          port: 6379      # Correct port
          password:       # Correct password
            valueFrom:
              envRef: REDIS_PASSWORD
    ```

### Template Errors

**Symptom:**

```
Error: template parsing failed
```

**Solutions:**

1.  **Check syntax**

    ```yaml
    # Wrong: Missing closing brace
    templateString: '{{ .Value }'

    # Correct
    templateString: '{{ .Value }}'
    ```
2.  **Handle nil values**

    ```yaml
    templateString: |
      {{ with .Data }}
        {{ .Field | default "N/A" }}
      {{ end }}
    ```
3.  **Validate JSON**

    ```yaml
    templateString: |
      {{ if .Payload }}
        {{ with $data := fromJSON .Payload }}
          {{ $data.field }}
        {{ end }}
      {{ end }}
    ```

## Debugging Techniques

#### Enable Debug Logging

```bash
export WH_DEBUG=true
./webhooked serve --config webhooked.yaml
```

#### Check Configuration

```bash
# Validate configuration
./webhooked --validate --config webhooked.yaml
```

#### Monitor Metrics

```bash
# Check metrics endpoint
curl http://localhost:8080/metrics | grep webhooked

# Key metrics to watch
webhooked_http_requests_total
webhooked_storage_errors_total
webhooked_security_rejections_total
```

#### Test Individual Components

```bash
# Test webhook endpoint
curl -i -X POST http://localhost:8080/webhooks/v1alpha2/test \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'

# Test health endpoint
curl http://localhost:8080/health

# Test readiness
curl http://localhost:8080/ready
```

## Error Messages

| Error                            | Meaning                | Solution                |
| -------------------------------- | ---------------------- | ----------------------- |
| `spec not found`                 | Webhook not configured | Check configuration     |
| `security validation failed`     | Auth failed            | Check credentials       |
| `rate limit exceeded`            | Too many requests      | Wait or increase limits |
| `connection refused`             | Storage down           | Check storage service   |
| `context deadline exceeded`      | Timeout                | Check storage service   |
| `invalid template`               | Template error         | Fix template syntax     |
| `environment variable not found` | Missing env var        | Export variable         |



## FAQ

<details>

<summary><strong>Q: Why am I getting 404 errors?</strong></summary>

A: Check that your URL includes `/webhooks/v1alpha2/` prefix.

</details>

<details>

<summary><strong>Q: How do I increase request limits?</strong></summary>

&#x20;A: Adjust `throttling.maxRequests` in configuration.

</details>

<details>

<summary><strong>Q: Can I disable authentication for testing?</strong> </summary>

A: Yes, use `security.type: noop` (never in production).

</details>

<details>

<summary><strong>Q: How do I debug template errors?</strong> </summary>

A: Enable debug mode with `WH_DEBUG=true`.

</details>

<details>

<summary><strong>Q: Why is storage failing?</strong> </summary>

A: Check connectivity, credentials, and firewall rules.

</details>

## Getting Help

1. **Check logs**: Enable debug logging
2. **Review configuration**: Validate YAML syntax
3. **Test components**: Isolate the issue
4. **GitHub Issues**: Report bugs
5. **Community**: Ask questions

***

_Most issues can be resolved by checking configuration and connectivity._
