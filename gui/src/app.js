import { CONFIG } from './config.js';
import { AudioCapture, AudioRecorder, OpenWakeWordDetector, FeedbackSound, OpenAITTS } from './audio/index.js';
import { LocalTranscriber, RemoteTranscriber } from './transcription/index.js';
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
        this.localTranscriber = null;
        this.remoteTranscriber = null;
        
        // Audio utilities
        this.feedbackSound = new FeedbackSound({ volume: 0.3 });
        this.tts = new OpenAITTS();
        this.wakeLock = new WakeLock();
        
        // State
        this.isRecording = false;
        this.wakeWordPhrase = CONFIG.wakeWord.defaultPhrase;
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
        const { sttMode, ttsEnabled } = this.settings;
        
        this.ui.setupSettingsUI(
            { sttMode, ttsEnabled, language: getLanguage() },
            (mode) => this._onSTTModeChange(mode),
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

    // Wake word models config - loaded from /models/wakewords.json
    wakeWordModels = [];

    async _initWakeWord() {
        await this._loadWakeWordConfig();
        await this._loadWakeWordModel();
    }

    async _loadWakeWordConfig() {
        try {
            const response = await fetch('/models/wakewords.json');
            if (response.ok) {
                const config = await response.json();
                this.wakeWordModels = config.models || [];
                this.settings.setValidWakeWordModels(this.wakeWordModels.map(m => m.id));
                
                // Render wake word model selector
                this.ui.renderWakeWordModels(
                    this.wakeWordModels,
                    this.settings.wakeWordModel,
                    (model) => this._onWakeWordModelChange(model)
                );
            }
        } catch (e) {
            console.warn('[WakeWord] Failed to load wakewords.json, using defaults');
        }
    }

    _getWakeWordModelConfig(modelId) {
        return this.wakeWordModels.find(m => m.id === modelId) || {
            id: modelId,
            name: modelId,
            file: `${modelId}.onnx`,
            threshold: 0.5,
            phrase: modelId
        };
    }

    async _loadWakeWordModel() {
        const modelId = this.settings.wakeWordModel;
        const modelConfig = this._getWakeWordModelConfig(modelId);
        
        const wakeWordModelPath = `/models/${modelConfig.file}`;
        this.wakeWordPhrase = modelConfig.phrase || modelConfig.name;
        
        this.ui.setStatus(t('status.loadingWakeWord'), 'loading');
        this.ui.showLoadingNotification('wakeword', t('notifications.wakeWordLoading'));
        
        try {
            this.wakeWordDetector = new OpenWakeWordDetector(
                { ...CONFIG.wakeWord, modelPath: wakeWordModelPath, threshold: modelConfig.threshold },
                () => this._onWakeWordDetected()
            );
            
            await this.wakeWordDetector.load();
            
            // Apply saved wake word setting
            const wakeWordEnabled = this.settings.wakeWordEnabled;
            this.wakeWordDetector.setEnabled(wakeWordEnabled);
            this.ui.setWakeWordEnabled(wakeWordEnabled, this.wakeWordPhrase);
            this.ui.setWakeWordToggle(wakeWordEnabled);
            
            this.ui.completeLoadingNotification('wakeword', t('notifications.wakeWordReady'));
        } catch (e) {
            console.error('[WakeWord]', e);
            this.ui.failLoadingNotification('wakeword', t('notifications.wakeWordFailed'));
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

    async _onWakeWordModelChange(model) {
        this.settings.wakeWordModel = model;
        
        // Stop current detector if running
        if (this.wakeWordDetector) {
            this.wakeWordDetector.stop();
            this.wakeWordDetector = null;
        }
        
        // Reload with new model
        await this._loadWakeWordModel();
        
        // Reconnect audio callback to new detector
        if (this.audioCapture && this.wakeWordDetector instanceof OpenWakeWordDetector) {
            this.audioCapture.onAudioData = (samples, sampleRate) => {
                this.wakeWordDetector.processAudio(samples, sampleRate);
            };
        }
        
        this._setReady();
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
            this.audioCapture = new AudioCapture(CONFIG.wakeWord);
            await this.audioCapture.start();
            
            this.waveform.setAnalyser(this.audioCapture.getAnalyser());
            this.waveform.start();
            
            // Connect audio to OpenWakeWord detector if applicable
            if (this.wakeWordDetector instanceof OpenWakeWordDetector) {
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
        return this.settings.sttMode === 'server' 
            ? this._transcribeRemote(blob) 
            : this._transcribeLocal(blob);
    }

    async _transcribeRemote(blob) {
        this.remoteTranscriber ??= new RemoteTranscriber(CONFIG.remote);
        return this.remoteTranscriber.transcribe(blob);
    }

    async _transcribeLocal(blob) {
        this.localTranscriber ??= new LocalTranscriber(CONFIG.whisper);
        
        if (!this.localTranscriber.isLoaded) {
            this.ui.setStatus(t('status.loadingWhisper'), 'loading');
            this.ui.showLoadingNotification('whisper', t('notifications.whisperLoading'));
            
            try {
                await this.localTranscriber.load();
                this.ui.completeLoadingNotification('whisper', t('notifications.whisperReady'));
            } catch (e) {
                console.error('[Whisper]', e);
                this.ui.failLoadingNotification('whisper', t('notifications.whisperFailed'));
                throw e;
            }
        }
        
        return this.localTranscriber.transcribe(blob, this.audioCapture.getAudioContext());
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

    _onSTTModeChange(mode) {
        this.settings.sttMode = mode;
    }

    _onTTSEnabledChange(enabled) {
        this.settings.ttsEnabled = enabled;
    }
}

// Bootstrap
const app = new MagecApp();
app.init();
