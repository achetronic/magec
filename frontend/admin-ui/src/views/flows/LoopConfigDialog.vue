<template>
  <AppDialog ref="dialogRef" title="Loop step settings" size="md" @save="save" @close="reset">
    <div class="space-y-5">
      <div>
        <FormLabel label="Max iterations" />
        <FormInput v-model.number="form.maxIterations" type="number" placeholder="0 (infinite)" />
        <p class="mt-1 text-[10px] text-arena-500">Hard safety cap. The loop never runs more than this many times. Always active, regardless of any early-exit option below. Use 0 to disable the cap (only safe combined with an early-exit option).</p>
      </div>
      <div class="border-t border-piedra-700/50 pt-4">
        <div class="flex items-center justify-between">
          <div>
            <FormLabel label="Early exit" />
            <p class="text-[10px] text-arena-500">Stop the loop before reaching the cap.</p>
          </div>
          <FormToggle v-model="form.earlyExitEnabled" />
        </div>
        <div v-if="form.earlyExitEnabled" class="mt-3 space-y-3">
          <SegmentedControl v-model="form.strategy" :options="strategyOptions" />
          <p class="text-[10px] text-arena-500">{{ strategyHint }}</p>
          <div v-if="form.strategy === 'expression'">
            <FormLabel label="CEL expression" :required="true" />
            <textarea
              v-model="form.exitWhen"
              rows="3"
              placeholder='state.approved == true'
              class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-xs font-mono focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
            />
            <p class="mt-1 text-[10px] text-arena-500">
              Evaluated against the shared <code class="text-arena-300">state</code> map after each iteration. The loop exits when this returns <code class="text-arena-300">true</code>.
              <br />
              See the <a href="https://playcel.undistro.io/" target="_blank" rel="noopener noreferrer" class="text-sol-400 hover:text-sol-300 underline">CEL playground</a> to test expressions.
            </p>
          </div>
        </div>
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormLabel from '../../components/FormLabel.vue'
import FormToggle from '../../components/FormToggle.vue'
import SegmentedControl from '../../components/SegmentedControl.vue'

const emit = defineEmits(['save'])

const dialogRef = ref(null)

// strategy is only meaningful when earlyExitEnabled is true. When the
// toggle is off, save() emits both exitLoop:false and exitWhen:"" so the
// backend treats the loop as cap-only.
const strategyOptions = [
  { value: 'agent',      label: 'Agent decides' },
  { value: 'expression', label: 'Expression' },
]

const form = reactive({
  maxIterations: 3,
  earlyExitEnabled: false,
  strategy: 'agent',
  exitWhen: '',
})

const strategyHint = computed(() => {
  if (form.strategy === 'agent') {
    return 'Every agent in the loop subtree gets the exit_loop tool. The cap still applies as a safety net.'
  }
  return 'A CEL expression on the shared flow state, evaluated after each iteration. The cap still applies as a safety net.'
})

function open(step = {}) {
  form.maxIterations = step.maxIterations ?? 3
  form.exitWhen = step.exitWhen || ''
  if (step.exitLoop) {
    form.earlyExitEnabled = true
    form.strategy = 'agent'
  } else if (step.exitWhen) {
    form.earlyExitEnabled = true
    form.strategy = 'expression'
  } else {
    form.earlyExitEnabled = false
    form.strategy = 'agent'
  }
  dialogRef.value?.open()
}

function reset() {
  form.maxIterations = 3
  form.earlyExitEnabled = false
  form.strategy = 'agent'
  form.exitWhen = ''
}

function save() {
  const payload = {
    maxIterations: Number.isFinite(form.maxIterations) ? Number(form.maxIterations) : 0,
    exitLoop: form.earlyExitEnabled && form.strategy === 'agent',
    exitWhen: form.earlyExitEnabled && form.strategy === 'expression' ? form.exitWhen.trim() : '',
  }
  emit('save', payload)
  dialogRef.value?.close()
}

defineExpose({ open })
</script>
