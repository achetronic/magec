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

import { createUserMessageHTML, createAIMessageHTML, createSessionItemHTML, escapeHtml, formatRelativeDate } from './templates.js';
import { t } from '../i18n/index.js';

// SVG icon paths
const ICONS = {
    alert: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
    info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    check: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    spinner: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15'
};

const icon = (path, color, extra = '') => `<svg class="w-4 h-4 text-${color} ${extra}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="${path}"/></svg>`;

// Notification type configurations
const NOTIFICATION_TYPES = {
    error:   { icon: icon(ICONS.alert, 'lava-400'),     bg: 'bg-lava-500/20' },
    warning: { icon: icon(ICONS.alert, 'sol-400'),      bg: 'bg-sol-500/20' },
    info:    { icon: icon(ICONS.info, 'atlantico-400'), bg: 'bg-atlantico-500/20' },
    loading: { icon: icon(ICONS.spinner, 'sol-400', 'animate-spin'), bg: 'bg-sol-500/20' },
    success: { icon: icon(ICONS.check, 'atlantico-400'), bg: 'bg-atlantico-500/20' }
};

// Status indicator configurations
const STATUS_COLORS = {
    listening: { ping: 'bg-atlantico-400', dot: 'bg-atlantico-500', anim: null },
    recording: { ping: 'bg-lava-400', dot: 'bg-lava-500', anim: 'animate-ping' },
    processing: { ping: 'bg-sol-400', dot: 'bg-sol-500', anim: 'animate-ping' },
    loading: { ping: 'bg-sol-400', dot: 'bg-sol-500', anim: 'animate-ping-slow' },
    default: { ping: 'bg-arena-400', dot: 'bg-arena-500', anim: null }
};

// Panel configurations
const PANELS = ['assistant', 'history', 'notifications', 'settings'];

export class UIController {
    constructor() {
        this.elements = this._getElements();
        this._centellaEnabled = false;
        this._notifications = [];
        this._notificationId = 0;
        this._loadingNotifications = {};
        this._startClock();
    }

    _getElements() {
        const $ = id => document.getElementById(id);
        return {
            // Status
            statusIndicator: $('statusIndicator'),
            statusText: $('statusText'),
            
            // Wake word
            wakeWordToggle: $('wakeWordToggle'),
            wakeWordText: $('wakeWordText'),
            
            // Conversation
            transcriptionContent: $('transcriptionContent'),
            transcriptionPlaceholder: $('transcriptionPlaceholder'),
            clearBtn: $('clearBtn'),
            copyBtn: $('copyBtn'),
            newSessionBtn: $('newSessionBtn'),
            newSessionBtnSidebar: $('newSessionBtnSidebar'),
            
            // Waveform & recording
            waveformCanvas: $('waveformCanvas'),
            centellaHint: $('centellaHint'),
            
            // Text input
            textInputForm: $('textInputForm'),
            textInput: $('textInput'),
            
            // Sidebar
            sidebar: $('sidebar'),
            sidebarOverlay: $('sidebarOverlay'),
            collapseSidebarBtn: $('collapseSidebarBtn'),
            menuBtn: $('menuBtn'),
            sessionList: $('sessionList'),
            
            // Navigation buttons
            centellaBtn: $('centellaBtn'),
            historyBtn: $('historyBtn'),
            notificationsBtn: $('notificationsBtn'),
            settingsBtn: $('settingsBtn'),
            
            // Panels
            panelAssistant: $('panelAssistant'),
            panelHistory: $('panelHistory'),
            panelNotifications: $('panelNotifications'),
            panelSettings: $('panelSettings'),
            
            // Back buttons
            backFromHistoryBtn: $('backFromHistoryBtn'),
            backFromNotificationsBtn: $('backFromNotificationsBtn'),
            backFromSettingsBtn: $('backFromSettingsBtn'),
            
            // Notifications
            notificationsBadge: $('notificationsBadge'),
            notificationsList: $('notificationsList'),
            notificationsPlaceholder: $('notificationsPlaceholder'),
            clearAllNotificationsBtn: $('clearAllNotificationsBtn'),
            
            // Clock
            footerClock: $('footerClock')
        };
    }

