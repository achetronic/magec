# Flow state & loop exit — Test plan

End-to-end checks for the `feature/flow-state` work (issue #36). Run these against a Magec instance with at least one LLM backend already configured.

> **Note on small models**: tool calls are sometimes flaky on `qwen3:8b` and similar small models. If a step misbehaves, switch the agents to a stronger backend or set `temperature: 0`.

---

## Setup

- Magec running (Docker compose, binary, whatever).
- At least one LLM backend in Admin UI → Backends.
- A direct client with a Bearer token (Admin UI → Clients) that you can put on `allowedAgents` per test. Easiest: create one client now and just toggle the agents/flows it can reach.

---

## Prueba 1 — Early exit por agente (`exit_loop`)

### Agentes a crear

Admin UI → tab **Agents** → crea estos dos. Mismo backend para los dos.

#### Agente `Generator`

| Campo         | Valor                            |
| ------------- | -------------------------------- |
| Name          | `Generator`                      |
| Description   | Generates a random number 1-10   |
| LLM Backend   | el que tengas                    |
| Output key    | _vacío_                          |

System prompt:

```
You are a number generator inside a workflow loop.

On every turn:
1. Pick a number between 1 and 10 (try to vary it across iterations).
2. Call the tool `set_state` with key="value" and value=<your number>.
3. Output ONLY the number, nothing else.

Do not call exit_loop. Do not call get_state. Just pick the number, write it to state, and reply.
```

#### Agente `Critic`

| Campo         | Valor                                                  |
| ------------- | ------------------------------------------------------ |
| Name          | `Critic`                                               |
| Description   | Decides whether the generated number is high enough    |
| LLM Backend   | el mismo                                               |
| Output key    | _vacío_                                                |

System prompt:

```
You are a critic inside a workflow loop. Your job: decide whether the number the Generator just produced is good enough.

On every turn:
1. Call `get_state` with key="value" to read the most recent number.
2. If the value is greater than or equal to 7, call `exit_loop` and reply "Approved: <value>".
3. Otherwise reply "Rejected: <value>, need higher" and do NOT call exit_loop.

Use the tools — do not guess. Always call get_state first.
```

### Flow `loop-exit-test`

Admin UI → tab **Flows** → New Flow.

Estructura:

```
loop-exit-test
└── Loop
    └── Sequential
        ├── Generator
        └── Critic   ← marca el response toggle (icono chat verde) en este
```

Cómo construirlo en el canvas:

1. Arrastra **Loop** dentro del root.
2. Arrastra **Sequential** dentro del Loop.
3. Arrastra dos **Agent** dentro del Sequential. Asigna `Generator` al primero, `Critic` al segundo.
4. En `Critic`, click en el icono de chat verde para marcarlo response agent (así sólo ves su salida, no la del Generator).
5. Click en el botón **Mode** del Loop:
   - Max iterations: `8`
   - Early exit: **on**
   - Strategy: **Agent decides**
   - Save.

### Cómo invocarlo

#### Opción A — Voice UI (más rápido)

1. Abre `http://localhost:8080` (Voice UI).
2. En el switcher de agente, selecciona `loop-exit-test`.
3. Escribe `go` en el input de texto.

#### Opción B — cURL directo

Sustituye `<flowID>` por el ID que verás cuando guardes el flow (Admin UI → flow → JSON), y `<client-token>` por el token de un cliente direct que tenga ese flow en `allowedAgents`.

```bash
# Crea sesión
curl -X POST http://localhost:8080/api/v1/agent/apps/<flowID>/users/test/sessions/sess1 \
  -H "Authorization: Bearer <client-token>" \
  -H "Content-Type: application/json" \
  -d '{}'

# Ejecuta
curl -N -X POST http://localhost:8080/api/v1/agent/run_sse \
  -H "Authorization: Bearer <client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "appName": "<flowID>",
    "userId": "test",
    "sessionId": "sess1",
    "newMessage": {"role": "user", "parts": [{"text": "go"}]}
  }'
```

### Qué observar

En los SSE events / Voice UI conversation:

- **Iteración 1**: `Generator` → `set_state(value, 4)` → `4`. `Critic` → `get_state(value)` → "Rejected: 4, need higher".
- **Iteración 2**: similar con otro número.
- **Iteración N** (cuando salga ≥7): `Generator` → `set_state(value, 8)` → `8`. `Critic` → `get_state(value)` → `exit_loop()` → "Approved: 8".
- **Loop corta inmediatamente**, no hay iteración N+1.

En Admin UI → tab **Conversations** abre la conversación → modo "admin" para ver todos los tool calls. Verifica:

- [ ] `set_state` se llama N veces.
- [ ] `get_state` se llama N veces.
- [ ] `exit_loop` se llama 1 vez (la última iter).
- [ ] Hay exactamente N iteraciones, no más.

### Comprobaciones extra

**State en Redis** (si Redis está configurado):

```bash
docker exec -it magec-redis redis-cli
> KEYS *sess1*
> HGETALL <session-key>
```

Verás `flow:value` con el último número.

**Cap funciona**: edita el flow, baja max iterations a `2`, modifica el system prompt del Critic para que NUNCA llame exit_loop. Invoca otra vez (sesión nueva). El loop debe parar a las 2 iteraciones aunque no haya exit_loop.

---

## Prueba 2 — Early exit por expresión CEL

### Agentes adicionales

Reusa `Generator` (sirve igual). Crea uno nuevo:

#### Agente `Validator`

| Campo         | Valor                                                          |
| ------------- | -------------------------------------------------------------- |
| Name          | `Validator`                                                    |
| Description   | Marks the generated value as valid or invalid based on parity  |
| LLM Backend   | el mismo                                                       |
| Output key    | _vacío_                                                        |

System prompt:

```
You are a validator inside a workflow loop.

On every turn:
1. Call `get_state` with key="value" to read the most recent number.
2. If the value is even, call `set_state` with key="valid" and value=true. Reply "Even: <value> — valid".
3. If the value is odd, call `set_state` with key="valid" and value=false. Reply "Odd: <value> — invalid".

Do NOT call exit_loop (you don't have it). Just write to state and reply.
```

### Flow `loop-cel-test`

```
loop-cel-test
└── Loop
    └── Sequential
        ├── Generator
        └── Validator   ← response agent
```

Loop config:

- Max iterations: `10`
- Early exit: **on**
- Strategy: **Expression**
- CEL expression: `state.valid == true`
- Save.

### Cómo invocarlo

Igual que en la prueba 1, cambiando el flow ID.

### Qué observar

- **Iteración 1**: `Generator` saca un impar (digamos 5). `set_state(value, 5)`. `Validator` lee 5 → impar → `set_state(valid, false)`. Reply "Odd: 5 — invalid".
- El evaluator CEL corre al final de la iteración: `state.valid == true` → false. Loop continúa.
- **Iteración 2**: `Generator` saca 8. `set_state(value, 8)`. `Validator` → `set_state(valid, true)`. Reply "Even: 8 — valid".
- Evaluator CEL: `state.valid == true` → **true**. Emite `Escalate`. Loop corta.

En la conversación verás eventos del agente sintético `<flowID>_0_exitwhen` (por la convención de naming que pusimos). Si es así, sabes que el evaluator se inyectó como último child del loop.

Checklist:

- [ ] `set_state(valid, …)` se llama una vez por iteración.
- [ ] El loop continúa mientras `valid == false`.
- [ ] Aparece un evento del agente sintético `*_exitwhen` cuando se cumple la expresión.
- [ ] El loop corta en cuanto `valid == true`.

### Comprobaciones extra

**Validación admin API** (que rechaza configs inválidas):

```bash
# Debería dar 400 — exclusión
curl -X POST http://localhost:8081/api/v1/admin/flows \
  -H "Authorization: Bearer <admin-password>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "bad",
    "root": {
      "type": "loop",
      "maxIterations": 5,
      "exitLoop": true,
      "exitWhen": "state.x == true",
      "steps": [{"type": "agent", "agentId": "<some-id>"}]
    }
  }'
```

Espera: `400 loop step: exitLoop and exitWhen are mutually exclusive`.

```bash
# Debería dar 400 — CEL inválido
curl -X POST http://localhost:8081/api/v1/admin/flows \
  -H "Authorization: Bearer <admin-password>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "bad-cel",
    "root": {
      "type": "loop",
      "maxIterations": 5,
      "exitWhen": "state.x ===",
      "steps": [{"type": "agent", "agentId": "<some-id>"}]
    }
  }'
```

Espera: `400 loop step: invalid CEL expression: ...`.

---

## Prueba 3 — Standalone agent NO tiene las tools

Invoca al agente `Generator` directamente (no a través de un flow). Su system prompt le dice que llame `set_state` — debería **fallar** o el modelo debería responder "no tengo esa tool".

1. Admin UI → Clients → crea o edita un cliente direct con `Generator` (el agente, no el flow) en `allowedAgents`.
2. Llámalo desde Voice UI o cURL como `appName: "Generator"` (su ID).
3. Verifica que `set_state` y `get_state` NO aparecen en el catálogo de tools (lo verás en el evento inicial de la conversación, o el modelo dirá que no las tiene).

Esto verifica el scoping (agentes standalone no reciben las tools de flow).

---

## Resumen de qué demuestra cada prueba

| Prueba                    | Demuestra                                                                                  |
| ------------------------- | ------------------------------------------------------------------------------------------ |
| **1**                     | `set_state`/`get_state` funcionan, `exit_loop` corta el loop, `MaxIterations` actúa de cap |
| **2**                     | El evaluator CEL se inyecta, evalúa state, emite Escalate cuando true, cap también actúa   |
| **3**                     | Las tools NO contaminan agentes standalone — scoping correcto                              |
| Validación admin (cURL)   | Backend rechaza configs inválidas en save time                                             |

---

## Si algo va mal

- **El modelo no llama las tools**: prueba con un backend más potente, baja la temperatura a 0, o haz el system prompt más explícito (ej. "you MUST call set_state").
- **El loop no corta con `exit_loop`**: revisa que `Mode → Early exit → Agent decides` esté guardado. En la conversación admin debería aparecer `exit_loop` como tool call.
- **El loop no corta con CEL**: comprueba que el `Validator` está escribiendo realmente. Mira en Redis o pon temporalmente `state.valid` en `outputKey` para que aparezca en los eventos.
- **400 al crear el flow**: revisa que sólo uno de `exitLoop` / `exitWhen` esté activo, y que el CEL devuelva `bool` (no `int`, no `string`).
- **Conversación cortada por contexto**: ContextGuard puede activarse si configuraste tope de tokens. Desactívalo en el agente para esta prueba.

---

## Cleanup

Después de probar:

- Borra los flows `loop-exit-test` y `loop-cel-test`.
- Borra los agentes `Generator`, `Critic`, `Validator` si no los vas a reusar.
- Borra el cliente direct si lo creaste sólo para esto.
