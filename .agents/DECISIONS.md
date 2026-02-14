# Technical Decisions

Decisiones técnicas tomadas por el owner del proyecto. Toda herramienta de IA que trabaje
en este repositorio **debe respetar estas decisiones** y no revertirlas sin aprobación
explícita.

---

## Middlewares en su propio paquete con httpsnoop

**Fecha**: 2026-02-13
**Estado**: Implementado

Los middlewares HTTP (`AccessLog`, `CORS`, `ClientAuth`) viven en `server/middleware/`,
**no** en `main.go`.

Para capturar status code y bytes escritos en el access log, se usa
[`httpsnoop.CaptureMetrics`](https://github.com/felixge/httpsnoop) en vez de un wrapper
propio sobre `http.ResponseWriter`. httpsnoop maneja correctamente `Hijacker`, `Flusher`,
`CloseNotifier`, `Pusher` y cualquier otra interfaz opcional sin código adicional.

**No usar**: `responseRecorder` custom ni wrappers manuales sobre `ResponseWriter`.

---

## Clientes unificados en server/clients/

**Fecha**: 2026-02-13
**Estado**: Implementado

Todos los tipos de cliente viven bajo `server/clients/`. Cada subtipo tiene su
subdirectorio con `spec.go` (JSON Schema del tipo) junto a su runtime:

```
server/clients/
├── provider.go          ← Provider interface + Schema type
├── registry.go          ← Register(), ValidateConfig(), All()
├── executor.go          ← lógica compartida (webhook + cron)
├── direct/
│   └── spec.go          ← Schema (sin runtime)
├── telegram/
│   ├── spec.go
│   └── bot.go
├── webhook/
│   ├── spec.go
│   └── handler.go
└── cron/
    ├── spec.go
    ├── cron.go
    └── scheduler.go
```

El paquete `server/client/` (singular) fue absorbido. La separación anterior
(schemas en `client/`, runtime en `clients/`) no era consistente.
El paquete `server/trigger/` fue eliminado. Webhook y cron son clientes igual que
Telegram — la separación anterior no era consistente con el dominio.

---

## Validación de JSON Schema con google/jsonschema-go

**Fecha**: 2026-02-13
**Estado**: Implementado

La validación de config de los client types usa
[`google/jsonschema-go`](https://github.com/google/jsonschema-go) en vez de lógica
manual. La librería:

- Es de Google, sin dependencias externas (solo stdlib)
- Soporta draft-07 y 2020-12 completos (`oneOf`, `const`, `required`, `enum`,
  `pattern`, `minLength`, `if/then/else`...)
- Valida directamente sobre `map[string]any`
- Incluye `ApplyDefaults` para poblar valores por defecto

**No usar**: validadores manuales de `required`/`oneOf` ni helpers tipo `matchOneOf`
o `jsonEqual`. Delegar siempre en la librería.

---

## Configuración de voz como bloque independiente

**Fecha**: 2026-02-14
**Estado**: Implementado

La configuración relacionada con voz (UI, ONNX runtime) vive en su propio bloque
`voice` en el YAML, **no** dentro de `server`. ONNX Runtime se usa para modelos de
voz de distintos tipos (wake word, VAD, embeddings), así que pertenece al dominio
de voz, no al de infraestructura HTTP.

```yaml
voice:
  ui:
    enabled: true          # Activa/desactiva Voice UI, rutas y static files
  onnxLibraryPath: ""      # Ruta a libonnxruntime.so (default: /usr/lib/libonnxruntime.so)
```

El struct en Go usa sub-structs: `Config.Voice.UI.Enabled` (*bool, default true)
y `Config.Voice.OnnxLibraryPath` (string).

**No poner**: campos de voz dentro de `Server` — ese bloque es solo red/puertos.
