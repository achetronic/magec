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
 * ServerWakeWordDetector - Wake word detection via WebSocket to server
 * Sends audio to the server for processing, reducing mobile CPU load
 */
export class ServerWakeWordDetector {
    constructor(config, onDetected) {
        this.config = {
            wsUrl: config.wsUrl || this._buildWsUrl(),
            ...config
        };
        
        this.onDetected = onDetected;
        this.onSilence = null;
        this.onModelsReceived = null;
        
        this.ws = null;
        this.isConnected = false;
        this.enabled = true;
        this.isLoaded = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        
        // Models from server
        this.models = [];
        this.activeModel = null;
        
        // Silence detection (still done locally for responsiveness)
        this.lastSpeechTime = 0;
        this.silenceTimeout = null;
        this.hasSpokenOnce = false;
        
        // Audio buffer for sending
        this.sampleRate = null;
    }

    _buildWsUrl() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        return `${protocol}//${window.location.host}/api/v1/wakeword`;
    }

    async load() {
        console.log('[ServerWakeWord] Connecting to server...');
        
        return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Connection timeout'));
            }, 5000);
            
            try {
                this.ws = new WebSocket(this.config.wsUrl);
                
                this.ws.onopen = () => {
                    console.log('[ServerWakeWord] Connected to server');
                    this.isConnected = true;
                    this.reconnectAttempts = 0;
                    // Don't resolve yet - wait for models message
                };
                
                this.ws.onclose = (event) => {
                    console.log('[ServerWakeWord] Disconnected', event.code, event.reason);
                    this.isConnected = false;
                    if (this.isLoaded) {
                        this._attemptReconnect();
                    }
                };
                
                this.ws.onerror = (error) => {
                    console.error('[ServerWakeWord] WebSocket error:', error);
                    if (!this.isLoaded) {
                        clearTimeout(timeout);
                        reject(new Error('Failed to connect to wake word server'));
                    }
                };
                
                this.ws.onmessage = (event) => {
                    const resolved = this._handleMessage(event.data, resolve);
                    if (resolved) {
                        clearTimeout(timeout);
                    }
                };
                
            } catch (e) {
                clearTimeout(timeout);
                console.error('[ServerWakeWord] Failed to create WebSocket:', e);
                reject(e);
            }
        });
    }

    _handleMessage(data, resolveLoad) {
        try {
            const msg = JSON.parse(data);
            
            switch (msg.type) {
                case 'models':
                    this.models = msg.data.models || [];
                    this.activeModel = msg.data.active;
                    console.log('[ServerWakeWord] Models received:', this.models.map(m => m.id), 'active:', this.activeModel);
                    
                    // Notify listener
                    this.onModelsReceived?.(this.models, this.activeModel);
                    
                    // Resolve load() on first models message
                    if (!this.isLoaded && resolveLoad) {
                        this.isLoaded = true;
                        resolveLoad(true);
                        return true;
                    }
                    break;
                    
                case 'detected':
                    console.log('[ServerWakeWord] ✅ DETECTED by server!', msg.data?.model);
                    this.onDetected?.();
                    break;
                    
                case 'error':
                    console.error('[ServerWakeWord] Server error:', msg.data);
                    break;
                    
                default:
                    console.log('[ServerWakeWord] Unknown message type:', msg.type);
            }
        } catch (e) {
            console.error('[ServerWakeWord] Failed to parse message:', e);
        }
        return false;
    }

    _attemptReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('[ServerWakeWord] Max reconnect attempts reached');
            return;
        }
        
        this.reconnectAttempts++;
        const delay = this.reconnectDelay * this.reconnectAttempts;
        
        console.log(`[ServerWakeWord] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
        
        setTimeout(() => {
            if (!this.isConnected && this.enabled) {
                this.load().catch(e => {
                    console.error('[ServerWakeWord] Reconnect failed:', e);
                });
            }
        }, delay);
    }

    _sendConfig() {
        if (!this.isConnected) return;
        
        const config = {
            type: 'config',
            data: {
                sampleRate: this.sampleRate,
                model: this.activeModel
            }
        };
        
        this.ws.send(JSON.stringify(config));
        console.log('[ServerWakeWord] Sent config:', config.data);
    }

    /**
     * Get available wake word models from server
     */
    getModels() {
        return this.models;
    }

    /**
     * Get the currently active model ID
     */
    getActiveModel() {
        return this.activeModel;
    }

    /**
     * Set the active wake word model
     */
    setModel(modelId) {
        if (!this.isConnected) return;
        
        const msg = {
            type: 'setModel',
            data: { model: modelId }
        };
        
        this.ws.send(JSON.stringify(msg));
        this.activeModel = modelId;
        console.log('[ServerWakeWord] Set model:', modelId);
    }

    /**
     * Get the phrase for the active model
     */
    getActivePhrase() {
        const model = this.models.find(m => m.id === this.activeModel);
        return model?.phrase || this.activeModel;
    }

    async processAudio(audioData, inputSampleRate) {
        // Update speech detection locally
        this._updateSpeechDetection(audioData);
        
        if (!this.isConnected || !this.enabled) return;
        
        // Send sample rate config if changed
        if (inputSampleRate !== this.sampleRate) {
            this.sampleRate = inputSampleRate;
            this._sendConfig();
        }
        
        // Send audio as binary (Float32Array)
        if (this.ws.readyState === WebSocket.OPEN) {
            // Convert Float32Array to ArrayBuffer and send
            this.ws.send(audioData.buffer);
        }
    }

    _updateSpeechDetection(audioData) {
        // Calculate RMS energy to detect speech
        let sum = 0;
        for (let i = 0; i < audioData.length; i++) {
            sum += audioData[i] * audioData[i];
        }
        const rms = Math.sqrt(sum / audioData.length);
        
        // Threshold for speech detection
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
        
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.isConnected = false;
    }
}
