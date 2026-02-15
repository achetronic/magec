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

---

## Store seeds para Docker

**Fecha**: 2026-02-14
**Estado**: Implementado

Al arrancar el contenedor Docker, el usuario elige un preset de datos iniciales
mediante la variable de entorno `MAGEC_SEED`:

| Valor | Comportamiento |
|-------|---------------|
| (vacío / no definido) | Store vacío — el usuario configura todo desde la Admin UI |
| `voice-ui` | Agente Magec + backend Ollama + STT/TTS + memory + cliente VoiceUI |
| `examples` | Todo lo anterior + Research Pipeline + Debate Arena + Software Factory + Telegram + webhooks |

Los ficheros seed viven en `data/seeds/`:
```
data/seeds/
├── voice-ui.json    ← mínimo para Voice UI funcional
└── examples.json    ← demo completa (equivale al store.json de desarrollo)
```

La lógica está en `docker/build/entrypoint.sh`: si `data/store.json` **no existe** y
`MAGEC_SEED` apunta a un seed válido, copia el seed como store inicial antes de
arrancar el servidor. El código Go no sabe nada de seeds — solo lee `store.json`.
Si el fichero ya existe, el seed se ignora (no se sobreescribe datos del usuario).

**No hacer**: Lógica de seeds dentro del código Go. Es responsabilidad del
entrypoint del contenedor, no de la aplicación.

Los seeds usan `${VARIABLES}` para credenciales (apiKey, botToken), expandidas
por `os.ExpandEnv()` al cargar el store.

**No hacer**: Seeds con credenciales hardcodeadas. Usar siempre `${VAR}`.

---

## Website estática para documentación

**Fecha**: 2026-02-14
**Estado**: Implementado

La documentación y landing page del proyecto vive en `website/` como HTML/CSS/JS
estático, preparada para GitHub Pages. No usa frameworks (ni React, ni Hugo, ni
Docusaurus) — solo archivos estáticos que se sirven directamente.

```
website/
├── index.html              ← Landing page con hero, features, arquitectura
├── docs.html               ← Documentación completa con sidebar navegable
├── css/
│   ├── tokens.css          ← Design tokens (colores Canarios, espaciado, etc.)
│   └── style.css           ← Estilos completos
├── js/
│   ├── centella.js         ← Orbe animado (versión decorativa del WaveformRenderer)
│   └── main.js             ← Nav, scroll, reveal animations, docs sidebar
└── assets/                 ← Logo, architecture SVG, screenshots, OG banner
```

Respeta la paleta del proyecto: piedra (volcanic stone), atlántico (cyan),
lava (red), sol (gold), arena (sand text). Dark mode only.

El README.md queda simplificado — highlights y quick start, apuntando a la web
para documentación detallada.

**No hacer**: Meter frameworks de build (Next, Astro, etc.). Es una web estática
que debe funcionar abriendo `index.html` o sirviéndola con cualquier servidor.

---

## Admin UI nunca accede a la User API

**Fecha**: 2026-02-14
**Estado**: Implementado

La Admin UI (puerto 8081) **nunca** debe acceder a la User API (puerto 8080) para
realizar operaciones. Toda la lógica debe ser directa a través de acceso interno
(Go structs, services, stores).

Ejemplo: para borrar una sesión de ADK, el admin handler llama directamente a
`sessionService.Delete()` — no hace HTTP al puerto 8080. Para listar conversaciones,
lee del `ConversationStore` — no llama a endpoints REST.

**Motivo**: El admin es un componente interno con acceso privilegiado. No debe
depender de la autenticación de clientes (`clientAuthMiddleware`) ni de la
disponibilidad de la User API. Si la User API está caída o mal configurada,
el admin debe seguir funcionando.

**No hacer**: `http.Get("http://127.0.0.1:8080/api/v1/...")` desde el admin handler.
Pasar siempre referencias directas a los services internos (session, memory, store).

---

## Memoria centralizada en el launcher

**Fecha**: 2026-02-14
**Estado**: Implementado

La configuración de session y long-term memory es **global**, no por agente.
El launcher de ADK acepta un único `session.Service` y un único `memory.Service`,
así que configurar memoria individualmente por agente es una ilusión — en la
práctica todos usan la misma.

La config global vive en `StoreData.Settings`:

```go
type Settings struct {
    SessionProvider  string `json:"sessionProvider,omitempty"`
    LongTermProvider string `json:"longTermProvider,omitempty"`
}
```

Los campos `AgentDefinition.Memory.Session` y `AgentDefinition.Memory.LongTerm`
se mantienen en el struct por backwards compatibility pero se ignoran. La UI ya no
los muestra en el formulario de agente.

**No hacer**: Configurar session/longterm memory a nivel de agente individual.
Si ADK mejora el launcher para soportar múltiples session services en el futuro,
se puede descentralizar.
