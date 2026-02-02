/*
 * Copyright 2025 Alby Hernández
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
    ttsUnavailable: "Síntesis de voz no disponible. El servidor TTS no está configurado o no responde.",
    wakeWordUnavailable: "Detección de wake word no disponible. Servidor no configurado.",
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
