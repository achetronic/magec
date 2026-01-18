// Traducciones en español (idioma predeterminado)
export default {
  app: {
    title: "Magec",
    subtitle: "Asistente de Voz",
  },

  status: {
    initializing: "Iniciando...",
    ready: "Listo",
    listening: "Escuchando...",
    recording: "Grabando...",
    processing: "Procesando...",
    thinking: "Pensando...",
    speaking: "Hablando...",
    loadingWakeWord: "Cargando wake word...",
    loadingWhisper: "Cargando Whisper local...",
    error: "Error",
  },

  assistant: {
    hint: "Toca a Magec para hablar",
    placeholder: "Tu conversación aparecerá aquí...",
    textInputPlaceholder: "Escribe un mensaje...",
    currentConversation: "Conversación actual",
  },

  sessions: {
    title: "Conversaciones",
    empty: "No hay conversaciones aún",
    emptyPreview: "Conversación vacía",
    new: "Nueva conversación",
    delete: "Eliminar",
  },

  settings: {
    title: "Ajustes",
    back: "Volver",
    savedAutomatically: "Los ajustes se guardan automáticamente",

    wakeWord: {
      title: "Palabra de activación",
      description: 'Di "{phrase}" para activar',
      disabled: "Desactivada",
      model: "Modelo",
    },

    stt: {
      title: "Reconocimiento de voz",
      server: "Servidor",
      serverDesc: "Más preciso",
      browser: "Navegador",
      browserDesc: "Más privado",
    },

    tts: {
      title: "Síntesis de voz",
    },

    language: {
      title: "Idioma",
      es: "Español",
      en: "English",
    },
  },

  notifications: {
    title: "Notificaciones",
    empty: "No hay notificaciones",
    clearAll: "Limpiar todas",
    delete: "Eliminar",
    seeConsole: "Ver detalles en consola",
    wakeWordLoading: "Cargando modelo Wake Word...",
    wakeWordReady: "Modelo Wake Word listo",
    wakeWordFailed: "Modelo Wake Word falló. Ver consola.",
    whisperLoading: "Cargando modelo Whisper (local)...",
    whisperReady: "Modelo Whisper listo (local)",
    whisperFailed: "Modelo Whisper falló. Ver consola.",
    ttsUnavailable: "Síntesis de voz no disponible. El servidor TTS no está configurado o no responde.",
  },

  actions: {
    copy: "Copiar",
    clear: "Limpiar vista",
    send: "Enviar",
  },


  time: {
    justNow: "Ahora mismo",
    minutesAgo: "hace {n}m",
    hoursAgo: "hace {n}h",
    daysAgo: "hace {n}d",
  },

  errors: {
    microphoneAccess: "No se pudo acceder al micrófono. Revisa los permisos.",
    connectionFailed: "Error de conexión",
    transcriptionFailed: "Error al transcribir",
    generic: "Ha ocurrido un error",
  },
};
