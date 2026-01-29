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

/**
 * OpenWakeWord detector using ONNX Runtime Web
 * Implements the full pipeline: audio → melspectrogram → embeddings → wake word detection
 */
export class OpenWakeWordDetector {
    constructor(config, onDetected) {
        this.config = {
            melspecModelPath: '/pretrained/mel-spectrogram.onnx',
            embeddingModelPath: '/pretrained/speech-embedding.onnx',
            wakeWordModelPath: config.modelPath,
            threshold: config.threshold || 0.5,
            cooldownMs: config.cooldownMs || 2000,
            sampleRate: 16000,
            frameSize: 1280, // 80ms at 16kHz
            ...config
        };
        
        this.onDetected = onDetected;
        this.onSilence = null;
        
        // ONNX sessions
        this.melspecSession = null;
        this.embeddingSession = null;
        this.wakeWordSession = null;
        
        // State
        this.enabled = true;
        this.isLoaded = false;
        this.lastDetectionTime = 0;
        this.isProcessing = false;
        
        // Audio buffer - accumulate raw samples
        this.audioBuffer = [];
        
        // Silence detection
        this.lastSpeechTime = 0;
        this.silenceTimeout = null;
        this.hasSpokenOnce = false;
    }

    async load() {
        if (typeof ort === 'undefined') {
            throw new Error('ONNX Runtime Web not loaded. Include ort.min.js before using OpenWakeWordDetector');
        }

        const modelName = this.config.wakeWordModelPath.split('/').pop().replace('.onnx', '');
        console.log('[OpenWakeWord] Loading models for:', modelName);

        try {
            // Configure ONNX Runtime
            ort.env.wasm.numThreads = 1;
            
            // Load all three models
            const [melspec, embedding, wakeWord] = await Promise.all([
                ort.InferenceSession.create(this.config.melspecModelPath, {
                    executionProviders: ['wasm']
                }),
                ort.InferenceSession.create(this.config.embeddingModelPath, {
                    executionProviders: ['wasm']
                }),
                ort.InferenceSession.create(this.config.wakeWordModelPath, {
                    executionProviders: ['wasm']
                })
            ]);

            this.melspecSession = melspec;
            this.embeddingSession = embedding;
            this.wakeWordSession = wakeWord;

            this.isLoaded = true;
            console.log('[OpenWakeWord] Models loaded successfully');
            return true;
        } catch (e) {
            console.error('[OpenWakeWord] Failed to load models:', e);
            throw e;
        }
    }

    async processAudio(audioData, inputSampleRate) {
        // Always check for speech (for silence detection) even when wake word is disabled
        this._updateSpeechDetection(audioData);
        
        if (!this.isLoaded || !this.enabled || this.isProcessing) return;
        
        // Resample if needed
        let audio16k;
        if (inputSampleRate !== this.config.sampleRate) {
            audio16k = this._resample(audioData, inputSampleRate, this.config.sampleRate);
        } else {
            audio16k = audioData;
        }

        // Convert to Int16 if Float32
        let audioInt16;
        if (audio16k instanceof Float32Array) {
            audioInt16 = new Int16Array(audio16k.length);
            for (let i = 0; i < audio16k.length; i++) {
                const val = Math.floor(audio16k[i] * 32767);
                audioInt16[i] = Math.max(-32768, Math.min(32767, val));
            }
        } else {
            audioInt16 = audio16k;
        }

        // Add to audio buffer
        for (let i = 0; i < audioInt16.length; i++) {
            this.audioBuffer.push(audioInt16[i]);
        }
        
        // Keep max ~5 seconds of audio (enough for processing)
        const maxAudioLength = this.config.sampleRate * 5;
        if (this.audioBuffer.length > maxAudioLength) {
            this.audioBuffer = this.audioBuffer.slice(-maxAudioLength);
        }

        // Process when we have enough audio (at least 2 seconds for good context)
        const minSamples = this.config.sampleRate * 2;
        if (this.audioBuffer.length >= minSamples) {
            await this._processBuffer();
        }
    }

    async _processBuffer() {
        if (this.isProcessing) return;
        this.isProcessing = true;
        
        try {
            // Get audio from buffer
            const audioInt16 = new Int16Array(this.audioBuffer);
            
            // Step 1: Compute melspectrogram
            const melspec = await this._getMelspectrogram(audioInt16);
            
            if (melspec.length < 76) {
                // Need at least 76 frames for embedding
                return;
            }
            
            // Step 2: Compute embeddings using sliding window
            const embeddings = await this._getEmbeddings(melspec);
            
            if (embeddings.length < 16) {
                // Need at least 16 embeddings for wake word model
                return;
            }
            
            // Step 3: Run wake word detection on last 16 embeddings
            const features = embeddings.slice(-16);
            const score = await this._runWakeWordModel(features);

            // Check for wake word detection
            if (score >= this.config.threshold) {
                const now = Date.now();
                if (now - this.lastDetectionTime >= this.config.cooldownMs) {
                    console.log(`[OpenWakeWord] Detected! Score: ${score.toFixed(2)}`);
                    this.lastDetectionTime = now;
                    // Clear buffer after detection
                    this.audioBuffer = [];
                    this.onDetected?.();
                }
            }
            
            // Trim audio buffer - keep last 2 seconds for sliding window
            const keepSamples = this.config.sampleRate * 2;
            if (this.audioBuffer.length > keepSamples) {
                this.audioBuffer = this.audioBuffer.slice(-keepSamples);
            }
            
        } catch (e) {
            console.error('[OpenWakeWord] Processing error:', e);
        } finally {
            this.isProcessing = false;
        }
    }

