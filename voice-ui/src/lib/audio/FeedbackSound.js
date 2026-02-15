export class FeedbackSound {
  constructor(options = {}) {
    this.audioContext = null
    this.volume = options.volume ?? 0.25
    this.enabled = true
  }

  _getContext() {
    if (!this.audioContext) {
      this.audioContext = new (window.AudioContext || window.webkitAudioContext)()
    }
    return this.audioContext
  }

  playWakeChime() {
    if (!this.enabled) return

    const ctx = this._getContext()
    const now = ctx.currentTime

    const reverb = this._createLushReverb(ctx)
    const reverbGain = ctx.createGain()
    reverbGain.gain.value = 0.45

    const dryGain = ctx.createGain()
    dryGain.gain.value = 0.55

    const master = ctx.createGain()
    master.gain.value = this.volume
    master.connect(ctx.destination)

    dryGain.connect(master)
    reverb.connect(reverbGain)
    reverbGain.connect(master)

    const notes = [
      { freq: 392.00, time: 0, dur: 0.5 },
      { freq: 587.33, time: 0.03, dur: 0.45 },
      { freq: 783.99, time: 0.06, dur: 0.4 },
      { freq: 1174.66, time: 0.09, dur: 0.35 },
    ]

    notes.forEach(note => {
      this._playCelestialTone(ctx, note.freq, now + note.time, note.dur, dryGain, reverb)
    })

    this._playShimmer(ctx, 2349.32, now + 0.06, 0.35, dryGain, reverb)
    this._playShimmer(ctx, 3135.96, now + 0.09, 0.3, dryGain, reverb)
  }

  playStopChime() {
    if (!this.enabled) return

    const ctx = this._getContext()
    const now = ctx.currentTime

    const reverb = this._createLushReverb(ctx)
    const reverbGain = ctx.createGain()
    reverbGain.gain.value = 0.4

    const dryGain = ctx.createGain()
    dryGain.gain.value = 0.6

    const master = ctx.createGain()
    master.gain.value = this.volume * 0.8
    master.connect(ctx.destination)

    dryGain.connect(master)
    reverb.connect(reverbGain)
    reverbGain.connect(master)

    const notes = [
      { freq: 783.99, time: 0, dur: 0.25 },
      { freq: 587.33, time: 0.08, dur: 0.3 },
    ]

    notes.forEach(note => {
      this._playCelestialTone(ctx, note.freq, now + note.time, note.dur, dryGain, reverb)
    })
  }

  _playCelestialTone(ctx, frequency, startTime, duration, dryNode, wetNode) {
    const osc = ctx.createOscillator()
    osc.type = 'sine'
    osc.frequency.value = frequency

    const vibrato = ctx.createOscillator()
    const vibratoGain = ctx.createGain()
    vibrato.frequency.value = 5.5
    vibratoGain.gain.value = frequency * 0.003
    vibrato.connect(vibratoGain)
    vibratoGain.connect(osc.frequency)
    vibrato.start(startTime)
    vibrato.stop(startTime + duration + 0.3)

    const harmonic = ctx.createOscillator()
    harmonic.type = 'sine'
    harmonic.frequency.value = frequency * 1.5

    const gain = ctx.createGain()
    const harmonicGain = ctx.createGain()

    gain.gain.setValueAtTime(0, startTime)
    gain.gain.linearRampToValueAtTime(0.4, startTime + 0.04)
    gain.gain.exponentialRampToValueAtTime(0.12, startTime + duration * 0.35)
    gain.gain.exponentialRampToValueAtTime(0.001, startTime + duration)

    harmonicGain.gain.setValueAtTime(0, startTime)
    harmonicGain.gain.linearRampToValueAtTime(0.08, startTime + 0.03)
    harmonicGain.gain.exponentialRampToValueAtTime(0.001, startTime + duration * 0.5)

    osc.connect(gain)
    harmonic.connect(harmonicGain)

    gain.connect(dryNode)
    gain.connect(wetNode)
    harmonicGain.connect(dryNode)
    harmonicGain.connect(wetNode)

    osc.start(startTime)
    osc.stop(startTime + duration + 0.1)
    harmonic.start(startTime)
    harmonic.stop(startTime + duration * 0.6)
  }

  _playShimmer(ctx, frequency, startTime, duration, dryNode, wetNode) {
    [-3, 3].forEach(detuneCents => {
      const osc = ctx.createOscillator()
      osc.type = 'sine'
      osc.frequency.value = frequency
      osc.detune.value = detuneCents

      const gain = ctx.createGain()
      gain.gain.setValueAtTime(0, startTime)
      gain.gain.linearRampToValueAtTime(0.03, startTime + 0.05)
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + duration)

      osc.connect(gain)
      gain.connect(dryNode)
      gain.connect(wetNode)

      osc.start(startTime)
      osc.stop(startTime + duration + 0.1)
    })
  }

  _createLushReverb(ctx) {
    const convolver = ctx.createConvolver()
    const rate = ctx.sampleRate
    const length = rate * 0.8
    const impulse = ctx.createBuffer(2, length, rate)

    for (let channel = 0; channel < 2; channel++) {
      const data = impulse.getChannelData(channel)
      for (let i = 0; i < length; i++) {
        const t = i / length
        const decay = Math.pow(1 - t, 2.0)
        const modulation = 1 + 0.1 * Math.sin(t * 50)
        data[i] = (Math.random() * 2 - 1) * decay * modulation * 0.5
      }
    }

    convolver.buffer = impulse
    return convolver
  }

  setEnabled(enabled) {
    this.enabled = enabled
  }

  setVolume(volume) {
    this.volume = Math.max(0, Math.min(1, volume))
  }
}
