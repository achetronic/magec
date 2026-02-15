# Contexto de Sesión

## Estado Actual

Admin UI migrada completamente de vanilla JS a **Vue 3 + Vite + Tailwind v4 + Pinia**. El editor de flows usa un **canvas interactivo** con pan/zoom, toolbar de bloques draggable, y nesting infinito de contenedores. No se usa Vue Flow — es HTML/CSS puro con `vuedraggable` para el drag & drop.

**Último refactor completado**: Triggers consolidados en Client types. Los webhooks y cron jobs ya no son entidades separadas — son tipos de client (`webhook`, `cron`), al igual que `telegram` o `direct`. El sistema de configuración **tanto de client types como de memory providers** migrado de `FieldSpec` custom a **OpenAPI JSON Schema** con extensiones (`x-entity`, `x-format`, `x-placeholder`). Los flows ahora se registran por UUID (no por nombre). El cron parser soporta atajos (`@daily`, `@hourly`, etc.).

## Trabajo Completado

### 1. Migración Admin UI a Vue 3

**Stack**: Vue 3 (Composition API, `<script setup>`), Vite, Tailwind CSS v4 (`@tailwindcss/vite`), Pinia, vuedraggable.

**Estructura**:
- `admin-ui/src/main.js` — App entry, Pinia
- `admin-ui/src/App.vue` — Layout, sidebar navigation, global ConfirmDialog/Toast/SearchPalette
- `admin-ui/src/style.css` — Tailwind v4 `@theme` con colores custom (piedra/atlantico/lava/sol/arena)
- `admin-ui/src/lib/api/` — Fetch wrapper + CRUD por recurso
- `admin-ui/src/lib/stores/data.js` — Pinia store central
- `admin-ui/src/components/` — Shared: AppDialog, Card, Badge, EmptyState, FormInput, FormSelect, FormLabel, DetailRow, Icon, OverviewBadges, ConfirmDialog, Sidebar, TopBar, Toast, Tooltip, SkeletonCard, SearchPalette
- `admin-ui/src/views/` — Una carpeta por entidad con `*List.vue` + `*Dialog.vue`

**Todas las entidades migradas y funcionando**: backends, memory, mcps, agents, clients, flows, commands.

### 2. Editor de Flows — Canvas con bloques anidados

**Decisión de diseño**: El usuario quería un editor visual donde arrastrar bloques y que se vea directamente la estructura del flow. Se evaluó Vue Flow (librería de grafos) pero se descartó porque el modelo de datos es un **árbol**, no un grafo — los bloques anidados son la representación natural.

**Componentes**:
- `FlowCanvas.vue` — Canvas con pan/zoom (pointer events + wheel), fondo de puntos, sidebar izquierda con toolbar draggable (Agent, Sequential, Parallel, Loop) + controles de zoom. Auto-centra el contenido al abrir. Cuando `root` es `null`, muestra picker de tipo raíz (3 botones: Sequential/Parallel/Loop).
- `FlowBlock.vue` — Componente recursivo. Dos modos:
  - **Agent**: bloque dorado con dropdown custom para seleccionar agente (no `<select>` nativo), botón ✕ para eliminar
  - **Container** (sequential/parallel/loop): header coloreado (azul/amarillo/rojo), body con `<draggable>` que permite reordenar y mover entre contenedores. Botones ↻ (ciclar tipo) y ✕. Loop muestra badge ×N clicable y texto "repeats".
- `FlowDialog.vue` — Nombre + descripción + FlowCanvas. El `root` empieza `null` (usuario elige tipo). `cleanStep()` elimina `__key` antes de enviar al API.
- `FlowsList.vue` — Grid de cards con edit/delete.

**Interacciones**:
- Drag desde toolbar → drop en cualquier contenedor (HTML5 drag con `stopPropagation` para evitar doble-add en contenedores anidados)
- Drag entre contenedores — vuedraggable con `group: 'flow-steps'`
- Pan canvas — pointer events en fondo vacío
- Zoom — Ctrl+scroll o botones +/−, con zoom al cursor
- Dropdown de agente custom con animación, descripción, checkmark

