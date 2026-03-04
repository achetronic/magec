<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Client' : 'New Client'" @save="save">
    <div class="space-y-4">
      <div class="flex items-center gap-4">
        <div class="grid grid-cols-2 gap-4 flex-1">
          <div>
            <FormLabel label="Name" :required="true" />
            <FormInput v-model="form.name" placeholder="my-client" :required="true" />
          </div>
          <div>
            <FormLabel label="Type" />
            <FormSelect v-model="form.type" @update:modelValue="onTypeChange">
              <option v-for="t in store.clientTypes" :key="t.type" :value="t.type">{{ t.displayName }}</option>
            </FormSelect>
          </div>
        </div>
        <label class="flex flex-col items-center gap-1 cursor-pointer flex-shrink-0 pt-1">
          <span class="text-[10px] text-arena-500">Enabled</span>
          <div class="relative">
            <input type="checkbox" v-model="form.enabled" class="sr-only peer" />
            <div class="w-9 h-5 bg-piedra-700 rounded-full peer-checked:bg-sol-500/60 transition-colors" />
            <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-arena-400 rounded-full peer-checked:translate-x-4 peer-checked:bg-white transition-transform" />
          </div>
        </label>
      </div>

      <div>
        <FormLabel label="Allowed Agents & Flows" />
        <div v-if="store.agents.length || store.flows.length" class="flex flex-wrap gap-1.5">
          <template v-for="(a, i) in store.agents" :key="a.id">
            <div
              v-if="showAllEntities || i < maxVisibleEntities"
              class="inline-flex items-center gap-1.5 rounded-lg border cursor-pointer transition-all text-xs"
              :class="form.allowedAgents.includes(a.id)
                ? 'bg-sol-500/10 border-sol-500/40 text-sol-300'
                : 'bg-piedra-800/60 border-piedra-700/50 text-arena-400 hover:border-piedra-600'"
            >
              <button
                v-if="hasDefaultAgent && form.allowedAgents.includes(a.id)"
                type="button"
                @click.stop="toggleDefault(a.id)"
                class="w-3 h-3 rounded border flex-shrink-0 transition-all ml-2"
                :class="form.config.defaultAgent === a.id
                  ? 'bg-sol-500 border-sol-500'
                  : 'bg-transparent border-sol-500/40 hover:border-sol-500'"
              />
              <span class="py-1 pr-2.5" :class="{'pl-1.5': !(hasDefaultAgent && form.allowedAgents.includes(a.id)), 'pl-1': hasDefaultAgent && form.allowedAgents.includes(a.id)}" @click="toggleAgent(a.id)">{{ a.name || a.id }}</span>
            </div>
          </template>
          <template v-for="(f, i) in store.flows" :key="f.id">
            <div
              v-if="showAllEntities || (store.agents.length + i) < maxVisibleEntities"
              class="inline-flex items-center gap-1.5 rounded-lg border cursor-pointer transition-all text-xs"
              :class="form.allowedAgents.includes(f.id)
                ? 'bg-rose-500/10 border-rose-500/40 text-rose-300'
                : 'bg-piedra-800/60 border-piedra-700/50 text-arena-400 hover:border-piedra-600'"
            >
              <button
                v-if="hasDefaultAgent && form.allowedAgents.includes(f.id)"
                type="button"
                @click.stop="toggleDefault(f.id)"
                class="w-3 h-3 rounded border flex-shrink-0 transition-all ml-2"
                :class="form.config.defaultAgent === f.id
                  ? 'bg-sol-500 border-sol-500'
                  : 'bg-transparent border-sol-500/40 hover:border-sol-500'"
              />
              <span class="py-1 pr-2.5" :class="{'pl-1.5': !(hasDefaultAgent && form.allowedAgents.includes(f.id)), 'pl-1': hasDefaultAgent && form.allowedAgents.includes(f.id)}" @click="toggleAgent(f.id)">⤳ {{ f.name || f.id }}</span>
            </div>
          </template>
          <button
            v-if="totalEntities > maxVisibleEntities"
            type="button"
            @click="showAllEntities = !showAllEntities"
            class="px-2.5 py-1 rounded-lg border border-piedra-700/50 text-[11px] text-arena-500 hover:text-arena-300 hover:border-piedra-600 transition-all cursor-pointer"
          >
            {{ showAllEntities ? 'Less' : `+${totalEntities - maxVisibleEntities} more` }}
          </button>
        </div>
        <p v-else class="text-xs text-arena-500">No agents or flows defined yet</p>
        <p class="text-[10px] text-arena-500 mt-1">Agents and flows this client can interact with.</p>
        <p v-if="hasDefaultAgent" class="text-[10px] text-arena-500 mt-0.5">Click the <span class="inline-block w-2 h-2 rounded bg-sol-500 align-middle mx-0.5" /> square on a selected agent to set it as default.</p>
      </div>

      <!-- Dynamic config from JSON Schema -->
      <template v-for="(propSchema, key) in mainProperties" :key="key">
        <!-- Boolean → toggle -->
        <div v-if="propSchema.type === 'boolean'">
          <label class="flex items-center gap-2 cursor-pointer">
            <div class="relative">
              <input type="checkbox" v-model="form.config[key]" class="sr-only peer" />
              <div class="w-9 h-5 bg-piedra-700 rounded-full peer-checked:bg-teal-500/60 transition-colors" />
              <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-arena-400 rounded-full peer-checked:translate-x-4 peer-checked:bg-white transition-transform" />
            </div>
            <span class="text-xs text-arena-300">{{ propSchema.title || key }}</span>
          </label>
          <p v-if="propSchema.description" class="text-[10px] text-arena-500 mt-1">{{ propSchema.description }}</p>
        </div>

        <!-- Entity reference → select from store -->
        <div v-else-if="propSchema['x-entity']">
          <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
          <FormSelect :modelValue="form.config[key] ?? ''" @update:modelValue="form.config[key] = $event">
            <option value="" disabled>Select a {{ propSchema.title?.toLowerCase() || key }}</option>
            <option v-for="item in entityItems(propSchema['x-entity'])" :key="item.id" :value="item.id">{{ item.name || item.id }}</option>
          </FormSelect>
        </div>

        <!-- Enum → select -->
        <div v-else-if="propSchema.enum">
          <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
          <FormSelect :modelValue="form.config[key] ?? propSchema.default ?? ''" @update:modelValue="form.config[key] = $event">
            <option v-for="o in propSchema.enum" :key="o" :value="o">{{ o }}</option>
          </FormSelect>
        </div>

        <!-- Array → text input rendered as comma-separated values -->
        <div v-else-if="propSchema.type === 'array'">
          <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
          <FormInput
            :modelValue="arrayToCSV(form.config[key])"
            @update:modelValue="form.config[key] = csvToArray(propSchema, $event)"
            :placeholder="propSchema['x-placeholder'] || ''"
          />
          <p v-if="propSchema.description" class="text-[10px] text-arena-500 mt-1">{{ propSchema.description }}</p>
        </div>

        <!-- String → text/password input -->
        <div v-else>
          <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
          <FormInput
            :modelValue="form.config[key] ?? propSchema.default ?? ''"
            @update:modelValue="form.config[key] = $event"
            :type="propSchema['x-format'] === 'password' ? 'password' : 'text'"
            :placeholder="propSchema['x-placeholder'] || ''"
          />
          <p v-if="propSchema.description" class="text-[10px] text-arena-500 mt-1">{{ propSchema.description }}</p>
        </div>
      </template>

      <!-- Options (responseMode, threadHistoryLimit, etc.) -->
      <details v-if="Object.keys(optionProperties).length" class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>Options</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4 space-y-4">
          <template v-for="(propSchema, key) in optionProperties" :key="key">
            <!-- Enum → select -->
            <div v-if="propSchema.enum">
              <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
              <FormSelect :modelValue="form.config[key] ?? propSchema.default ?? ''" @update:modelValue="form.config[key] = $event">
                <option v-for="o in propSchema.enum" :key="o" :value="o">{{ o }}</option>
              </FormSelect>
              <p v-if="propSchema.description" class="text-[10px] text-arena-500 mt-1">{{ propSchema.description }}</p>
            </div>
            <!-- Default: string/number input -->
            <div v-else>
              <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
              <FormInput
                :modelValue="form.config[key] ?? propSchema.default ?? ''"
                @update:modelValue="form.config[key] = $event"
                :placeholder="propSchema['x-placeholder'] || ''"
              />
              <p v-if="propSchema.description" class="text-[10px] text-arena-500 mt-1">{{ propSchema.description }}</p>
            </div>
          </template>
        </div>
      </details>

      <!-- Permissions (allowedUsers / allowedChannels / allowedChats) -->
      <details v-if="Object.keys(permissionProperties).length" class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>Permissions</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4 space-y-4">
          <template v-for="(propSchema, key) in permissionProperties" :key="key">
            <div v-if="key === 'allowedChats'">
              <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
              <div class="space-y-2">
                <div
                  v-for="(rule, idx) in getAllowedChatRows()"
                  :key="idx"
                  class="grid grid-cols-[1fr_1fr_auto] gap-2"
                >
                  <FormInput
                    :modelValue="rule.chatId"
                    @update:modelValue="updateAllowedChatRule(idx, 'chatId', $event)"
                    placeholder="Chat ID (e.g. -1001234567890)"
                  />
                  <FormInput
                    :modelValue="rule.threadId"
                    @update:modelValue="updateAllowedChatRule(idx, 'threadId', $event)"
                    placeholder="Thread ID (optional)"
                  />
                  <button
                    type="button"
                    @click="removeAllowedChatRule(idx)"
                    title="Remove rule"
                    aria-label="Remove rule"
                    class="px-2.5 py-2 bg-piedra-800 hover:bg-lava-500/20 border border-lava-500/40 rounded-lg text-lava-300 hover:text-lava-200 transition-colors"
                  >
                    <Icon name="trash" size="md" />
                  </button>
                </div>
                <button
                  type="button"
                  @click="addAllowedChatRule()"
                  class="px-3 py-2 bg-piedra-800 hover:bg-piedra-700 border border-piedra-700 rounded-lg text-xs text-arena-300 transition-colors"
                >
                  + Add chat rule
                </button>
              </div>
              <p v-if="propSchema.description" class="text-[10px] text-arena-500 mt-1">{{ propSchema.description }}</p>
            </div>
            <div v-else>
              <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
              <FormInput
                :modelValue="arrayToCSV(form.config[key])"
                @update:modelValue="form.config[key] = csvToArray(propSchema, $event)"
                :placeholder="propSchema['x-placeholder'] || ''"
              />
              <p v-if="propSchema.description" class="text-[10px] text-arena-500 mt-1">{{ propSchema.description }}</p>
            </div>
          </template>
        </div>
      </details>

      <!-- Token (edit only) -->
      <div v-if="isEdit && form.token">
        <FormLabel label="Token" />
        <div class="flex gap-2">
          <FormInput :modelValue="form.token" :type="tokenVisible ? 'text' : 'password'" :readonly="true" mono input-class="select-all" />
          <button type="button" @click="tokenVisible = !tokenVisible" class="px-3 py-2 bg-piedra-800 hover:bg-piedra-700 border border-piedra-700 rounded-lg text-xs text-arena-300 transition-colors flex-shrink-0">
            <Icon name="eye" size="md" />
          </button>
          <button type="button" @click="copyToken" class="px-3 py-2 bg-piedra-800 hover:bg-piedra-700 border border-piedra-700 rounded-lg text-xs text-arena-300 transition-colors flex-shrink-0">
            <Icon name="copy" size="md" />
          </button>
          <button type="button" @click="regenerateToken" class="px-3 py-2 bg-piedra-800 hover:bg-lava-500/20 border border-piedra-700 rounded-lg text-xs text-arena-300 transition-colors flex-shrink-0">
            <Icon name="refresh" size="md" />
          </button>
          <button
            v-if="form.type === 'telegram'"
            type="button"
            @click="sendTelegramTest"
            :disabled="isTestingTelegram"
            class="px-3 py-2 bg-piedra-800 hover:bg-teal-500/20 disabled:opacity-50 disabled:cursor-not-allowed border border-piedra-700 rounded-lg text-xs text-arena-300 transition-colors flex-shrink-0"
          >
            {{ isTestingTelegram ? 'Testing...' : 'Test' }}
          </button>
        </div>
        <p class="text-[10px] text-arena-500 mt-1">Use as <code class="text-arena-400">Authorization: Bearer &lt;token&gt;</code></p>

        <!-- Webhook endpoint hint -->
        <p v-if="form.type === 'webhook'" class="text-[10px] text-arena-500 mt-2">
          Endpoint: <code class="text-arena-400">POST /api/v1/webhooks/{{ editId }}</code>
        </p>
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, computed, inject } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { clientsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'
import Icon from '../../components/Icon.vue'
import { buildAllowedChatsPayload } from './allowedChatsRules.js'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)
const tokenVisible = ref(false)
const isTestingTelegram = ref(false)
const showAllEntities = ref(false)
const maxVisibleEntities = 6

