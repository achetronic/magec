# Contexto de Sesión

## Estado Actual

Todo funciona. El servidor compila y corre en `:8080` (voice-ui) y `:8081` (admin-ui). Multi-agente, hot-reload, rename con cascade, device auth, y voice endpoints — todo implementado y verificado.

## Trabajo Completado en Sesiones Recientes

### Multi-agente (servidor)
- `agent.New()` acepta `[]store.AgentDefinition`, crea un LLM agent por definición
- ADK `NewMultiLoader(root, ...others)` enruta por `appName`

### Multi-agente (voice-ui)
- `AgentClient.js`, `SessionService.js`, `OpenAITTS.js`, `RemoteTranscriber.js` tienen `setAgent(agentId)`
- `app.js` propaga el agente al cambiar en el dropdown + crea nueva sesión
- `config.js`: `appName` cambiado de `'magec_agent'` a `'default'`

### Hot-reload de agentes
- Store tiene `OnChange()` channel que se dispara en cada `persist()`
- `agentRouterHandler` con `sync.RWMutex` reconstruye el agente con debounce de 500ms

### Voice endpoints rediseñados
- `/api/v1/voice/{agentId}/speech` — proxy TTS (resuelve backend dinámicamente por agente)
- `/api/v1/voice/{agentId}/transcription` — proxy STT (resuelve backend dinámicamente por agente)
- `/api/v1/voice/events` — WebSocket (era `/api/v1/voice-events`)
- Los proxies ahora envían el `apiKey` del backend como `Authorization: Bearer` al upstream

### Rename con cascade
- 6 métodos: `RenameBackend`, `RenameMemoryProvider`, `RenameMCPServer`, `RenameAgent`, `RenameDevice`, `RenameCronJob`
- Mapa de cascada:
  - Backend → Agent.LLM/TTS/Transcription.Backend + MemoryProvider.Embedding.Backend
  - MemoryProvider → Agent.Memory.Session/LongTerm
  - MCPServer → Agent.MCPServers[]
  - Agent → Device.DefaultAgent + Device.AllowedAgents[] + CronJob.AgentID
- Admin handlers detectan rename: si `body.Name != urlName`, llaman `store.Rename*()` primero

### Wake word model name
- Campo `Name` añadido a `WakewordModelInfo` en `voice/handler.go`
- Mensaje de capabilities incluye `name`, `id`, y `phrase`

### Bug fix: 401 en voice speech
- El 401 NO venía del `deviceAuthMiddleware` — el device auth pasaba bien
- El error `"Missing or invalid API key"` venía del backend TTS upstream (OpenAI Edge TTS)
- Fix: `serveSpeechProxy` y `serveTranscriptionProxy` ahora envían `Authorization: Bearer <apiKey>` al upstream

### Admin UI modal fix
- Cancel/close buttons tenían `type="submit"` dentro de `<form method="dialog">`
- Los inputs `required` impedían cerrar el modal con Cancel/X si estaban vacíos
- Fix: `formnovalidate` en los 12 botones de cancel/close

## Estado del Store Actual

- Agents: `magec` (Magec), `itahisa` (Itahisa) — ambos usan `Ollama` con `qwen2.5:7b`
- Backends: `Ollama` (openai, localhost:11434), `Parakeet TDT`, `OpenAI Edge TTS`
- Memory: `Redis` (session), `Postgres` (longterm con nomic-embed-text)
- Devices: `opo` con token `mgc_...`, defaultAgent `magec`, allowedAgents `[magec, itahisa]`

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

1. Probar el flujo completo de voz con cambio de agentes (voice-ui)
2. Verificar que TTS y transcripción usan los backends correctos por agente
3. Cualquier otra tarea que el usuario decida