**Dependencias eliminadas**: `@vue-flow/core`, `@vue-flow/background`, `@vue-flow/controls`, `@vue-flow/minimap`
**Dependencia añadida**: `vuedraggable@next` (wrapper de sortablejs para Vue 3)

### 3. Limpieza de Go backend

- `FlowDefinition.Layout` eliminado del struct (ya no se usa, el editor es determinista)
- `FlowDefinition.Description` sigue presente y funcional

### 4. OutputKey de ADK — Implementado

ADK `OutputKey` permite que un agente guarde su output final en el session state bajo una clave, para que otros agentes en el flow puedan referenciarlo (ej. `{generated_code}` en instrucciones).

**Diseño**: `OutputKey` vive en `AgentDefinition` (no en `FlowStep`) porque define cómo un agente publica su resultado — es una propiedad del agente, no del flow.

**Cambios**:
- `server/store/types.go` — `AgentDefinition.OutputKey string` (json: `outputKey`, omitempty)
- `server/agent/agent.go` — Pasa `OutputKey: agentDef.OutputKey` a `llmagent.Config` al crear el agente. Sin `agentConfigMap` — los flows usan agentes pre-built directamente
- `server/agent/flow.go` — Simplificado a solo `(flow, agentMap)`. Busca agentes pre-built del mapa, sin lógica de rebuild
- `server/api/admin/flows.go` — Sin validación de OutputKey (ya no existe en FlowStep)
- `admin-ui/src/views/agents/AgentDialog.vue` — Campo OutputKey con FormInput y texto explicativo
- `admin-ui/src/views/flows/FlowBlock.vue` — Eliminado campo outputKey del bloque de agente
- `admin-ui/src/views/flows/FlowDialog.vue` — `cleanStep()` ya no preserva outputKey; help text actualizado

### 5. Triggers→Clients Consolidación + JSON Schema

**Decisión de diseño**: Cron y webhook no son entidades separadas (Triggers) — son **client types**. Un client ES una identidad (token + permisos). Un cron/webhook necesita auth para llamar al agent API, así que ES un client. La entidad `Command` sobrevive como prompt reutilizable independiente.

**Nuevos client types**:

| Type | Descripción | JSON Schema config |
|------|-------------|-------------------|
| `direct` | Voice-UI tablets, apps sin config extra | `{}` (vacío) |
| `telegram` | Bot de Telegram | `botToken` (required, x-format:password), `allowedUsers`, `allowedChats`, `responseMode` (enum) |
| `cron` | Tarea programada | `schedule` (required), `commandId` (required, x-entity:commands) |
| `webhook` | Endpoint HTTP | `oneOf`: passthrough=true (sin commandId) XOR passthrough=false + commandId (x-entity:commands) |

**Cambios clave**:

- **`server/client/provider.go`** — `FieldSpec` eliminado. Nuevo: `Schema = map[string]interface{}`. Interface: `ConfigSchema() Schema`
- **`server/client/registry.go`** — `ValidateRequired()` → `ValidateConfig()` con soporte `oneOf`, `required`, `properties` recursivo
- **`server/client/{direct,cron,webhook}/`** — Nuevos providers con JSON Schema
- **`server/client/telegram/`** — Reescrito para usar `ConfigSchema()` en vez de `ConfigFields()`
- **`server/api/admin/clients.go`** — `ClientTypeInfo.Fields []FieldSpec` → `ConfigSchema client.Schema`
- **`server/store/types.go`** — `ClientConfig` con `Cron *CronClientConfig` y `Webhook *WebhookClientConfig`
- **`server/store/store.go`** — Triggers CRUD eliminado. Nuevas migraciones: `migrateTriggersToClients()`, `migrateDeviceToDirectType()`
- **`server/api/admin/handler.go`** — Rutas de triggers eliminadas
- **`server/trigger/executor.go`** — `RunTrigger()` → `RunClient()`, ejecuta contra TODOS los `allowedAgents`
- **`server/trigger/scheduler.go`** — Filtra clients tipo `cron` del store
- **`server/trigger/webhook.go`** — Auth via Bearer token del client (no secret separado)