    // ==================== Clock ====================
    
    _startClock() {
        this._updateClock();
        setInterval(() => this._updateClock(), 1000);
    }
    
    _updateClock() {
        const now = new Date();
        const time = now.toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit' });
        const date = now.toLocaleDateString('es-ES', { weekday: 'long', day: 'numeric', month: 'long' });
        const formatted = `${time} · ${date.charAt(0).toUpperCase() + date.slice(1)}`;
        
        if (this.elements.footerClock) this.elements.footerClock.textContent = formatted;
    }

    // ==================== Status ====================
    
    setStatus(text, type = 'default') {
        this.elements.statusText.textContent = text;
        this._updateStatusIndicator(type);
    }

    _updateStatusIndicator(type) {
        const indicator = this.elements.statusIndicator;
        if (!indicator) return;
        
        const ping = indicator.querySelector('span:first-child');
        const dot = indicator.querySelector('.relative');
        
        const allColors = Object.values(STATUS_COLORS).flatMap(c => [c.ping, c.dot]);
        ping?.classList.remove(...allColors, 'animate-ping', 'animate-ping-slow');
        dot?.classList.remove(...allColors);
        
        const colors = STATUS_COLORS[type] || STATUS_COLORS.default;
        ping?.classList.add(colors.ping);
        if (colors.anim) ping?.classList.add(colors.anim);
        dot?.classList.add(colors.dot);
    }

    // ==================== Notifications ====================
    
    showError(msg) {
        this.addNotification('error', msg, { showConsoleHint: true });
    }

    addNotification(type, message, options = {}) {
        const { showConsoleHint = false } = options;
        const notification = {
            id: ++this._notificationId,
            type,
            message,
            timestamp: new Date(),
            showConsoleHint
        };
        this._notifications.unshift(notification);
        this._syncNotifications();
        return notification.id;
    }

    removeNotification(id) {
        this._notifications = this._notifications.filter(n => n.id !== id);
        this._syncNotifications();
    }

    clearAllNotifications() {
        this._notifications = [];
        this._syncNotifications();
    }

    _syncNotifications() {
        this._renderNotifications();
        this._updateNotificationBadge();
    }

    showLoadingNotification(key, message) {
        if (this._loadingNotifications[key]) {
            this.removeNotification(this._loadingNotifications[key]);
        }
        this._loadingNotifications[key] = this.addNotification('loading', message);
    }

    completeLoadingNotification(key, message) {
        this._clearLoadingNotification(key);
        this.addNotification('success', message);
    }

    failLoadingNotification(key, message) {
        this._clearLoadingNotification(key);
        this.addNotification('error', message, { showConsoleHint: true });
    }

    _clearLoadingNotification(key) {
        if (this._loadingNotifications[key]) {
            this.removeNotification(this._loadingNotifications[key]);
            delete this._loadingNotifications[key];
        }
    }

    _renderNotifications() {
        const { notificationsList: list, notificationsPlaceholder: placeholder } = this.elements;
        if (!list) return;

        list.querySelectorAll('.notification-item').forEach(el => el.remove());

        if (this._notifications.length === 0) {
            if (placeholder) placeholder.style.display = 'block';
            return;
        }

        if (placeholder) placeholder.style.display = 'none';
        
        for (const notification of this._notifications) {
            list.insertAdjacentHTML('beforeend', this._createNotificationHTML(notification));
        }

        list.querySelectorAll('[data-delete-notification]').forEach(btn => {
            btn.onclick = (e) => {
                e.stopPropagation();
                this.removeNotification(parseInt(btn.dataset.deleteNotification, 10));
            };
        });
    }

