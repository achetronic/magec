<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex flex-col lg:flex-row lg:items-center gap-3 mb-2">
      <div class="flex items-center gap-3 flex-1 min-w-0">
        <button
          @click="$emit('back')"
          class="w-7 h-7 rounded-lg flex items-center justify-center text-arena-400 hover:text-arena-200 hover:bg-piedra-800/80 transition-colors flex-shrink-0"
        >
          <Icon name="back" size="sm" />
        </button>
        <div class="flex-1 min-w-0 space-y-1.5">
          <div class="flex items-center gap-2">
            <h2 class="text-sm font-semibold text-arena-200 truncate">
              {{ conversation?.agentName || conversation?.agentId || 'Conversation' }}
            </h2>
            <Badge v-if="conversation?.summary" variant="green">summarized</Badge>
            <!-- Info popover -->
            <div class="relative group/info flex-shrink-0">
              <Icon name="eye" size="sm" class="text-arena-600 group-hover/info:text-arena-400 transition-colors cursor-default" />
              <div class="absolute left-0 top-full mt-1.5 z-50 hidden group-hover/info:block">
                <div class="bg-piedra-900 border border-piedra-700/50 rounded-lg shadow-xl p-3 space-y-1.5 min-w-52">
                  <div v-if="conversation?.startedAt" class="flex items-center gap-2 text-[10px]">
                    <span class="text-arena-600 w-14 flex-shrink-0">Started</span>
                    <span class="text-arena-400 tabular-nums">{{ formatTime(conversation.startedAt) }}</span>
                  </div>
                  <div v-if="conversation?.userId" class="flex items-center gap-2 text-[10px]">
                    <span class="text-arena-600 w-14 flex-shrink-0">User</span>
                    <span class="text-arena-400 truncate">{{ conversation.userId }}</span>
                  </div>
                  <div v-if="conversation?.sessionId" class="flex items-center gap-2 text-[10px]">
                    <span class="text-arena-600 w-14 flex-shrink-0">Session</span>
                    <span class="text-arena-400 font-mono truncate">{{ conversation.sessionId }}</span>
                  </div>
                  <div v-if="totalMessages" class="flex items-center gap-2 text-[10px]">
                    <span class="text-arena-600 w-14 flex-shrink-0">Messages</span>
                    <span class="text-arena-400 tabular-nums">{{ totalMessages }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-1.5">
            <Badge v-if="conversation?.source" variant="muted" class="!py-0">{{ formatSource(conversation.source) }}</Badge>
            <Badge v-if="conversation?.flowId" variant="muted" class="!py-0">Flow</Badge>
            <Badge v-if="conversation?.clientName" variant="muted" class="!py-0">{{ conversation.clientName }}</Badge>
          </div>
        </div>
      </div>

      <!-- Controls -->
      <div class="flex items-center gap-1.5 flex-shrink-0 overflow-x-auto">
          <!-- Auto-refresh -->
        <div class="flex items-center bg-piedra-800 rounded-lg border border-piedra-700/50 p-0.5">
          <button
            v-for="opt in [{ label: 'Off', ms: 0 }, { label: '5s', ms: 5000 }, { label: '30s', ms: 30000 }]"
            :key="opt.ms"
            @click="setAutoRefresh(opt.ms)"
            class="px-2.5 py-1 text-[10px] font-medium rounded-md transition-colors"
            :class="autoRefreshMs === opt.ms ? 'bg-piedra-700 text-arena-100' : 'text-arena-500 hover:text-arena-300'"
          >{{ opt.label }}</button>
        </div>

        <!-- Manual refresh -->
        <button
          @click="manualRefresh"
          class="p-1.5 hover:bg-piedra-800 rounded-lg transition-colors"
          title="Refresh messages"
        >
          <Icon
            name="refresh"
            size="sm"
            class="transition-all duration-400"
            :class="refreshPulse ? 'text-arena-200 rotate-180' : 'text-arena-500'"
          />
        </button>

        <div class="w-px h-4 bg-piedra-700/50" />

        <!-- Perspective toggle -->
        <div v-if="pairId || conversation?.perspective" class="flex items-center bg-piedra-800 rounded-lg border border-piedra-700/50 p-0.5">
          <button
            @click="switchPerspective('user')"
            :disabled="!canSwitch('user')"
            class="px-2.5 py-1 text-[10px] font-medium rounded-md transition-colors"
            :class="activePerspective === 'user'
              ? 'bg-piedra-700 text-arena-100'
              : pairId ? 'text-arena-500 hover:text-arena-300 cursor-pointer' : 'text-arena-600 cursor-not-allowed'"
          >User</button>
          <button
            @click="switchPerspective('admin')"
            :disabled="!canSwitch('admin')"
            class="px-2.5 py-1 text-[10px] font-medium rounded-md transition-colors"
            :class="activePerspective === 'admin'
              ? 'bg-piedra-700 text-arena-100'
              : pairId ? 'text-arena-500 hover:text-arena-300 cursor-pointer' : 'text-arena-600 cursor-not-allowed'"
          >Admin</button>
        </div>

        <!-- View toggle -->
        <div class="flex items-center bg-piedra-800 rounded-lg border border-piedra-700/50 p-0.5">
          <button
            @click="showRaw = false"
            class="px-2.5 py-1 text-[10px] font-medium rounded-md transition-colors"
            :class="!showRaw ? 'bg-piedra-700 text-arena-100' : 'text-arena-500 hover:text-arena-300'"
          >Messages</button>
          <button
            @click="showRaw = true"
            class="px-2.5 py-1 text-[10px] font-medium rounded-md transition-colors"
            :class="showRaw ? 'bg-piedra-700 text-arena-100' : 'text-arena-500 hover:text-arena-300'"
          >Raw</button>
        </div>

        <!-- Actions -->
        <button
          @click="handleExportPDF"
          class="flex items-center gap-1 p-1.5 hover:bg-piedra-800 rounded-lg transition-colors group/btn"
          title="Export as PDF"
        >
          <Icon name="download" size="sm" class="text-arena-500 group-hover/btn:text-arena-300 transition-colors" />
          <span class="text-[10px] font-medium text-arena-500 group-hover/btn:text-arena-300 transition-colors">PDF</span>
        </button>
        <button
          v-if="conversation?.sessionId"
          @click="handleResetSession"
          class="flex items-center gap-1 p-1.5 hover:bg-lava-500/10 rounded-lg transition-colors group/btn"
          title="Reset ADK session"
        >
          <Icon name="close" size="sm" class="text-arena-500 group-hover/btn:text-lava-400 transition-colors" />
          <span class="text-[10px] font-medium text-arena-500 group-hover/btn:text-lava-400 transition-colors">Session</span>
        </button>
        <button
          @click="handleDelete"
          class="p-1.5 hover:bg-lava-500/10 rounded-lg transition-colors group/btn"
          title="Delete conversation"
        >
          <Icon name="trash" size="sm" class="text-arena-500 group-hover/btn:text-lava-400 transition-colors" />
        </button>
      </div>
    </div>

    <!-- Summary -->
    <div v-if="conversation?.summary" class="bg-green-500/5 border border-green-500/20 rounded-xl p-4">
      <div class="flex items-center gap-2 mb-2">
        <div class="w-5 h-5 rounded-md bg-green-500/15 flex items-center justify-center">
          <Icon name="command" size="xs" class="text-green-400" />
        </div>
        <span class="text-[10px] font-semibold text-green-300 uppercase tracking-wider">Context Summary</span>
      </div>
      <div class="text-xs text-arena-300 leading-relaxed prose-content" v-html="renderMarkdown(conversation.summary)" />
    </div>

    <!-- Loading -->
    <div v-if="loading && !messages.length" class="flex items-center justify-center py-20">
      <Icon name="refresh" size="lg" class="text-arena-500 animate-spin" />
    </div>

    <!-- Raw events -->
    <div v-else-if="showRaw && rawEvents.length" class="space-y-2">
      <div class="bg-piedra-900 border border-piedra-700/50 rounded-xl overflow-hidden">
        <div class="px-4 py-2 border-b border-piedra-700/50 flex items-center justify-between">
          <span class="text-[10px] font-semibold text-arena-400 uppercase tracking-wider">
            Raw ADK Events ({{ rawEvents.length }})
          </span>
          <button @click="copyRaw" class="text-[10px] text-arena-500 hover:text-arena-300 transition-colors">
            Copy JSON
          </button>
        </div>
        <pre class="p-4 text-[11px] text-arena-300 font-mono overflow-x-auto max-h-[70vh] leading-relaxed">{{ formatJSON(rawEvents) }}</pre>
      </div>
    </div>

    <!-- Messages -->
    <div v-else-if="messages.length" class="space-y-1 py-4">
      <template v-for="(msg, i) in displayMessages" :key="msg._key">
        <!-- Timestamp separator -->
        <div
          v-if="i === 0 || shouldShowTimestamp(displayMessages[i - 1], msg)"
          class="flex items-center gap-3 py-2"
        >
          <div class="flex-1 h-px bg-piedra-700/40" />
          <span class="text-[9px] text-arena-600 tabular-nums flex-shrink-0">{{ formatMessageTime(msg.timestamp) }}</span>
          <div class="flex-1 h-px bg-piedra-700/40" />
        </div>

        <!-- Message row -->
        <div class="group/msg flex gap-3 py-3 px-3 -mx-3 rounded-lg hover:bg-piedra-800/30 transition-colors">
          <!-- Role indicator -->
          <div class="flex-shrink-0 mt-0.5">
            <div
              class="w-6 h-6 rounded-full flex items-center justify-center text-[9px] font-bold"
              :class="msg.role === 'user'
                ? 'bg-teal-500/15 text-teal-400'
                : 'bg-piedra-700/60 text-arena-400'"
            >{{ msg.role === 'user' ? 'U' : 'A' }}</div>
          </div>

          <!-- Content -->
          <div class="flex-1 min-w-0">
            <!-- Author line -->
            <div class="flex items-baseline gap-2 mb-0.5">
              <span class="text-[10px] text-arena-500">{{ msg.role === 'user' ? (conversation.userId || 'User') : (store.agentLabel(msg.agent) || conversation.agentName || 'Assistant') }}</span>
              <span class="text-[9px] text-arena-600 tabular-nums opacity-0 group-hover/msg:opacity-100 transition-opacity">{{ formatMessageTime(msg.timestamp) }}</span>
            </div>

            <!-- Text -->
            <div
              v-if="msg.content"
              class="text-[13px] leading-[1.7] text-arena-300 prose-content"
              v-html="renderMarkdown(msg.content)"
            />

            <!-- Tool calls -->
            <div v-if="msg.toolCalls?.length" class="mt-2 space-y-1">
              <button
                v-for="(tc, j) in msg.toolCalls"
                :key="j"
                class="w-full text-left rounded-lg transition-colors"
                :class="expandedTools[toolKey(msg._key, j)]
                  ? 'bg-indigo-500/5 border border-indigo-500/20'
                  : 'bg-piedra-800/40 border border-piedra-700/30 hover:border-piedra-600/50'"
                @click="toggleTool(msg._key, j)"
              >
                <div class="flex items-center gap-1.5 px-2.5 py-1.5">
                  <Icon name="bolt" size="xs" class="text-indigo-400/70 flex-shrink-0" />
                  <span class="text-[10px] font-medium text-indigo-300/80">{{ tc.name }}</span>
                  <span v-if="tc.args && !tc.result && !expandedTools[toolKey(msg._key, j)]" class="text-[9px] text-arena-600 truncate flex-1 ml-1 font-mono">{{ truncateJSON(tc.args) }}</span>
                  <Icon
                    :name="expandedTools[toolKey(msg._key, j)] ? 'chevronDown' : 'chevronRight'"
                    size="xs"
                    class="text-arena-600 flex-shrink-0 ml-auto"
                  />
                </div>
                <div v-if="expandedTools[toolKey(msg._key, j)]" class="px-2.5 pb-2 border-t border-piedra-700/20 mt-0.5 pt-1.5" @click.stop>
                  <pre v-if="tc.args" class="text-[10px] text-arena-400 font-mono overflow-x-auto max-h-48 leading-relaxed select-text">{{ formatJSON(tc.args) }}</pre>
                  <pre v-if="tc.result" class="text-[10px] text-green-400/60 font-mono overflow-x-auto max-h-48 leading-relaxed mt-1.5 select-text">{{ formatJSON(tc.result) }}</pre>
                </div>
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- Load older -->
      <button
        v-if="hasOlderMessages"
        @click="loadOlder"
        :disabled="loadingOlder"
        class="w-full py-2 text-[10px] font-medium text-arena-400 hover:text-arena-200 bg-piedra-800/60 hover:bg-piedra-800 border border-piedra-700/50 rounded-xl transition-colors disabled:opacity-50 mt-2"
      >
        <span v-if="loadingOlder">Loading…</span>
        <span v-else>Load older messages ({{ totalMessages - messages.length }} remaining)</span>
      </button>
    </div>

    <!-- Empty -->
    <EmptyState
      v-else-if="!loading"
      title="No messages"
      subtitle="This conversation has no message data"
      icon="chat"
      color="teal"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, inject, onBeforeUnmount } from 'vue'
