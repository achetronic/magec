# Contexto de Sesión

## Estado Actual

Todo funciona. El servidor compila y corre en `:8080` (voice-ui) y `:8081` (admin-ui). Multi-agente, hot-reload, rename con cascade, client auth, y voice endpoints — todo implementado y verificado.

## Trabajo Completado en Esta Sesión

### Client entity (reemplaza Device + Telegram config de agente)

- Nueva entidad `ClientDefinition` en `store/types.go` con `Name`, `Type`, `Token`, `AllowedAgents`, `Enabled`, `Config`
- `ClientConfig` struct con campos opcionales por plataforma: `Telegram`, `Discord`, `Slack`
- `TelegramClientConfig`, `DiscordClientConfig`, `SlackClientConfig` structs tipados
- `TelegramConfig` eliminado de `AgentDefinition` — la config de Telegram ahora vive en `Client.Config.Telegram`
- `Device` struct eliminado, reemplazado por `ClientDefinition`
- `StoreData.Devices` → `StoreData.Clients`
- CRUD completo en store: `ListClients`, `GetClient`, `GetClientByToken`, `CreateClient`, `UpdateClient`, `DeleteClient`, `RegenerateClientToken`, `RenameClient`
- Migración automática `Device` → `Client` en `loadFromDisk()` para store.json legacy
- `allowedAgents[0]` es el default agent implícito (no hay campo `defaultAgent` separado)

### Client type registry (`server/client/`)
- Patrón idéntico a `server/memory/`: `Provider` interface + registro global
- `Provider.ConfigFields()` retorna specs de campos para formularios dinámicos
- Tipos: `device` (sin config extra), `telegram` (botToken, allowedUsers, allowedChats, responseMode)
- Blank imports en `main.go` para auto-registro

### Admin API actualizada
- `/clients` CRUD reemplaza `/devices`
- `GET /clients/types` retorna tipos registrados con field specs (como `/memory/types`)
- `POST /clients/{name}/regenerate-token` regenera token
- Validación de tipo contra registry en create
- Overview: `clientsCount` en vez de `devicesCount`

### Auth middleware
- `deviceAuthMiddleware` → `clientAuthMiddleware`
- `X-Device-Name` header → `X-Client-Name`
- Misma lógica: si no hay clients → open mode, si hay → requiere Bearer token
- `/api/v1/device/info` mantiene URL para backward compat con voice-ui

### Telegram startup
- Ya no se lee de `defaultAgent.Telegram` en main.go
- Itera `ListClients()`, filtra `type == "telegram"`, arranca un bot por client
- Cada bot de Telegram se asocia al primer agente del client (`allowedAgents[0]`)

### Admin UI
- Tab "Devices" → "Clients"
- Formulario dinámico basado en `type`: al seleccionar "telegram" aparecen los campos de Telegram (fetched desde `/clients/types`)
- Sección de Telegram eliminada del diálogo de agente
- Badge "TG" eliminado de la vista de agente

### Rename cascade
- `RenameAgent` actualiza `Client.AllowedAgents[]` (antes era `Device.DefaultAgent` + `Device.AllowedAgents[]`)

## Estado del Store Actual

- Agents: `magec` (Magec), `itahisa` (Itahisa) — ambos usan `Ollama` con `qwen2.5:7b`
- Backends: `Ollama` (openai, localhost:11434), `Parakeet TDT`, `OpenAI Edge TTS`
- Memory: `Redis` (session), `Postgres` (longterm con nomic-embed-text)
- Clients: migrado automáticamente desde Devices al cargar store.json legacy

## Comandos

```bash
cd server && go build ./...   # Compilar
make dev                      # Compilar + ejecutar
make swagger                  # Regenerar Swagger docs
```

## Entorno

- Go 1.25.5 (GOPATH=GOROOT warning es cosmético)
- ADK v0.3.0, adk-utils-go v0.1.7
- ONNX Runtime en `./bin/lib/onnxruntime-linux-x64-1.23.2/lib/libonnxruntime.so`

## Siguiente Sesión

1. Regenerar Swagger docs (`make swagger`) para reflejar Client endpoints
2. Probar flujo completo: crear client tipo device en admin-ui, emparejar voice-ui
3. Probar flujo Telegram: crear client tipo telegram en admin-ui, verificar que el bot arranca
4. Implementar Command entity (prompts con triggers cron/webhook) — futuro
5. Actualizar AGENTS.md con la nueva arquitectura de Client
