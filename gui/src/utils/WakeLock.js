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