import { marked } from 'marked'
import { conversationsApi } from '../../lib/api/index.js'
import { useDataStore } from '../../lib/stores/data.js'
import Badge from '../../components/Badge.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'

const MSG_PAGE_SIZE = 50

const props = defineProps({
  conversationId: { type: String, default: '' },
})

const emit = defineEmits(['back', 'deleted', 'navigate'])

const store = useDataStore()
const toast = inject('toast')
const requestDelete = inject('requestDelete')

const conversation = ref(null)
const messages = ref([])
const rawEvents = ref([])
const totalMessages = ref(0)
const loading = ref(false)
const loadingOlder = ref(false)
const showRaw = ref(false)
const expandedTools = reactive({})
const pairId = ref(null)
const autoRefreshMs = ref(0)
const refreshPulse = ref(false)

let keyCounter = 0
let autoRefreshTimer = null

const activePerspective = computed(() => conversation.value?.perspective || 'user')
const hasOlderMessages = computed(() => messages.value.length < totalMessages.value)
const displayMessages = computed(() => [...messages.value].reverse())

function canSwitch(target) {
  if (activePerspective.value === target) return false
  return !!pairId.value
}

function switchPerspective(target) {
  if (!canSwitch(target)) return
  emit('navigate', pairId.value)
}

function shouldShowTimestamp(prev, curr) {
  if (!prev?.timestamp || !curr?.timestamp) return false
  return Math.abs(new Date(curr.timestamp) - new Date(prev.timestamp)) > 60000
}