    _createNotificationHTML({ id, type, message, timestamp, showConsoleHint }) {
        const config = NOTIFICATION_TYPES[type] || NOTIFICATION_TYPES.info;
        
        const consoleHintHTML = showConsoleHint ? `
            <button onclick="console.info('Press F12 or Cmd+Option+J to open DevTools')" 
                class="text-xs text-arena-500 hover:text-arena-400 underline decoration-dotted underline-offset-2 mt-1 cursor-help"
                title="Press F12 or Cmd+Option+J">
                ${t('notifications.seeConsole')}
            </button>
        ` : '';
        
        return `
            <div class="notification-item group flex items-start gap-3 p-3 ${config.bg} rounded-xl transition-all hover:bg-opacity-30">
                <div class="flex-shrink-0 mt-0.5">${config.icon}</div>
                <div class="flex-1 min-w-0">
                    <p class="text-sm text-arena-200 break-words">${escapeHtml(message)}</p>
                    <div class="flex items-center gap-2">
                        <p class="text-xs text-arena-500 mt-1">${formatRelativeDate(timestamp)}</p>
                        ${consoleHintHTML}
                    </div>
                </div>
                <button data-delete-notification="${id}" 
                    class="flex-shrink-0 p-1 opacity-0 group-hover:opacity-100 hover:bg-piedra-700/50 rounded transition-all"
                    title="${t('notifications.delete')}">
                    <svg class="w-4 h-4 text-arena-500 hover:text-arena-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12"/>
                    </svg>
                </button>
            </div>
        `;
    }

    _updateNotificationBadge() {
        const badge = this.elements.notificationsBadge;
        if (!badge) return;
        
        const count = this._notifications.length;
        badge.classList.toggle('hidden', count === 0);
        badge.textContent = count > 99 ? '99+' : count;
    }

    // ==================== Recording ====================
    
    setRecordButtonEnabled(enabled) {
        this._centellaEnabled = enabled;
        const canvas = this.elements.waveformCanvas;
        if (canvas) {
            canvas.style.cursor = enabled ? 'pointer' : 'not-allowed';
            canvas.style.opacity = enabled ? '1' : '0.5';
        }
    }

    setRecordingState(isRecording) {
        const hint = this.elements.centellaHint;
        if (hint) {
            hint.style.opacity = isRecording ? '0' : '1';
        }
    }

    getCanvas() {
        return this.elements.waveformCanvas;
    }

    // ==================== Wake Word ====================
    
    setWakeWordEnabled(enabled, phrase = 'Hey Buddy') {
        const text = this.elements.wakeWordText;
        if (text) {
            text.textContent = enabled 
                ? t('settings.wakeWord.description', { phrase }) 
                : t('settings.wakeWord.disabled');
            text.classList.toggle('text-arena-600', !enabled);
        }
    }

    setWakeWordToggle(enabled) {
        if (this.elements.wakeWordToggle) {
            this.elements.wakeWordToggle.checked = enabled;
        }
    }

    setWakeWordPhrase(phrase) {
        if (this.elements.wakeWordToggle?.checked) {
            this.elements.wakeWordText.textContent = t('settings.wakeWord.description', { phrase });
        }
    }

    // ==================== Conversation ====================
    
    appendTranscription(text) {
        this._appendMessage(createUserMessageHTML(text));
    }

    appendAIResponse(text) {
        this._appendMessage(createAIMessageHTML(text));
    }

    _appendMessage(html) {
        const content = this.elements.transcriptionContent;
        content.insertAdjacentHTML('beforeend', html);
        this.elements.transcriptionPlaceholder.style.display = 'none';
        this.elements.clearBtn.disabled = false;
        this.elements.copyBtn.disabled = false;
        this._pruneEntries();
        content.scrollTop = content.scrollHeight;
    }

    _pruneEntries(maxEntries = 100) {
        const entries = this.elements.transcriptionContent.querySelectorAll('.flex.gap-3');
        const toRemove = entries.length - maxEntries;
        for (let i = 0; i < toRemove; i++) {
            entries[i].remove();
        }
    }

