const STORAGE_PREFIX = 'magec_settings'

const DEFAULT_SETTINGS = {
  tts: { enabled: true },
  wakeWord: { enabled: true, model: 'oye-magec' },
  spokesperson: { agentId: null }
}

export class SettingsManager {
  constructor(agentId = null) {
    this._agentId = agentId
    this._settings = this._load()
    this._validWakeWordModels = null
  }

  get agentId() {
    return this._agentId
  }

  switchAgent(agentId) {
    this._agentId = agentId
    this._settings = this._load()
  }

  _storageKey() {
    return this._agentId
      ? `${STORAGE_PREFIX}_${this._agentId}`
      : STORAGE_PREFIX
  }

  _load() {
    try {
      const stored = localStorage.getItem(this._storageKey())
      if (stored) {
        return this._merge(DEFAULT_SETTINGS, JSON.parse(stored))
      }
    } catch {}
    return this._deepCopy(DEFAULT_SETTINGS)
  }

  _deepCopy(obj) {
    return JSON.parse(JSON.stringify(obj))
  }

  setValidWakeWordModels(modelIds) {
    this._validWakeWordModels = modelIds
    this._validateWakeWordModel()
  }

  _validateWakeWordModel() {
    if (!this._validWakeWordModels) return

    if (!this._validWakeWordModels.includes(this._settings.wakeWord.model)) {
      this._settings.wakeWord.model = this._validWakeWordModels[0] || DEFAULT_SETTINGS.wakeWord.model
      this._save()
    }
  }

  _merge(defaults, stored) {
    const result = { ...defaults }
    for (const key of Object.keys(defaults)) {
      if (stored[key] !== undefined) {
        if (typeof defaults[key] === 'object' && !Array.isArray(defaults[key])) {
          result[key] = { ...defaults[key], ...stored[key] }
        } else {
          result[key] = stored[key]
        }
      }
    }
    return result
  }

  _save() {
    try {
      localStorage.setItem(this._storageKey(), JSON.stringify(this._settings))
    } catch {}
  }

  get ttsEnabled() { return this._settings.tts.enabled }
  set ttsEnabled(value) { this._settings.tts.enabled = value; this._save() }

  get wakeWordEnabled() { return this._settings.wakeWord.enabled }
  set wakeWordEnabled(value) { this._settings.wakeWord.enabled = value; this._save() }

  get wakeWordModel() { return this._settings.wakeWord.model }
  set wakeWordModel(value) { this._settings.wakeWord.model = value; this._save() }

  get spokesperson() { return this._settings.spokesperson.agentId }
  set spokesperson(value) { this._settings.spokesperson.agentId = value; this._save() }
}
