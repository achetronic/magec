export class VoiceEventsClient {
  constructor(config = {}) {
    this.config = {
      wsUrl: config.wsUrl || this._buildWsUrl(),
      ...config
    }

    this.onWakeword = null
    this.onSpeechStart = null
    this.onSpeechEnd = null
    this.onCapabilities = null
    this.onError = null

    this.ws = null
    this.isConnected = false
    this.isLoaded = false
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectDelay = 1000

    this.capabilities = null
    this.wakewordModels = []
    this.activeWakeword = null
    this.vadEnabled = false
    this.vadSilenceTimeout = 2000

    this.sampleRate = null
  }

  _buildWsUrl() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}/api/v1/voice/events`
  }

  async load() {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Connection timeout'))
      }, 5000)

      try {
        this.ws = new WebSocket(this.config.wsUrl)

        this.ws.onopen = () => {
          this.isConnected = true
          this.reconnectAttempts = 0
        }

        this.ws.onclose = (event) => {
          this.isConnected = false
          if (this.isLoaded) {
            this._attemptReconnect()
          }
        }

        this.ws.onerror = (error) => {
          this.onError?.(error)
          if (!this.isLoaded) {
            clearTimeout(timeout)
            reject(new Error('Failed to connect to voice events server'))
          }
        }

        this.ws.onmessage = (event) => {
          const resolved = this._handleMessage(event.data, resolve)
          if (resolved) {
            clearTimeout(timeout)
          }
        }
      } catch (e) {
        clearTimeout(timeout)
        reject(e)
      }
    })
  }

  _handleMessage(data, resolveLoad) {
    try {
      const msg = JSON.parse(data)

      switch (msg.type) {
        case 'capabilities':
          this._handleCapabilities(msg.data)
          if (!this.isLoaded && resolveLoad) {
            this.isLoaded = true
            resolveLoad(true)
            return true
          }
          break
        case 'wakeword':
          this.onWakeword?.(msg.data?.model)
          break
        case 'speech_start':
          this.onSpeechStart?.()
          break
        case 'speech_end':
          this.onSpeechEnd?.()
          break
        case 'error':
          this.onError?.(msg.data)
          break
      }
    } catch (e) {
      console.error('[VoiceEvents] Failed to parse message:', e)
    }
    return false
  }

  _handleCapabilities(capabilities) {
    this.capabilities = capabilities

    if (capabilities.wakewords) {
      this.wakewordModels = capabilities.wakewords.models || []
      this.activeWakeword = capabilities.wakewords.active
    }

    if (capabilities.vad) {
      this.vadEnabled = capabilities.vad.enabled
      this.vadSilenceTimeout = capabilities.vad.silenceTimeout || 2000
    }

    this.onCapabilities?.(capabilities)
  }

  _attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) return

    this.reconnectAttempts++
    const delay = this.reconnectDelay * this.reconnectAttempts

    setTimeout(() => {
      if (!this.isConnected) {
        this.load().catch(() => {})
      }
    }, delay)
  }

  _sendConfig() {
    if (!this.isConnected) return

    const config = {
      type: 'config',
      data: {
        sampleRate: this.sampleRate,
        model: this.activeWakeword
      }
    }

    this.ws.send(JSON.stringify(config))
  }

  getWakewordModels() {
    return this.wakewordModels
  }

  getActiveWakeword() {
    return this.activeWakeword
  }

  setWakewordModel(modelId) {
    if (!this.isConnected) return

    this.ws.send(JSON.stringify({
      type: 'setModel',
      data: { model: modelId }
    }))
    this.activeWakeword = modelId
  }

  getActivePhrase() {
    const model = this.wakewordModels.find(m => m.id === this.activeWakeword)
    return model?.phrase || this.activeWakeword
  }

  isVADEnabled() {
    return this.vadEnabled
  }

  getVADSilenceTimeout() {
    return this.vadSilenceTimeout
  }

  async processAudio(audioData, inputSampleRate) {
    if (!this.isConnected) return

    if (inputSampleRate !== this.sampleRate) {
      this.sampleRate = inputSampleRate
      this._sendConfig()
    }

    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(audioData.buffer)
    }
  }

  isReady() {
    return this.isConnected && this.isLoaded
  }

  stop() {
    this.isLoaded = false
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.isConnected = false
  }
}
