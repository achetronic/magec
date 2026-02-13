<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Client' : 'New Client'" @save="save">
    <div class="space-y-4">
      <div class="flex items-center gap-4">
        <div class="grid grid-cols-2 gap-4 flex-1">
          <div>
            <FormLabel label="Name" :required="true" />
            <FormInput v-model="form.name" placeholder="tablet-cocina" :required="true" />
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
        <FormLabel label="Allowed Agents" />
        <div v-if="store.agents.length" class="flex flex-wrap gap-1.5">
          <label
            v-for="a in store.agents" :key="a.id"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg border cursor-pointer transition-all text-xs"
            :class="form.allowedAgents.includes(a.id)
              ? 'bg-sol-500/10 border-sol-500/40 text-sol-300'
              : 'bg-piedra-800/60 border-piedra-700/50 text-arena-400 hover:border-piedra-600'"
          >
            <input type="checkbox" :value="a.id" v-model="form.allowedAgents" class="hidden" />
            <span>{{ a.name || a.id }}</span>
          </label>
        </div>
        <p v-else class="text-xs text-arena-500">No agents defined yet</p>
        <p class="text-[10px] text-arena-500 mt-1">First selected agent is the default.</p>
      </div>

      <!-- Dynamic config fields -->
      <div v-for="f in currentFields" :key="f.key">
        <FormLabel :label="f.label" :required="f.required" />
        <FormSelect v-if="f.type === 'select'" :modelValue="form.config[f.key] ?? f.default ?? ''" @update:modelValue="form.config[f.key] = $event">
          <option v-for="o in (f.options || '').split(',')" :key="o" :value="o">{{ o }}</option>
        </FormSelect>
        <FormInput v-else
          :modelValue="form.config[f.key] ?? f.default ?? ''"
          @update:modelValue="form.config[f.key] = $event"
          :type="f.type === 'password' ? 'password' : 'text'"
          :placeholder="f.placeholder || ''"
        />
      </div>

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

const form = reactive({
  name: '',
  type: 'device',
  enabled: true,
  allowedAgents: [],
  config: {},
  token: '',
})

const currentFields = computed(() => {
  const t = store.clientTypes.find(t => t.type === form.type)
  return t?.fields || []
})

function onTypeChange() {
  form.config = {}
}

function open(client = null) {
  isEdit.value = !!client
  editId.value = client?.id || null
  form.name = client?.name || ''
  form.type = client?.type || 'device'
  form.enabled = client?.enabled ?? true
  form.allowedAgents = [...(client?.allowedAgents || [])]
  form.config = { ...(client?.config?.[client?.type] || {}) }
  form.token = client?.token || ''
  tokenVisible.value = false
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
  const typeInfo = store.clientTypes.find(t => t.type === form.type)
  if (typeInfo?.fields?.length) {
    const typeCfg = {}
    for (const f of typeInfo.fields) {
      const val = form.config[f.key]
      if (val?.toString().trim()) {
        if (f.key === 'allowedUsers' || f.key === 'allowedChats') {
          typeCfg[f.key] = val.toString().split(',').map(s => s.trim()).filter(Boolean).map(Number).filter(n => !isNaN(n))
        } else {
          typeCfg[f.key] = val.toString().trim()
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
