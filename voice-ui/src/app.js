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

import { CONFIG } from './config.js';
import { AudioCapture, AudioRecorder, VoiceEventsClient, FeedbackSound, OpenAITTS } from './audio/index.js';
import { RemoteTranscriber } from './transcription/index.js';
import { UIController, WaveformRenderer } from './ui/index.js';
import { SessionManager, SessionService } from './session/index.js';
import { AgentClient, clientAuth } from './api/index.js';
import { SettingsManager } from './settings/index.js';
import { errorHandler } from './errors/index.js';
import { initLanguage, setLanguage, getLanguage, t, onLanguageChange } from './i18n/index.js';
import { WakeLock } from './utils/index.js';

class MagecApp {
    constructor() {
        // Core services
        this.ui = new UIController();
        this.waveform = new WaveformRenderer(this.ui.getCanvas());
        this.sessionManager = new SessionManager();
        this.sessionService = new SessionService();
        this.agentClient = new AgentClient();
        this.settings = new SettingsManager();
        
        // Audio components (lazy initialized)
        this.audioCapture = null;
        this.audioRecorder = null;
        this.voiceEvents = null;
        this.transcriber = null;
        
        // Audio utilities
        this.feedbackSound = new FeedbackSound({ volume: 0.3 });
        this.tts = new OpenAITTS();
        this.wakeLock = new WakeLock();
        
        // State
        this.isRecording = false;
        this.wakeWordPhrase = '';
        this.wakeWordEnabled = true;
        this.vadEnabled = false;
    }

    // ==================== Initialization ====================

    async init() {
        initLanguage();

        const needsPairing = await this._checkClientPairing();
        if (needsPairing) {
            this._showPairingScreen();
            return;
        }

        this._startApp();
    }

    async _checkClientPairing() {
        const paired = await clientAuth.checkPairing();
        if (paired) return false;
        if (!clientAuth.token) {
            const infoRes = await fetch('/api/v1/client/info');
            if (infoRes.ok) {
                const info = await infoRes.json();
                if (!info.paired && info.paired !== undefined) {
                    return true;
                }
            }
            return false;
        }
        return true;
    }

