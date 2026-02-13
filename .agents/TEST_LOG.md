# Test Log — 2026-02-13

Pruebas de integración contra el servidor real con credenciales reales.

---

## 1. Webhook → Flow (passthrough)

**Client**: `External Prompt (Flow)` (b8fa980c)
**Target**: Flow `Flujito Agéntico` (977e9ae5) → Sequential(Root → Parallel(Itahisa, Alby))

```bash
POST http://localhost:8080/api/v1/webhooks/b8fa980c-e82e-4e6c-8b83-3bc83f3e879a
Authorization: Bearer mgc_d4b544250cc2dc37f236c59825ed5d2ee723cbba
Content-Type: application/json

{"prompt": "Dime algo interesante sobre la energía renovable en Canarias. Responde brevemente."}
```

**Status**: 200 OK (16.7s)

**Respuesta**:
```
Canarias es un laboratorio natural fascinante para las energías renovables. Aquí tienes un dato
destacado:

**El Hierro** se convirtió en 2014 en la primera isla del mundo con objetivo de autoabastecerse
al 100% con renovables, gracias a la central hidroeólica de Gorona del Viento. Combina
aerogeneradores con un sistema de bombeo hidráulico que almacena energía excedente bombeando agua
a un depósito elevado, que luego se turbina cuando no hay viento suficiente.

En días óptimos, la isla ha funcionado semanas enteras sin quemar una gota de combustible fósil —
un logro notable para un sistema eléctrico aislado donde no puedes importar electricidad del
continente cuando te falta.

---

Canarias es tan buen sitio para renovables que casi parece que alguien pensó antes de construir
hoteles a lo loco.

Un dato concreto: **El Hierro** tiene la central hidroeólica **Gorona del Viento**, que combina
eólica con un sistema hidráulico de bombeo. Cuando sobra viento, bombean agua a un embalse alto;
cuando falta viento, turbina ese agua para generar electricidad. Resultado: ha llegado a operar
**semanas enteras al 100% con renovables**, en un sistema eléctrico aislado, sin tirar de
península ni leches. Para una isla pequeña, es casi ciencia ficción bien hecha.

---

Mira, lo de **El Hierro** es un caso de estudio brutal para cualquier ingeniero de sistemas. La
central de Gorona del Viento es básicamente un problema de control resuelto con elegancia:

- **5 aerogeneradores** generan cuando hay viento
- **El excedente** bombea agua a un depósito a 700m de altura (energía potencial almacenada)
- **Cuando no hay viento**, el agua baja por turbinas y genera electricidad

Es almacenamiento de energía sin baterías de litio, usando la gravedad. Simple, robusto, y con
una vida útil de décadas.

El dato curioso: en 2019 la isla estuvo **25 días consecutivos** funcionando 100% renovable. Para
un sistema aislado sin conexión a ninguna red continental, eso es un logro técnico serio.

Canarias en general tiene un potencial absurdo — sol, viento, y ahora se está explorando energía
de las olas. El problema sigue siendo el mismo de siempre: almacenamiento y gestión de la
intermitencia.
```

**Análisis**: Se distinguen 3 bloques separados por `---`:
1. **Root** (Claude Opus) — Párrafo coordinador, tono neutro y enciclopédico.
2. **Itahisa** (GPT-5.1) — Tono canario informal ("sin tirar de península ni leches").
3. **Alby** (Claude Opus) — Enfoque pragmático de ingeniería, lista de viñetas, datos concretos.

---

## 2. Webhook → Agents passthrough (Magec + Itahisa)

**Client**: `External Prompt` (7b7d8126)
**Target**: Magec (5b95e520) + Itahisa (60f40ce6)

```bash
POST http://localhost:8080/api/v1/webhooks/7b7d8126-72c3-4e8e-b894-1a920101d22c
Authorization: Bearer mgc_26af5acc12c41a327e14949805618eb0ce8060d7
Content-Type: application/json

{"prompt": "Di hola en una frase corta."}
```

