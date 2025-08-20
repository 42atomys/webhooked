# Sourcing (Valuable)

## Introduction

Webhooked's Valuable system provides flexible value sourcing, allowing configuration values to come from multiple sources including environment variables, files, and static references. This enables secure secret management and environment-specific configurations.

## Overview

The Valuable pattern allows any configuration value to be sourced from:

```yaml
# Direct value
secret: "direct-value"

# Listing value
secret:
  values: ["array", "value"]

# Environment variable
secret:
  valueFrom:
    envRef: ENV_VAR_NAME

# File reference
secret:
  valueFrom:
    fileRef: /path/to/file

# Static reference
secret:
  valueFrom:
    staticRef: "static,comma,separated"
```

## Value Sources

### Direct value

The simplest form - values directly in configuration

```yaml
storage:
  - type: redis
    specs:
      host: localhost     # Direct value
      port: 6379          # Direct value
      database: 0         # Direct value
```

### Environment Variables

Source values from environment variables:

```yaml
security:
  type: github
  specs:
    secret: # Environment Sourced
      valueFrom:
        envRef: GITHUB_WEBHOOK_SECRET

storage:
  - type: redis
    specs:
      host: # Environment Sourced
        valueFrom:
          envRef: REDIS_HOST
      password: # Environment Sourced
        valueFrom:
          envRef: REDIS_PASSWORD
```

### File References

Read values from files (useful for secrets):

```yaml
security:
  type: custom
  specs:
    apiKey: # File Sourced
      valueFrom:
        fileRef: /run/secrets/api-key

storage:
  - type: postgres
    specs:
      databaseUrl: # File Sourced
        valueFrom:
          fileRef: /run/secrets/db-url
```

### Value Sources Ordering

When you provide multiples references, a priority will be applied as following:

1. Values List
2. Direct Value
3. StaticRef
4. EnvRef
5. FileRef

## Security Best Practices

:white\_check\_mark: <mark style="color:$success;">**GOOD**</mark>: Use external source for secrets

```yaml
secret:
  valueFrom:
    envRef: WEBHOOK_SECRET
```

:x: <mark style="color:red;">**BAD**</mark>: Hardcoded secret

```yaml
secret: "my-secret-key"  # Don't do this!
```

## Use cases

### Secret Rotation

Sometimes, you must handle a rotation in your webhook secrets, you can provide both as a comma separated values in multiples ways:

```yaml
security:
  type: github
  specs:
    secret:
      valueFrom:
        fileRef: /run/secrets/api-key
        # -- OR --
        envRef: WEBHOOK_SECRET
        # -- OR --
        staticRef: "old_secret,new_secret"  # Don't do this, this is an example
```

The content of  `/run/secrets/api-key` or `WEBHOOK_SECRET` can be `old_secret,new_secret`, when the configuration are reloaded, both secrets are accepted as valide out of the box.

### With HashiCorp Vault

```yaml
# Mount Vault secrets as files
storage:
  specs:
    password:
      valueFrom:
        fileRef: /vault/secrets/database-password
```

### With Azure Key Vault

```yaml
# Using CSI driver to mount secrets
storage:
  specs:
    apiKey:
      valueFrom:
        fileRef: /mnt/secrets-store/api-key
```

More will coming... If you found a mising case, don't hesitate to [open an issue](https://github.com/42atomys/webhooked/issues) !