function toolKey(msgIndex, toolIndex) {
  return `${msgIndex}_${toolIndex}`
}

function toggleTool(msgIndex, toolIndex) {
  const key = toolKey(msgIndex, toolIndex)
  expandedTools[key] = !expandedTools[key]
}

function truncateJSON(obj) {
  try {
    const s = JSON.stringify(obj)
    return s.length > 80 ? s.slice(0, 80) + '...' : s
  } catch {
    return String(obj)
  }
}

marked.setOptions({ breaks: true, gfm: true })

function tagMessages(msgs) {
  return msgs.map(m => ({ ...m, _key: `msg_${keyCounter++}` }))
}

async function loadConversation(id) {
  if (!id) return
  loading.value = true
  pairId.value = null
  try {
    const result = await conversationsApi.get(id, { msgLimit: MSG_PAGE_SIZE, msgOffset: 0 })
    conversation.value = result.conversation
    totalMessages.value = result.totalMessages || 0
    messages.value = tagMessages(result.conversation?.messages || [])
    rawEvents.value = result.conversation?.rawEvents || []

    conversationsApi.findPair(id).then(r => {
      pairId.value = r.pairId || null
    }).catch(() => {})
  } catch (e) {
    toast.error(e.message)
  } finally {
    loading.value = false
  }
}

