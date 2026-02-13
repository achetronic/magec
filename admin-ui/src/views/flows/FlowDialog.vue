<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Flow' : 'New Flow'" size="2xl" @save="save">
    <div class="space-y-4">
      <div class="grid grid-cols-3 gap-4">
        <div>
          <FormLabel label="Name" :required="true" />
          <FormInput v-model="form.name" placeholder="my-workflow" :required="true" />
        </div>
        <div class="col-span-2">
          <FormLabel label="Description" />
          <FormInput v-model="form.description" placeholder="What this flow does..." />
        </div>
      </div>
      <FlowCanvas
        v-model="form.root"
        :agents="store.agents"
      />
      <details class="group text-arena-500">
        <summary class="text-[10px] font-medium cursor-pointer select-none hover:text-arena-300 transition-colors">
          How does the flow editor work?
        </summary>
        <div class="mt-2 text-[10px] leading-relaxed space-y-2 text-arena-500/80">
          <p>Drag blocks from the left sidebar into the canvas to build your workflow.</p>
          <div class="grid grid-cols-2 gap-x-4 gap-y-1.5">
            <div><span class="text-atlantico-400 font-semibold">Sequential</span> — runs steps one after another, in order.</div>
            <div><span class="text-sol-400 font-semibold">Parallel</span> — runs all steps at the same time.</div>
            <div><span class="text-lava-400 font-semibold">Loop</span> — repeats its steps N times.</div>
            <div><span class="text-sol-400 font-semibold">Agent</span> — an AI agent that processes input.</div>
          </div>
          <div class="flex items-start gap-1.5 pt-1 border-t border-piedra-700/30">
            <svg class="w-3.5 h-3.5 text-green-400 flex-shrink-0 mt-px" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 8.25h9m-9 3H12m-9.75 1.51c0 1.6 1.123 2.994 2.707 3.227 1.087.16 2.185.283 3.293.369V21l4.076-4.076a1.526 1.526 0 0 1 1.037-.443 48.2 48.2 0 0 0 5.887-.512c1.584-.233 2.707-1.626 2.707-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.4 48.4 0 0 0 12 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018Z" />
            </svg>
            <span>Each agent has a <span class="text-green-400 font-semibold">response</span> toggle. Only agents with this active will be included in the flow output. If none are marked, all agent outputs are returned.</span>
          </div>
        </div>
      </details>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { flowsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormLabel from '../../components/FormLabel.vue'
import FlowCanvas from './FlowCanvas.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

const form = reactive({
  name: '',
  description: '',
  root: null,
})

function open(flow = null) {
  isEdit.value = !!flow
  editId.value = flow?.id || null
  form.name = flow?.name || ''
  form.description = flow?.description || ''
  form.root = flow ? JSON.parse(JSON.stringify(flow.root)) : null
  dialogRef.value?.open()
}

async function save() {
  const data = {
    name: form.name.trim(),
    description: form.description.trim(),
    root: cleanStep(form.root),
  }
  try {
    if (isEdit.value) {
      await flowsApi.update(editId.value, data)
    } else {
      await flowsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

function cleanStep(step) {
  const clean = { type: step.type }
  if (step.type === 'agent') {
    clean.agentId = step.agentId
    if (step.responseAgent) clean.responseAgent = true
  } else {
    clean.steps = (step.steps || []).map(cleanStep)
    if (step.type === 'loop') {
      clean.maxIterations = step.maxIterations || 0
    }
  }
  return clean
}

defineExpose({ open })
</script>