**Status**: 200 OK

**Respuesta**:
```
Hola.
```

**Nota**: Solo respondió Itahisa. Magec falla por error de TLS en su MCP server
(`hass-mcp.tools.internal.place` cert mismatch) — issue de infraestructura, no de código.

---

## 3. Webhook → Fixed command (Itahisa + Alby)

**Client**: `Simple Receipt` (7c2c9b65)
**Target**: Itahisa (60f40ce6) + Alby (76649961)
**Command**: "Tell me a receipt with 4 or less common ingredients"

```bash
POST http://localhost:8080/api/v1/webhooks/7c2c9b65-6c7c-4437-a50b-856d4f2307c4
Authorization: Bearer mgc_ef1b9dd605b6cf0783fd41a2820a6ab635f91057
Content-Type: application/json

{}
```

**Status**: 200 OK

**Respuesta**:
```
Tortilla francesa. Poor man's protein packet.

**Ingredients (4 common things):**
1. Eggs
2. Salt
3. Olive oil (or whatever oil you have)
4. Optional: black pepper or some herb (parsley, oregano, whatever's dying in your fridge)

**Steps:**
1. Crack 2–3 eggs in a bowl. Add a pinch of salt (and pepper/herbs if using).
2. Beat with a fork until the mixture looks uniform, no big snot-like streaks.
3. Heat a bit of oil in a pan over medium heat. If it's smoking, you've gone full arsonist: lower it.
...
```

**Nota**: El body es `{}` porque es fixed-command (passthrough=false). El prompt viene del
Command referenciado, no del body del request.

---

## 4. Cron — `@daily` shorthand

**Client**: `Daily News` (fd98998b)
**Schedule**: `@daily` (expande a `0 0 * * *`)
**Command**: "Sumarize today's news"
**Target**: Root (bf3ad28d)

**Antes del fix**:
```
level=WARN msg="Invalid schedule" client="Daily News" schedule="@daily" error="expected 5 fields, got 1"
```

**Después del fix**:
```
level=DEBUG msg="Scheduler reloaded" cronClients=1
```

No WARN. El cron se cargó correctamente. Próxima ejecución: 00:00 UTC del día siguiente.

---

## 5. Flow webhook — `responseAgent` filter (Root only)

**Client**: `External Prompt (Flow)` (b8fa980c)
**Target**: Flow `Flujito Agéntico` (977e9ae5) — Sequential(Parallel(Itahisa, Alby) → Root[responseAgent=true])

```bash
POST http://localhost:8080/api/v1/webhooks/b8fa980c-e82e-4e6c-8b83-3bc83f3e879a
Authorization: Bearer mgc_d4b544250cc2dc37f236c59825ed5d2ee723cbba
Content-Type: application/json

{"prompt": "Qué diferencia hay entre Zigbee y Z-Wave para domótica?"}
```

**Status**: 200 OK (17.0s)

**Respuesta** (812 chars, 1 bloque):
```
**Zigbee vs Z-Wave** — resumen práctico:

| | **Zigbee** | **Z-Wave** |
|---|---|---|
| **Frecuencia** | 2,4 GHz (como WiFi) → más interferencias posibles | ~868 MHz (EU) / ~908 MHz (US) → mejor penetración de paredes |
| **Red** | Mesh, hasta ~65.000 dispositivos | Mesh, máximo ~232 dispositivos |
| **Ecosistema** | Abierto, más barato, más variedad, pero compatibilidad variable | Propietario con certificación obligatoria → "simplemente funciona" entre marcas |
| **Consumo** | Muy bajo | Muy bajo |

**Conclusión práctica:**
- **Zigbee** → si buscas variedad y precio
- **Z-Wave** → si prefieres menos interferencias y compatibilidad garantizada

Si empiezas de cero, vale la pena mirar dispositivos compatibles con **Matter**, que busca unificar ambos mundos.
```

