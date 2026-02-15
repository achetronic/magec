import { CONFIG } from '../config.js'
import { stripMetadata } from '../utils/metadata.js'

export class SessionService {
  constructor() {
    this.baseUrl = CONFIG.agent.baseUrl
    this.appName = CONFIG.agent.appName
    this.userId = CONFIG.agent.defaultUserId
  }

  setAgent(agentId) {
    this.appName = agentId
  }

  async listSessions() {
    try {
      const response = await fetch(`${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions`)
      if (!response.ok) return []
      return await response.json() || []
    } catch {
      return []
    }
  }

  async getSession(sessionId) {
    try {
      const response = await fetch(
        `${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions/${sessionId}`
      )
      if (!response.ok) return null
      return await response.json()
    } catch {
      return null
    }
  }

  async deleteSession(sessionId) {
    try {
      const response = await fetch(
        `${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions/${sessionId}`,
        { method: 'DELETE' }
      )
      return response.ok
    } catch {
      return false
    }
  }

  extractMessages(session) {
    if (!session?.events) return []

    const messages = []
    for (const event of session.events) {
      if (event.content?.role && event.content?.parts?.[0]?.text) {
        messages.push({
          role: event.content.role,
          text: stripMetadata(event.content.parts[0].text)
        })
      }
    }
    return messages
  }

  getSessionPreview(session) {
    const messages = this.extractMessages(session)
    if (messages.length === 0) return 'Empty conversation'

    const firstUserMessage = messages.find(m => m.role === 'user')
    if (firstUserMessage) {
      const text = firstUserMessage.text
      return text.length > 50 ? text.substring(0, 50) + '...' : text
    }

    return 'Conversation'
  }
}
