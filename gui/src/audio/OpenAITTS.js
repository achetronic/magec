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
 * OpenAI-compatible TTS client.
 * Uses the server's /api/v1/tts/ endpoint which proxies to the configured TTS backend.
 * All configuration (model, voice, speed, etc.) is handled server-side.
 */
export class OpenAITTS {
    constructor() {
        this._audio = null;
        this._speaking = false;
        this._abortController = null;
        this._available = null; // null = unknown, true/false after check
    }

    /**
     * Check if TTS backend is available.
     * @returns {Promise<boolean>}
     */
    async checkAvailable() {
        try {
            const response = await fetch('/api/v1/tts/speech', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ input: 'test' }),
            });
            this._available = response.ok;
        } catch {
            this._available = false;
        }
        return this._available;
    }

    /**
     * Returns cached availability status.
     */
    isAvailable() {
        return this._available === true;
    }

    /**
     * Clean text for TTS (remove markdown formatting).
     */
    _cleanText(text) {
        return text
            .replace(/\*\*/g, '')
            .replace(/\*/g, '')
            .replace(/`/g, '')
            .replace(/#{1,6}\s/g, '')
            .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
            .replace(/\s+/g, ' ')
            .trim();
    }

    /**
     * Speak text using the TTS API.
     * @param {string} text - Text to speak
     * @returns {Promise<void>}
     */
    async speak(text) {
        const cleanedText = this._cleanText(text);
        if (!cleanedText) {
            return;
        }

        this.stop();
        this._abortController = new AbortController();

        try {
            const response = await fetch('/api/v1/tts/speech', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    input: cleanedText,
                }),
                signal: this._abortController.signal,
            });

            if (!response.ok) {
                const error = await response.text().catch(() => '');
                throw new Error(`TTS request failed: ${response.status} ${error}`);
            }

            const audioBlob = await response.blob();
            const audioUrl = URL.createObjectURL(audioBlob);
            
            await this._playAudio(audioUrl);
            
            URL.revokeObjectURL(audioUrl);
        } catch (e) {
            if (e.name === 'AbortError') {
                return;
            }
            console.error('[OpenAITTS] Error:', e);
            this._speaking = false;
            throw e;
        }
    }

    /**
     * Play audio from URL.
     */
    _playAudio(url) {
        return new Promise((resolve, reject) => {
            this._audio = new Audio(url);
            this._speaking = true;

            this._audio.onended = () => {
                this._speaking = false;
                this._audio = null;
                resolve();
            };

            this._audio.onerror = () => {
                this._speaking = false;
                this._audio = null;
                reject(new Error('Audio playback failed'));
            };

            this._audio.play().catch((e) => {
                this._speaking = false;
                this._audio = null;
                reject(e);
            });
        });
    }

    /**
     * Stop current speech.
     */
    stop() {
        if (this._abortController) {
            this._abortController.abort();
            this._abortController = null;
        }
        
        if (this._audio) {
            this._audio.pause();
            this._audio.currentTime = 0;
            this._audio = null;
        }
        this._speaking = false;
    }

    /**
     * Check if currently speaking.
     */
    isSpeaking() {
        return this._speaking;
    }

    /**
     * TTS is always supported (server-side).
     */
    isSupported() {
        return true;
    }

    /**
     * TTS is always ready (no client-side initialization needed).
     */
    isReady() {
        return true;
    }
}
