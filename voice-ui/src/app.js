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
import { AudioCapture, AudioRecorder, ServerWakeWordDetector, FeedbackSound, OpenAITTS } from './audio/index.js';
import { RemoteTranscriber } from './transcription/index.js';
import { UIController, WaveformRenderer } from './ui/index.js';
import { SessionManager, SessionService } from './session/index.js';
import { AgentClient } from './api/index.js';
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
        this.wakeWordDetector = null;
        this.transcriber = null;
        
        // Audio utilities
        this.feedbackSound = new FeedbackSound({ volume: 0.3 });
        this.tts = new OpenAITTS();
        this.wakeLock = new WakeLock();
        
        // State
        this.isRecording = false;
        this.wakeWordPhrase = '';
    }

    // ==================== Initialization ====================

    async init() {
        // Initialize i18n
        initLanguage();
        
        // Keep screen awake on mobile
        this.wakeLock.enable();
        
        // Connect error handler to UI notifications
        errorHandler.setNotificationCallback((type, message) => {
            this.ui.addNotification(type, message, { showConsoleHint: true });
        });
        
        this._setupEventHandlers();
        this._initSettings();
        await this._checkTTSAvailability();
        await this._initWakeWord();
        await this._initSession();
        this._setReady();
        await this._startListening();
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

    _onLanguageChange() {
        // Update dynamic UI elements that aren't covered by data-i18n
        this.ui.setWakeWordEnabled(
            this.wakeWordDetector?.isEnabled() ?? true,
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

    // ==================== Wake Word ====================

    // Wake word models config - received from server
    wakeWordModels = [];

    async _initWakeWord() {
        await this._loadWakeWordModel();
    }

    async _loadWakeWordModel() {
        this.ui.setStatus(t('status.loadingWakeWord'), 'loading');
        this.ui.showLoadingNotification('wakeword', t('notifications.wakeWordLoading'));
        
        try {
            // Use server-side wake word detection only
            const serverDetector = new ServerWakeWordDetector(
                {},
                () => this._onWakeWordDetected()
            );
            
            // Set up callback to receive models from server
            serverDetector.onModelsReceived = (models, activeModel) => {
                this.wakeWordModels = models;
                this.settings.setValidWakeWordModels(models.map(m => m.id));
                
                // Update phrase from active model
                const activeConfig = models.find(m => m.id === activeModel);
                if (activeConfig) {
                    this.wakeWordPhrase = activeConfig.phrase || activeConfig.name;
                }
                
                // Render wake word model selector
                this.ui.renderWakeWordModels(
                    models,
                    activeModel,
                    (model) => this._onWakeWordModelChange(model)
                );
            };
            
            await serverDetector.load();
            this.wakeWordDetector = serverDetector;
            
            // Get phrase from active model
            this.wakeWordPhrase = serverDetector.getActivePhrase();
            console.log('[WakeWord] Using server-side detection, phrase:', this.wakeWordPhrase);
            
            // Apply saved wake word setting
            const wakeWordEnabled = this.settings.wakeWordEnabled;
            this.wakeWordDetector.setEnabled(wakeWordEnabled);
            this.ui.setWakeWordEnabled(wakeWordEnabled, this.wakeWordPhrase);
            this.ui.setWakeWordToggle(wakeWordEnabled);
            
            this.ui.completeLoadingNotification('wakeword', t('notifications.wakeWordReady'));
        } catch (e) {
            // Server wake word not available - disable wake word entirely
            console.warn('[WakeWord] Server not available, wake word disabled:', e.message);
            this.wakeWordDetector = null;
            this.wakeWordModels = [];
            this.ui.setWakeWordEnabled(false, '');
            this.ui.setWakeWordToggle(false);
            this.ui.disableWakeWordToggle();
            this.ui.hideWakeWordModelSelector();
            this.ui.failLoadingNotification('wakeword', t('notifications.wakeWordUnavailable'));
        }
    }

    _onWakeWordDetected() {
        this._startRecording();
    }

    _setWakeWordEnabled(enabled) {
        this.wakeWordDetector?.setEnabled(enabled);
        this.settings.wakeWordEnabled = enabled;
        this.ui.setWakeWordEnabled(enabled, this.wakeWordPhrase);
    }

    async _onWakeWordModelChange(modelId) {
        if (!this.wakeWordDetector) return;
        
        // Tell server to change model
        this.wakeWordDetector.setModel(modelId);
        this.settings.wakeWordModel = modelId;
        
        // Update phrase
        const modelConfig = this.wakeWordModels.find(m => m.id === modelId);
        this.wakeWordPhrase = modelConfig?.phrase || modelId;
        this.ui.setWakeWordEnabled(this.settings.wakeWordEnabled, this.wakeWordPhrase);
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

    async _refreshSessionList() {
        const sessions = await this.sessionService.listSessions();
        const currentSessionId = this.sessionManager.getCurrentSessionId();
        
        const enrichedSessions = await Promise.all(
            sessions.map(async (session) => ({
                id: session.id,
                preview: this.sessionService.getSessionPreview(
                    await this.sessionService.getSession(session.id)
                ),
                createdAt: this.sessionManager.getSessionHistory()
                    .find(s => s.id === session.id)?.createdAt || Date.now()
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
            
            // Connect audio to server wake word detector if enabled
            if (this.wakeWordDetector instanceof ServerWakeWordDetector) {
                this.audioCapture.onAudioData = (samples, sampleRate) => {
                    this.wakeWordDetector.processAudio(samples, sampleRate);
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
        this.wakeWordDetector?.setEnabled(false);
        this.ui.setStatus(t('status.recording'), 'recording');
        this.ui.setRecordingState(true);
        this.waveform.setRecording(true);
        
        this.audioRecorder = new AudioRecorder(this.audioCapture.getMicStream());
        this.audioRecorder.onRecordingComplete = (blob) => this._processRecording(blob);
        this.audioRecorder.start();
        
        this.wakeWordDetector.onSilence = () => this._stopRecording();
        this.wakeWordDetector.startSilenceDetection(2000);
    }

    _stopRecording() {
        if (!this.isRecording) return;
        
        this.isRecording = false;
        this.feedbackSound.playStopChime();
        this.ui.setStatus(t('status.processing'), 'processing');
        this.ui.setRecordingState(false);
        this.waveform.setRecording(false);
        this.wakeWordDetector?.stopSilenceDetection();
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
            this.wakeWordDetector?.setEnabled(this.settings.wakeWordEnabled);
            this.ui.setRecordButtonEnabled(true);
            this._setReady();
        }
    }

    // ==================== Transcription ====================

    async _transcribe(blob) {
        this.transcriber ??= new RemoteTranscriber(CONFIG.transcription);
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