    _showPairingScreen() {
        const screen = document.getElementById('pairingScreen');
        const mainApp = document.getElementById('mainApp');
        const input = document.getElementById('pairingTokenInput');
        const btn = document.getElementById('pairingConnectBtn');
        const error = document.getElementById('pairingError');

        screen.classList.remove('hidden');
        mainApp.classList.add('hidden');

        input.addEventListener('input', () => {
            btn.disabled = !input.value.trim();
            error.classList.add('hidden');
        });

        input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !btn.disabled) btn.click();
        });

        btn.addEventListener('click', async () => {
            btn.disabled = true;
            btn.textContent = 'Conectando...';
            error.classList.add('hidden');

            const ok = await clientAuth.pair(input.value.trim());
            if (ok) {
                screen.classList.add('hidden');
                mainApp.classList.remove('hidden');
                this._startApp();
            } else {
                error.classList.remove('hidden');
                btn.disabled = false;
                btn.textContent = 'Emparejar';
                input.value = '';
                input.focus();
            }
        });

        input.focus();
    }

    async _startApp() {
        this.wakeLock.enable();

        errorHandler.setNotificationCallback((type, message) => {
            this.ui.addNotification(type, message, { showConsoleHint: true });
        });

        this._selectedAgent = clientAuth.defaultAgent || null;
        this.settings = new SettingsManager(this._selectedAgent);

        if (this._selectedAgent) {
            this.agentClient.setAgent(this._selectedAgent);
            this.sessionService.setAgent(this._selectedAgent);
            this.tts.setAgent(this._selectedAgent);
        }

        this._setupEventHandlers();
        this._initSettings();
        this._setupAgentSwitcher();
        await this._checkTTSAvailability();
        await this._initWakeWord();
        await this._initSession();
        this._setReady();
        await this._startListening();
    }

    _setupAgentSwitcher() {
        const agents = clientAuth.allowedAgents;
        const switcher = document.getElementById('agentSwitcher');
        const btn = document.getElementById('agentSwitcherBtn');
        const dropdown = document.getElementById('agentDropdown');
        const list = document.getElementById('agentDropdownList');
        if (!switcher || !btn || !dropdown || !list) return;

        if (!agents || agents.length <= 1) {
            switcher.classList.add('hidden');
            return;
        }

        switcher.classList.remove('hidden');
        this._selectedAgent = clientAuth.defaultAgent;

        const renderList = () => {
            list.innerHTML = agents.map(a => {
                const active = a.id === this._selectedAgent;
                return `
                <button data-agent-id="${a.id}" class="w-full flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-left transition-colors ${active ? 'bg-sol-500/10' : 'hover:bg-piedra-800'}">
                    <div class="w-6 h-6 rounded-md ${active ? 'bg-sol-500/20' : 'bg-piedra-800'} flex items-center justify-center flex-shrink-0">
                        <span class="text-[10px] font-bold ${active ? 'text-sol-400' : 'text-arena-500'}">${(a.name || a.id).charAt(0).toUpperCase()}</span>
                    </div>
                    <div class="min-w-0 flex-1">
                        <p class="text-xs ${active ? 'text-arena-100 font-medium' : 'text-arena-300'} truncate">${a.name || a.id}</p>
                    </div>
                    ${active ? '<svg class="w-3.5 h-3.5 text-sol-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>' : ''}
                </button>`;
            }).join('');

            list.querySelectorAll('[data-agent-id]').forEach(el => {
                el.addEventListener('click', () => {
                    this._selectedAgent = el.dataset.agentId;
                    this.agentClient.setAgent(this._selectedAgent);
                    this.sessionService.setAgent(this._selectedAgent);
                    this.tts.setAgent(this._selectedAgent);
                    if (this.transcriber) this.transcriber.setAgent(this._selectedAgent);
                    this.settings.switchAgent(this._selectedAgent);
                    this._applySettingsToUI();
                    dropdown.classList.add('hidden');
                    renderList();
                    this.sessionManager.newSession();
                    console.log('[Client] Agent switched to:', this._selectedAgent);
                });
            });
        };

        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const isOpen = !dropdown.classList.contains('hidden');
            dropdown.classList.toggle('hidden', isOpen);
            if (!isOpen) renderList();
        });

        document.addEventListener('click', (e) => {
            if (!switcher.contains(e.target)) {
                dropdown.classList.add('hidden');
            }
        });

        renderList();
    }

    _setupEventHandlers() {
        this.ui.onRecordClick(() => this._toggleRecording());
        this.ui.onClearClick(() => this.ui.clearTranscription());
        this.ui.onCopyClick(() => this.ui.copyTranscription());
        this.ui.onWakeWordToggle((enabled) => this._setWakeWordEnabled(enabled));
        this.ui.onNewSessionClick(() => this.sessionManager.newSession());
        this.ui.onTextSubmit((text) => this._handleTextInput(text));
        this.ui.setupSidebarToggle();
        this.ui.setupPanelNavigation();
        
        // Language change listener
        onLanguageChange(() => this._onLanguageChange());
    }

    _initSettings() {
        const { ttsEnabled } = this.settings;
        
        this.ui.setupSettingsUI(
            { ttsEnabled, language: getLanguage() },
            (enabled) => this._onTTSEnabledChange(enabled),
            (lang) => setLanguage(lang)
        );
    }

    _applySettingsToUI() {
        const ttsToggle = document.getElementById('ttsToggle');
        if (ttsToggle) ttsToggle.checked = this.settings.ttsEnabled;

        this.wakeWordEnabled = this.settings.wakeWordEnabled;
        this.ui.setWakeWordToggle(this.wakeWordEnabled);
        this.ui.setWakeWordEnabled(this.wakeWordEnabled, this.wakeWordPhrase);
    }

    _onLanguageChange() {
        // Update dynamic UI elements that aren't covered by data-i18n
        this.ui.setWakeWordEnabled(
            this.wakeWordEnabled,
            this.wakeWordPhrase
        );
    }

    async _checkTTSAvailability() {
        const available = await this.tts.checkAvailable();
        this.ui.setTTSAvailable(available);
        if (!available) {
            this.settings.ttsEnabled = false;
            this.ui.addNotification('warning', t('notifications.ttsUnavailable'));
        }
    }

    _setReady() {
        this.ui.setRecordButtonEnabled(true);
        this.ui.setStatus(t('status.ready'), 'listening');
    }

    // ==================== Voice Events (Wake Word + VAD) ====================

    // Wake word models config - received from server
    wakeWordModels = [];

    async _initWakeWord() {
        await this._initVoiceEvents();
    }

    async _initVoiceEvents() {
        this.ui.setStatus(t('status.loadingWakeWord'), 'loading');
        this.ui.showLoadingNotification('wakeword', t('notifications.wakeWordLoading'));
        
        try {
            const voiceEvents = new VoiceEventsClient();
            
            // Set up event handlers
            voiceEvents.onWakeword = (model) => this._onWakeWordDetected(model);
            voiceEvents.onSpeechStart = () => this._onSpeechStart();
            voiceEvents.onSpeechEnd = () => this._onSpeechEnd();
            voiceEvents.onCapabilities = (caps) => this._onCapabilitiesReceived(caps);
            
            await voiceEvents.load();
            this.voiceEvents = voiceEvents;
            
            // Store VAD state
            this.vadEnabled = voiceEvents.isVADEnabled();
            console.log('[VoiceEvents] Connected, VAD enabled:', this.vadEnabled);
            
            // Get phrase from active model
            this.wakeWordPhrase = voiceEvents.getActivePhrase();
            console.log('[VoiceEvents] Using server-side detection, phrase:', this.wakeWordPhrase);
            
            // Apply saved wake word setting
            this.wakeWordEnabled = this.settings.wakeWordEnabled;
            this.ui.setWakeWordEnabled(this.wakeWordEnabled, this.wakeWordPhrase);
            this.ui.setWakeWordToggle(this.wakeWordEnabled);
            
            this.ui.completeLoadingNotification('wakeword', t('notifications.wakeWordReady'));
        } catch (e) {
            // Server not available - disable voice events entirely
            console.warn('[VoiceEvents] Server not available:', e.message);
            this.voiceEvents = null;
            this.wakeWordModels = [];
            this.vadEnabled = false;
            this.wakeWordEnabled = false;
            this.ui.setWakeWordEnabled(false, '');
            this.ui.setWakeWordToggle(false);
            this.ui.disableWakeWordToggle();
            this.ui.hideWakeWordModelSelector();
            this.ui.failLoadingNotification('wakeword', t('notifications.wakeWordUnavailable'));
        }
    }

    _onCapabilitiesReceived(caps) {
        // Update wake word models
        if (caps.wakewords) {
            this.wakeWordModels = caps.wakewords.models || [];
            this.settings.setValidWakeWordModels(this.wakeWordModels.map(m => m.id));
            
            // Update phrase from active model
            const activeModel = caps.wakewords.active;
            const activeConfig = this.wakeWordModels.find(m => m.id === activeModel);
            if (activeConfig) {
                this.wakeWordPhrase = activeConfig.phrase || activeConfig.id;
            }
            
            // Render wake word model selector
            this.ui.renderWakeWordModels(
                this.wakeWordModels,
                activeModel,
                (model) => this._onWakeWordModelChange(model)
            );
        }
        
        // Update VAD state
        if (caps.vad) {
            this.vadEnabled = caps.vad.enabled;
            console.log('[VoiceEvents] VAD enabled:', this.vadEnabled, 'timeout:', caps.vad.silenceTimeout);
        }
    }

    _onWakeWordDetected(model) {
        if (!this.wakeWordEnabled) return;
        this._startRecording();
    }

    _onSpeechStart() {
        // Speech detected - could be used for UI feedback
        console.log('[VoiceEvents] Speech detected');
    }

    _onSpeechEnd() {
        // Speech ended - stop recording if we're recording and VAD is enabled
        if (this.isRecording && this.vadEnabled) {
            console.log('[VoiceEvents] Speech ended, stopping recording');
            this._stopRecording();
        }
    }

    _setWakeWordEnabled(enabled) {
        this.wakeWordEnabled = enabled;
        this.settings.wakeWordEnabled = enabled;
        this.ui.setWakeWordEnabled(enabled, this.wakeWordPhrase);
    }

    async _onWakeWordModelChange(modelId) {
        if (!this.voiceEvents) return;
        
        // Tell server to change model
        this.voiceEvents.setWakewordModel(modelId);
        this.settings.wakeWordModel = modelId;
        
        // Update phrase
        const modelConfig = this.wakeWordModels.find(m => m.id === modelId);
        this.wakeWordPhrase = modelConfig?.phrase || modelId;
        this.ui.setWakeWordEnabled(this.wakeWordEnabled, this.wakeWordPhrase);
    }

    // ==================== Session Management ====================

    async _initSession() {
        this.sessionManager.onSessionChange = (sessionId) => this._onSessionChange(sessionId);
        const sessionId = this.sessionManager.init();
        await this.agentClient.createSession(sessionId);
        await this._refreshSessionList();
    }

    _onSessionChange(sessionId) {
        this.agentClient.createSession(sessionId);
        this.ui.clearTranscription();
        this._refreshSessionList();
    }

    _parseSessionTimestamp(id) {
        const parts = (id || '').split('_');
        if (parts.length >= 2) {
            const ts = parseInt(parts[1], 36);
            if (!isNaN(ts) && ts > 0) return ts;
        }
        return 0;
    }

    async _refreshSessionList() {
        const sessions = await this.sessionService.listSessions();
        const currentSessionId = this.sessionManager.getCurrentSessionId();
        const history = this.sessionManager.getSessionHistory();
        
        const enrichedSessions = await Promise.all(
            sessions.map(async (session) => ({
                id: session.id,
                preview: this.sessionService.getSessionPreview(
                    await this.sessionService.getSession(session.id)
                ),
                createdAt: this._parseSessionTimestamp(session.id)
                    || history.find(s => s.id === session.id)?.createdAt
                    || 0
            }))
        );
        
        enrichedSessions.sort((a, b) => b.createdAt - a.createdAt);
        
        this.ui.renderSessionList(
            enrichedSessions,
            currentSessionId,
            (id) => this._selectSession(id),
            (id) => this._deleteSession(id)
        );
    }

    async _selectSession(sessionId) {
        if (sessionId === this.sessionManager.getCurrentSessionId()) {
            this.ui.closeSidebar();
            return;
        }
        
        const session = await this.sessionService.getSession(sessionId);
        if (!session) {
            // Error already handled by SessionService if it was a server error
            return;
        }
        
        this.sessionManager.currentSessionId = sessionId;
        this.ui.loadConversation(this.sessionService.extractMessages(session));
        await this._refreshSessionList();
        this.ui.closeSidebar();
    }

    async _deleteSession(sessionId) {
        if (sessionId === this.sessionManager.getCurrentSessionId()) {
            this.sessionManager.newSession();
        }
        await this.sessionService.deleteSession(sessionId);
        await this._refreshSessionList();
    }

    // ==================== Audio Recording ====================

    async _startListening() {
        try {
            this.audioCapture = new AudioCapture();
            await this.audioCapture.start();
            
            this.waveform.setAnalyser(this.audioCapture.getAnalyser());
            this.waveform.start();
            
            // Connect audio to voice events server
            if (this.voiceEvents) {
                this.audioCapture.onAudioData = (samples, sampleRate) => {
                    this.voiceEvents.processAudio(samples, sampleRate);
                };
            }
        } catch (e) {
            console.error('[Microphone]', e);
            this.ui.showError(t('errors.microphoneAccess'));
        }
    }

    _toggleRecording() {
        this.isRecording ? this._stopRecording() : this._startRecording();
    }

    _startRecording() {
        if (this.isRecording) return;
        
        this.isRecording = true;
        this.tts.stop();
        this.feedbackSound.playWakeChime();
        // Temporarily disable wake word detection while recording
        const prevWakeWordEnabled = this.wakeWordEnabled;
        this.wakeWordEnabled = false;
        this.ui.setStatus(t('status.recording'), 'recording');
        this.ui.setRecordingState(true);
        this.waveform.setRecording(true);
        
        this.audioRecorder = new AudioRecorder(this.audioCapture.getMicStream());
        this.audioRecorder.onRecordingComplete = (blob) => {
            // Restore wake word enabled state
            this.wakeWordEnabled = prevWakeWordEnabled;
            this._processRecording(blob);
        };
        this.audioRecorder.start();
        
        // If VAD is not enabled, use a fallback timeout
        if (!this.vadEnabled) {
            // Fallback: stop after 10 seconds max
            this._recordingTimeout = setTimeout(() => {
                if (this.isRecording) {
                    console.log('[Recording] Fallback timeout, stopping');
                    this._stopRecording();
                }
            }, 10000);
        }
    }

    _stopRecording() {
        if (!this.isRecording) return;
        
        this.isRecording = false;
        this.feedbackSound.playStopChime();
        this.ui.setStatus(t('status.processing'), 'processing');
        this.ui.setRecordingState(false);
        this.waveform.setRecording(false);
        
        // Clear fallback timeout
        if (this._recordingTimeout) {
            clearTimeout(this._recordingTimeout);
            this._recordingTimeout = null;
        }
        
        this.audioRecorder?.stop();
    }

    async _processRecording(blob) {
        this.ui.setRecordButtonEnabled(false);

        try {
            const text = await this._transcribe(blob);
            if (text) {
                this.ui.appendTranscription(text);
                await this._sendToAgent(text);
            }
        } catch (e) {
            console.error('[Transcription]', e);
        } finally {
            this.ui.setRecordButtonEnabled(true);
            this._setReady();
        }
    }

    // ==================== Transcription ====================

    async _transcribe(blob) {
        if (!this.transcriber) {
            this.transcriber = new RemoteTranscriber(CONFIG.transcription);
            if (this._selectedAgent) this.transcriber.setAgent(this._selectedAgent);
        }
        return this.transcriber.transcribe(blob);
    }

    // ==================== Agent Communication ====================

    async _handleTextInput(text) {
        this.ui.appendTranscription(text);
        await this._sendToAgent(text);
    }

    async _sendToAgent(message) {
        this.tts.stop();
        this.ui.setStatus(t('status.thinking'), 'processing');
        
        let responses = [];
        try {
            const sessionId = this.sessionManager.getCurrentSessionId();
            responses = await this.agentClient.sendMessage(sessionId, message);
        } catch (e) {
            console.error('[Agent]', e);
            this.ui.appendAIResponse(t('errors.generic'));
            this.ui.setStatus(t('status.ready'), 'listening');
            return;
        }
        
        this.ui.setStatus(t('status.ready'), 'listening');
        this._refreshSessionList();
        
        for (const response of responses) {
            this.ui.appendAIResponse(response);
            try {
                await this._speak(response);
            } catch (e) {
                console.warn('[TTS]', e);
                this.tts.stop();
            }
        }
    }

    async _speak(text) {
        if (!this.settings.ttsEnabled) {
            return;
        }
        await this.tts.speak(text);
    }

    // ==================== Settings ====================

    _onTTSEnabledChange(enabled) {
        this.settings.ttsEnabled = enabled;
    }
}

// Bootstrap
const app = new MagecApp();
app.init();
