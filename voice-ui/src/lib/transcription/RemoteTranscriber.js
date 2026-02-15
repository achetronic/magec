import { AudioConverter } from '../audio/AudioConverter.js'

export class RemoteTranscriber {
  constructor(config) {
    this.config = config
    this._agentId = 'default'
  }

  setAgent(agentId) {
    this._agentId = agentId
  }

  _transcriptionUrl() {
    return `/api/v1/voice/${this._agentId}/transcription`
  }

  async transcribe(blob) {
    const wavBlob = await AudioConverter.blobToWav(blob)

    const formData = new FormData()
    formData.append('file', wavBlob, 'audio.wav')
    formData.append('model', this.config.model)
    formData.append('language', 'es')

    const response = await fetch(this._transcriptionUrl(), {
      method: 'POST',
      body: formData
    })

    if (!response.ok) {
      throw new Error(`Transcription error: ${response.status}`)
    }

    const result = await response.json()
    return result.text?.trim() || ''
  }
}
