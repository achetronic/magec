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

// English translations
export default {
  app: {
    title: "Magec",
    subtitle: "Voice UI",
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
    ttsUnavailable: "Text-to-speech unavailable. TTS server is not configured or not responding.",
    wakeWordUnavailable: "Wake word detection unavailable. Server not configured.",
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

  pairing: {
    subtitle: "Enter the device token to pair",
    connect: "Pair",
    connecting: "Connecting...",
    error: "Invalid token or device disabled",
    hint: "Generate the token in Admin > Devices",
  },

  errors: {
    microphoneAccess: "Could not access microphone. Check permissions.",
    connectionFailed: "Connection error",
    transcriptionFailed: "Transcription failed",
    generic: "An error occurred",
  },
};