const totalEntities = computed(() => store.agents.length + store.flows.length)

const form = reactive({
  name: '',
  type: 'direct',
  enabled: true,
  allowedAgents: [],
  config: {},
  token: '',
})

const currentSchema = computed(() => {
  const t = store.clientTypes.find(t => t.type === form.type)
  return t?.configSchema || {}
})

const allProperties = computed(() => {
  return currentSchema.value.properties || {}
})

const activeOneOfBranch = computed(() => {
  const branches = currentSchema.value.oneOf
  if (!branches) return null
  for (const branch of branches) {
    const props = branch.properties || {}
    let match = true
    for (const [key, schema] of Object.entries(props)) {
      if ('const' in schema) {
        const val = form.config[key] ?? getDefault(key)
        if (!jsonEqual(val, schema.const)) {
          match = false
          break
        }
      }
    }
    if (match) return branch
  }
  return null
})

const visibleProperties = computed(() => {
  const props = allProperties.value
  const branch = activeOneOfBranch.value
  // defaultAgent is managed via the agent chips UI, not the dynamic form
  const skip = new Set(['defaultAgent'])
  if (!branch) return Object.fromEntries(Object.entries(props).filter(([k]) => !skip.has(k)))

  const branchProps = branch.properties || {}
  const result = {}
  for (const [key, schema] of Object.entries(props)) {
    if (skip.has(key)) continue
    const branchSchema = branchProps[key]
    if (branchSchema && 'const' in branchSchema) {
      result[key] = schema
      continue
    }
    const isExcluded = isExcludedByOtherBranches(key)
    if (!isExcluded || key in branchProps) {
      result[key] = schema
    }
  }
  return result
})

