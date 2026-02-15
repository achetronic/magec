export class OpenAITTS {
  constructor() {
    this._audio = null
    this._speaking = false
    this._abortController = null
    this._available = null
    this._agentId = 'default'
  }

  setAgent(agentId) {
    this._agentId = agentId
  }

  _speechUrl() {
    return `/api/v1/voice/${this._agentId}/speech`
  }

  async checkAvailable() {
    try {
      const response = await fetch(this._speechUrl(), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ input: 'test' }),
      })
      this._available = response.ok
    } catch {
      this._available = false
    }
    return this._available
  }

  isAvailable() {
    return this._available === true
  }

  _cleanText(text) {
    return text
      .replace(/\*\*/g, '')
      .replace(/\*/g, '')
      .replace(/`/g, '')
      .replace(/#{1,6}\s/g, '')
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
      .replace(/\s+/g, ' ')
      .trim()
  }

  async speak(text) {
    const cleanedText = this._cleanText(text)
    if (!cleanedText) return

    this.stop()
    this._abortController = new AbortController()

    try {
      const response = await fetch(this._speechUrl(), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ input: cleanedText }),
        signal: this._abortController.signal,
      })

      if (!response.ok) {
        const error = await response.text().catch(() => '')
        throw new Error(`TTS request failed: ${response.status} ${error}`)
      }

      const audioBlob = await response.blob()
      const audioUrl = URL.createObjectURL(audioBlob)

      await this._playAudio(audioUrl)

      URL.revokeObjectURL(audioUrl)
    } catch (e) {
      if (e.name === 'AbortError') return
      this._speaking = false
      throw e
    }
  }

  _playAudio(url) {
    return new Promise((resolve, reject) => {
      this._audio = new Audio(url)
      this._speaking = true

      this._audio.onended = () => {
        this._speaking = false
        this._audio = null
        resolve()
      }

      this._audio.onerror = () => {
        this._speaking = false
        this._audio = null
        reject(new Error('Audio playback failed'))
      }

      this._audio.play().catch((e) => {
        this._speaking = false
        this._audio = null
        reject(e)
      })
    })
  }

  stop() {
    if (this._abortController) {
      this._abortController.abort()
      this._abortController = null
    }

    if (this._audio) {
      this._audio.pause()
      this._audio.currentTime = 0
      this._audio = null
    }
    this._speaking = false
  }

  isSpeaking() {
    return this._speaking
  }
}
