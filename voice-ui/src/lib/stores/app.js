import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { CONFIG } from '../config.js'
import { clientAuth, AgentClient } from '../api/index.js'
import { SessionManager, SessionService } from '../session/index.js'
import { SettingsManager } from '../settings/SettingsManager.js'
import { AudioCapture, AudioRecorder, VoiceEventsClient, FeedbackSound, OpenAITTS } from '../audio/index.js'
import { RemoteTranscriber } from '../transcription/RemoteTranscriber.js'
import { WakeLock } from '../utils/WakeLock.js'
import { t, initLanguage, setLanguage, getLanguage } from '../i18n/index.js'

export const useAppStore = defineStore('app', () => {
  const activePanel = ref('assistant')
  const status = ref({ text: t('status.initializing'), type: 'default' })
  const isRecording = ref(false)
  const centellaEnabled = ref(false)
  const isPaired = ref(false)
  const showPairing = ref(false)

  const selectedAgent = ref(null)
  const allowedAgents = ref([])
  const spokesperson = ref(null)

  const messages = ref([])
  const sessions = ref([])
  const currentSessionId = ref(null)

  const notifications = ref([])
  let notificationId = 0
  const loadingNotifications = {}

  const wakeWordEnabled = ref(true)
  const wakeWordPhrase = ref('')
  const wakeWordModels = ref([])
  const activeWakeWordModel = ref(null)
  const wakeWordAvailable = ref(true)

  const ttsEnabled = ref(true)
  const ttsAvailable = ref(true)

  const sidebarOpen = ref(false)
  const spokespersonPanelOpen = ref(false)

  let agentClient = null
  let sessionManager = null
  let sessionService = null
  let settings = null
  let audioCapture = null
  let audioRecorder = null
  let voiceEvents = null
  let transcriber = null
  let feedbackSound = null
  let tts = null
  let wakeLock = null
  let prevWakeWordEnabled = true
  let recordingTimeout = null

  async function init() {
    initLanguage()

    const needsPairing = await _checkClientPairing()
    if (needsPairing) {
      showPairing.value = true
      return
    }

    await _startApp()
  }

  async function _checkClientPairing() {
    const paired = await clientAuth.checkPairing()
    if (paired) return false
    if (!clientAuth.token) {
      try {
        const infoRes = await fetch('/api/v1/client/info')
        if (infoRes.ok) {
          const info = await infoRes.json()
          if (!info.paired && info.paired !== undefined) return true
        }
      } catch {}
      return false
    }
    return true
  }

  async function pair(token) {
    const ok = await clientAuth.pair(token)
    if (ok) {
      showPairing.value = false
      isPaired.value = true
      await _startApp()
    }
    return ok
  }

  async function _startApp() {
    isPaired.value = true

    wakeLock = new WakeLock()
    wakeLock.enable()

    feedbackSound = new FeedbackSound({ volume: 0.3 })
    tts = new OpenAITTS()
    agentClient = new AgentClient()
    sessionManager = new SessionManager()
    sessionService = new SessionService()

    selectedAgent.value = clientAuth.defaultAgent || null
    allowedAgents.value = clientAuth.allowedAgents || []
    settings = new SettingsManager(selectedAgent.value)

    _resolveSpokesperson()

    if (selectedAgent.value) {
      const voiceAgentId = spokesperson.value || selectedAgent.value
      agentClient.setAgent(selectedAgent.value)
      sessionService.setAgent(selectedAgent.value)
      tts.setAgent(voiceAgentId)
    }

    ttsEnabled.value = settings.ttsEnabled
    wakeWordEnabled.value = settings.wakeWordEnabled

    await _checkTTSAvailability()
    await _initVoiceEvents()
    await _initSession()
    _setReady()
    await _startListening()
  }

  async function _checkTTSAvailability() {
    const available = await tts.checkAvailable()
    ttsAvailable.value = available
    if (!available) {
      settings.ttsEnabled = false
      ttsEnabled.value = false
      addNotification('warning', t('notifications.ttsUnavailable'))
    }
  }

  async function _initVoiceEvents() {
    setStatus(t('status.loadingWakeWord'), 'loading')
    showLoadingNotification('wakeword', t('notifications.wakeWordLoading'))

    try {
      const ve = new VoiceEventsClient()

      ve.onWakeword = () => {
        if (wakeWordEnabled.value) startRecording()
      }
      ve.onSpeechStart = () => {}
      ve.onSpeechEnd = () => {
        if (isRecording.value && ve.isVADEnabled()) {
          stopRecording()
        }
      }
      ve.onCapabilities = (caps) => _onCapabilities(caps)

      await ve.load()
      voiceEvents = ve

      wakeWordPhrase.value = ve.getActivePhrase()
      wakeWordEnabled.value = settings.wakeWordEnabled

      completeLoadingNotification('wakeword', t('notifications.wakeWordReady'))
    } catch {
      voiceEvents = null
      wakeWordModels.value = []
      wakeWordEnabled.value = false
      wakeWordAvailable.value = false
      failLoadingNotification('wakeword', t('notifications.wakeWordUnavailable'))
    }
  }

  function _onCapabilities(caps) {
    if (caps.wakewords) {
      wakeWordModels.value = caps.wakewords.models || []
      settings.setValidWakeWordModels(wakeWordModels.value.map(m => m.id))

      const activeModel = caps.wakewords.active
      activeWakeWordModel.value = activeModel
      const activeConfig = wakeWordModels.value.find(m => m.id === activeModel)
      if (activeConfig) {
        wakeWordPhrase.value = activeConfig.phrase || activeConfig.id
      }
    }
  }

  async function _initSession() {
    sessionManager.onSessionChange = (id) => _onSessionChange(id)
    const id = sessionManager.init()
    currentSessionId.value = id
    await agentClient.createSession(id)
    await refreshSessionList()
  }

  function _onSessionChange(id) {
    currentSessionId.value = id
    agentClient.createSession(id)
    messages.value = []
    refreshSessionList()
  }

  function parseSessionTimestamp(id) {
    const parts = (id || '').split('_')
    if (parts.length >= 2) {
      const ts = parseInt(parts[1], 36)
      if (!isNaN(ts) && ts > 0) return ts
    }
    return 0
  }

  async function refreshSessionList() {
    const list = await sessionService.listSessions()
    const history = sessionManager.getSessionHistory()

    const enriched = await Promise.all(
      list.map(async (session) => ({
        id: session.id,
        preview: sessionService.getSessionPreview(
          await sessionService.getSession(session.id)
        ),
        createdAt: parseSessionTimestamp(session.id)
          || history.find(s => s.id === session.id)?.createdAt
          || 0
      }))
    )

    enriched.sort((a, b) => b.createdAt - a.createdAt)
    sessions.value = enriched
  }

  async function selectSession(sessionId) {
    if (sessionId === sessionManager.getCurrentSessionId()) {
      sidebarOpen.value = false
      return
    }

    const session = await sessionService.getSession(sessionId)
    if (!session) return

    sessionManager.currentSessionId = sessionId
    currentSessionId.value = sessionId
    messages.value = sessionService.extractMessages(session).map(m => ({
      role: m.role === 'user' ? 'user' : 'ai',
      text: m.text
    }))
    await refreshSessionList()
    sidebarOpen.value = false
  }

  async function deleteSession(sessionId) {
    if (sessionId === sessionManager.getCurrentSessionId()) {
      sessionManager.newSession()
    }
    await sessionService.deleteSession(sessionId)
    await refreshSessionList()
  }

  function newSession() {
    sessionManager.newSession()
  }

  function _setReady() {
    centellaEnabled.value = true
    setStatus(t('status.ready'), 'listening')
  }

  async function _startListening() {
    try {
      audioCapture = new AudioCapture()
      await audioCapture.start()

      if (voiceEvents) {
        audioCapture.onAudioData = (samples, sampleRate) => {
          voiceEvents.processAudio(samples, sampleRate)
        }
      }
    } catch {
      addNotification('error', t('errors.microphoneAccess'))
    }
  }

  function getAnalyser() {
    return audioCapture?.getAnalyser()
  }

  function toggleRecording() {
    isRecording.value ? stopRecording() : startRecording()
  }

  function startRecording() {
    if (isRecording.value || !audioCapture) return

    isRecording.value = true
    tts?.stop()
    feedbackSound?.playWakeChime()
    prevWakeWordEnabled = wakeWordEnabled.value
    wakeWordEnabled.value = false
    setStatus(t('status.recording'), 'recording')

    audioRecorder = new AudioRecorder(audioCapture.getMicStream())
    audioRecorder.onRecordingComplete = (blob) => {
      wakeWordEnabled.value = prevWakeWordEnabled
      _processRecording(blob)
    }
    audioRecorder.start()

    if (!voiceEvents?.isVADEnabled()) {
      recordingTimeout = setTimeout(() => {
        if (isRecording.value) stopRecording()
      }, 10000)
    }
  }

  function stopRecording() {
    if (!isRecording.value) return

    isRecording.value = false
    feedbackSound?.playStopChime()
    setStatus(t('status.processing'), 'processing')

    if (recordingTimeout) {
      clearTimeout(recordingTimeout)
      recordingTimeout = null
    }

    audioRecorder?.stop()
  }

  async function _processRecording(blob) {
    centellaEnabled.value = false

    try {
      const text = await _transcribe(blob)
      if (text) {
        messages.value.push({ role: 'user', text })
        await _sendToAgent(text)
      }
    } catch {
      addNotification('warning', t('errors.transcriptionUnavailable'))
    }

    centellaEnabled.value = true
    _setReady()
  }

  async function _transcribe(blob) {
    if (!transcriber) {
      transcriber = new RemoteTranscriber(CONFIG.transcription)
    }
    const voiceAgentId = spokesperson.value || selectedAgent.value
    if (voiceAgentId) transcriber.setAgent(voiceAgentId)
    return transcriber.transcribe(blob)
  }

  async function sendTextMessage(text) {
    messages.value.push({ role: 'user', text })
    await _sendToAgent(text)
  }

  async function _sendToAgent(message) {
    tts?.stop()
    setStatus(t('status.thinking'), 'processing')

    let responses = []
    try {
      const sessionId = sessionManager.getCurrentSessionId()
      responses = await agentClient.sendMessage(sessionId, message)
    } catch {
      messages.value.push({ role: 'ai', text: t('errors.generic') })
      setStatus(t('status.ready'), 'listening')
      return
    }

    setStatus(t('status.ready'), 'listening')
    refreshSessionList()

    for (const response of responses) {
      messages.value.push({ role: 'ai', text: response })
      if (ttsEnabled.value) {
        try {
          await tts.speak(response)
        } catch {
          tts?.stop()
          addNotification('warning', t('errors.ttsUnavailable'))
        }
      }
    }
  }

  function switchAgent(agentId) {
    selectedAgent.value = agentId
    agentClient.setAgent(agentId)
    sessionService.setAgent(agentId)
    settings.switchAgent(agentId)

    _resolveSpokesperson()
    const voiceAgentId = spokesperson.value || agentId
    tts.setAgent(voiceAgentId)
    if (transcriber) transcriber.setAgent(voiceAgentId)

    ttsEnabled.value = settings.ttsEnabled
    wakeWordEnabled.value = settings.wakeWordEnabled

    sessionManager.newSession()
  }

  function switchSpokesperson(agentId) {
    spokesperson.value = agentId
    settings.spokesperson = agentId
    const voiceAgentId = agentId || selectedAgent.value
    tts.setAgent(voiceAgentId)
    if (transcriber) transcriber.setAgent(voiceAgentId)
  }

  function _resolveSpokesperson() {
    const current = _getActiveAgentInfo()
    if (!current || current.type !== 'flow') {
      spokesperson.value = null
      return
    }
    const candidates = (current.agents || []).filter(a => a.responseAgent)
    const saved = settings.spokesperson
    if (saved && candidates.some(a => a.id === saved)) {
      spokesperson.value = saved
    } else if (candidates.length > 0) {
      spokesperson.value = candidates[0].id
      settings.spokesperson = candidates[0].id
    } else if (current.agents?.length > 0) {
      spokesperson.value = current.agents[0].id
      settings.spokesperson = current.agents[0].id
    } else {
      spokesperson.value = null
    }
  }

  function _getActiveAgentInfo() {
    return allowedAgents.value.find(a => a.id === selectedAgent.value) || null
  }

  const activeAgentInfo = computed(() => _getActiveAgentInfo())
  const spokespersonCandidates = computed(() => {
    const info = _getActiveAgentInfo()
    if (!info || info.type !== 'flow') return []
    const resp = (info.agents || []).filter(a => a.responseAgent)
    return resp.length > 0 ? resp : (info.agents || [])
  })
  const spokespersonName = computed(() => {
    if (!spokesperson.value) return null
    const info = _getActiveAgentInfo()
    if (!info?.agents) return null
    const agent = info.agents.find(a => a.id === spokesperson.value)
    return agent?.name || null
  })

  function setWakeWordModel(modelId) {
    if (!voiceEvents) return
    voiceEvents.setWakewordModel(modelId)
    settings.wakeWordModel = modelId
    activeWakeWordModel.value = modelId

    const modelConfig = wakeWordModels.value.find(m => m.id === modelId)
    wakeWordPhrase.value = modelConfig?.phrase || modelId
  }

  function toggleWakeWord(enabled) {
    wakeWordEnabled.value = enabled
    settings.wakeWordEnabled = enabled
  }

  function toggleTTS(enabled) {
    ttsEnabled.value = enabled
    settings.ttsEnabled = enabled
  }

  function changeLanguage(lang) {
    setLanguage(lang)
  }

  function setStatus(text, type = 'default') {
    status.value = { text, type }
  }

  function switchPanel(panel) {
    activePanel.value = panel
  }

  function clearMessages() {
    messages.value = []
  }

  function copyMessages() {
    const text = messages.value.map(m => m.text).join('\n')
    navigator.clipboard.writeText(text)
  }

  function addNotification(type, message) {
    notifications.value.unshift({
      id: ++notificationId,
      type,
      message,
      timestamp: new Date()
    })
  }

  function removeNotification(id) {
    notifications.value = notifications.value.filter(n => n.id !== id)
  }

  function clearAllNotifications() {
    notifications.value = []
  }

  function showLoadingNotification(key, message) {
    if (loadingNotifications[key]) {
      removeNotification(loadingNotifications[key])
    }
    notificationId++
    loadingNotifications[key] = notificationId
    notifications.value.unshift({
      id: notificationId,
      type: 'loading',
      message,
      timestamp: new Date()
    })
  }

  function completeLoadingNotification(key, message) {
    _clearLoadingNotification(key)
    addNotification('success', message)
  }

  function failLoadingNotification(key, message) {
    _clearLoadingNotification(key)
    addNotification('error', message)
  }

  function _clearLoadingNotification(key) {
    if (loadingNotifications[key]) {
      removeNotification(loadingNotifications[key])
      delete loadingNotifications[key]
    }
  }

  const notificationCount = computed(() => notifications.value.length)
  const hasMessages = computed(() => messages.value.length > 0)

  return {
    activePanel,
    status,
    isRecording,
    centellaEnabled,
    isPaired,
    showPairing,
    selectedAgent,
    allowedAgents,
    spokesperson,
    activeAgentInfo,
    spokespersonCandidates,
    spokespersonName,
    messages,
    sessions,
    currentSessionId,
    notifications,
    notificationCount,
    hasMessages,
    wakeWordEnabled,
    wakeWordPhrase,
    wakeWordModels,
    activeWakeWordModel,
    wakeWordAvailable,
    ttsEnabled,
    ttsAvailable,
    sidebarOpen,
    spokespersonPanelOpen,

    init,
    pair,
    switchAgent,
    switchSpokesperson,
    switchPanel,
    toggleRecording,
    startRecording,
    stopRecording,
    sendTextMessage,
    clearMessages,
    copyMessages,
    newSession,
    selectSession,
    deleteSession,
    refreshSessionList,
    setWakeWordModel,
    toggleWakeWord,
    toggleTTS,
    changeLanguage,
    addNotification,
    removeNotification,
    clearAllNotifications,
    getAnalyser,
  }
})
