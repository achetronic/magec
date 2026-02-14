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
            <label
              v-if="showAllEntities || i < maxVisibleEntities"
              class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg border cursor-pointer transition-all text-xs"
              :class="form.allowedAgents.includes(a.id)
                ? 'bg-sol-500/10 border-sol-500/40 text-sol-300'
                : 'bg-piedra-800/60 border-piedra-700/50 text-arena-400 hover:border-piedra-600'"
            >
              <input type="checkbox" :value="a.id" v-model="form.allowedAgents" class="hidden" />
              <span>{{ a.name || a.id }}</span>
            </label>
          </template>
          <template v-for="(f, i) in store.flows" :key="f.id">
            <label
              v-if="showAllEntities || (store.agents.length + i) < maxVisibleEntities"
              class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg border cursor-pointer transition-all text-xs"
              :class="form.allowedAgents.includes(f.id)
                ? 'bg-rose-500/10 border-rose-500/40 text-rose-300'
                : 'bg-piedra-800/60 border-piedra-700/50 text-arena-400 hover:border-piedra-600'"
            >
              <input type="checkbox" :value="f.id" v-model="form.allowedAgents" class="hidden" />
              <span>⤳ {{ f.name || f.id }}</span>
            </label>
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
        <p class="text-[10px] text-arena-500 mt-1">Agents and flows this client can interact with. Cron/webhook clients run commands against all selected items.</p>
      </div>

      <!-- Dynamic config from JSON Schema -->
      <template v-for="(propSchema, key) in visibleProperties" :key="key">
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

const emit = defineEmits(['saved'])
const toast = inject('toast')
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)
const tokenVisible = ref(false)
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
  if (!branch) return props

  const branchProps = branch.properties || {}
  const result = {}
  for (const [key, schema] of Object.entries(props)) {
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
  form.token = client?.token || ''
  tokenVisible.value = false
  showAllEntities.value = false
  dialogRef.value?.open()
}

function copyToken() {
  navigator.clipboard.writeText(form.token)
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

async function save() {
  const config = {}
  const schema = currentSchema.value
  const props = schema.properties || {}

  if (Object.keys(props).length) {
    const typeCfg = {}
    for (const [key, propSchema] of Object.entries(props)) {
      const val = form.config[key]
      if (propSchema.type === 'boolean') {
        typeCfg[key] = !!val
      } else if (val?.toString().trim()) {
        if (key === 'allowedUsers' || key === 'allowedChats') {
          typeCfg[key] = val.toString().split(',').map(s => s.trim()).filter(Boolean).map(Number).filter(n => !isNaN(n))
        } else {
          typeCfg[key] = val.toString().trim()
        }
      }
    }
    config[form.type] = typeCfg
  }

  const data = {
    name: form.name.trim(),
    type: form.type,
    allowedAgents: form.allowedAgents,
    enabled: form.enabled,
    config,
  }
  try {
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
