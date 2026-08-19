<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Backend' : 'New Backend'" @save="save">
    <div class="space-y-4">
      <div>
        <FormLabel label="Name" :required="true" />
        <FormInput v-model="form.name" placeholder="ollama" :required="true" />
      </div>
      <div>
        <FormLabel label="Type" />
        <FormSelect v-model="form.type">
          <option value="openai">OpenAI-compatible</option>
          <option value="anthropic">Anthropic</option>
          <option value="gemini">Gemini</option>
        </FormSelect>
      </div>
      <div v-if="form.type === 'openai'">
        <FormLabel label="API" />
        <FormSelect v-model="form.api">
          <option value="completions">Chat Completions</option>
          <option value="responses">Responses (coming soon)</option>
        </FormSelect>
        <p class="text-[10px] text-arena-500 mt-1">Wire API for OpenAI-compatible backends. Responses is not supported yet; the value is stored for when it lands.</p>
      </div>
      <div>
        <FormLabel label="URL" />
        <FormInput v-model="form.url" placeholder="http://localhost:11434/v1" />
      </div>
      <div>
        <FormLabel label="API Key" />
        <FormInput v-model="form.apiKey" type="password" placeholder="sk-..." />
      </div>
      <div>
        <FormLabel label="Headers" />
        <div class="space-y-2">
          <div v-for="(h, i) in form.headers" :key="i" class="flex gap-2 items-center">
            <input
              v-model="h.key"
              placeholder="Authorization"
              class="flex-1 bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-1.5 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
            />
            <input
              v-model="h.value"
              placeholder="Bearer sk-..."
              class="flex-[2] bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-1.5 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
            />
            <button @click="form.headers.splice(i, 1)" class="p-1.5 hover:bg-piedra-800 rounded-lg text-arena-400 hover:text-lava-400 flex-shrink-0" title="Remove header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M18 6L6 18M6 6l12 12"/></svg>
            </button>
          </div>
          <button @click="form.headers.push({ key: '', value: '' })" class="text-xs text-sol-400 hover:text-sol-500 transition-colors">
            + Add header
          </button>
        </div>
        <p class="text-[10px] text-arena-500 mt-1">Extra HTTP headers sent with every request to this backend. Agent-level headers override these.</p>
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'
import { backendsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

const form = reactive({
  name: '',
  type: 'openai',
  api: 'completions',
  url: '',
  apiKey: '',
  headers: [],
})

function headersToList(obj) {
  if (!obj || !Object.keys(obj).length) return []
  return Object.entries(obj).map(([key, value]) => ({ key, value }))
}

function listToHeaders(list) {
  const obj = {}
  for (const h of list) {
    const k = h.key.trim()
    if (k) obj[k] = h.value
  }
  return Object.keys(obj).length ? obj : undefined
}

function open(backend = null) {
  isEdit.value = !!backend
  editId.value = backend?.id || null
  form.name = backend?.name || ''
  form.type = backend?.type || 'openai'
  form.api = backend?.api || 'completions'
  form.url = backend?.url || ''
  form.apiKey = backend?.apiKey || ''
  form.headers = headersToList(backend?.headers)
  dialogRef.value?.open()
}

async function save() {
  const data = { name: form.name, type: form.type, url: form.url, apiKey: form.apiKey }
  if (form.type === 'openai') data.api = form.api
  const headers = listToHeaders(form.headers)
  if (headers) data.headers = headers
  try {
    if (isEdit.value) {
      await backendsApi.update(editId.value, data)
    } else {
      await backendsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
