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

// Mantiene la pantalla encendida usando la Screen Wake Lock API
export class WakeLock {
    constructor() {
        this._wakeLock = null;
        this._enabled = false;
    }

    // Comprueba si el navegador soporta Wake Lock
    static isSupported() {
        return 'wakeLock' in navigator;
    }

    // Solicita mantener la pantalla encendida
    async enable() {
        if (!WakeLock.isSupported()) {
            console.warn('[WakeLock] Not supported in this browser');
            return false;
        }

        try {
            this._wakeLock = await navigator.wakeLock.request('screen');
            this._enabled = true;
            
            // Re-adquirir si la página vuelve a ser visible
            this._wakeLock.addEventListener('release', () => {
                this._enabled = false;
            });
            
            document.addEventListener('visibilitychange', () => this._onVisibilityChange());
            
            console.log('[WakeLock] Screen wake lock acquired');
            return true;
        } catch (e) {
            console.warn('[WakeLock] Failed to acquire:', e.message);
            return false;
        }
    }

    // Libera el wake lock
    async disable() {
        if (this._wakeLock) {
            await this._wakeLock.release();
            this._wakeLock = null;
            this._enabled = false;
            console.log('[WakeLock] Screen wake lock released');
        }
    }

    // Re-adquiere el wake lock cuando la página vuelve a ser visible
    async _onVisibilityChange() {
        if (document.visibilityState === 'visible' && this._enabled === false && this._wakeLock === null) {
            await this.enable();
        }
    }

    isEnabled() {
        return this._enabled;
    }
}