const PERMISSION_KEYS = new Set(['allowedUsers', 'allowedChannels', 'allowedChats'])
const OPTION_KEYS = new Set(['responseMode', 'threadHistoryLimit'])

const mainProperties = computed(() =>
  Object.fromEntries(Object.entries(visibleProperties.value).filter(([k]) => !PERMISSION_KEYS.has(k) && !OPTION_KEYS.has(k)))
)

const permissionProperties = computed(() =>
  Object.fromEntries(Object.entries(visibleProperties.value).filter(([k]) => PERMISSION_KEYS.has(k)))
)

const optionProperties = computed(() =>
  Object.fromEntries(Object.entries(visibleProperties.value).filter(([k]) => OPTION_KEYS.has(k)))
)

function isExcludedByOtherBranches(key) {
  const branches = currentSchema.value.oneOf
  if (!branches) return false
  for (const branch of branches) {
    if (branch === activeOneOfBranch.value) continue
    const req = branch.required || []
    if (req.includes(key)) return true
  }
  return false
}

const hasDefaultAgent = computed(() => 'defaultAgent' in allProperties.value)

function toggleAgent(id) {
  const idx = form.allowedAgents.indexOf(id)
  if (idx === -1) {
    form.allowedAgents.push(id)
  } else {
    form.allowedAgents.splice(idx, 1)
    if (form.config.defaultAgent === id) form.config.defaultAgent = ''
  }
}

