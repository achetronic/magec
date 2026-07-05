---
title: "Admin Password"
---

By default, the Admin UI and Admin API are open. Anyone who can reach port `8081` can view and modify your entire configuration. Setting an admin password adds authentication to all admin endpoints.

## Enabling authentication

Add `adminPassword` to the server section of your `config.yaml`:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  adminPort: 8081
  adminPassword: ${MAGEC_ADMIN_PASSWORD}
```

Then set the environment variable before starting Magec:

```bash
export MAGEC_ADMIN_PASSWORD="your-strong-password"
magec --config config.yaml
```

When the password is set, every request to `/api/` endpoints requires a `Authorization: Bearer <password>` header. The Admin UI handles this automatically: it shows a login screen on first load and sends the header on every subsequent request.

## How it works

| Aspect | Detail |
|--------|--------|
| **Header** | `Authorization: Bearer <password>` on all `/api/` requests |
| **Comparison** | Constant-time (`crypto/subtle`), immune to timing attacks |
| **Rate limiting** | 5 failed attempts per minute per IP address |
| **Static files** | The Admin UI itself (HTML, CSS, JS) is always accessible. Only API calls require auth |
| **CORS preflight** | `OPTIONS` requests bypass authentication |
| **Auth check** | `GET /api/v1/admin/auth/check`: returns `200` if the password is correct, `401` otherwise |

## Login screen

When authentication is enabled, the Admin UI shows a password prompt before anything else:

- The password is stored **in memory only**. Closing the browser tab requires re-authentication
- No cookies, no `localStorage`, no session tokens
- The UI calls `/api/v1/admin/auth/check` on load to determine if authentication is needed

## Docker and Kubernetes

Pass the password as an environment variable. Never hardcode it in `config.yaml`:

### Docker Compose

```yaml
services:
  magec:
    image: ghcr.io/achetronic/magec:latest
    environment:
      MAGEC_ADMIN_PASSWORD: "${MAGEC_ADMIN_PASSWORD}"
```

### Kubernetes

```yaml
env:
  - name: MAGEC_ADMIN_PASSWORD
    valueFrom:
      secretKeyRef:
        name: magec-secrets
        key: admin-password
```

## What happens without a password

If `adminPassword` is not set (or is empty), the server starts normally with a warning in the logs:

```
WARN  Admin API is unprotected — set adminPassword in config.yaml
```

All admin endpoints work without authentication. This is fine for local development but should never be used in production or anywhere the admin port is network-accessible.

## Secrets encryption

The admin password also serves as the encryption key for [Secrets](/docs/secrets/). When a password is set, secret values are encrypted at rest using AES-256-GCM with a PBKDF2-derived key. Without a password, secrets are stored in plain text and a warning is logged.

{{< callout type="info" >}}
The admin password is a single shared credential. There are no user accounts or roles. If you need network-level isolation, restrict access to the admin port using firewall rules or a reverse proxy.
{{< /callout >}}
