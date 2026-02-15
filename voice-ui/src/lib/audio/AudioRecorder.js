export class AudioRecorder {
  constructor(micStream) {
    this.micStream = micStream
    this.mediaRecorder = null
    this.recordedChunks = []
    this.isRecording = false
    this.onRecordingComplete = null
  }

  start() {
    if (this.isRecording) return

    this.isRecording = true
    this.recordedChunks = []

    this.mediaRecorder = new MediaRecorder(this.micStream)
    this.mediaRecorder.ondataavailable = (e) => {
      if (e.data.size > 0) {
        this.recordedChunks.push(e.data)
      }
    }
    this.mediaRecorder.onstop = () => this._processRecording()
    this.mediaRecorder.start(100)
  }

  stop() {
    if (!this.isRecording) return

    this.isRecording = false
    if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
      this.mediaRecorder.stop()
    }
  }

  getIsRecording() {
    return this.isRecording
  }

  _processRecording() {
    const blob = new Blob(this.recordedChunks, { type: 'audio/webm' })
    if (blob.size >= 1000 && this.onRecordingComplete) {
      this.onRecordingComplete(blob)
    }
  }
}
