<template>
  <AppDialog ref="dialogRef" :title="dialogTitle" @save="save">
    <div class="space-y-4">
      <div class="flex items-center gap-1 p-0.5 rounded-lg bg-piedra-800 w-fit">
        <button
          v-for="cat in categories" :key="cat.value"
          type="button"
          @click="setCategory(cat.value)"
          :disabled="isEdit"
          class="px-3 py-1.5 text-xs font-medium rounded-md transition-colors cursor-pointer disabled:cursor-not-allowed"
          :class="form.category === cat.value ? 'bg-piedra-700 text-arena-100' : 'text-arena-500 hover:text-arena-300'"
        >{{ cat.label }}</button>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <FormLabel label="Name" :required="true" />
          <FormInput v-model="form.name" placeholder="My Memory" :required="true" />
        </div>
        <div>
          <FormLabel label="Type" />
          <FormSelect v-model="form.type" @update:modelValue="onTypeChange">
            <option v-for="t in typesInCategory" :key="t.type" :value="t.type">{{ t.displayName }}</option>
          </FormSelect>
        </div>
      </div>

      <!-- Dynamic config from JSON Schema -->
      <template v-for="(propSchema, key) in visibleProperties" :key="key">
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

        <div v-else-if="propSchema['x-entity']">
          <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
          <FormSelect :modelValue="form.config[key] ?? ''" @update:modelValue="form.config[key] = $event">
            <option value="" disabled>Select a {{ propSchema.title?.toLowerCase() || key }}</option>
            <option v-for="item in entityItems(propSchema['x-entity'])" :key="item.id" :value="item.id">{{ item.name || item.id }}</option>
          </FormSelect>
        </div>

        <div v-else-if="propSchema.enum">
          <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
          <FormSelect :modelValue="form.config[key] ?? propSchema.default ?? ''" @update:modelValue="form.config[key] = $event">
            <option v-for="o in propSchema.enum" :key="o" :value="o">{{ o }}</option>
          </FormSelect>
        </div>

        <div v-else>
          <FormLabel :label="propSchema.title || key" :required="isFieldRequired(key)" />
          <FormInput
            :modelValue="form.config[key] ?? propSchema.default ?? ''"
            @update:modelValue="form.config[key] = $event"
            :type="propSchema['x-format'] === 'password' ? 'password' : 'text'"
            :placeholder="propSchema['x-placeholder'] || ''"
            :mono="key === 'connectionString'"
          />
          <p v-if="propSchema.description" class="text-[10px] text-arena-500 mt-1">{{ propSchema.description }}</p>
        </div>
      </template>

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
import { ref, reactive, computed, inject } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { memoryApi } from '../../lib/api/index.js'
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

const categories = [
  { value: 'session', label: 'Session' },
  { value: 'longterm', label: 'Long-term' },
]

const dialogTitle = computed(() => isEdit.value ? 'Edit Provider' : 'New Provider')

function setCategory(cat) {
  if (isEdit.value) return
  form.category = cat
  const types = typesInCategory.value
  form.type = types[0]?.type || ''
  form.config = {}
  onTypeChange()
}

const typesInCategory = computed(() =>
  store.memoryTypes.filter(t => t.categories?.includes(form.category))
)

const currentSchema = computed(() => {
  const t = store.memoryTypes.find(t => t.type === form.type)
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
  const props = allProperties.value
  for (const [key, schema] of Object.entries(props)) {
    if ('default' in schema) {
      form.config[key] = schema.default
    }
  }
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
  for (const [key, propSchema] of Object.entries(visibleProperties.value)) {
    const val = form.config[key]
    if (typeof val === 'boolean') {
      data.config[key] = val
    } else if (typeof val === 'string' && val.trim()) {
      data.config[key] = val.trim()
    } else if (val !== undefined && val !== null && val !== '') {
      data.config[key] = val
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
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
