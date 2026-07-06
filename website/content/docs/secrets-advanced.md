---
title: "Advanced Secrets"
---

This page covers the internal details of how Magec handles secrets: the expansion pipeline, encryption at rest, compatibility with external environment variables, and recovery procedures. For a practical guide on creating and using secrets, see [Secrets](/docs/secrets/).

Secrets are named values (API keys, tokens, passwords) that you manage through the Admin UI and reference anywhere in your configuration using `${KEY}` syntax. They provide a single place to store sensitive data instead of scattering environment variables across your deployment.

## How secrets work

Each secret has a **key** (like `OPENAI_API_KEY`) and a **value** (the actual API key). When Magec starts, it loads all secrets and injects them as environment variables. Any `${VAR}` reference in `config.yaml` or `store.json` that matches a secret key gets expanded to its value.

This means a backend configured with:

```json
{
  "apiKey": "${OPENAI_API_KEY}"
}
```

…will resolve to the actual API key if you've created a secret with key `OPENAI_API_KEY`.

### The expansion pipeline

On startup, Magec processes secrets before anything else:

1. Load `store.json` from disk (raw, unexpanded)
2. Decrypt any encrypted secret values (if `encryptionKey` is set)
3. Inject all secret key-value pairs into the process environment via `os.Setenv()`
4. Expand all `${VAR}` references across the entire store using `os.ExpandEnv()`
5. Continue with the fully resolved configuration

This two-pass approach means secrets can reference other environment variables, and the rest of the store can reference secrets. Everything resolves in a single consistent pass.

## Encryption at rest

When `server.encryptionKey` is set in `config.yaml`, secret values are encrypted before writing to `store.json`:

| Aspect | Detail |
|--------|--------|
| **Algorithm** | AES-256-GCM (authenticated encryption) |
| **Key derivation** | PBKDF2 with 100,000 iterations, SHA-256 |
| **Source key** | The `encryptionKey` from `config.yaml` |
| **Format** | Encrypted values are stored as `enc:v1:<base64>` (clearly distinguishable from plain text) |

Without an encryption key, secrets are stored as plain text in `store.json` and a warning is logged:

```
WARN  Secrets are stored without encryption — set server.encryptionKey in config
```

## Compatibility with external environment variables

Secrets don't replace environment variables. They complement them. The expansion pipeline merges both sources:

- **Secrets** are injected via `os.Setenv()` before expansion
- **External env vars** (from Docker, Kubernetes, systemd, shell) are already in the environment
- `os.ExpandEnv()` resolves all `${VAR}` references from the combined environment

This means you can use both approaches interchangeably:

| Source | How to set | Available as `${VAR}` |
|--------|-----------|----------------------|
| Magec secret | Admin UI → Secrets | Yes |
| Docker env | `docker run -e KEY=value` | Yes |
| Docker Compose env | `environment:` in `compose.yaml` | Yes |
| Kubernetes env | `env:` in Pod spec | Yes |
| Kubernetes secret | `secretKeyRef` in Pod spec | Yes |
| Shell export | `export KEY=value` | Yes |

If a secret and an external env var have the same key, the **secret takes precedence**: `os.Setenv()` overwrites the existing value.

{{< callout type="info" >}}
The main advantage of Magec secrets over external environment variables is that you can manage them through the Admin UI without restarting the server or redeploying. Add an API key, and it's available immediately on the next store reload.
{{< /callout >}}

## Recovery

### Changing the encryption key

If you change the encryption key, existing encrypted secrets cannot be decrypted. The derived key will be different. Before changing it:

1. Note down all secret values (they're write-only, so you'll need them from their original source)
2. Change `encryptionKey` in `config.yaml`
3. Restart Magec. Encrypted values will fail to decrypt and remain as `enc:v1:...` strings
4. Re-create or update each secret with its plain-text value. They'll be re-encrypted with the new key

### Losing the encryption key

If you lose the encryption key entirely, the same applies. Encrypted values are unrecoverable without the original key. You'll need to re-enter all secret values from their original sources.

### Removing the encryption key

If you remove `encryptionKey` from `config.yaml`:

1. Existing encrypted secrets cannot be decrypted. You'll need to re-create them (they'll be stored as plain text)
2. New secrets will be stored without encryption and a warning will be logged

{{< callout type="info" >}}
The `encryptionKey` is independent from `adminPassword`. You can have admin authentication without encryption, or encryption without authentication. Using both is recommended for production.
{{< /callout >}}