function toggleDefault(id) {
  form.config.defaultAgent = form.config.defaultAgent === id ? '' : id
}

function isFieldRequired(key) {
  const topRequired = currentSchema.value.required || []
  if (topRequired.includes(key)) return true
  const branch = activeOneOfBranch.value
  if (branch) {
    const branchRequired = branch.required || []
    if (branchRequired.includes(key)) return true
  }
  return false
}

function getDefault(key) {
  const prop = allProperties.value[key]
  if (!prop) return undefined
  if ('default' in prop) return prop.default
  if (prop.type === 'boolean') return false
  return undefined
}

function entityItems(entityKey) {
  const map = {
    commands: store.commands,
    agents: store.agents,
    backends: store.backends,
    memory: store.memory,
    mcps: store.mcps,
    flows: store.flows,
  }
  return map[entityKey] || []
}

function jsonEqual(a, b) {
  return JSON.stringify(a) === JSON.stringify(b)
}

function arrayToCSV(val) {
  if (Array.isArray(val)) return val.join(', ')
  return val ?? ''
}

function csvToArray(propSchema, val) {
  const itemType = propSchema.items?.type
  const parts = val.toString().split(',').map(s => s.trim()).filter(Boolean)
  if (itemType === 'integer' || itemType === 'number') {
    return parts.map(Number).filter(n => !isNaN(n))
  }
  return parts
}

function parseLegacyAllowedChatRule(value) {
  const match = value?.toString().match(/^\s*(-?\d+)(?:\s*-\s*(\d+))?\s*$/)
  if (!match) return null
  return { chatId: match[1], threadId: match[2] || '' }
}

function normalizeAllowedChatRows(val) {
  if (!Array.isArray(val)) return []
  return val.map(rule => {
    if (typeof rule === 'string') {
      const parsed = parseLegacyAllowedChatRule(rule)
      if (!parsed) return null
      return parsed
    }
    if (rule && typeof rule === 'object') {
      return {
        chatId: rule.chatId !== undefined && rule.chatId !== null ? String(rule.chatId) : '',
        threadId: rule.threadId !== undefined && rule.threadId !== null ? String(rule.threadId) : '',
      }
    }
    return null
  }).filter(Boolean)
}

