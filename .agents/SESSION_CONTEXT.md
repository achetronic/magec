# Contexto de Sesión

## Estado Actual

Admin UI migrada completamente de vanilla JS a **Vue 3 + Vite + Tailwind v4 + Pinia**. El editor de flows usa un **canvas interactivo** con pan/zoom, toolbar de bloques draggable, y nesting infinito de contenedores. No se usa Vue Flow — es HTML/CSS puro con `vuedraggable` para el drag & drop.

## Trabajo Completado

### 1. Migración Admin UI a Vue 3

**Stack**: Vue 3 (Composition API, `<script setup>`), Vite, Tailwind CSS v4 (`@tailwindcss/vite`), Pinia, vuedraggable.

**Estructura**:
- `admin-ui/src/main.js` — App entry, Pinia
- `admin-ui/src/App.vue` — Layout, tabs via `location.hash`, global ConfirmDialog
- `admin-ui/src/style.css` — Tailwind v4 `@theme` con colores custom (piedra/atlantico/lava/sol/arena)
- `admin-ui/src/lib/api/` — Fetch wrapper + CRUD por recurso
- `admin-ui/src/lib/stores/data.js` — Pinia store central
- `admin-ui/src/components/` — Shared: AppDialog, Card, Badge, EmptyState, FormInput, FormSelect, FormLabel, DetailRow, Icon, OverviewBadges, ConfirmDialog
- `admin-ui/src/views/` — Una carpeta por entidad con `*List.vue` + `*Dialog.vue`

**Todas las entidades migradas y funcionando**: backends, memory, mcps, agents, clients, crons, flows.

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
- `server/admin/flows.go` — Sin validación de OutputKey (ya no existe en FlowStep)
- `admin-ui/src/views/agents/AgentDialog.vue` — Campo OutputKey con FormInput y texto explicativo
- `admin-ui/src/views/flows/FlowBlock.vue` — Eliminado campo outputKey del bloque de agente
- `admin-ui/src/views/flows/FlowDialog.vue` — `cleanStep()` ya no preserva outputKey; help text actualizado

### 5. Commands + Triggers — Implementado (CRUD + UI, scheduler/webhook pendientes)

Separación de prompts (Commands) y automatización (Triggers):

**Command**: Prompt reutilizable con nombre, descripción, prompt, y agente por defecto (opcional).

**Trigger**: Dos tipos:
- **Cron**: schedule + command + agent + client (token auth). Scheduler aún no implementado.
- **Webhook**: endpoint único por trigger. Modo fijo (command + agent) o passthrough (prompt viene del request body). Secret auto-generado para auth entrante.

**Migración**: Los CronJob legacy se convierten automáticamente a Command + Trigger tipo cron en `migrateCronsToTriggers()`.

**Cambios**:
- `server/store/types.go` — `Command`, `Trigger`, `CronConfig`, `WebhookConfig`. CronJob mantenido como legacy
- `server/store/store.go` — CRUD para Commands y Triggers. Migración automática CronJob→Command+Trigger
- `server/admin/commands.go` — CRUD handlers
- `server/admin/triggers.go` — CRUD handlers con validación por tipo
- `admin-ui/src/views/commands/` — CommandsList + CommandDialog (color: indigo)
- `admin-ui/src/views/triggers/` — TriggersList + TriggerDialog (color: teal, tipo toggle chips)
- `admin-ui/src/App.vue` — Tabs actualizados: Crons reemplazado por Commands + Triggers
- `.agents/ENTITY_COLORS.md` — Commands = indigo, Triggers = teal (antes "Crons")

### 6. Admin UI Polish — Completado

9 mejoras de UX implementadas:

**Componentes nuevos**:
- `Toast.vue` — Notificaciones animadas (success/error/info) en esquina inferior derecha. Teleport to body, TransitionGroup.
- `SkeletonCard.vue` — Placeholders con `animate-pulse` durante carga inicial. Soporta grid y stacked.
- `Tooltip.vue` — Tooltip CSS-only con `group-hover` y flecha. Max 240px, multi-line.
- `SearchPalette.vue` — Cmd+K search modal con navegación por teclado (↑↓ Enter Esc), busca por nombre/descripción en las 8 entidades. Icono y color por entidad.

**Cambios en componentes existentes**:
- `EmptyState.vue` — Reescrito: ahora acepta `icon`, `color`, `actionLabel` props. Icono grande coloreado + botón CTA.
- `App.vue` — Transition `section` mode out-in, mobile backdrop, SearchPalette, Toast, provide `toast`/`registerNew`.
- `Sidebar.vue` — Prop `mobileOpen`: fixed overlay z-40 en mobile, hidden en desktop.
- `TopBar.vue` — Botón search con hint ⌘K, hamburger menu (mobile only), emits `search`/`menu`.

**Cambios en todas las vistas (9 List + 9 Dialog)**:
- List views: `inject('toast')`, `inject('registerNew')`, SkeletonCard durante loading, EmptyState con icon/color/CTA.
- Dialog views: `inject('toast')`, todos los `alert()` reemplazados por `toast.error()`.
- BackendsList, McpsList, TriggersList, ClientsList: Tooltip en badges de cross-referencia.
- TriggersList: Status dot (green/gray) en icono de trigger card.

**Keyboard shortcuts** (App.vue global):
- `n` → crear nueva entidad (delegado al view activo via `registerNew`)
- `r` → refresh store
- `Cmd+K` / `Ctrl+K` → abrir búsqueda global
- Inactivos cuando el foco está en inputs/textareas/dialogs

## Lo que NO se ha tocado

- **Swagger docs**: No regeneradas tras los cambios
- **Consolidar `/types` en `/schemas/{entity}`**: Discutido pero no implementado
- **Cron scheduler**: Triggers tipo cron son solo datos — no hay motor de ejecución
- **Webhook handler**: Triggers tipo webhook no tienen endpoint HTTP aún

## Datos de Prueba

- 2 flows: "Test Sequential Flow" (seq: Magec→Itahisa), "Complex Flow" (seq: Root→parallel(Magec,Itahisa)→loop×3(Root))
- 3 agents: Magec, Itahisa, Root
- 3 backends: Ollama, Parakeet TDT, OpenAI Edge TTS

## Comandos

```bash
cd admin-ui && npx vite build    # Build admin UI (~1.2s)
cd admin-ui && npx vite          # Dev server con hot-reload (proxy a :8081)
make build                       # Build admin-ui + Go binary
make dev                         # Build todo + arrancar server
cd server && go build ./...      # Solo Go
```

## Entorno

- Go 1.25.5 (GOPATH=GOROOT warning es cosmético)
- Node 22+ con Vite 7.3, Vue 3.5, Tailwind 4.1
- ADK v0.4.0, adk-utils-go v0.2.0
- Server: :8080 (main) + :8081 (admin UI + admin API)
- Store: `data/store.json`
- Errores de Telegram son por tokens fake en datos de prueba