    async _getMelspectrogram(audioInt16) {
        // Convert Int16 to Float32 for ONNX
        const audioFloat = new Float32Array(audioInt16.length);
        for (let i = 0; i < audioInt16.length; i++) {
            audioFloat[i] = audioInt16[i];
        }

        // Create input tensor [1, N]
        const inputTensor = new ort.Tensor('float32', audioFloat, [1, audioFloat.length]);
        
        // Run inference
        const results = await this.melspecSession.run({ 'input': inputTensor });
        const output = results[Object.keys(results)[0]];
        
        // Output shape is [1, 1, frames, 32] - batch, channels, frames, mel_bins
        const numFrames = output.dims[2];
        const numBins = output.dims[3];
        
        // Apply transform: melspec / 10 + 2
        const data = new Float32Array(output.data.length);
        for (let i = 0; i < output.data.length; i++) {
            data[i] = output.data[i] / 10 + 2;
        }

        // Reshape to array of [32] frames
        const melspec = [];
        for (let i = 0; i < numFrames; i++) {
            const frame = new Float32Array(numBins);
            for (let j = 0; j < numBins; j++) {
                frame[j] = data[i * numBins + j];
            }
            melspec.push(frame);
        }
        
        return melspec;
    }

    async _getEmbeddings(melspec) {
        // Create windows of 76 frames with step size 8
        const windows = [];
        for (let i = 0; i <= melspec.length - 76; i += 8) {
            windows.push(melspec.slice(i, i + 76));
        }

        if (windows.length === 0) {
            return [];
        }

        // Process all windows in batch
        const batchSize = windows.length;
        const flatData = new Float32Array(batchSize * 76 * 32);
        
        let idx = 0;
        for (let b = 0; b < batchSize; b++) {
            for (let i = 0; i < 76; i++) {
                for (let j = 0; j < 32; j++) {
                    flatData[idx++] = windows[b][i][j];
                }
            }
        }

        const inputTensor = new ort.Tensor('float32', flatData, [batchSize, 76, 32, 1]);
        const results = await this.embeddingSession.run({ 'input_1': inputTensor });
        const output = results[Object.keys(results)[0]];
        
        // Reshape to [batch, 96]
        const embeddings = [];
        for (let b = 0; b < batchSize; b++) {
            const emb = new Float32Array(96);
            for (let j = 0; j < 96; j++) {
                emb[j] = output.data[b * 96 + j];
            }
            embeddings.push(emb);
        }
        
        return embeddings;
    }

    async _runWakeWordModel(features) {
        // Create input tensor [1, 16, 96]
        const flatData = new Float32Array(16 * 96);
        let idx = 0;
        for (let i = 0; i < 16; i++) {
            for (let j = 0; j < 96; j++) {
                flatData[idx++] = features[i][j];
            }
        }

        const inputTensor = new ort.Tensor('float32', flatData, [1, 16, 96]);
        
        // Get input name from model
        const inputName = this.wakeWordSession.inputNames[0];
        const feeds = {};
        feeds[inputName] = inputTensor;
        
        const results = await this.wakeWordSession.run(feeds);
        const output = results[Object.keys(results)[0]];
        
        return output.data[0];
    }

    _resample(audioData, fromRate, toRate) {
        if (fromRate === toRate) return audioData;
        
        const ratio = toRate / fromRate;
        const newLength = Math.round(audioData.length * ratio);
        const result = new Float32Array(newLength);
        
        for (let i = 0; i < newLength; i++) {
            const srcIdx = i / ratio;
            const srcIdxFloor = Math.floor(srcIdx);
            const srcIdxCeil = Math.min(srcIdxFloor + 1, audioData.length - 1);
            const t = srcIdx - srcIdxFloor;
            
            // Linear interpolation
            result[i] = audioData[srcIdxFloor] * (1 - t) + audioData[srcIdxCeil] * t;
        }
        
        return result;
    }

    _updateSpeechDetection(audioData) {
        // Calculate RMS energy to detect speech
        let sum = 0;
        for (let i = 0; i < audioData.length; i++) {
            sum += audioData[i] * audioData[i];
        }
        const rms = Math.sqrt(sum / audioData.length);
        
        // Threshold for speech detection (adjust if needed)
        const speechThreshold = 0.01;
        if (rms > speechThreshold) {
            this.lastSpeechTime = Date.now();
            this.hasSpokenOnce = true;
        }
    }

    startSilenceDetection(silenceMs = 2000) {
        this.hasSpokenOnce = false;
        this.lastSpeechTime = Date.now();
        this.silenceTimeout = setInterval(() => {
            if (!this.hasSpokenOnce) {
                this.lastSpeechTime = Date.now();
                return;
            }
            
            const silenceDuration = Date.now() - this.lastSpeechTime;
            if (silenceDuration >= silenceMs) {
                this.stopSilenceDetection();
                this.onSilence?.();
            }
        }, 100);
    }

    stopSilenceDetection() {
        if (this.silenceTimeout) {
            clearInterval(this.silenceTimeout);
            this.silenceTimeout = null;
        }
        this.hasSpokenOnce = false;
    }

    setEnabled(enabled) {
        this.enabled = enabled;
    }

    isEnabled() {
        return this.enabled;
    }

    stop() {
        this.stopSilenceDetection();
        this.enabled = false;
        this.isLoaded = false;
        this.audioBuffer = [];
        this.melspecSession = null;
        this.embeddingSession = null;
        this.wakeWordSession = null;
    }
}
