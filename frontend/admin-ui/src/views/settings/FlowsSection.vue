<template>
  <div class="max-w-2xl space-y-6">
    <!-- Header -->
    <div>
      <h3 class="text-sm font-semibold text-arena-100">Flows</h3>
      <p class="text-xs text-arena-500 mt-1">Behaviour and limits for workflow flows.</p>
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
            Modules available to code nodes. All enabled by default; click to disable what this deployment should not allow.
          </p>

          <div class="flex flex-wrap gap-2 mt-4">
            <button
              v-for="mod in ALL_MODULES"
              :key="mod"
              type="button"
              @click="toggleLibrary(mod)"
              :title="isSensitive(mod) ? 'Network or filesystem access' : undefined"
              class="px-2.5 py-1 text-[11px] font-medium rounded-lg border transition-all cursor-pointer select-none font-mono"
              :class="isLibraryEnabled(mod)
                ? 'bg-piedra-800 text-arena-200 border-piedra-700/40 hover:border-piedra-600'
                : 'bg-piedra-900 text-arena-600 border-piedra-800 line-through hover:text-arena-500'"
            >
              {{ mod }}<span v-if="isSensitive(mod)" class="ml-1 text-arena-500 no-underline">&bull;</span>
            </button>
          </div>

          <p class="text-[11px] text-arena-500 mt-4">
            See the
            <a href="https://github.com/1set/starlet#libraries"
               target="_blank" rel="noopener"
               class="text-sol-400 hover:text-sol-300 underline underline-offset-2 decoration-sol-400/40 hover:decoration-sol-300">library reference</a>
            for what each module provides.
          </p>
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

          <div class="flex items-center gap-3 mt-3">
            <button
              @click="onSaveLimits"
              :disabled="!limitsDirty || savingLimits"
              class="px-4 py-1.5 bg-blue-500/15 hover:bg-blue-500/25 text-blue-300 text-xs font-medium rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {{ savingLimits ? 'Saving...' : 'Save' }}
            </button>
            <button
              v-if="limitsDirty"
              @click="onResetLimits"
              class="text-[11px] text-arena-500 hover:text-arena-300 transition-colors"
            >
              Discard
            </button>
          </div>
        </div>
      </div>
    </Card>
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

const disabledLibraries = ref([])
let saveInFlight = false
let pendingSave = false

const initialTimeoutEnabled = ref(true)
const initialTimeoutMs = ref(5000)
const initialOutputEnabled = ref(true)
const initialMaxOutputBytes = ref(1048576)

const timeoutEnabled = ref(true)
const timeoutMs = ref(5000)
const outputEnabled = ref(true)
const maxOutputBytes = ref(1048576)
const savingLimits = ref(false)

const limitsDirty = computed(() => {
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
  disabledLibraries.value = disabledLibraries.value.includes(mod)
    ? disabledLibraries.value.filter(m => m !== mod)
    : [...disabledLibraries.value, mod]

  persistLibraries()
}

async function persistLibraries() {
  if (saveInFlight) {
    pendingSave = true
    return
  }

  saveInFlight = true
  try {
    do {
      pendingSave = false
      const current = (await settingsApi.get()) || {}
      const flows = { ...(current.flows || {}), disabledLibraries: [...disabledLibraries.value] }
      await settingsApi.update({ ...current, flows })
    } while (pendingSave)
  } catch (e) {
    toast.error('Failed to update libraries: ' + e.message)
    await load()
  } finally {
    saveInFlight = false
  }
}

async function load() {
  try {
    const settings = await settingsApi.get()
    const flows = settings?.flows || {}

    disabledLibraries.value = [...(flows.disabledLibraries || [])]

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

async function onSaveLimits() {
  savingLimits.value = true
  try {
    const current = (await settingsApi.get()) || {}
    const flows = {
      ...(current.flows || {}),
      executionTimeoutMs: timeoutEnabled.value ? Number(timeoutMs.value) || 0 : 0,
      maxOutputBytes: outputEnabled.value ? Number(maxOutputBytes.value) || 0 : 0
    }
    await settingsApi.update({ ...current, flows })

    initialTimeoutEnabled.value = timeoutEnabled.value
    initialTimeoutMs.value = timeoutMs.value
    initialOutputEnabled.value = outputEnabled.value
    initialMaxOutputBytes.value = maxOutputBytes.value

    toast.success('Execution limits saved')
  } catch (e) {
    toast.error('Save failed: ' + e.message)
  } finally {
    savingLimits.value = false
  }
}

function onResetLimits() {
  timeoutEnabled.value = initialTimeoutEnabled.value
  timeoutMs.value = initialTimeoutMs.value
  outputEnabled.value = initialOutputEnabled.value
  maxOutputBytes.value = initialMaxOutputBytes.value
}

onMounted(load)
</script>
