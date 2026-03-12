<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Skill' : 'New Skill'" size="lg" @save="save">
    <div class="space-y-4">
      <SegmentedControl v-model="mode" :options="modeOptions" />

      <template v-if="mode === 'manual'">
        <div>
          <FormLabel label="Name" :required="true" />
          <FormInput v-model="form.name" placeholder="greeting-skill" :required="true" />
        </div>
        <div>
          <FormLabel label="Description" />
          <FormInput v-model="form.description" placeholder="What this skill does..." />
        </div>
        <div>
          <FormLabel label="Instructions" :required="true" />
          <textarea v-model="form.instructions" rows="6" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y" placeholder="Step-by-step instructions for this skill..." required />
        </div>

        <details class="group border border-piedra-700/40 rounded-xl" open>
          <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
            <span>References ({{ allFiles.length }})</span>
            <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
          </summary>
          <div class="px-4 pb-4 space-y-3">
            <div
              @dragover.prevent="dragOver = true"
              @dragleave.prevent="dragOver = false"
              @drop.prevent="onDrop"
              @click="fileInput?.click()"
              class="flex flex-col items-center justify-center gap-2 py-6 border-2 border-dashed rounded-xl cursor-pointer transition-colors"
              :class="dragOver
                ? 'border-cyan-500/50 bg-cyan-500/5'
                : 'border-piedra-700/40 hover:border-piedra-600 bg-piedra-800/30'"
            >
              <Icon name="download" size="lg" class="text-arena-500" />
              <p class="text-xs text-arena-400">Drop files here or <span class="text-cyan-400 underline">browse</span></p>
              <p class="text-[10px] text-arena-600">Schemas, templates, documentation — any text file</p>
            </div>
            <input ref="fileInput" type="file" multiple class="hidden" @change="onFileSelect" />

            <div v-if="allFiles.length" class="space-y-1.5">
              <div v-for="item in allFiles" :key="item.filename" class="flex items-center justify-between gap-3 px-3 py-2 bg-piedra-800/50 rounded-lg">
                <div class="flex items-center gap-2.5 min-w-0">
                  <Icon name="command" size="sm" class="text-cyan-400 flex-shrink-0" />
                  <div class="min-w-0">
                    <p class="text-xs font-medium text-arena-200 truncate">{{ item.filename }}</p>
                    <p class="text-[10px] text-arena-500">{{ formatSize(item.size) }}</p>
                  </div>
                </div>
                <div class="flex items-center gap-1 flex-shrink-0">
                  <button v-if="item.saved" type="button" @click.stop="downloadFile(item)" class="p-1 hover:bg-piedra-700 rounded-lg transition-colors cursor-pointer" title="Download">
                    <Icon name="download" size="sm" class="text-arena-500 hover:text-cyan-400" />
                  </button>
                  <button type="button" @click="removeFile(item)" class="p-1 hover:bg-piedra-700 rounded-lg transition-colors cursor-pointer" title="Remove">
                    <Icon name="trash" size="sm" class="text-arena-500 hover:text-lava-400" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </details>
      </template>

      <template v-if="mode === 'package'">
        <div
          @dragover.prevent="pkgDragOver = true"
          @dragleave.prevent="pkgDragOver = false"
          @drop.prevent="onPackageDrop"
          @click="pkgInput?.click()"
          class="flex flex-col items-center justify-center gap-3 py-12 border-2 border-dashed rounded-xl cursor-pointer transition-colors"
          :class="pkgDragOver
            ? 'border-cyan-500/50 bg-cyan-500/5'
            : 'border-piedra-700/40 hover:border-piedra-600 bg-piedra-800/30'"
        >
          <template v-if="!packageFile">
            <Icon name="upload" size="lg" class="text-arena-500" />
            <p class="text-xs text-arena-400">Drop <span class="text-cyan-400">.zip</span> or <span class="text-cyan-400">.tar.gz</span> here or <span class="text-cyan-400 underline">browse</span></p>
            <p class="text-[10px] text-arena-600">Must contain a SKILL.md at the root</p>
          </template>
          <template v-else>
            <Icon name="command" size="lg" class="text-cyan-400" />
            <p class="text-xs text-arena-200 font-medium">{{ packageFile.name }}</p>
            <p class="text-[10px] text-arena-500">{{ formatSize(packageFile.size) }}</p>
            <button type="button" @click.stop="packageFile = null" class="text-[10px] text-lava-400 hover:text-lava-300 transition-colors">Remove</button>
          </template>
        </div>
        <input ref="pkgInput" type="file" accept=".zip,.tar.gz,.tgz" class="hidden" @change="onPackageSelect" />
      </template>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { skillsApi } from '../../lib/api/index.js'
