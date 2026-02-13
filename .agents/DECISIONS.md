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

## Clientes en server/clients/

**Fecha**: 2026-02-13
**Estado**: Implementado

Todos los tipos de cliente (webhook, cron, telegram) viven bajo `server/clients/`:

```
server/clients/
├── executor.go          ← lógica compartida (webhook + cron)
├── webhook/webhook.go
├── cron/
│   ├── cron.go
│   └── scheduler.go
└── telegram/telegram.go
```

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
