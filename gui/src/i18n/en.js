// English translations
export default {
  app: {
    title: "Magec",
    subtitle: "Voice Assistant",
  },

  status: {
    initializing: "Initializing...",
    ready: "Ready",
    listening: "Listening...",
    recording: "Recording...",
    processing: "Processing...",
    thinking: "Thinking...",
    speaking: "Speaking...",
    loadingWakeWord: "Loading wake word...",
    loadingWhisper: "Loading local Whisper...",
    error: "Error",
  },

  assistant: {
    hint: "Tap Magec to speak",
    placeholder: "Your conversation will appear here...",
    textInputPlaceholder: "Type a message...",
    currentConversation: "Current conversation",
  },

  sessions: {
    title: "Conversations",
    empty: "No conversations yet",
    emptyPreview: "Empty conversation",
    new: "New conversation",
    delete: "Delete",
  },

  settings: {
    title: "Settings",
    back: "Back",
    savedAutomatically: "Settings are saved automatically",

    wakeWord: {
      title: "Wake word",
      description: 'Say "{phrase}" to activate',
      disabled: "Disabled",
      model: "Model",
    },

    stt: {
      title: "Speech recognition",
      server: "Server",
      serverDesc: "More accurate",
      browser: "Browser",
      browserDesc: "More private",
    },

    tts: {
      title: "Text to speech",
    },

    language: {
      title: "Language",
      es: "Español",
      en: "English",
    },
  },

  notifications: {
    title: "Notifications",
    empty: "No notifications",
    clearAll: "Clear all",
    delete: "Delete",
    seeConsole: "See details in console",
    wakeWordLoading: "Loading Wake Word model...",
    wakeWordReady: "Wake Word model ready",
    wakeWordFailed: "Wake Word model failed. Check console.",
    whisperLoading: "Loading Whisper model (local)...",
    whisperReady: "Whisper model ready (local)",
    whisperFailed: "Whisper model failed. Check console.",
    ttsUnavailable: "Text-to-speech unavailable. TTS server is not configured or not responding.",
  },

  actions: {
    copy: "Copy",
    clear: "Clear view",
    send: "Send",
  },


  time: {
    justNow: "Just now",
    minutesAgo: "{n}m ago",
    hoursAgo: "{n}h ago",
    daysAgo: "{n}d ago",
  },

  errors: {
    microphoneAccess: "Could not access microphone. Check permissions.",
    connectionFailed: "Connection error",
    transcriptionFailed: "Transcription failed",
    generic: "An error occurred",
  },
};