import { getAuthHeaders } from '../../lib/auth.js'
import { useDataStore } from '../../lib/stores/data.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormLabel from '../../components/FormLabel.vue'
import Icon from '../../components/Icon.vue'
import SegmentedControl from '../../components/SegmentedControl.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)
const fileInput = ref(null)
const pkgInput = ref(null)
const dragOver = ref(false)
const pkgDragOver = ref(false)
const mode = ref('manual')
const modeOptions = [
  { label: 'Manual', value: 'manual' },
  { label: 'Package', value: 'package' },
]
const packageFile = ref(null)

const savedRefs = ref([])
const pendingFiles = ref([])

const allFiles = computed(() => [
  ...savedRefs.value.map(r => ({ ...r, saved: true })),
  ...pendingFiles.value.map(f => ({ filename: f.name, size: f.size, saved: false })),
])

const form = ref({
  name: '',
  description: '',
  instructions: '',
})

function open(skill = null) {
  isEdit.value = !!skill
  editId.value = skill?.id || null
  mode.value = 'manual'
  packageFile.value = null
  form.value = {
    name: skill?.name || '',
    description: skill?.description || '',
    instructions: skill?.instructions || '',
  }
  savedRefs.value = [...(skill?.references || [])]
  pendingFiles.value = []
  dialogRef.value?.open()
}

function addFiles(files) {
  for (const file of files) {
    const exists = allFiles.value.some(r => r.filename === file.name)
    if (exists) {
      toast.error(`"${file.name}" already exists`)
      continue
    }
    pendingFiles.value.push(file)
  }
}

function onDrop(e) {
  dragOver.value = false
  addFiles([...(e.dataTransfer?.files || [])])
}

function onFileSelect(e) {
  addFiles([...(e.target.files || [])])
  e.target.value = ''
}

function onPackageDrop(e) {
  pkgDragOver.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) setPackageFile(file)
}

function onPackageSelect(e) {
  const file = e.target.files?.[0]
  if (file) setPackageFile(file)
  e.target.value = ''
}

function setPackageFile(file) {
  const name = file.name.toLowerCase()
  if (!name.endsWith('.zip') && !name.endsWith('.tar.gz') && !name.endsWith('.tgz')) {
    toast.error('Only .zip and .tar.gz files are supported')
    return
  }
  packageFile.value = file
}

async function removeFile(item) {
  if (item.saved) {
    try {
      await skillsApi.deleteReference(editId.value, item.filename)
      savedRefs.value = savedRefs.value.filter(r => r.filename !== item.filename)
    } catch (e) {
      toast.error(e.message)
      return
    }
  } else {
    pendingFiles.value = pendingFiles.value.filter(f => f.name !== item.filename)
  }
}

async function downloadFile(item) {
  try {
    const url = skillsApi.referenceUrl(editId.value, item.filename)
    const res = await fetch(url, { headers: { ...getAuthHeaders() } })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const blob = await res.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = item.filename
    a.click()
    URL.revokeObjectURL(a.href)
  } catch (e) {
    toast.error(e.message)
  }
}

function formatSize(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function save() {
  try {
    if (mode.value === 'package') {
      if (!packageFile.value) {
        toast.error('Select a package file')
        return
      }
      let skillId = editId.value
      if (!isEdit.value) {
        const created = await skillsApi.create({ name: packageFile.value.name.replace(/\.(zip|tar\.gz|tgz)$/i, ''), instructions: '(pending package upload)' })
        skillId = created.id
      }
      await skillsApi.uploadPackage(skillId, packageFile.value)
    } else {
      const data = {
        name: form.value.name.trim(),
        description: form.value.description.trim(),
        instructions: form.value.instructions.trim(),
      }
      let skillId = editId.value
      if (isEdit.value) {
        await skillsApi.update(skillId, data)
      } else {
        const created = await skillsApi.create(data)
        skillId = created.id
      }
      for (const file of pendingFiles.value) {
        await skillsApi.uploadReference(skillId, file)
      }
    }

    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