    clearTranscription() {
        this.elements.transcriptionContent.querySelectorAll('.flex.gap-3').forEach(el => el.remove());
        this.elements.transcriptionPlaceholder.style.display = 'block';
        this.elements.clearBtn.disabled = true;
        this.elements.copyBtn.disabled = true;
    }

    copyTranscription() {
        const entries = this.elements.transcriptionContent.querySelectorAll('.flex.gap-3 p');
        const text = Array.from(entries).map(e => e.textContent).join('\n');
        navigator.clipboard.writeText(text);
    }

    loadConversation(messages) {
        this.clearTranscription();
        if (!messages?.length) return;
        
        for (const { role, text } of messages) {
            if (role === 'user') this.appendTranscription(text);
            else if (role === 'model') this.appendAIResponse(text);
        }
    }

    // ==================== Sessions ====================
    
    renderSessionList(sessions, currentSessionId, onSelect, onDelete) {
        const list = this.elements.sessionList;
        if (!list) return;

        if (!sessions?.length) {
            list.innerHTML = `<p class="text-arena-500 text-sm text-center py-4">${t('sessions.empty')}</p>`;
            return;
        }

        list.innerHTML = sessions
            .map(session => createSessionItemHTML(session, session.id === currentSessionId))
            .join('');

        list.querySelectorAll('[data-session-id]').forEach(btn => {
            btn.onclick = () => onSelect(btn.dataset.sessionId);
        });
        
        list.querySelectorAll('[data-delete-session]').forEach(btn => {
            btn.onclick = (e) => {
                e.stopPropagation();
                onDelete(btn.dataset.deleteSession);
            };
        });
    }

    // ==================== Sidebar ====================
    
    closeSidebar() {
        this.elements.sidebar?.classList.add('sidebar-closed');
        this.elements.sidebar?.classList.remove('sidebar-open');
        this.elements.sidebarOverlay?.classList.add('hidden');
    }

    toggleSidebar() {
        const { sidebar, sidebarOverlay } = this.elements;
        const isMobile = window.innerWidth < 1024;
        
        if (isMobile) {
            const isOpen = sidebar.classList.contains('sidebar-open');
            sidebar.classList.toggle('sidebar-open', !isOpen);
            sidebar.classList.toggle('sidebar-closed', isOpen);
            sidebarOverlay?.classList.toggle('hidden', isOpen);
        } else {
            sidebar.classList.toggle('sidebar-collapsed');
        }
    }

    setupSidebarToggle() {
        this.elements.menuBtn?.addEventListener('click', () => this.toggleSidebar());
        this.elements.collapseSidebarBtn?.addEventListener('click', () => this.toggleSidebar());
        this.elements.sidebarOverlay?.addEventListener('click', () => this.closeSidebar());
    }

    // ==================== Panel Navigation ====================
    
    setupPanelNavigation() {
        const { centellaBtn, historyBtn, notificationsBtn, settingsBtn,
                backFromHistoryBtn, backFromNotificationsBtn, backFromSettingsBtn,
                clearAllNotificationsBtn } = this.elements;
        
        centellaBtn?.addEventListener('click', () => this.switchPanel('assistant'));
        historyBtn?.addEventListener('click', () => this.switchPanel('history'));
        notificationsBtn?.addEventListener('click', () => this.switchPanel('notifications'));
        settingsBtn?.addEventListener('click', () => this.switchPanel('settings'));
        
        backFromHistoryBtn?.addEventListener('click', () => this.switchPanel('assistant'));
        backFromNotificationsBtn?.addEventListener('click', () => this.switchPanel('assistant'));
        backFromSettingsBtn?.addEventListener('click', () => this.switchPanel('assistant'));
        
        clearAllNotificationsBtn?.addEventListener('click', () => this.clearAllNotifications());
    }

