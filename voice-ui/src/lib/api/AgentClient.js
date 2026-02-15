import { CONFIG } from '../config.js'

export class AgentClient {
  constructor() {
    this.baseUrl = CONFIG.agent.baseUrl
    this.appName = CONFIG.agent.appName
    this.userId = CONFIG.agent.defaultUserId
  }

  setAgent(agentId) {
    this.appName = agentId
  }

  async createSession(sessionId) {
    try {
      const response = await fetch(`${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions/${sessionId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({})
      })
      return response !== null
    } catch {
      return false
    }
  }

  async ensureSession(sessionId) {
    const exists = await this.sessionExists(sessionId)
    if (!exists) {
      return this.createSession(sessionId)
    }
    return true
  }

  async sessionExists(sessionId) {
    try {
      const response = await fetch(`${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions/${sessionId}`)
      return response.ok
    } catch {
      return false
    }
  }

  async sendMessage(sessionId, message) {
    await this.ensureSession(sessionId)

    const response = await fetch(`${this.baseUrl}/run`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        appName: this.appName,
        userId: this.userId,
        sessionId: sessionId,
        newMessage: {
          role: 'user',
          parts: [{ text: message }]
        }
      })
    })

    if (!response.ok) {
      const errorText = await response.text().catch(() => '')
      const error = new Error(errorText || `Server error: ${response.status}`)
      error.status = response.status
      throw error
    }

    return this._extractResponses(await response.json())
  }

  _extractResponses(result) {
    const responses = []

    if (Array.isArray(result)) {
      for (const event of result) {
        if (event.content?.parts?.[0]?.text) {
          responses.push(event.content.parts[0].text)
        }
      }
    } else if (result.content?.parts?.[0]?.text) {
      responses.push(result.content.parts[0].text)
    }

    return responses
  }
}
