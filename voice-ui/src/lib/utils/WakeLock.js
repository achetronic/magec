export class WakeLock {
  constructor() {
    this._wakeLock = null
    this._enabled = false
  }

  static isSupported() {
    return 'wakeLock' in navigator
  }

  async enable() {
    if (!WakeLock.isSupported()) return false

    try {
      this._wakeLock = await navigator.wakeLock.request('screen')
      this._enabled = true

      this._wakeLock.addEventListener('release', () => {
        this._enabled = false
      })

      document.addEventListener('visibilitychange', () => this._onVisibilityChange())
      return true
    } catch {
      return false
    }
  }

  async disable() {
    if (this._wakeLock) {
      await this._wakeLock.release()
      this._wakeLock = null
      this._enabled = false
    }
  }

  async _onVisibilityChange() {
    if (document.visibilityState === 'visible' && this._enabled === false && this._wakeLock === null) {
      await this.enable()
    }
  }

  isEnabled() {
    return this._enabled
  }
}