**Análisis**: Una sola respuesta sintetizada por Root. No se ven las respuestas individuales de Itahisa ni Alby.
El `responseAgent: true` en el step de Root filtra el event stream por `event.author == "bf3ad28d..."`.

---

## 6. Flow webhook — sin `responseAgent` (control, backwards-compat)

**Client**: Mismo (b8fa980c)
**Target**: Mismo flujo pero con `responseAgent` removido de Root

**Status**: 200 OK (13.1s)

**Respuesta** (1121 chars, 3 bloques separados por `---`):
```
Block 1 (185 chars): Itahisa — tono cortante, "bajo bitrate", "a lo bruto"
Block 2 (638 chars): Alby — detalle industrial, IEEE 802.15.4, tabla comparativa
Block 3 (254 chars): Root — síntesis pero indistinguible de los otros dos
```

**Análisis**: Sin `responseAgent`, todas las respuestas de todos los agentes se concatenan (comportamiento por defecto).
3 bloques = 3 agentes respondiendo. Confirma backwards compatibility.

---

## 7. Regular agent webhook (non-flow, control)

**Client**: `External Prompt` (7b7d8126)
**Target**: Magec (5b95e520) + Itahisa (60f40ce6) — no es un flujo

```bash
POST http://localhost:8080/api/v1/webhooks/7b7d8126-72c3-4e8e-b894-1a920101d22c
Authorization: Bearer mgc_26af5acc12c41a327e14949805618eb0ce8060d7
Content-Type: application/json

{"prompt": "Di hola en una frase muy corta."}
```

**Status**: 200 OK (2.9s)

**Respuesta**:
```
Hola.
```

**Nota**: Solo respondió Itahisa. Magec falla por MCP TLS (issue conocida). El path `responseFilter=nil` para agentes regulares (no flujos) funciona correctamente.

---

## Resumen

| # | Test | Prompt | Status | Tiempo | Agentes | Notas |
|---|------|--------|--------|--------|---------|-------|
| 1 | Flow webhook | "Dime algo sobre energía renovable en Canarias" | 200 | 16.7s | Root→Parallel(Itahisa,Alby) | 3 bloques (pre-responseAgent) |
| 2 | Passthrough webhook | "Di hola en una frase corta" | 200 | ~3s | Itahisa (Magec falló MCP) | |
| 3 | Fixed-command webhook | (del Command: "Tell me a receipt...") | 200 | ~8s | Itahisa + Alby | |
| 4 | Cron @daily | N/A (schedule parse test) | OK | N/A | Cargado sin warnings | |
| 5 | Flow + responseAgent | "Zigbee vs Z-Wave" | 200 | 17.0s | Root only (filtered) | 1 bloque, sintetizado |
| 6 | Flow sin responseAgent | "Qué es Zigbee" | 200 | 13.1s | Itahisa+Alby+Root | 3 bloques (backwards compat) |
| 7 | Regular agent webhook | "Di hola" | 200 | 2.9s | Itahisa | responseFilter=nil path |

## Equivalente con curl (referencia)

Estos tests se hicieron con `urllib` de Python porque `curl` no está disponible en el entorno.
El equivalente curl sería:

```bash
# Webhook passthrough (flow)
curl -X POST http://localhost:8080/api/v1/webhooks/b8fa980c-e82e-4e6c-8b83-3bc83f3e879a \
  -H "Authorization: Bearer mgc_d4b544250cc2dc37f236c59825ed5d2ee723cbba" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Dime algo interesante sobre la energía renovable en Canarias."}'

# Webhook fixed command
curl -X POST http://localhost:8080/api/v1/webhooks/7c2c9b65-6c7c-4437-a50b-856d4f2307c4 \
  -H "Authorization: Bearer mgc_ef1b9dd605b6cf0783fd41a2820a6ab635f91057" \
  -H "Content-Type: application/json" \
  -d '{}'
```
