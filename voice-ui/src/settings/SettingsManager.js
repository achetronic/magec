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

const STORAGE_PREFIX = 'magec_settings';

const DEFAULT_SETTINGS = {
    tts: {
        enabled: true
    },
    wakeWord: {
        enabled: true,
        model: 'oye-magec'
    }
};

export class SettingsManager {
    constructor(agentId = null) {
        this._agentId = agentId;
        this._settings = this._load();
        this._validWakeWordModels = null;
    }

    get agentId() {
        return this._agentId;
    }

    switchAgent(agentId) {
        this._agentId = agentId;
        this._settings = this._load();
    }

    _storageKey() {
        return this._agentId
            ? `${STORAGE_PREFIX}_${this._agentId}`
            : STORAGE_PREFIX;
    }

    _load() {
        try {
            const stored = localStorage.getItem(this._storageKey());
            if (stored) {
                const parsed = JSON.parse(stored);
                return this._merge(DEFAULT_SETTINGS, parsed);
            }
        } catch (e) {
            console.warn('Failed to load settings:', e);
        }
        return this._deepCopy(DEFAULT_SETTINGS);
    }

    _deepCopy(obj) {
        return JSON.parse(JSON.stringify(obj));
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
            localStorage.setItem(this._storageKey(), JSON.stringify(this._settings));
        } catch (e) {
            console.warn('Failed to save settings:', e);
        }
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