**Frontend**:
- **`ClientDialog.vue`** — Reescrito: renderiza formularios dinámicamente desde JSON Schema. Soporta `oneOf` (visibilidad condicional), `x-entity` (select de entidades del store), `enum` (select), `boolean` (toggle), `x-format:password`
- **`ClientsList.vue`** — Badges de tipo (teal para cron/webhook, lava para direct/telegram), muestra schedule, passthrough, commandRef
- **Triggers eliminados de**: App.vue, Sidebar.vue, SearchPalette.vue, TopBar.vue, data.js, api/index.js
- **`admin-ui/src/views/triggers/`** — Directorio eliminado

**Cadena de migración** (en `loadFromDisk`):
1. `devices → clients` (legacy)
2. `cronJobs → triggers` (legacy)
3. `triggers → clients` (`migrateTriggersToClients`)
4. `device → direct` (`migrateDeviceToDirectType`)
5. `migrateIDs` (UUIDs)

**Extensiones JSON Schema custom**:
- `x-entity: "commands"` — UI renderiza select poblado desde store.commands
- `x-format: "password"` — UI usa input type password
- `x-placeholder: "..."` — Placeholder text para inputs

### 6. Webhook Endpoint + Swagger

**Webhook endpoint** documentado en swagger de userapi:
- `POST /api/v1/webhooks/{clientId}` en puerto 8080
- Auth: `Authorization: Bearer <mgc_token>` (token del client)
- Modo passthrough: body `{"prompt": "texto"}` — el prompt viene del request
- Modo fixed command: body vacío/ignorado — el prompt viene del Command referenciado
- Ejecuta contra todos los agents en `allowedAgents` del client
- Bypass `clientAuthMiddleware` — auth propia en webhook.go

**Swagger userapi** (`server/api/user/docs/`): regenerado con endpoint webhook.
**Swagger admin**: NO regenerado tras los últimos cambios (pendiente).

### 7. Admin UI Polish — Completado

9 mejoras de UX implementadas:

**Componentes nuevos**:
- `Toast.vue` — Notificaciones animadas (success/error/info) en esquina inferior derecha. Teleport to body, TransitionGroup.
- `SkeletonCard.vue` — Placeholders con `animate-pulse` durante carga inicial. Soporta grid y stacked.
- `Tooltip.vue` — Tooltip CSS-only con `group-hover` y flecha. Max 240px, multi-line.
- `SearchPalette.vue` — Cmd+K search modal con navegación por teclado (↑↓ Enter Esc), busca por nombre/descripción en las 7 entidades. Icono y color por entidad.

**Cambios en componentes existentes**:
- `EmptyState.vue` — Reescrito: ahora acepta `icon`, `color`, `actionLabel` props. Icono grande coloreado + botón CTA.
- `App.vue` — Transition `section` mode out-in, mobile backdrop, SearchPalette, Toast, provide `toast`/`registerNew`.
- `Sidebar.vue` — Prop `mobileOpen`: fixed overlay z-40 en mobile, hidden en desktop. 3 grupos: Infraestructura, Agentes, Conexiones.
- `TopBar.vue` — Botón search con hint ⌘K, hamburger menu (mobile only), emits `search`/`menu`. 3 stats: agents, backends, clients.

**Cambios en todas las vistas (7 List + 7 Dialog)**:
- List views: `inject('toast')`, `inject('registerNew')`, SkeletonCard durante loading, EmptyState con icon/color/CTA.
- Dialog views: `inject('toast')`, todos los `alert()` reemplazados por `toast.error()`.
- BackendsList, McpsList, ClientsList: Tooltip en badges de cross-referencia.

**Keyboard shortcuts** (App.vue global):
- `n` → crear nueva entidad (delegado al view activo via `registerNew`)
- `r` → refresh store
- `Cmd+K` / `Ctrl+K` → abrir búsqueda global
- Inactivos cuando el foco está en inputs/textareas/dialogs

