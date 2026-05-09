<template>
  <AppDialog ref="dialogRef" title="Upload Skill" size="lg">
    <div class="space-y-4">
      <!-- The viewer is opinionated by upstream contract: ADK's
           skilltoolset only accepts the layout from the Agent Skills
           specification (SKILL.md at root + references/, assets/,
           scripts/ subtrees). We surface that to the operator before
           the dropzone instead of letting them upload something we'd
           reject with a 400, and link the spec so the choice is
           informed rather than guessed. -->
      <!-- Inline intro: title-weight first line tells the operator
           what to drop, sub-line points at the spec they need to
           follow. No card / box — the modal itself is the canvas. -->
      <div class="space-y-1">
        <p class="text-[13px] text-arena-200 leading-6">
          Drop a <code class="px-1 py-0.5 rounded bg-piedra-800 text-arena-200 font-mono text-[11px]">SKILL.md</code>
          or a packaged skill
          (<code class="px-1 py-0.5 rounded bg-piedra-800 text-arena-200 font-mono text-[11px]">.zip</code>
          / <code class="px-1 py-0.5 rounded bg-piedra-800 text-arena-200 font-mono text-[11px]">.tar.gz</code>).
        </p>
        <p class="text-[11px] text-arena-500 leading-6">
          Skills must follow the
          <a href="https://agentskills.io/specification"
             target="_blank" rel="noopener"
             class="text-sol-400 hover:text-sol-300 underline underline-offset-2 decoration-sol-400/40 hover:decoration-sol-300">official Agent Skills specification</a>.
        </p>
      </div>

      <div
        @dragover.prevent="dragOver = true"
        @dragleave.prevent="dragOver = false"
        @drop.prevent="onDrop"
        @click="fileInput?.click()"
        class="flex flex-col items-center justify-center gap-3 py-12 border-2 border-dashed rounded-xl cursor-pointer transition-colors"
        :class="dragOver
          ? 'border-sol-400/50 bg-sol-500/5'
          : 'border-piedra-700/40 hover:border-piedra-600 bg-piedra-800/30'"
      >
        <template v-if="!file">
          <Icon name="upload" size="lg" class="text-arena-500" />
          <p class="text-xs text-arena-400">Drop a file here or <span class="text-sol-400 underline">browse</span></p>
          <p class="text-[10px] text-arena-600">SKILL.md, .zip, .tar.gz</p>
        </template>
        <template v-else>
          <Icon name="command" size="lg" class="text-sol-400" />
          <p class="text-xs text-arena-200 font-medium">{{ file.name }}</p>
          <p class="text-[10px] text-arena-500">{{ formatSize(file.size) }}</p>
          <button type="button" @click.stop="clearFile" class="text-[10px] text-lava-400 hover:text-lava-300 transition-colors">Remove</button>
        </template>
      </div>
      <input ref="fileInput" type="file" accept=".md,.markdown,.zip,.tar.gz,.tgz" class="hidden" @change="onSelect" />

      <div v-if="errorMsg" class="border border-lava-500/40 bg-lava-500/5 rounded-lg px-3 py-2">
        <p class="text-[11px] text-lava-300 whitespace-pre-line">{{ errorMsg }}</p>
        <p v-if="conflict" class="text-[10px] text-arena-400 mt-2">
          A skill with this slug already exists. Enable replace mode below to overwrite (existing agent links are preserved).
        </p>
      </div>

      <div v-if="file" class="border border-piedra-700/40 rounded-xl px-4 py-3 space-y-2">
        <label class="flex items-center gap-2 text-[11px] text-arena-300">
          <input type="checkbox" v-model="replace" class="accent-sol-500" />
          Replace if a skill with this slug already exists
        </label>
        <p class="text-[10px] text-arena-500">
          When enabled, an upload whose <code>frontmatter.name</code> matches an existing skill replaces its package on disk while keeping the same store ID — agents that link to the skill stay linked.
        </p>
      </div>
    </div>
    <template #footer>
      <button type="button" @click="close" class="px-4 py-2 text-sm text-arena-400 hover:text-arena-200 hover:bg-piedra-800 rounded-lg transition-colors">
        Cancel
      </button>
      <button type="button" :disabled="uploading" @click="save" class="px-4 py-2 bg-sol-500 hover:bg-sol-600 disabled:opacity-50 disabled:cursor-not-allowed text-piedra-950 text-sm font-medium rounded-lg transition-colors">
        {{ uploading ? 'Uploading…' : (replace ? 'Upload (replace)' : 'Upload') }}
      </button>
    </template>
  </AppDialog>
</template>

<script setup>
import { ref, inject } from 'vue'
import { skillsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import Icon from '../../components/Icon.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const dialogRef = ref(null)
const fileInput = ref(null)
const dragOver = ref(false)
const file = ref(null)
const replace = ref(false)
const errorMsg = ref('')
const conflict = ref(false)
const uploading = ref(false)

function open() {
  file.value = null
  replace.value = false
  errorMsg.value = ''
  conflict.value = false
  uploading.value = false
  dialogRef.value?.open()
}

function close() {
  dialogRef.value?.close()
}

function clearFile() {
  file.value = null
  errorMsg.value = ''
  conflict.value = false
}

function onDrop(e) {
  dragOver.value = false
  const f = e.dataTransfer?.files?.[0]
  if (f) setFile(f)
}

function onSelect(e) {
  const f = e.target.files?.[0]
  if (f) setFile(f)
  e.target.value = ''
}

function setFile(f) {
  const lower = f.name.toLowerCase()
  const ok =
    lower.endsWith('.md') ||
    lower.endsWith('.markdown') ||
    lower.endsWith('.zip') ||
    lower.endsWith('.tar.gz') ||
    lower.endsWith('.tgz')
  if (!ok) {
    toast.error('Only SKILL.md, .zip and .tar.gz uploads are accepted')
    return
  }
  file.value = f
  errorMsg.value = ''
  conflict.value = false
}

function formatSize(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function save() {
  if (!file.value) {
    errorMsg.value = 'Select a file first'
    return
  }
  errorMsg.value = ''
  conflict.value = false
  uploading.value = true
  try {
    await skillsApi.upload(file.value, { replace: replace.value })
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    if (e.status === 409) {
      conflict.value = true
      errorMsg.value = e.message
    } else {
      errorMsg.value = e.message
    }
  } finally {
    uploading.value = false
  }
}

defineExpose({ open })
</script>