function getAllowedChatRows() {
  return normalizeAllowedChatRows(form.config.allowedChats)
}

function setAllowedChatRows(rows) {
  form.config.allowedChats = rows
}

function addAllowedChatRule() {
  const rows = getAllowedChatRows()
  rows.push({ chatId: '', threadId: '' })
  setAllowedChatRows(rows)
}

function removeAllowedChatRule(index) {
  const rows = getAllowedChatRows()
  rows.splice(index, 1)
  setAllowedChatRows(rows)
}

function updateAllowedChatRule(index, field, value) {
  const rows = getAllowedChatRows()
  if (!rows[index]) return
  rows[index][field] = value?.toString().trim() ?? ''
  setAllowedChatRows(rows)
}

function onTypeChange() {
  form.config = {}
  const props = allProperties.value
  for (const [key, schema] of Object.entries(props)) {
    if ('default' in schema) {
      form.config[key] = schema.default
    }
  }
}

function open(client = null) {
  isEdit.value = !!client
  editId.value = client?.id || null
  form.name = client?.name || ''
  form.type = client?.type || 'direct'
  form.enabled = client?.enabled ?? true
  form.allowedAgents = [...(client?.allowedAgents || [])]
  form.config = { ...(client?.config?.[client?.type] || {}) }
  if (form.type === 'telegram') {
    form.config.allowedChats = normalizeAllowedChatRows(form.config.allowedChats)
  }
  form.token = client?.token || ''
  tokenVisible.value = false
  showAllEntities.value = false
  dialogRef.value?.open()
}

function copyToken() {
  const text = form.token
  if (navigator.clipboard?.writeText) {
    navigator.clipboard.writeText(text).then(
      () => toast.success('Token copied'),
      () => { fallbackCopy(text); toast.success('Token copied') }
    )
  } else {
    fallbackCopy(text)
    toast.success('Token copied')
  }
}

function fallbackCopy(text) {
  const ta = document.createElement('textarea')
  ta.value = text
  ta.style.position = 'fixed'
  ta.style.opacity = '0'
  document.body.appendChild(ta)
  ta.select()
  document.execCommand('copy')
  document.body.removeChild(ta)
}

async function regenerateToken() {
  if (!editId.value) return
  if (!confirm('Regenerate token? The old token will stop working immediately.')) return
  try {
    const updated = await clientsApi.regenerateToken(editId.value)
    form.token = updated.token
    await store.refresh()
  } catch (e) {
    toast.error(e.message)
  }
}

async function sendTelegramTest() {
  if (!editId.value || form.type !== 'telegram' || isTestingTelegram.value) return
  isTestingTelegram.value = true
  try {
    const telegramConfig = buildTypeConfig().telegram || {}
    const result = await clientsApi.telegramTest(editId.value, { config: telegramConfig })
    if (result.failed > 0) {
      toast.error(`Test completed with errors. Sent ${result.sent}/${result.attempted}.`)
      return
    }
    toast.success(`Test message sent to ${result.sent} destination(s).`)
  } catch (e) {
    toast.error(e.message)
  } finally {
    isTestingTelegram.value = false
  }
}

function buildTypeConfig() {
  const config = {}
  const schema = currentSchema.value
  const props = schema.properties || {}

  if (Object.keys(props).length) {
    const typeCfg = {}
    for (const [key, propSchema] of Object.entries(props)) {
      const val = form.config[key]
      if (propSchema.type === 'boolean') {
        typeCfg[key] = !!val
      } else if (propSchema.type === 'array') {
        if (key === 'allowedChats') {
          const rules = buildAllowedChatsPayload(val)
          if (rules.length) {
            typeCfg[key] = rules
          }
        } else if (Array.isArray(val) && val.length) {
          typeCfg[key] = val
        }
      } else if (propSchema.type === 'integer' || propSchema.type === 'number') {
        const n = Number(val)
        if (!isNaN(n) && val !== '' && val !== null && val !== undefined) {
          typeCfg[key] = propSchema.type === 'integer' ? Math.trunc(n) : n
        }
      } else if (val?.toString().trim()) {
        typeCfg[key] = val.toString().trim()
      }
    }
    config[form.type] = typeCfg
  }

  return config
}

async function save() {
  try {
    const config = buildTypeConfig()
    const data = {
      name: form.name.trim(),
      type: form.type,
      allowedAgents: form.allowedAgents,
      enabled: form.enabled,
      config,
    }

    if (isEdit.value) {
      await clientsApi.update(editId.value, data)
    } else {
      await clientsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
