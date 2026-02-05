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
 * VoiceEventsClient - Voice events detection via WebSocket to server
 * Handles wake word detection and VAD (Voice Activity Detection)
 * Sends audio to the server for processing, reducing mobile CPU load
 */
export class VoiceEventsClient {
    constructor(config = {}) {
        this.config = {
            wsUrl: config.wsUrl || this._buildWsUrl(),
            ...config
        };
        
        // Event callbacks
        this.onWakeword = null;
        this.onSpeechStart = null;
        this.onSpeechEnd = null;
        this.onCapabilities = null;
        this.onError = null;
        
        this.ws = null;
        this.isConnected = false;
        this.isLoaded = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        
        // Capabilities from server
        this.capabilities = null;
        this.wakewordModels = [];
        this.activeWakeword = null;
        this.vadEnabled = false;
        this.vadSilenceTimeout = 2000;
        
        // Audio config
        this.sampleRate = null;
    }

    _buildWsUrl() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        return `${protocol}//${window.location.host}/api/v1/voice-events`;
    }

    async load() {
        console.log('[VoiceEvents] Connecting to server...');
        
        return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Connection timeout'));
            }, 5000);
            
            try {
                this.ws = new WebSocket(this.config.wsUrl);
                
                this.ws.onopen = () => {
                    console.log('[VoiceEvents] Connected to server');
                    this.isConnected = true;
                    this.reconnectAttempts = 0;
                    // Don't resolve yet - wait for capabilities message
                };
                
                this.ws.onclose = (event) => {
                    console.log('[VoiceEvents] Disconnected', event.code, event.reason);
                    this.isConnected = false;
                    if (this.isLoaded) {
                        this._attemptReconnect();
                    }
                };
                
                this.ws.onerror = (error) => {
                    console.error('[VoiceEvents] WebSocket error:', error);
                    this.onError?.(error);
                    if (!this.isLoaded) {
                        clearTimeout(timeout);
                        reject(new Error('Failed to connect to voice events server'));
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
                console.error('[VoiceEvents] Failed to create WebSocket:', e);
                reject(e);
            }
        });
    }

    _handleMessage(data, resolveLoad) {
        try {
            const msg = JSON.parse(data);
            
            switch (msg.type) {
                case 'capabilities':
                    this._handleCapabilities(msg.data);
                    
                    // Resolve load() on first capabilities message
                    if (!this.isLoaded && resolveLoad) {
                        this.isLoaded = true;
                        resolveLoad(true);
                        return true;
                    }
                    break;
                    
                case 'wakeword':
                    this.onWakeword?.(msg.data?.model);
                    break;
                    
                case 'speech_start':
                    this.onSpeechStart?.();
                    break;
                    
                case 'speech_end':
                    this.onSpeechEnd?.();
                    break;
                    
                case 'error':
                    console.error('[VoiceEvents] Server error:', msg.data);
                    this.onError?.(msg.data);
                    break;
                    
                default:
                    console.log('[VoiceEvents] Unknown message type:', msg.type);
            }
        } catch (e) {
            console.error('[VoiceEvents] Failed to parse message:', e);
        }
        return false;
    }

    _handleCapabilities(capabilities) {
        this.capabilities = capabilities;
        
        // Extract wakeword info
        if (capabilities.wakewords) {
            this.wakewordModels = capabilities.wakewords.models || [];
            this.activeWakeword = capabilities.wakewords.active;
        }
        
        // Extract VAD info
        if (capabilities.vad) {
            this.vadEnabled = capabilities.vad.enabled;
            this.vadSilenceTimeout = capabilities.vad.silenceTimeout || 2000;
        }
        
        console.log('[VoiceEvents] Capabilities received:', {
            wakewords: this.wakewordModels.map(m => m.id),
            activeWakeword: this.activeWakeword,
            vadEnabled: this.vadEnabled,
            vadSilenceTimeout: this.vadSilenceTimeout
        });
        
        this.onCapabilities?.(capabilities);
    }

    _attemptReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('[VoiceEvents] Max reconnect attempts reached');
            return;
        }
        
        this.reconnectAttempts++;
        const delay = this.reconnectDelay * this.reconnectAttempts;
        
        console.log(`[VoiceEvents] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
        
        setTimeout(() => {
            if (!this.isConnected) {
                this.load().catch(e => {
                    console.error('[VoiceEvents] Reconnect failed:', e);
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
                model: this.activeWakeword
            }
        };
        
        this.ws.send(JSON.stringify(config));
        console.log('[VoiceEvents] Sent config:', config.data);
    }

    /**
     * Get available wake word models from server
     */
    getWakewordModels() {
        return this.wakewordModels;
    }

    /**
     * Get the currently active wake word model ID
     */
    getActiveWakeword() {
        return this.activeWakeword;
    }

    /**
     * Set the active wake word model
     */
    setWakewordModel(modelId) {
        if (!this.isConnected) return;
        
        const msg = {
            type: 'setModel',
            data: { model: modelId }
        };
        
        this.ws.send(JSON.stringify(msg));
        this.activeWakeword = modelId;
        console.log('[VoiceEvents] Set wakeword model:', modelId);
    }

    /**
     * Get the phrase for the active wake word model
     */
    getActivePhrase() {
        const model = this.wakewordModels.find(m => m.id === this.activeWakeword);
        return model?.phrase || this.activeWakeword;
    }

    /**
     * Check if VAD is enabled on the server
     */
    isVADEnabled() {
        return this.vadEnabled;
    }

    /**
     * Get the VAD silence timeout in milliseconds
     */
    getVADSilenceTimeout() {
        return this.vadSilenceTimeout;
    }

    /**
     * Process audio samples and send to server
     * @param {Float32Array} audioData - Audio samples
     * @param {number} inputSampleRate - Sample rate of the audio
     */
    async processAudio(audioData, inputSampleRate) {
        if (!this.isConnected) return;
        
        // Send sample rate config if changed
        if (inputSampleRate !== this.sampleRate) {
            this.sampleRate = inputSampleRate;
            this._sendConfig();
        }
        
        // Send audio as binary (Float32Array)
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(audioData.buffer);
        }
    }

    /**
     * Check if connected to server
     */
    isReady() {
        return this.isConnected && this.isLoaded;
    }

    /**
     * Stop and disconnect
     */
    stop() {
        this.isLoaded = false;
        
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.isConnected = false;
    }
}

// Keep backward compatibility alias
export { VoiceEventsClient as ServerWakeWordDetector };
