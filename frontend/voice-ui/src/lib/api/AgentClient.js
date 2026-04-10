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

  async sendMessage(sessionId, message, fileParts = []) {
    await this.ensureSession(sessionId)

    const parts = []
    if (message) {
      parts.push({ text: message })
    } else if (fileParts && fileParts.length > 0) {
      parts.push({ text: " " }) // Ensure at least a blank text part exists for LLM context
    }
    for (const fp of fileParts) {
      parts.push(fp)
    }

    const response = await fetch(`${this.baseUrl}/run`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        appName: this.appName,
        userId: this.userId,
        sessionId: sessionId,
        newMessage: {
          role: 'user',
          parts: parts
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
        if (!event.content?.parts?.length) continue

        for (const part of event.content.parts) {
          if (part.text) {
            responses.push({ type: 'text', text: part.text })
          } else if (part.functionCall) {
            responses.push({
              type: 'tool',
              name: part.functionCall.name,
              args: part.functionCall.args
            })
          } else if (part.functionResponse) {
            responses.push({
              type: 'tool_result',
              name: part.functionResponse.name,
              result: part.functionResponse.response
            })
          }
        }
      }
    } else if (result.content?.parts) {
      for (const part of result.content.parts) {
        if (part.text) {
          responses.push({ type: 'text', text: part.text })
        } else if (part.functionCall) {
          responses.push({
            type: 'tool',
            name: part.functionCall.name,
            args: part.functionCall.args
          })
        } else if (part.functionResponse) {
          responses.push({
            type: 'tool_result',
            name: part.functionResponse.name,
            result: part.functionResponse.response
          })
        }
      }
    }

    // Merge adjacent tool calls and results
    const merged = []
    for (const r of responses) {
      if (r.type === 'text') {
        merged.push(r)
      } else if (r.type === 'tool') {
        merged.push({ type: 'tool', name: r.name, args: r.args, result: null })
      } else if (r.type === 'tool_result') {
        const last = merged[merged.length - 1]
        if (last && last.type === 'tool' && last.name === r.name) {
          last.result = r.result
        } else {
          merged.push({ type: 'tool', name: r.name, args: null, result: r.result })
        }
      }
    }

    return merged
  }
}
