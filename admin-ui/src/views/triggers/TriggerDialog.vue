<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Trigger' : 'New Trigger'" @save="save">
    <div class="space-y-4">
      <div class="flex items-center gap-4">
        <div class="flex-1">
          <FormLabel label="Name" :required="true" />
          <FormInput v-model="form.name" placeholder="nightly-report" :required="true" />
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
        <FormLabel label="Description" />
        <FormInput v-model="form.description" placeholder="What this trigger does..." />
      </div>

      <div>
        <FormLabel label="Type" :required="true" />
        <div class="flex gap-2">
          <button type="button" v-for="t in triggerTypes" :key="t.value"
            @click="form.type = t.value"
            class="flex-1 px-3 py-2 text-xs font-medium rounded-lg border transition-all cursor-pointer text-center"
            :class="form.type === t.value
              ? 'bg-teal-500/15 text-teal-300 border-teal-500/30'
              : 'bg-piedra-800 text-arena-500 border-piedra-700/40 hover:border-piedra-600 hover:text-arena-300'"
          >{{ t.label }}</button>
        </div>
      </div>

      <!-- Cron config -->
      <div v-if="form.type === 'cron'">
        <FormLabel label="Schedule" :required="true" />
        <FormInput v-model="form.schedule" placeholder="0 9 * * *" :required="true" mono />
        <p class="text-[10px] text-arena-500 mt-1">Standard cron expression (min hour day month weekday)</p>
      </div>

      <!-- Webhook config -->
      <div v-if="form.type === 'webhook'" class="space-y-3">
        <label class="flex items-center gap-2 cursor-pointer">
          <div class="relative">
            <input type="checkbox" v-model="form.passthrough" class="sr-only peer" />
            <div class="w-9 h-5 bg-piedra-700 rounded-full peer-checked:bg-teal-500/60 transition-colors" />
            <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-arena-400 rounded-full peer-checked:translate-x-4 peer-checked:bg-white transition-transform" />
          </div>
          <span class="text-xs text-arena-300">Passthrough mode</span>
        </label>
        <p class="text-[10px] text-arena-500">When enabled, the prompt comes from the webhook request body instead of a command.</p>
      </div>

      <!-- Command (not shown for passthrough webhooks) -->
      <div v-if="!(form.type === 'webhook' && form.passthrough)">
        <FormLabel label="Command" :required="true" />
        <FormSelect v-model="form.commandId">
          <option value="" disabled>Select a command</option>
          <option v-for="c in store.commands" :key="c.id" :value="c.id">{{ c.name }}</option>
        </FormSelect>
      </div>

      <!-- Agent -->
      <div v-if="!(form.type === 'webhook' && form.passthrough)">
        <FormLabel label="Agent" :required="true" />
        <FormSelect v-model="form.agentId">
          <option value="" disabled>Select an agent</option>
          <option v-for="a in store.agents" :key="a.id" :value="a.id">{{ a.name || a.id }}</option>
        </FormSelect>
      </div>

      <!-- Agent for passthrough (optional override) -->
      <div v-if="form.type === 'webhook' && form.passthrough">
        <FormLabel label="Default Agent (optional)" />
        <FormSelect v-model="form.agentId">
          <option value="">(from request)</option>
          <option v-for="a in store.agents" :key="a.id" :value="a.id">{{ a.name || a.id }}</option>
        </FormSelect>
        <p class="text-[10px] text-arena-500 mt-1">If empty, the request body must include the agent ID.</p>
      </div>

      <!-- Client (auth token) -->
      <div>
        <FormLabel label="Client (auth)" />
        <FormSelect v-model="form.clientId">
          <option value="">(none)</option>
          <option v-for="cl in enabledClients" :key="cl.id" :value="cl.id">{{ cl.name }}</option>
        </FormSelect>
        <p class="text-[10px] text-arena-500 mt-1">The client whose token will be used to authenticate against the agent API.</p>
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { triggersApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'

const emit = defineEmits(['saved'])
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

const triggerTypes = [
  { value: 'cron', label: 'Cron' },
  { value: 'webhook', label: 'Webhook' },
]

const form = reactive({
  name: '',
  description: '',
  type: 'cron',
  enabled: true,
  schedule: '',
  passthrough: false,
  commandId: '',
  agentId: '',
  clientId: '',
})

const enabledClients = computed(() => store.clients.filter(c => c.enabled))

function open(trigger = null) {
  isEdit.value = !!trigger
  editId.value = trigger?.id || null
  form.name = trigger?.name || ''
  form.description = trigger?.description || ''
  form.type = trigger?.type || 'cron'
  form.enabled = trigger?.enabled ?? true
  form.schedule = trigger?.cron?.schedule || ''
  form.passthrough = trigger?.webhook?.passthrough || false
  form.commandId = trigger?.commandId || ''
  form.agentId = trigger?.agentId || ''
  form.clientId = trigger?.clientId || ''
  dialogRef.value?.open()
}

async function save() {
  const data = {
    name: form.name.trim(),
    description: form.description.trim(),
    type: form.type,
    enabled: form.enabled,
    agentId: form.agentId,
    commandId: form.commandId,
    clientId: form.clientId,
  }

  if (form.type === 'cron') {
    data.cron = { schedule: form.schedule.trim() }
  } else if (form.type === 'webhook') {
    data.webhook = { passthrough: form.passthrough }
    if (form.passthrough) {
      data.commandId = ''
    }
  }

  try {
    if (isEdit.value) {
      await triggersApi.update(editId.value, data)
    } else {
      await triggersApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    alert('Error: ' + e.message)
  }
}

defineExpose({ open })
</script>
