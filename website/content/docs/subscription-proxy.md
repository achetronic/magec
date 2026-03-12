---
title: "Using a subscription proxy"
---

Some AI providers offer paid subscriptions (e.g., Claude Max/Pro, ChatGPT Plus) that are separate from their API billing. A **subscription proxy** lets you route Magec's API requests through your existing subscription instead of paying for API access separately.

[CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) is an open-source proxy that translates subscription OAuth credentials into a standard API. It supports multiple providers and can be used with any Magec backend type that has a configurable `url` field.

## How it works

```
Magec  ──▶  CLIProxyAPI (localhost:8317)  ──▶  Provider API
              ▲
              │ OAuth credentials
              │ from your subscription
```

Magec talks to CLIProxyAPI as if it were the provider's official API. CLIProxyAPI authenticates using the OAuth tokens from your subscription login and forwards the request upstream.

## Setting up CLIProxyAPI

### Docker (recommended)

If you used the interactive installer and enabled the subscription proxy, CLIProxyAPI is already running. Otherwise, add it to your `docker-compose.yaml`:

```yaml
services:
  cliproxyapi:
    image: eceasy/cli-proxy-api:latest
    ports:
      - "54545:54545"   # OAuth callback (required for login only)
    volumes:
      - ./cliproxyapi/config.yaml:/CLIProxyAPI/config.yaml
      - cliproxyapi_auth:/CLIProxyAPI/auth
    restart: unless-stopped

volumes:
  cliproxyapi_auth:
```

Create the config file at `cliproxyapi/config.yaml`:

```yaml
host: "0.0.0.0"
port: 8317
auth-dir: "/CLIProxyAPI/auth"
api-keys:
  - "sk-magec-local"
```

### Binary

Download the latest release from [CLIProxyAPI releases](https://github.com/router-for-me/CLIProxyAPI/releases) and run it directly.

## Logging in to a provider

CLIProxyAPI requires you to authenticate with your provider's subscription account. The login command starts a temporary OAuth callback server, so you must **stop the running CLIProxyAPI service first** to free its ports:

```bash
docker compose stop cliproxyapi
docker compose run --rm --service-ports cliproxyapi /CLIProxyAPI/CLIProxyAPI --no-browser --<provider-login>
docker compose up -d cliproxyapi
```

Replace `--<provider-login>` with the login flag for your provider (see examples below).

The command prints an authorization URL — open it in your browser, authorize, and wait for the callback. Your credentials are stored in the `cliproxyapi_auth` volume and persist across restarts.

## Provider examples

### Claude (Anthropic)

Login flag: `--claude-login` ([docs](https://help.router-for.me/configuration/provider/claude-code.html))

```bash
docker compose run --rm --service-ports cliproxyapi /CLIProxyAPI/CLIProxyAPI --no-browser --claude-login
```

Then create an `anthropic` backend in the Admin UI:

| Field | Value |
|-------|-------|
| Name | `Claude (Subscription)` |
| Type | `anthropic` |
| URL | `http://cliproxyapi:8317` (Docker) or `http://localhost:8317` (local) |
| API Key | `sk-magec-local` |

{{< callout type="warning" >}}
The **URL** field is required. Without it, Magec sends requests to the default Anthropic API and authentication will fail.
{{< /callout >}}

### Gemini (Google)

Login flag: `--login` ([docs](https://help.router-for.me/configuration/provider/gemini-cli.html))

```bash
docker compose run --rm --service-ports cliproxyapi /CLIProxyAPI/CLIProxyAPI --no-browser --login
```

Then create an `openai` backend in the Admin UI (CLIProxyAPI exposes Gemini models through an OpenAI-compatible API):

| Field | Value |
|-------|-------|
| Name | `Gemini (Subscription)` |
| Type | `openai` |
| URL | `http://cliproxyapi:8317/v1` (Docker) or `http://localhost:8317/v1` (local) |
| API Key | `sk-magec-local` |

{{< callout type="info" >}}
CLIProxyAPI supports additional providers like [Codex](https://help.router-for.me/configuration/provider/codex.html) and [Antigravity](https://help.router-for.me/configuration/provider/antigravity.html). See the [full provider list](https://help.router-for.me/) for all available login commands and configuration options.
{{< /callout >}}

## Verifying the setup

After logging in, check that the proxy is working:

```bash
curl http://localhost:8317/v1/models -H "X-Api-Key: sk-magec-local"
```

If you see a list of models, the proxy is ready. Assign the backend to any agent and start chatting.

{{< callout type="info" >}}
The interactive installer can set up CLIProxyAPI automatically — just answer "yes" when asked about the subscription proxy. See [CLIProxyAPI on GitHub](https://github.com/router-for-me/CLIProxyAPI) for full configuration options.
{{< /callout >}}

{{< callout type="info" >}}
CLIProxyAPI is a third-party project not affiliated with any AI provider. You are responsible for ensuring your usage complies with each provider's terms of service.
{{< /callout >}}
