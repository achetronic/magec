import { CONFIG } from '../config.js'

export class SessionManager {
  constructor(config = {}) {
    this.storageKey = config.storageKey || CONFIG.session.storageKey
    this.autoRotateMinutes = config.autoRotateMinutes || CONFIG.session.autoRotateMinutes
    this.maxStoredSessions = config.maxStoredSessions || CONFIG.session.maxStoredSessions

    this.currentSessionId = null
    this.rotationTimer = null
    this.onSessionChange = null
  }

  init() {
    const stored = this._loadFromStorage()

    if (stored.currentSessionId && stored.currentSessionCreatedAt) {
      const elapsed = Date.now() - stored.currentSessionCreatedAt
      const maxAge = this.autoRotateMinutes * 60 * 1000

      if (elapsed < maxAge) {
        this.currentSessionId = stored.currentSessionId
        this._scheduleRotation(maxAge - elapsed)
        return this.currentSessionId
      }
    }

    return this.newSession()
  }

  newSession() {
    this.currentSessionId = this._generateSessionId()
    this._saveSession(this.currentSessionId)
    this._scheduleRotation()

    if (this.onSessionChange) {
      this.onSessionChange(this.currentSessionId)
    }

    return this.currentSessionId
  }

  getCurrentSessionId() {
    return this.currentSessionId
  }

  getSessionHistory() {
    const stored = this._loadFromStorage()
    return stored.sessions || []
  }

  _generateSessionId() {
    const timestamp = Date.now().toString(36)
    const random = Math.random().toString(36).substring(2, 8)
    return `session_${timestamp}_${random}`
  }

  _scheduleRotation(delayMs = null) {
    if (this.rotationTimer) {
      clearTimeout(this.rotationTimer)
    }

    const delay = delayMs || this.autoRotateMinutes * 60 * 1000

    this.rotationTimer = setTimeout(() => {
      this.newSession()
    }, delay)
  }

  _saveSession(sessionId) {
    const stored = this._loadFromStorage()

    stored.sessions = stored.sessions || []
    stored.sessions.unshift({
      id: sessionId,
      createdAt: Date.now()
    })

    if (stored.sessions.length > this.maxStoredSessions) {
      stored.sessions = stored.sessions.slice(0, this.maxStoredSessions)
    }

    stored.currentSessionId = sessionId
    stored.currentSessionCreatedAt = Date.now()

    localStorage.setItem(this.storageKey, JSON.stringify(stored))
  }

  _loadFromStorage() {
    try {
      const data = localStorage.getItem(this.storageKey)
      return data ? JSON.parse(data) : {}
    } catch {
      return {}
    }
  }

  destroy() {
    if (this.rotationTimer) {
      clearTimeout(this.rotationTimer)
      this.rotationTimer = null
    }
  }
}
