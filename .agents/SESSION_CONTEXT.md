# Contexto de Sesión - Wake Word Server-Side

## Estado Actual
Implementación completa. Modelos movidos a `/models` y `/pretrained` en raíz del proyecto.

## Estructura de Modelos

```
magec/
├── models/
│   ├── wakewords.yaml      # Configuración de modelos wake word
│   ├── oye-magec.onnx      # Modelo "Oye Magec"
│   └── magec.onnx          # Modelo "Magec"
├── pretrained/
│   ├── mel-spectrogram.onnx
│   ├── speech-embedding.onnx
│   └── silero-vad.onnx
└── voice-ui/
    ├── models/             # Para client-side (backwards compat)
    └── pretrained/         # Para client-side (backwards compat)
```

## Configuración

```yaml
# config.yaml
wakeWord:
  enabled: true
  modelsPath: models           # Directorio de modelos wake word
  pretrainedPath: pretrained   # Directorio de modelos pretrained
  # onnxLibraryPath: /usr/lib/libonnxruntime.so
```

```yaml
# models/wakewords.yaml
models:
  - id: oye-magec
    name: Oye Magec
    file: oye-magec.onnx
    phrase: Oye Magec
    threshold: 0.5
  - id: magec
    name: Magec
    file: magec.onnx
    phrase: Magec
    threshold: 0.3
```

## Archivos Modificados

### Backend
- `server/config/config.go` - Nueva estructura `WakeWordConfig`, función `LoadWakeWordModels()`
- `server/main.go` - Carga modelos desde `wakewords.yaml`
- `server/wakeword/` - Sin cambios

### Frontend
- `voice-ui/src/app.js` - Solo server-side, sin fallback
- `voice-ui/src/audio/ServerWakeWordDetector.js` - Recibe modelos del servidor
- `voice-ui/src/ui/UIController.js` - `disableWakeWordToggle()`, `hideWakeWordModelSelector()`

### Configuración
- `models/wakewords.yaml` (NUEVO) - Configuración de modelos
- `config.example.yaml` - Añadida sección `wakeWord`
- `.gitignore` - Actualizado para nuevas rutas
- `.dockerignore` - Actualizado para nuevas rutas
- `Dockerfile` - Copia modelos a `/app/models` y `/app/pretrained`
- `scripts/download-model.go` - Descarga a `pretrained/` (raíz)

## Siguiente Paso

1. Descargar ONNX Runtime
2. Configurar `wakeWord.enabled: true` y `onnxLibraryPath`
3. Probar
