<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Secret' : 'New Secret'" @save="save">
    <div class="space-y-4">
      <div>
        <FormLabel label="Name" :required="true" />
        <FormInput v-model="form.name" placeholder="OpenAI API Key" :required="true" />
      </div>
      <div>
        <FormLabel label="Environment Variable Key" :required="true" />
        <FormInput v-model="form.key" placeholder="OPENAI_API_KEY" :mono="true" @input="enforceUpperSnake" :required="true" />
        <p class="text-[10px] text-arena-500 mt-1">Use UPPER_SNAKE_CASE. Reference as <code class="text-arena-400">${{ '{' + form.key + '}' }}</code> in any field.</p>
      </div>
      <div>
        <FormLabel label="Value" :required="!isEdit" />
        <div class="relative">
          <FormInput
            v-model="form.value"
            :type="showValue ? 'text' : 'password'"
            :placeholder="isEdit ? '(unchanged)' : 'sk-...'"
            :mono="true"
            :required="!isEdit"
          />
          <button
            type="button"
            @click="showValue = !showValue"
            class="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-piedra-700 rounded"
            :title="showValue ? 'Hide' : 'Show'"
          >
            <Icon :name="showValue ? 'eye' : 'eye'" size="sm" class="text-arena-500" />
          </button>
        </div>
        <p v-if="isEdit" class="text-[10px] text-arena-500 mt-1">Leave empty to keep the current value.</p>
      </div>
      <div>
        <FormLabel label="Description" />
        <FormInput v-model="form.description" placeholder="Production key for GPT-4o" />
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'
import { secretsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormLabel from '../../components/FormLabel.vue'
import Icon from '../../components/Icon.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)
const showValue = ref(false)

const form = reactive({
  name: '',
  key: '',
  value: '',
  description: '',
})

function enforceUpperSnake() {
  form.key = form.key.toUpperCase().replace(/[^A-Z0-9_]/g, '_')
}

function open(secret = null) {
  isEdit.value = !!secret
  editId.value = secret?.id || null
  form.name = secret?.name || ''
  form.key = secret?.key || ''
  form.value = ''
  form.description = secret?.description || ''
  showValue.value = false
  dialogRef.value?.open()
}

async function save() {
  const data = {
    name: form.name.trim(),
    key: form.key.trim(),
    description: form.description.trim(),
  }

  if (form.value.trim()) {
    data.value = form.value.trim()
  }

  if (!isEdit.value && !data.value) {
    toast.error('Value is required for new secrets')
    return
  }

  try {
    if (isEdit.value) {
      await secretsApi.update(editId.value, data)
    } else {
      await secretsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