async function loadOlder() {
  if (!props.conversationId || loadingOlder.value) return
  loadingOlder.value = true
  try {
    const msgOffset = messages.value.length
    const result = await conversationsApi.get(props.conversationId, {
      msgLimit: MSG_PAGE_SIZE,
      msgOffset,
    })
    const olderMsgs = result.conversation?.messages || []
    if (olderMsgs.length) {
      messages.value = [...tagMessages(olderMsgs), ...messages.value]
      rawEvents.value = [...(result.conversation?.rawEvents || []), ...rawEvents.value]
    }
  } catch (e) {
    toast.error(e.message)
  } finally {
    loadingOlder.value = false
  }
}

function manualRefresh() {
  if (!props.conversationId) return
  triggerPulse()
  loadConversation(props.conversationId)
}

function triggerPulse() {
  refreshPulse.value = true
  setTimeout(() => { refreshPulse.value = false }, 400)
}

function setAutoRefresh(ms) {
  clearInterval(autoRefreshTimer)
  autoRefreshTimer = null
  autoRefreshMs.value = ms
  if (ms > 0) {
    autoRefreshTimer = setInterval(() => {
      triggerPulse()
      loadConversation(props.conversationId)
    }, ms)
  }
}

watch(() => props.conversationId, (id) => {
  showRaw.value = false
  setAutoRefresh(0)
  Object.keys(expandedTools).forEach(k => delete expandedTools[k])
  messages.value = []
  rawEvents.value = []
  totalMessages.value = 0
  keyCounter = 0
  loadConversation(id)
}, { immediate: true })

