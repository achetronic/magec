<template>
  <div class="max-w-2xl space-y-6">
    <!-- Header -->
    <div>
      <h3 class="text-sm font-semibold text-arena-100">Flows</h3>
      <p class="text-xs text-arena-500 mt-1">Starlark code-node libraries and execution limits.</p>
    </div>

    <!-- Script Libraries -->
    <Card color="blue">
      <div class="flex items-start gap-4">
        <div class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 bg-blue-500/10">
          <Icon name="flow" size="md" class="text-blue-400" />
        </div>
        <div class="flex-1 min-w-0">
          <h4 class="text-[13px] font-medium text-arena-100">Script libraries</h4>
          <p class="text-xs text-arena-500 mt-0.5">
            Modules available to code nodes. All enabled by default; turn off what this deployment should not allow.
          </p>

          <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2 mt-4">
            <div
              v-for="mod in ALL_MODULES"
              :key="mod"
              class="flex items-center justify-between p-2 rounded-lg bg-piedra-950/40 border border-piedra-800/60"
            >
              <div
                v-if="isSensitive(mod)"
                class="flex items-center gap-1.5 min-w-0"
                title="Network or filesystem access"
              >
                <svg
                  class="w-3.5 h-3.5 text-amber-400 flex-shrink-0"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fill-rule="evenodd"
                    d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                    clip-rule="evenodd"
                  />
                </svg>
                <span class="font-mono text-xs text-amber-400 truncate">{{ mod }}</span>
              </div>
              <div v-else class="flex items-center gap-1.5 min-w-0">
                <span class="font-mono text-xs text-arena-300 truncate">{{ mod }}</span>
              </div>
              <FormToggle
                :model-value="isLibraryEnabled(mod)"
                @update:model-value="toggleLibrary(mod)"
              />
            </div>
          </div>
        </div>
      </div>
    </Card>

    <!-- Execution Limits -->
    <Card color="blue">
      <div class="flex items-start gap-4">
        <div class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 bg-blue-500/10">
          <Icon name="bolt" size="md" class="text-blue-400" />
        </div>
        <div class="flex-1 min-w-0 space-y-4">
          <div>
            <h4 class="text-[13px] font-medium text-arena-100">Execution limits</h4>
            <p class="text-xs text-arena-500 mt-0.5">
              Control execution time and output size of code nodes.
            </p>
          </div>

          <!-- Execution timeout row -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between">
              <span class="text-xs font-medium text-arena-200">Execution timeout (ms)</span>
              <div class="flex items-center gap-2">
                <span class="text-[11px] text-arena-500">Enabled</span>
                <FormToggle v-model="timeoutEnabled" />
              </div>
            </div>
            <FormInput
              type="number"
              v-model="timeoutMs"
              :disabled="!timeoutEnabled"
              placeholder="5000"
              mono
              input-class="disabled:opacity-40 disabled:cursor-not-allowed"
            />
          </div>

          <!-- Max output size row -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between">
              <span class="text-xs font-medium text-arena-200">Max output size (bytes)</span>
              <div class="flex items-center gap-2">
                <span class="text-[11px] text-arena-500">Enabled</span>
                <FormToggle v-model="outputEnabled" />
              </div>
            </div>
            <FormInput
              type="number"
              v-model="maxOutputBytes"
              :disabled="!outputEnabled"
              placeholder="1048576"
              mono
              input-class="disabled:opacity-40 disabled:cursor-not-allowed"
            />
          </div>

          <p class="text-[11px] text-arena-500 mt-1">
            0 disables the limit.
          </p>
        </div>
      </div>
    </Card>

    <!-- Actions -->
    <div class="flex items-center gap-3 mt-3">
      <button
        @click="onSave"
        :disabled="!dirty || saving"
        class="px-4 py-1.5 bg-blue-500/15 hover:bg-blue-500/25 text-blue-300 text-xs font-medium rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
      >
        {{ saving ? 'Saving...' : 'Save' }}
      </button>
      <button
        v-if="dirty"
        @click="onReset"
        class="text-[11px] text-arena-500 hover:text-arena-300 transition-colors"
      >
        Discard
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { settingsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import FormInput from '../../components/FormInput.vue'
import FormToggle from '../../components/FormToggle.vue'
import Icon from '../../components/Icon.vue'

const toast = inject('toast')

const ALL_MODULES = [
  'atom', 'base64', 'csv', 'file', 'go_idiomatic', 'hashlib', 'http', 'json',
  'log', 'math', 'net', 'path', 'random', 're', 'regex', 'runtime', 'serial',
  'stats', 'string', 'struct', 'time'
]

const SENSITIVE_MODULES = ['http', 'net', 'file', 'path', 'runtime']

const isSensitive = (mod) => SENSITIVE_MODULES.includes(mod)

const initialDisabledLibraries = ref([])
const initialTimeoutEnabled = ref(true)
const initialTimeoutMs = ref(5000)
const initialOutputEnabled = ref(true)
const initialMaxOutputBytes = ref(1048576)

const disabledLibraries = ref([])
const timeoutEnabled = ref(true)
const timeoutMs = ref(5000)
const outputEnabled = ref(true)
const maxOutputBytes = ref(1048576)
const saving = ref(false)

const dirty = computed(() => {
  const set1 = new Set(disabledLibraries.value)
  const set2 = new Set(initialDisabledLibraries.value)
  if (set1.size !== set2.size) return true
  for (const item of set1) {
    if (!set2.has(item)) return true
  }

  if (timeoutEnabled.value !== initialTimeoutEnabled.value) return true
  if (timeoutEnabled.value) {
    if (Number(timeoutMs.value) !== Number(initialTimeoutMs.value)) return true
  }

  if (outputEnabled.value !== initialOutputEnabled.value) return true
  if (outputEnabled.value) {
    if (Number(maxOutputBytes.value) !== Number(initialMaxOutputBytes.value)) return true
  }

  return false
})

const isLibraryEnabled = (mod) => !disabledLibraries.value.includes(mod)

function toggleLibrary(mod) {
  if (disabledLibraries.value.includes(mod)) {
    disabledLibraries.value = disabledLibraries.value.filter(m => m !== mod)
  } else {
    disabledLibraries.value = [...disabledLibraries.value, mod]
  }
}

async function load() {
  try {
    const settings = await settingsApi.get()
    const flows = settings?.flows || {}

    const libs = flows.disabledLibraries || []
    initialDisabledLibraries.value = [...libs]
    disabledLibraries.value = [...libs]

    const tMs = flows.executionTimeoutMs !== undefined ? flows.executionTimeoutMs : 5000
    if (tMs > 0) {
      initialTimeoutEnabled.value = true
      initialTimeoutMs.value = tMs
    } else {
      initialTimeoutEnabled.value = false
      initialTimeoutMs.value = 5000
    }
    timeoutEnabled.value = initialTimeoutEnabled.value
    timeoutMs.value = initialTimeoutMs.value

    const oB = flows.maxOutputBytes !== undefined ? flows.maxOutputBytes : 1048576
    if (oB > 0) {
      initialOutputEnabled.value = true
      initialMaxOutputBytes.value = oB
    } else {
      initialOutputEnabled.value = false
      initialMaxOutputBytes.value = 1048576
    }
    outputEnabled.value = initialOutputEnabled.value
    maxOutputBytes.value = initialMaxOutputBytes.value
  } catch (e) {
    toast.error('Failed to load settings: ' + e.message)
  }
}

async function onSave() {
  saving.value = true
  try {
    const current = (await settingsApi.get()) || {}
    const flows = {
      disabledLibraries: [...disabledLibraries.value],
      executionTimeoutMs: timeoutEnabled.value ? Number(timeoutMs.value) || 0 : 0,
      maxOutputBytes: outputEnabled.value ? Number(maxOutputBytes.value) || 0 : 0
    }
    const next = { ...current, flows }
    await settingsApi.update(next)

    initialDisabledLibraries.value = [...disabledLibraries.value]
    initialTimeoutEnabled.value = timeoutEnabled.value
    initialTimeoutMs.value = timeoutMs.value
    initialOutputEnabled.value = outputEnabled.value
    initialMaxOutputBytes.value = maxOutputBytes.value

    toast.success('Flows settings saved')
  } catch (e) {
    toast.error('Save failed: ' + e.message)
  } finally {
    saving.value = false
  }
}

function onReset() {
  disabledLibraries.value = [...initialDisabledLibraries.value]
  timeoutEnabled.value = initialTimeoutEnabled.value
  timeoutMs.value = initialTimeoutMs.value
  outputEnabled.value = initialOutputEnabled.value
  maxOutputBytes.value = initialMaxOutputBytes.value
}

onMounted(load)
</script>
