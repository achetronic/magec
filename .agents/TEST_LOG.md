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

## Resumen

| Test | Prompt | Status | Tiempo | Agentes |
|------|--------|--------|--------|---------|
| Flow webhook | "Dime algo sobre energía renovable en Canarias" | 200 | 16.7s | Root→Parallel(Itahisa,Alby) |
| Passthrough webhook | "Di hola en una frase corta" | 200 | ~3s | Itahisa (Magec falló MCP) |
| Fixed-command webhook | (del Command: "Tell me a receipt...") | 200 | ~8s | Itahisa + Alby |
| Cron @daily | N/A (schedule parse test) | OK | N/A | Cargado sin warnings |

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