onBeforeUnmount(() => {
  clearInterval(autoRefreshTimer)
})

function renderMarkdown(text) {
  if (!text) return ''
  try {
    return marked.parse(text)
  } catch {
    return text.replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/\n/g, '<br>')
  }
}

function formatJSON(obj) {
  try {
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(obj)
  }
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) +
    ' ' + d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatMessageTime(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatSource(source) {
  const map = {
    'voice-ui': 'Voice UI',
    telegram: 'Telegram',
    executor: 'Executor',
    flow: 'Flow',
    direct: 'Direct',
    cron: 'Cron',
    webhook: 'Webhook',
  }
  return map[source] || source
}

function copyRaw() {
  if (!rawEvents.value.length) return
  navigator.clipboard.writeText(JSON.stringify(rawEvents.value, null, 2))
  toast.success('Copied to clipboard')
}

function handleDelete() {
  requestDelete('Delete this conversation? This cannot be undone.', async () => {
    try {
      await conversationsApi.delete(props.conversationId)
      toast.success('Conversation deleted')
      emit('deleted')
      emit('back')
    } catch (e) {
      toast.error(e.message)
    }
  })
}

function handleResetSession() {
  requestDelete('Reset the ADK session for this conversation? The user will start a fresh session on their next message. The audit log is preserved.', async () => {
    try {
      await conversationsApi.resetSession(props.conversationId)
      toast.success('Session reset successfully')
    } catch (e) {
      toast.error(e.message)
    }
  })
}

async function handleExportPDF() {
  if (!conversation.value) return
  toast.success('Preparing PDF…')

  try {
    const ids = [props.conversationId]
    if (pairId.value) ids.push(pairId.value)

    const perspectives = await Promise.all(
      ids.map(async (id) => {
        const result = await conversationsApi.get(id, { msgLimit: 9999, msgOffset: 0 })
        return result.conversation
      })
    )

    perspectives.sort((a, b) => (a.perspective === 'user' ? -1 : 1))

    const title = conversation.value.agentName || conversation.value.agentId || 'Conversation'
    const meta = [
      conversation.value.source && formatSource(conversation.value.source),
      conversation.value.flowName,
      conversation.value.clientName,
      conversation.value.userId && `User: ${conversation.value.userId}`,
      conversation.value.sessionId && `Session: ${conversation.value.sessionId.slice(0, 12)}`,
      formatTime(conversation.value.startedAt),
    ].filter(Boolean).join(' · ')

    const renderMessages = (msgs) => msgs.map((m) => {
      const author = m.role === 'user' ? 'User' : (store.agentLabel(m.agent) || conversation.value.agentName || 'Assistant')
      const time = formatMessageTime(m.timestamp)
      const toolsHtml = (m.toolCalls || []).map(tc =>
        `<div class="tool"><span class="tool-name">⚡ ${tc.name}</span>${tc.args ? `<pre>${JSON.stringify(tc.args, null, 2)}</pre>` : ''}${tc.result ? `<pre class="tool-result">${JSON.stringify(tc.result, null, 2)}</pre>` : ''}</div>`
      ).join('')
      return `<div class="msg ${m.role}"><div class="msg-header"><span class="author">${author}</span><span class="time">${time}</span></div>${m.content ? `<div class="content">${marked.parse(m.content)}</div>` : ''}${toolsHtml}</div>`
    }).join('')

    const sections = perspectives.map((p) => {
      const label = p.perspective === 'user' ? 'User Perspective' : 'Admin Perspective'
      const msgs = p.messages || []
      return `<div class="section"><h2>${label} <span class="msg-count">(${msgs.length} messages)</span></h2>${msgs.length ? renderMessages(msgs) : '<p class="empty">No messages</p>'}</div>`
    }).join('')

    const html = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>${title}</title><style>
      * { margin: 0; padding: 0; box-sizing: border-box; }
      body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; font-size: 11px; color: #1a1a1a; padding: 32px; max-width: 800px; margin: 0 auto; }
      h1 { font-size: 16px; font-weight: 600; margin-bottom: 4px; }
      .meta { font-size: 10px; color: #666; margin-bottom: 24px; }
      .section { margin-bottom: 32px; }
      .section h2 { font-size: 12px; font-weight: 600; color: #333; border-bottom: 1px solid #ddd; padding-bottom: 6px; margin-bottom: 12px; }
      .msg-count { font-weight: 400; color: #999; }
      .msg { padding: 8px 0; border-bottom: 1px solid #f0f0f0; }
      .msg:last-child { border-bottom: none; }
      .msg-header { margin-bottom: 3px; }
      .author { font-size: 10px; font-weight: 600; color: #555; }
      .msg.user .author { color: #0d9488; }
      .time { font-size: 9px; color: #999; margin-left: 8px; }
      .content { font-size: 11px; line-height: 1.6; color: #333; }
      .content p { margin-bottom: 4px; }
      .content pre { background: #f5f5f5; border: 1px solid #e5e5e5; border-radius: 4px; padding: 8px; overflow-x: auto; font-size: 10px; margin: 4px 0; }
      .content code { background: #f5f5f5; padding: 1px 4px; border-radius: 2px; font-size: 10px; }
      .content pre code { background: none; padding: 0; }
      .tool { margin: 4px 0; padding: 6px 8px; background: #f8f8ff; border: 1px solid #e8e8f0; border-radius: 4px; }
      .tool-name { font-size: 10px; font-weight: 500; color: #6366f1; }
      .tool pre { font-size: 9px; margin-top: 4px; background: #fff; }
      .tool-result { color: #059669; }
      .empty { color: #999; font-style: italic; }
      @media print { body { padding: 16px; } }
    </style></head><body><h1>${title}</h1><p class="meta">${meta}</p>${sections}</body></html>`

    const win = window.open('', '_blank')
    win.document.write(html)
    win.document.close()
    win.onload = () => { win.print() }
  } catch (e) {
    toast.error('Export failed: ' + e.message)
  }
}
</script>

<style>
.prose-content {
  line-height: 1.7;
}

.prose-content p {
  margin-bottom: 0.4em;
}

.prose-content p:last-child {
  margin-bottom: 0;
}

.prose-content pre {
  background-color: rgba(42, 42, 47, 0.8);
  border: 1px solid rgba(61, 61, 68, 0.4);
  border-radius: 0.5rem;
  padding: 0.75rem 1rem;
  overflow-x: auto;
  margin: 0.5em 0;
  font-size: 0.75rem;
  line-height: 1.5;
}

.prose-content code {
  background-color: rgba(42, 42, 47, 0.6);
  padding: 0.1em 0.35em;
  border-radius: 0.25rem;
  font-size: 0.85em;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.prose-content pre code {
  background: none;
  padding: 0;
  border-radius: 0;
  font-size: inherit;
}

.prose-content h1, .prose-content h2, .prose-content h3,
.prose-content h4, .prose-content h5, .prose-content h6 {
  font-weight: 600;
  margin-top: 0.75em;
  margin-bottom: 0.25em;
  color: #f5f5f4;
}

.prose-content h1 { font-size: 1.125rem; }
.prose-content h2 { font-size: 1rem; }
.prose-content h3 { font-size: 0.875rem; }

.prose-content ul, .prose-content ol {
  padding-left: 1.5em;
  margin: 0.375em 0;
}

.prose-content li {
  margin: 0.125em 0;
}

.prose-content ul {
  list-style-type: disc;
}

.prose-content ol {
  list-style-type: decimal;
}

.prose-content blockquote {
  border-left: 3px solid rgba(56, 188, 216, 0.3);
  padding-left: 0.75em;
  margin: 0.5em 0;
  color: #a8a29e;
  font-style: italic;
}

.prose-content a {
  color: #38bcd8;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.prose-content a:hover {
  color: #7dd3e8;
}

.prose-content table {
  width: 100%;
  border-collapse: collapse;
  margin: 0.5em 0;
  font-size: 0.8em;
}

.prose-content th, .prose-content td {
  border: 1px solid rgba(61, 61, 68, 0.4);
  padding: 0.375em 0.5em;
  text-align: left;
}

.prose-content th {
  background-color: rgba(42, 42, 47, 0.6);
  font-weight: 600;
  color: #f5f5f4;
}

.prose-content hr {
  border: none;
  border-top: 1px solid rgba(61, 61, 68, 0.4);
  margin: 0.75em 0;
}

.prose-content img {
  max-width: 100%;
  border-radius: 0.5rem;
}

.prose-content strong {
  font-weight: 600;
  color: #e7e5e4;
}

.prose-content em {
  font-style: italic;
}
</style>
