/*
 * Copyright 2025 Alby Hernández
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

export class AudioCapture {
    constructor() {
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
        
        // Note: We don't force sampleRate in getUserMedia constraints because
        // many mobile devices ignore it and use their native rate (often 48kHz).
        // The actual sample rate is handled by AudioContext and passed to consumers.
        this.micStream = await navigator.mediaDevices.getUserMedia({
            audio: {
                channelCount: 1,
                echoCancellation: false,
                noiseSuppression: false,
                autoGainControl: false
            }
        });
        
        // Don't force sample rate on AudioContext either - let the browser use
        // the native rate. Mobile browsers often ignore requested sample rates.
        // Resampling to 16kHz for wake word detection is handled by the server.
        this.audioContext = new AudioContext();
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
        return this.audioContext?.sampleRate;
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
