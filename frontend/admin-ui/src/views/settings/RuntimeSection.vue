<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <div class="max-w-2xl space-y-6">
    <!-- Header -->
    <div>
      <h3 class="text-sm font-semibold text-arena-100">Runtime</h3>
      <p class="text-xs text-arena-500 mt-1">Low-level paths and behaviour of the platform.</p>
    </div>

    <Card color="blue">
      <div class="flex items-start gap-4">
        <div class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 bg-blue-500/10">
          <Icon name="infrastructure" size="md" class="text-blue-400" />
        </div>
        <div class="flex-1 min-w-0">
          <h4 class="text-[13px] font-medium text-arena-100">Temporary directory</h4>
          <p class="text-xs text-arena-500 mt-0.5">
            Where transient files are written. Leave empty to use the operating system default.
          </p>

          <FormInput
            v-model="temporaryDir"
            placeholder="/app/data/temporary"
            mono
            input-class="mt-3"
          />

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
      </div>
    </Card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { settingsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import FormInput from '../../components/FormInput.vue'
import Icon from '../../components/Icon.vue'

const toast = inject('toast')

const initial = ref('')
const temporaryDir = ref('')
const saving = ref(false)

const dirty = computed(() => temporaryDir.value !== initial.value)

async function load() {
  try {
    const settings = await settingsApi.get()
    initial.value = settings?.temporaryDir || ''
    temporaryDir.value = initial.value
  } catch (e) {
    toast.error('Failed to load settings: ' + e.message)
  }
}

async function onSave() {
  saving.value = true
  try {
    const current = (await settingsApi.get()) || {}
    const next = { ...current, temporaryDir: temporaryDir.value.trim() }
    if (!next.temporaryDir) delete next.temporaryDir
    await settingsApi.update(next)
    initial.value = next.temporaryDir || ''
    toast.success('Runtime settings saved')
  } catch (e) {
    toast.error('Save failed: ' + e.message)
  } finally {
    saving.value = false
  }
}

function onReset() {
  temporaryDir.value = initial.value
}

onMounted(load)
</script>
