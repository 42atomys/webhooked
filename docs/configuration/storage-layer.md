# Storage Layer

## Introduction

Webhooked supports multiple storage backends for persisting webhook data. You can write to one or multiple backends simultaneously, with each backend supporting its own formatting and transformation.

## Storage Context Variables

Variables available in security conditions:

| Variable             | Description                                            | Example                                   |
| -------------------- | ------------------------------------------------------ | ----------------------------------------- |
| `.SpecName`          | Name of the current spec as defined in config          | `"user-events"`                           |
| `.SpecEntrypointURL` | EntrypointURL of the current spec as defined in config | `"/user-events"`                          |
| `.ConnID`            | Unique connection ID                                   | `123549841`                               |
| `.ConnTime`          | Connection established time                            | time.Time object `2025-08-20T21:10:00Z`   |
| `.Host`              | Host header of request                                 | `"example.com"`                           |
| `.IsTLS`             | Whether request is HTTPS                               | `true`                                    |
| `.Method`            | HTTP method used                                       | `"POST"`                                  |
| `.Payload`           | Raw request body                                       | `{"data": "value"}`                       |
| `.QueryArgs`         | Query parameters object                                | `fasthttp.Args{"id":"123","token":"abc"}` |
| `.RemoteAddr`        | Remote Addr                                            | `"192.168.1.10:54321"`                    |
| `.RemoteIP`          | Remote network address                                 | `"192.168.1.1"`                           |
| `.RequestTime`       | Time when request was received                         | time.Time object `2025-08-20T21:10:00Z`   |
| `.Request`           | Full `fasthttp.Request` object                         | `&fasthttp.Request{...}`                  |
| `.URI`               | Request URI                                            | `"/webhooks/..."`                         |
| `.UserAgent`         | Client User-Agent header                               | `"Mozilla/5.0 (X11; Linux x86_64)"`       |
| `.StorageType`       | Type as defined in spec                                | `"redis"`                                 |

## Storages Providers

### Redis

```yaml
storage:
  - type: redis
    specs:
      host: localhost # Valuable - see doc "Sourcing (Valuable)" for more info
      port: 6379      # Valuable - see doc "Sourcing (Valuable)" for more info
      username: user  # Valuable - see doc "Sourcing (Valuable)" for more info
      password: pass  # Valuable - see doc "Sourcing (Valuable)" for more info
      database: 0     # integer
      key: "webhooks:events" # string
```

<table><thead><tr><th width="113">Variable</th><th width="98">Type</th><th width="112" data-type="checkbox">Valuable</th><th width="123" data-type="checkbox">Formatting</th><th>Description</th></tr></thead><tbody><tr><td><code>host</code></td><td>string</td><td>true</td><td>false</td><td>Host of the redis server</td></tr><tr><td><code>port</code></td><td>integer</td><td>true</td><td>false</td><td>Port of the redis server</td></tr><tr><td><code>username</code></td><td>string</td><td>true</td><td>false</td><td>Username to connect to redis</td></tr><tr><td><code>password</code></td><td>string</td><td>true</td><td>false</td><td>Password to connect to redis</td></tr><tr><td><code>database</code></td><td>integer</td><td>false</td><td>false</td><td>Database to use (default to <code>0</code>)</td></tr><tr><td><code>key</code></td><td>string</td><td>false</td><td>false</td><td>Key used to store with <code>RPush</code> command</td></tr></tbody></table>

### PostgreSQL

```yaml
storage:
  - type: postgres
    specs:
      databaseUrl: postgres://user:password@localhost:5432/webhooks?sslmode=disable
      query: |
        INSERT INTO webhook_events (name, payload, received_at)
        VALUES (:name, :payload, NOW())
      args:
        name: "{{ .WebhookName }}"
        payload: "{{ .Payload }}"
```

<table><thead><tr><th width="113">Variable</th><th width="98">Type</th><th width="112" data-type="checkbox">Valuable</th><th width="122" data-type="checkbox">Formatting</th><th>Description</th></tr></thead><tbody><tr><td><code>databaseUrl</code></td><td>string</td><td>true</td><td>false</td><td>Database URL of the postgres server</td></tr><tr><td><code>query</code></td><td>string</td><td>false</td><td>false</td><td>The query to execute with named arguments</td></tr><tr><td><code>args</code></td><td>map</td><td>false</td><td>true</td><td>Map of naed arguments used in the query</td></tr></tbody></table>

### RabbitMQ Storage

```yaml
storage:
  - type: rabbitmq
    specs:
      databaseUrl: amqp://user:password@localhost:5672/
      queueName: webhook-events
      maxAttempt: 10
      durable: true
      deleteWhenUnused: false
      exclusive: false
      noWait: false
      exchange: webhooks
      contentType: text
      mandatory: false
      immediate: false
```

<table><thead><tr><th width="162">Variable</th><th width="98">Type</th><th width="112" data-type="checkbox">Valuable</th><th width="122" data-type="checkbox">Formatting</th><th>Description</th></tr></thead><tbody><tr><td><code>databaseUrl</code></td><td>string</td><td>true</td><td>false</td><td>Database URL of the rabbitmq server</td></tr><tr><td><code>queueName</code></td><td>string</td><td>false</td><td>false</td><td>Name of the queue</td></tr><tr><td><code>maxAttempt</code></td><td>integer</td><td>false</td><td>false</td><td>Maximum retry attempts</td></tr><tr><td><code>durable</code></td><td>bool</td><td>false</td><td>false</td><td>Survive broker restart</td></tr><tr><td><code>deleteWhenUnused</code></td><td>bool</td><td>false</td><td>false</td><td>Don't delete when unused</td></tr><tr><td><code>exclusive</code></td><td>bool</td><td>false</td><td>false</td><td>Not exclusive to connection</td></tr><tr><td><code>noWait</code></td><td>bool</td><td>false</td><td>false</td><td>Wait for confirmation</td></tr><tr><td><code>exchange</code></td><td>string</td><td>false</td><td>false</td><td>Exchange to bind the queue</td></tr><tr><td><code>contentType</code></td><td>string</td><td>false</td><td>false</td><td>Content type for messages</td></tr><tr><td><code>mandatory</code></td><td>bool</td><td>false</td><td>false</td><td>Return message if not routable</td></tr><tr><td><code>immediate</code></td><td>bool</td><td>false</td><td>false</td><td>Deliver immediately or fail</td></tr></tbody></table>

{% hint style="warning" %}
IMPORTANT: Configuration of the queue and exchange must match your existing declaration. If no queue are present, webhooked try to defined it.
{% endhint %}

### NoOp

NoOp storage is for testing and development where persistence isn't needed.

```yaml
storage:
  - type: noop
```

## Storage Formatting

Each storage can have its own data transformation.

```yaml
storage:
  # Raw JSON to Redis
  - type: redis
    specs: # ...
    formatting:
      templateString: "{{ .Payload }}"
  
  # Structured data to RabbitMQ
  - type: rabbitmq
    specs: # ...
    formatting:
      templateString: |
        {
          "id": "{{ (fromJSON .Payload).id }}",
          "type": "{{ (fromJSON .Payload).type }}",
          "timestamp": "{{ now }}"
        }
```

{% hint style="info" %}
See more about formating on the dedicated page : [formatting.md](formatting.md "mention")
{% endhint %}
