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

import { CONFIG } from '../config.js';

export class SessionManager {
    constructor(config = {}) {
        this.storageKey = config.storageKey || CONFIG.session.storageKey;
        this.autoRotateMinutes = config.autoRotateMinutes || CONFIG.session.autoRotateMinutes;
        this.maxStoredSessions = config.maxStoredSessions || CONFIG.session.maxStoredSessions;
        
        this.currentSessionId = null;
        this.rotationTimer = null;
        this.onSessionChange = null;
    }

    init() {
        const stored = this._loadFromStorage();
        
        // Check if current session is still valid (not expired)
        if (stored.currentSessionId && stored.currentSessionCreatedAt) {
            const elapsed = Date.now() - stored.currentSessionCreatedAt;
            const maxAge = this.autoRotateMinutes * 60 * 1000;
            
            if (elapsed < maxAge) {
                this.currentSessionId = stored.currentSessionId;
                // Schedule rotation for remaining time
                const remaining = maxAge - elapsed;
                this._scheduleRotation(remaining);
                return this.currentSessionId;
            }
        }
        
        // Create new session
        return this.newSession();
    }

    newSession() {
        this.currentSessionId = this._generateSessionId();
        this._saveSession(this.currentSessionId);
        this._scheduleRotation();
        
        if (this.onSessionChange) {
            this.onSessionChange(this.currentSessionId);
        }
        
        return this.currentSessionId;
    }

    getCurrentSessionId() {
        return this.currentSessionId;
    }

    getSessionHistory() {
        const stored = this._loadFromStorage();
        return stored.sessions || [];
    }

    _generateSessionId() {
        const timestamp = Date.now().toString(36);
        const random = Math.random().toString(36).substring(2, 8);
        return `session_${timestamp}_${random}`;
    }

    _scheduleRotation(delayMs = null) {
        if (this.rotationTimer) {
            clearTimeout(this.rotationTimer);
        }
        
        const delay = delayMs || this.autoRotateMinutes * 60 * 1000;
        
        this.rotationTimer = setTimeout(() => {
            console.log('Auto-rotating session after', this.autoRotateMinutes, 'minutes');
            this.newSession();
        }, delay);
    }

    _saveSession(sessionId) {
        const stored = this._loadFromStorage();
        
        // Add to history
        stored.sessions = stored.sessions || [];
        stored.sessions.unshift({
            id: sessionId,
            createdAt: Date.now()
        });
        
        // Trim old sessions
        if (stored.sessions.length > this.maxStoredSessions) {
            stored.sessions = stored.sessions.slice(0, this.maxStoredSessions);
        }
        
        stored.currentSessionId = sessionId;
        stored.currentSessionCreatedAt = Date.now();
        
        localStorage.setItem(this.storageKey, JSON.stringify(stored));
    }

    _loadFromStorage() {
        try {
            const data = localStorage.getItem(this.storageKey);
            return data ? JSON.parse(data) : {};
        } catch (e) {
            console.error('Failed to load sessions from storage:', e);
            return {};
        }
    }

    destroy() {
        if (this.rotationTimer) {
            clearTimeout(this.rotationTimer);
            this.rotationTimer = null;
        }
    }
}
