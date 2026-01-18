const STORAGE_KEY = 'magec_settings';

const DEFAULT_SETTINGS = {
    stt: {
        mode: 'server'  // 'server' | 'local'
    },
    tts: {
        enabled: true
    },
    wakeWord: {
        enabled: true,
        model: 'oye-magec'
    }
};

export class SettingsManager {
    constructor() {
        this._settings = this._load();
        this._validWakeWordModels = null;
    }

    _load() {
        try {
            const stored = localStorage.getItem(STORAGE_KEY);
            if (stored) {
                const parsed = JSON.parse(stored);
                return this._merge(DEFAULT_SETTINGS, parsed);
            }
        } catch (e) {
            console.warn('Failed to load settings:', e);
        }
        return { ...DEFAULT_SETTINGS };
    }

    setValidWakeWordModels(modelIds) {
        this._validWakeWordModels = modelIds;
        this._validateWakeWordModel();
    }

    _validateWakeWordModel() {
        if (!this._validWakeWordModels) return;
        
        if (!this._validWakeWordModels.includes(this._settings.wakeWord.model)) {
            console.warn(`Invalid wake word model "${this._settings.wakeWord.model}", resetting to default`);
            this._settings.wakeWord.model = this._validWakeWordModels[0] || DEFAULT_SETTINGS.wakeWord.model;
            this._save();
        }
    }

    _merge(defaults, stored) {
        const result = { ...defaults };
        for (const key of Object.keys(defaults)) {
            if (stored[key] !== undefined) {
                if (typeof defaults[key] === 'object' && !Array.isArray(defaults[key])) {
                    result[key] = { ...defaults[key], ...stored[key] };
                } else {
                    result[key] = stored[key];
                }
            }
        }
        return result;
    }

    _save() {
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(this._settings));
        } catch (e) {
            console.warn('Failed to save settings:', e);
        }
    }

    // STT
    get sttMode() {
        return this._settings.stt.mode;
    }

    set sttMode(value) {
        this._settings.stt.mode = value;
        this._save();
    }

    // TTS
    get ttsEnabled() {
        return this._settings.tts.enabled;
    }

    set ttsEnabled(value) {
        this._settings.tts.enabled = value;
        this._save();
    }

    // Wake word
    get wakeWordEnabled() {
        return this._settings.wakeWord.enabled;
    }

    set wakeWordEnabled(value) {
        this._settings.wakeWord.enabled = value;
        this._save();
    }

    get wakeWordModel() {
        return this._settings.wakeWord.model;
    }

    set wakeWordModel(value) {
        this._settings.wakeWord.model = value;
        this._save();
    }
}