## Lo que NO se ha tocado

- **Memory provider migration a JSON Schema**: El paquete memory sigue usando su propio sistema `FieldSpec`. Podría migrarse a JSON Schema también, pero fuera de scope
- **Admin API Swagger regeneration**: Solo userapi swagger fue regenerado. El swagger de admin en `server/api/admin/docs/` aún referencia `client.FieldSpec` viejo — necesita regeneración via `swag init --dir ./api/admin`
- **Consolidar `/types` en `/schemas/{entity}`**: Discutido pero no implementado

## Datos de Prueba

- 2 flows: "Test Sequential Flow" (seq: Magec→Itahisa), "Complex Flow" (seq: Root→parallel(Magec,Itahisa)→loop×3(Root))
- 3 agents: Magec, Itahisa, Root
- 3 backends: Ollama, Parakeet TDT, OpenAI Edge TTS
- Clients migrados de triggers anteriores (tipo cron/webhook) + direct + telegram

## Comandos

```bash
cd admin-ui && npx vite build    # Build admin UI (~1.3s)
cd admin-ui && npx vite          # Dev server con hot-reload (proxy a :8081)
make build                       # Build admin-ui + Go binary
make dev                         # Build todo + arrancar server
cd server && go build ./...      # Solo Go

# Regenerar swagger userapi
cd server && go run github.com/swaggo/swag/cmd/swag init --dir ./api/user --generalInfo doc.go --output ./api/user/docs --parseDependency --parseInternal --instanceName userapi

# Regenerar swagger admin (PENDIENTE de ejecutar)
cd server && go run github.com/swaggo/swag/cmd/swag init --dir ./api/admin --generalInfo doc.go --output ./api/admin/docs --parseDependency --parseInternal
```

## Entorno

- Go 1.25.5 (GOPATH=GOROOT warning es cosmético)
- Node 22+ con Vite 7.3, Vue 3.5, Tailwind 4.1
- ADK v0.4.0, adk-utils-go v0.2.0
- Server: :8080 (main/user API + webhooks) + :8081 (admin UI + admin API)
- Store: `data/store.json` (entities), `data/conversations.json` (conversation audit)
- Errores de Telegram son por tokens fake en datos de prueba
- Cron warning `"@daily"` — el parser solo soporta expresiones de 5 campos, no shorthand

### 8. Conversation Audit — Completado

**Feature**: Menú "Conversations" en la admin UI bajo el grupo "Auditoría" que permite ver todas las conversaciones de los agentes con soporte de markdown/código.

**Arquitectura**:
- **Captura via middleware** (`server/middleware/recorder.go`): `ConversationRecorder` y `ConversationRecorderSSE` envuelven el handler ADK y capturan todas las llamadas a `/run` y `/run_sse`. La fuente (voice-ui, telegram, cron, webhook) se identifica por el Bearer token del client.
- **Store independiente** (`server/store/conversations.go`): Persistencia en `data/conversations.json`, separado del store principal para no bloatear `data/store.json` ni disparar hot-reloads.
- **Admin API** (`server/api/admin/conversations.go`): CRUD completo + stats + summary update.
- **No en Pinia**: Las conversaciones se fetchean on-demand via API, no se cargan al init.

**Componentes UI**:
- `ConversationsList.vue` — Lista filtrable por agente/source, badges de color por source, timestamps relativos, "Clear All".
- `ConversationDetail.vue` — Vista chat con markdown rendering (`marked`), tool calls colapsables, toggle raw ADK events con copy-to-clipboard, sección de summary.
- `ConversationsView.vue` — Wrapper padre que alterna entre lista y detalle.

**Modelo de datos preparado para summarización**: Los campos `Summary` y `ParentID` en `Conversation` están listos para la futura feature de auto-summarización cuando se agota la context window del agente.

**Dependencia nueva**: `marked` (npm) para renderizar markdown a HTML.

**Color de entidad**: Teal, bajo el grupo "Auditoría" en el sidebar.
