# Contexto de Sesión

## Estado Actual

Implementando **Flows** (composiciones multi-agente) y **migración completa a UUID v4 inmutables**. El servidor compila y corre. La Admin UI tiene la tab "Flows" funcional con editor tree-view recursivo. Se acaba de corregir un syntax error en app.js (brace extra) — **el usuario necesita recargar la página para verificar que funciona**.

## Trabajo Completado en Esta Sesión

### 1. Migración completa a UUID v4 (google/uuid)

**Problema**: Los IDs eran hex strings de 32 chars sin guiones. El usuario pidió UUID v4 estándar con guiones.

**Cambios**:
- `server/store/types.go`: `generateID()` ahora usa `github.com/google/uuid` → `uuid.New().String()` (formato `550e8400-e29b-41d4-a716-446655440000`)
- `server/store/store.go`: `isHexID()` renombrado a `isUUID()` con regex `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
- Dependencia añadida: `github.com/google/uuid`

### 2. Migración de cross-references (name→ID)

**Problema**: El back-fill original solo generaba IDs para entidades sin ID, pero NO migraba las cross-references que seguían usando nombres (ej: `agent.llm.backend = "Ollama"` en vez del UUID del backend).

**Solución**: Función `migrateIDs()` en `store.go` con 2 fases:
- **Fase 1**: Genera UUID para toda entidad sin ID válido (incluye agents con IDs legacy como `"magec"`)
- **Fase 2**: Construye mapas `name→ID` y reescribe todas las cross-references:
  - `agent.llm.backend`, `agent.transcription.backend`, `agent.tts.backend` → backend ID
  - `agent.memory.session`, `agent.memory.longTerm` → memory provider ID
  - `agent.mcpServers[]` → MCP server IDs
  - `memoryProvider.embedding.backend` → backend ID
  - `cronJob.agentId` → agent ID
  - `client.allowedAgents[]` → agent IDs
- Es **idempotente**: usa `isUUID()` para detectar si un campo ya fue migrado

**Todas las funciones `Rename*WithCascade` eliminadas** (6 funciones): con IDs inmutables, renombrar es solo un PUT con nuevo `name`.

### 3. Flows — Composiciones Multi-Agente

**Modelo de datos** (`server/store/types.go`):
- `FlowDefinition`: `ID`, `Name`, `Root FlowStep`
- `FlowStep` (recursivo): `Type` (agent/sequential/parallel/loop), `AgentID`, `MaxIterations`, `Steps []FlowStep`
- Constantes: `FlowStepAgent`, `FlowStepSequential`, `FlowStepParallel`, `FlowStepLoop`
- `StoreData.Flows []FlowDefinition` añadido

**Store CRUD** (`server/store/store.go`):
- `ListFlows`, `GetFlow`, `CreateFlow`, `UpdateFlow`, `DeleteFlow`
- Nil-slice init en `loadFromDisk()` y `New()`
- Flow ID migration en `migrateIDs()`

**Admin API** (`server/admin/flows.go`, `server/admin/handler.go`):
- `GET/POST /flows`, `GET/PUT/DELETE /flows/{id}`
- `validateFlowStep()` — validación recursiva del árbol (agent necesita agentId, containers necesitan steps, etc.)

**Motor de ejecución** (`server/agent/flow.go`):
- `BuildFlowAgent(flow, agentMap)` — traduce `FlowDefinition` a árbol ADK
- Recursivo: agent→`llmagent`, sequential→`sequentialagent.New()`, parallel→`parallelagent.New()`, loop→`loopagent.New()`
- Imports: `google.golang.org/adk/agent/workflowagents/{sequentialagent,parallelagent,loopagent}`

**Integración** (`server/agent/agent.go`, `server/main.go`):
- `agent.New()` acepta `flows []store.FlowDefinition` como nuevo parámetro
- Construye map `storeAgentID → adkAgent` durante iteración
- Después del loop de agents, itera flows y llama `BuildFlowAgent()`
- Flow agents se registran como `otherAgents` en el `MultiLoader`
- Hot-reload funciona: al crear/editar flow via API, se reconstruye todo

**Admin UI**:
- `admin-ui/src/api.js`: `listFlows`, `getFlow`, `createFlow`, `updateFlow`, `deleteFlow`
- `admin-ui/index.html`: Tab "Flows", panel `panelFlows`, dialog `flowDialog` con editor tree-view
- `admin-ui/src/app.js`: 
  - `this.flows = []` en constructor
  - `_renderFlows()` — lista de flows con resumen visual del árbol
  - `_flowStepSummary()` — genera representación inline del árbol
  - `showFlowDialog()` / `editFlow()` / `saveFlow()` — CRUD
  - `_renderStepEditor()` — editor recursivo: seleccionar tipo, elegir agent, añadir/eliminar steps
  - `_changeStepType()`, `_setStepAgent()`, `_setStepMaxIter()`, `_addChildStep()`, `_removeStep()`, `_getStepAtPath()`
  - Delete confirmation actualizado para incluir `flow`

## Estado Verificado

- Server compila limpio (`go build ./...`)
- Server arranca y hot-reload funciona
- Flows API probada: 2 flows de prueba creados (sequential simple + complex con parallel+loop anidados)
- Logs muestran `Flow initialized id=xxx name="Test Sequential Flow"` y `Flow initialized id=xxx name="Complex Flow"`
- **Se corrigió un `}` extra en app.js línea 857** que causaba syntax error — pendiente de verificar por el usuario recargando

## Datos de Prueba en store.json

- 2 flows de prueba creados:
  - "Test Sequential Flow": sequential con Magec → Itahisa
  - "Complex Flow": sequential con Magec → parallel(Magec, Itahisa) → loop(Itahisa, x3)

## Pendiente (TODOs)

1. **Consolidar endpoints `*/types` en `/schemas/{entity}`** — discutido con el usuario, pendiente de implementar
2. **Mejorar UI de Flows** — el editor tree-view funciona pero es básico. Futuro: drag-and-drop con canvas visual
3. **El input `agentId` en el agent dialog** del HTML es vestigial (ID es auto-generated) — debería ocultarse

## Comandos

```bash
cd server && go build ./...   # Compilar
make dev                      # Compilar + ejecutar (puertos 8080 + 8081)
go mod tidy                   # Limpiar dependencias
```

## Entorno

- Go 1.25.5 (GOPATH=GOROOT warning es cosmético)
- ADK v0.3.0, adk-utils-go v0.1.7
- google/uuid para generación de IDs
- Server corre en :8080 (API principal) y :8081 (admin UI + admin API)
- Store: `data/store.json` (auto-migra al arrancar)
- Backup del store original: `data/store.json.bak`