    switchPanel(panel) {
        const btnMap = { assistant: 'centellaBtn', history: 'historyBtn', notifications: 'notificationsBtn', settings: 'settingsBtn' };
        for (const p of PANELS) {
            const isActive = p === panel;
            const panelKey = `panel${p.charAt(0).toUpperCase() + p.slice(1)}`;
            this.elements[panelKey]?.classList.toggle('hidden', !isActive);
            this.elements[panelKey]?.classList.toggle('flex', isActive);
            this.elements[btnMap[p]]?.classList.toggle('bg-piedra-800', isActive);
        }
    }

    // ==================== Event Handlers ====================
    
    onRecordClick(callback) {
        this.elements.waveformCanvas?.addEventListener('click', () => {
            if (this._centellaEnabled) callback();
        });
    }

    onClearClick(callback) {
        this.elements.clearBtn.onclick = callback;
    }

    onCopyClick(callback) {
        this.elements.copyBtn.onclick = callback;
    }

    onWakeWordToggle(callback) {
        this.elements.wakeWordToggle.onchange = (e) => callback(e.target.checked);
    }

    onNewSessionClick(callback) {
        this.elements.newSessionBtn?.addEventListener('click', callback);
        this.elements.newSessionBtnSidebar?.addEventListener('click', callback);
    }

    onTextSubmit(callback) {
        this.elements.textInputForm?.addEventListener('submit', (e) => {
            e.preventDefault();
            const text = this.elements.textInput?.value?.trim();
            if (text) {
                callback(text);
                this.elements.textInput.value = '';
            }
        });
    }

    // ==================== Settings ====================
    
    renderWakeWordModels(models, selectedModel, onChange) {
        const container = document.getElementById('wakeWordModelContainer');
        if (!container) return;
        
        container.innerHTML = models.map(model => `
            <label class="relative">
                <input type="radio" name="wakeWordModel" value="${model.id}" 
                       ${model.id === selectedModel ? 'checked' : ''} class="peer sr-only">
                <div class="p-3 rounded-lg border border-piedra-600/50 bg-piedra-700/30 cursor-pointer transition-all peer-checked:border-atlantico-500 peer-checked:bg-atlantico-500/10 hover:border-piedra-500">
                    <div class="text-sm font-medium text-arena-100 mb-0.5">${model.name}</div>
                    <div class="text-xs text-arena-500 italic">${model.phrase}</div>
                </div>
            </label>
        `).join('');
        
        this._setupRadioGroup('wakeWordModel', selectedModel, onChange);
    }
    
    setupSettingsUI(settings, onSTTChange, onTTSEnabledChange, onLanguageChange) {
        this._setupRadioGroup('sttMode', settings.sttMode, onSTTChange);
        this._setupRadioGroup('language', settings.language, onLanguageChange);
        
        // TTS toggle
        const ttsToggle = document.getElementById('ttsToggle');
        if (ttsToggle) {
            ttsToggle.checked = settings.ttsEnabled;
            ttsToggle.addEventListener('change', () => {
                onTTSEnabledChange(ttsToggle.checked);
            });
        }
    }

    /**
     * Disable TTS toggle when backend is unavailable.
     */
    setTTSAvailable(available) {
        const ttsToggle = document.getElementById('ttsToggle');
        if (ttsToggle) {
            ttsToggle.disabled = !available;
            if (!available) {
                ttsToggle.checked = false;
            }
        }
        
        const ttsLabel = document.querySelector('label[for="ttsToggle"]');
        if (ttsLabel) {
            ttsLabel.classList.toggle('opacity-50', !available);
            ttsLabel.classList.toggle('cursor-not-allowed', !available);
        }
    }

    _setupRadioGroup(name, initialValue, onChange) {
        document.querySelectorAll(`input[name="${name}"]`).forEach(radio => {
            radio.checked = radio.value === initialValue;
            radio.addEventListener('change', () => {
                if (radio.checked) onChange(radio.value);
            });
        });
    }
}
