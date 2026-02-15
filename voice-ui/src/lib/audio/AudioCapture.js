export class AudioCapture {
  constructor() {
    this.audioContext = null
    this.micStream = null
    this.analyser = null
    this.workletNode = null
    this.onAudioData = null
  }

  async start() {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      throw new Error('Microphone not available. Make sure you are using HTTPS.')
    }

    this.micStream = await navigator.mediaDevices.getUserMedia({
      audio: {
        channelCount: 1,
        echoCancellation: false,
        noiseSuppression: false,
        autoGainControl: false
      }
    })

    this.audioContext = new AudioContext()
    const source = this.audioContext.createMediaStreamSource(this.micStream)

    this.analyser = this.audioContext.createAnalyser()
    this.analyser.fftSize = 256
    source.connect(this.analyser)

    await this.audioContext.audioWorklet.addModule('/audio-processor.worklet.js')
    this.workletNode = new AudioWorkletNode(this.audioContext, 'audio-capture-processor')

    this.workletNode.port.onmessage = (event) => {
      if (this.onAudioData) {
        this.onAudioData(event.data.samples, event.data.sampleRate)
      }
    }

    source.connect(this.workletNode)
    this.workletNode.connect(this.audioContext.destination)
  }

  getAnalyser() {
    return this.analyser
  }

  getAudioContext() {
    return this.audioContext
  }

  getMicStream() {
    return this.micStream
  }

  getSampleRate() {
    return this.audioContext?.sampleRate
  }

  stop() {
    if (this.workletNode) {
      this.workletNode.disconnect()
      this.workletNode = null
    }
    if (this.micStream) {
      this.micStream.getTracks().forEach(track => track.stop())
      this.micStream = null
    }
    if (this.audioContext) {
      this.audioContext.close()
      this.audioContext = null
    }
  }
}
