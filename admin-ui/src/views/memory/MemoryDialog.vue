<template>
  <AppDialog ref="dialogRef" :title="dialogTitle" @save="save">
    <div class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <FormLabel label="Name" :required="true" />
          <FormInput v-model="form.name" placeholder="redis-session" :required="true" />
        </div>
        <div>
          <FormLabel label="Type" />
          <FormSelect v-model="form.type" @update:modelValue="onTypeChange">
            <option v-for="t in typesInCategory" :key="t.type" :value="t.type">{{ t.displayName }}</option>
          </FormSelect>
        </div>
      </div>

      <!-- Dynamic config fields -->
      <div v-for="f in currentFields" :key="f.key" class="space-y-1">
        <FormLabel :label="f.label" :required="f.required" />
        <FormInput
          :modelValue="form.config[f.key] ?? f.default ?? ''"
          @update:modelValue="form.config[f.key] = $event"
          :type="f.type === 'password' ? 'password' : 'text'"
          :placeholder="f.placeholder || ''"
          :mono="f.key === 'connectionString'"
        />
      </div>

      <!-- Embedding (longterm only) -->
      <fieldset v-if="form.category === 'longterm'" class="border border-piedra-700/40 rounded-xl p-4 space-y-3">
        <legend class="text-xs font-medium text-arena-400 px-1.5">Embedding</legend>
        <p class="text-[11px] text-arena-500 -mt-1">Required for semantic search in long-term memory.</p>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <FormLabel label="Backend" />
            <FormSelect v-model="form.embeddingBackend">
              <option value="">(none)</option>
              <option v-for="b in store.backends" :key="b.id" :value="b.id">{{ b.name }}</option>
            </FormSelect>
          </div>
          <div>
            <FormLabel label="Model" />
            <FormInput v-model="form.embeddingModel" placeholder="nomic-embed-text" />
          </div>
        </div>
      </fieldset>
    </div>

    <template #footer>
      <button
        v-if="isEdit"
        type="button"
        @click="testConnection"
        :disabled="testLoading"
        class="flex items-center gap-1.5 px-3 py-2 text-xs rounded-lg border transition-colors"
        :class="testClass"
      >
        <Icon name="bolt" size="xs" />
        <span>{{ testLabel }}</span>
      </button>
      <div class="flex-1" />
      <button type="button" @click="dialogRef?.close()" class="px-4 py-2 text-sm text-arena-400 hover:text-arena-200 hover:bg-piedra-800 rounded-lg transition-colors">
        Cancel
      </button>
      <button type="button" @click="save" class="px-4 py-2 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-sm font-medium rounded-lg transition-colors">
        Save
      </button>
    </template>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { memoryApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'
import Icon from '../../components/Icon.vue'

const emit = defineEmits(['saved'])
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)
const testLoading = ref(false)
const testResult = ref(null)

const form = reactive({
  name: '',
  type: '',
  category: 'session',
  config: {},
  embeddingBackend: '',
  embeddingModel: '',
})

const dialogTitle = computed(() => {
  const catLabel = form.category === 'session' ? 'Session Provider' : 'Long-Term Provider'
  return isEdit.value ? `Edit ${catLabel}` : `New ${catLabel}`
})

const typesInCategory = computed(() =>
  store.memoryTypes.filter(t => t.categories?.includes(form.category))
)

const currentFields = computed(() => {
  const t = store.memoryTypes.find(t => t.type === form.type)
  return t?.fields || []
})

const testLabel = computed(() => {
  if (testLoading.value) return 'Testing...'
  if (!testResult.value) return 'Test Connection'
  return testResult.value.healthy ? '✓ Connected' : `✗ ${testResult.value.detail}`
})

const testClass = computed(() => {
  if (testResult.value?.healthy) return 'text-green-400 border-green-500/30'
  if (testResult.value && !testResult.value.healthy) return 'text-lava-400 border-lava-500/30'
  return 'text-arena-400 border-piedra-700 hover:text-arena-200 hover:bg-piedra-800'
})

function onTypeChange() {
  form.config = {}
  testResult.value = null
}

function open(mem = null, category = null) {
  isEdit.value = !!mem
  editId.value = mem?.id || null
  form.category = mem?.category || category || 'session'
  form.name = mem?.name || ''
  const types = typesInCategory.value
  form.type = mem?.type || types[0]?.type || ''
  form.config = { ...(mem?.config || {}) }
  form.embeddingBackend = mem?.embedding?.backend || ''
  form.embeddingModel = mem?.embedding?.model || ''
  testResult.value = null
  testLoading.value = false
  dialogRef.value?.open()
}

async function testConnection() {
  if (!editId.value) return
  testLoading.value = true
  testResult.value = null
  try {
    testResult.value = await memoryApi.checkHealth(editId.value)
  } catch {
    testResult.value = { healthy: false, detail: 'Check failed' }
  } finally {
    testLoading.value = false
  }
}

async function save() {
  const data = {
    name: form.name,
    type: form.type,
    category: form.category,
    config: {},
  }
  for (const f of currentFields.value) {
    if (form.config[f.key]?.trim?.()) {
      data.config[f.key] = form.config[f.key].trim()
    }
  }
  if (form.category === 'longterm' && form.embeddingBackend) {
    data.embedding = { backend: form.embeddingBackend, model: form.embeddingModel.trim() }
  }
  try {
    if (isEdit.value) {
      await memoryApi.update(editId.value, data)
    } else {
      await memoryApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    alert('Error: ' + e.message)
  }
}

defineExpose({ open })
</script>
