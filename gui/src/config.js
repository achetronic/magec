export const CONFIG = {
    wakeWord: {
        modelPath: '/models/hey-buddy.onnx',
        metadataPath: '/models/wakeword.json',
        defaultPhrase: 'Hey Buddy',
        threshold: 0.5,  // Lower = more sensitive
        cooldownMs: 2000,
        // AudioCapture still needs these for waveform visualization
        sampleRate: 16000,
        bufferTime: 1.5
    },
    whisper: {
        model: 'Xenova/whisper-small',
        language: 'spanish',
        task: 'transcribe'
    },
    remote: {
        url: '/api/v1/transcription/v1/audio/transcriptions',
        model: 'parakeet'
    },
    agent: {
        baseUrl: '/api/v1/agent',
        appName: 'magec_agent',
        defaultUserId: 'default_user'
    },
    session: {
        storageKey: 'magec_sessions',
        autoRotateMinutes: 30,
        maxStoredSessions: 50
    }
};
