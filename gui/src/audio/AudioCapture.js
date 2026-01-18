export class AudioCapture {
    constructor(config) {
        this.config = config;
        this.audioContext = null;
        this.micStream = null;
        this.analyser = null;
        this.workletNode = null;
        this.onAudioData = null;
    }

    async start() {
        if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
            throw new Error('Microphone not available. Make sure you are using HTTPS.');
        }
        
        this.micStream = await navigator.mediaDevices.getUserMedia({
            audio: {
                channelCount: 1,
                sampleRate: this.config.sampleRate,
                echoCancellation: true,
                noiseSuppression: true
            }
        });
        
        this.audioContext = new AudioContext({ sampleRate: this.config.sampleRate });
        const source = this.audioContext.createMediaStreamSource(this.micStream);
        
        this.analyser = this.audioContext.createAnalyser();
        this.analyser.fftSize = 256;
        source.connect(this.analyser);
        
        // Use AudioWorklet for capturing raw audio data
        await this.audioContext.audioWorklet.addModule('/src/audio/audio-processor.worklet.js');
        this.workletNode = new AudioWorkletNode(this.audioContext, 'audio-capture-processor');
        
        this.workletNode.port.onmessage = (event) => {
            if (this.onAudioData) {
                this.onAudioData(event.data.samples, event.data.sampleRate);
            }
        };
        
        source.connect(this.workletNode);
        this.workletNode.connect(this.audioContext.destination);
    }

    getAnalyser() {
        return this.analyser;
    }

    getAudioContext() {
        return this.audioContext;
    }

    getMicStream() {
        return this.micStream;
    }

    getSampleRate() {
        return this.audioContext?.sampleRate || this.config.sampleRate;
    }

    stop() {
        if (this.workletNode) {
            this.workletNode.disconnect();
            this.workletNode = null;
        }
        if (this.micStream) {
            this.micStream.getTracks().forEach(track => track.stop());
            this.micStream = null;
        }
        if (this.audioContext) {
            this.audioContext.close();
            this.audioContext = null;
        }
    }
}
