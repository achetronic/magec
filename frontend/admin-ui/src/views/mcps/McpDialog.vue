<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit MCP Server' : 'New MCP Server'" @save="save">
    <div class="space-y-4">
      <div>
        <FormLabel label="Name" :required="true" />
        <FormInput v-model="form.name" placeholder="home-assistant" :required="true" />
      </div>
      <div>
        <FormLabel label="Type" />
        <FormSelect v-model="form.type">
          <option value="http">HTTP</option>
          <option value="stdio">Stdio</option>
        </FormSelect>
      </div>
      <div v-if="form.type === 'http'">
        <FormLabel label="Endpoint" />
        <FormInput v-model="form.endpoint" placeholder="http://localhost:8080/mcp" />
      </div>
      <div v-if="form.type === 'http'">
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
      </div>
      <div v-if="form.type === 'http'">
        <label class="flex items-center gap-2 cursor-pointer">
          <div class="relative">
            <input type="checkbox" v-model="form.insecure" class="sr-only peer" />
            <div class="w-9 h-5 bg-piedra-700 rounded-full peer-checked:bg-sol-500/60 transition-colors" />
            <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-arena-400 rounded-full peer-checked:translate-x-4 peer-checked:bg-white transition-transform" />
          </div>
          <span class="text-sm text-arena-300">Skip TLS verification</span>
        </label>
        <p class="text-[10px] text-arena-500 mt-1 ml-11">Allow connections to HTTPS endpoints with self-signed or invalid certificates.</p>
      </div>
      <template v-if="form.type === 'stdio'">
        <div>
          <FormLabel label="Command" />
          <FormInput v-model="form.command" placeholder="uvx" />
        </div>
        <div>
          <FormLabel label="Args (comma-separated)" />
          <FormInput v-model="form.argsStr" placeholder="mcp-server-sqlite, --db-path, /data/db" />
        </div>
      </template>
      <div>
        <FormLabel label="System Prompt" />
        <textarea
          v-model="form.systemPrompt"
          rows="2"
          class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y"
          placeholder="Instructions for the LLM about this MCP..."
        />
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'
import { mcpsApi } from '../../lib/api/index.js'
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
  type: 'http',
  endpoint: '',
  headers: [],
  insecure: false,
  command: '',
  argsStr: '',
  systemPrompt: '',
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

function open(mcp = null) {
  isEdit.value = !!mcp
  editId.value = mcp?.id || null
  form.name = mcp?.name || ''
  form.type = mcp?.type || 'http'
  form.endpoint = mcp?.endpoint || ''
  form.headers = headersToList(mcp?.headers)
  form.insecure = mcp?.insecure || false
  form.command = mcp?.command || ''
  form.argsStr = (mcp?.args || []).join(', ')
  form.systemPrompt = mcp?.systemPrompt || ''
  dialogRef.value?.open()
}

async function save() {
  const data = { name: form.name, type: form.type, systemPrompt: form.systemPrompt.trim() }
  if (form.type === 'http') {
    data.endpoint = form.endpoint.trim()
    const headers = listToHeaders(form.headers)
    if (headers) data.headers = headers
    if (form.insecure) data.insecure = true
  } else {
    data.command = form.command.trim()
    data.args = form.argsStr ? form.argsStr.split(',').map(s => s.trim()).filter(Boolean) : []
  }
  try {
    if (isEdit.value) {
      await mcpsApi.update(editId.value, data)
    } else {
      await mcpsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
