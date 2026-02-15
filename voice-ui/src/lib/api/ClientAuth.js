const STORAGE_KEY = 'magec_client_token'

class ClientAuth {
  constructor() {
    this._token = localStorage.getItem(STORAGE_KEY) || ''
    this._clientInfo = null
    this._patchFetch()
  }

  get token() { return this._token }
  get isPaired() { return !!this._clientInfo?.paired }
  get clientName() { return this._clientInfo?.name || '' }
  get defaultAgent() { return this._clientInfo?.defaultAgent || '' }
  get allowedAgents() { return this._clientInfo?.allowedAgents || [] }

  setToken(token) {
    this._token = token
    localStorage.setItem(STORAGE_KEY, token)
  }

  clearToken() {
    this._token = ''
    localStorage.removeItem(STORAGE_KEY)
    this._clientInfo = null
  }

  async checkPairing() {
    try {
      const headers = {}
      if (this._token) {
        headers['Authorization'] = `Bearer ${this._token}`
      }
      const res = await fetch('/api/v1/client/info', { headers })
      if (res.status === 401) {
        this.clearToken()
        return false
      }
      if (!res.ok) return false
      this._clientInfo = await res.json()
      return this._clientInfo.paired === true
    } catch {
      return false
    }
  }

  async pair(token) {
    this.setToken(token)
    const ok = await this.checkPairing()
    if (!ok) {
      this.clearToken()
    }
    return ok
  }

  _patchFetch() {
    const originalFetch = window.fetch
    const self = this
    window.fetch = function (input, init = {}) {
      const url = typeof input === 'string' ? input : input?.url || ''
      if (self._token && url.startsWith('/api/')) {
        init = { ...init }
        init.headers = { ...init.headers }
        if (!init.headers['Authorization']) {
          init.headers['Authorization'] = `Bearer ${self._token}`
        }
      }
      return originalFetch.call(this, input, init)
    }
  }
}

export const clientAuth = new ClientAuth()
