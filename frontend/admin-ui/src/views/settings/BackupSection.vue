<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <div class="max-w-2xl space-y-6">
    <!-- Header -->
    <div>
      <h3 class="text-sm font-semibold text-arena-100">Backup & Restore</h3>
      <p class="text-xs text-arena-500 mt-1">Download a full backup of your data or restore from a previous backup file.</p>
    </div>

    <!-- Backup -->
    <Card color="blue">
      <div class="flex items-start gap-4">
        <div class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 bg-blue-500/10">
          <Icon name="download" size="md" class="text-blue-400" />
        </div>
        <div class="flex-1 min-w-0">
          <h4 class="text-[13px] font-medium text-arena-100">Download Backup</h4>
          <p class="text-xs text-arena-500 mt-0.5">
            Exports all agents, flows, skills, backends, clients, secrets, commands, memory providers, and conversations as a <code class="text-arena-400 bg-piedra-800/80 px-1 rounded">.tar.gz</code> archive.
          </p>
          <button
            @click="onBackup"
            :disabled="backupLoading"
            class="mt-3 px-4 py-1.5 bg-blue-500/15 hover:bg-blue-500/25 text-blue-300 text-xs font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ backupLoading ? 'Downloading...' : 'Download Backup' }}
          </button>
        </div>
      </div>
    </Card>

    <!-- Restore -->
    <Card color="blue">
      <div class="flex items-start gap-4">
        <div class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 bg-blue-500/10">
          <Icon name="upload" size="md" class="text-blue-400" />
        </div>
        <div class="flex-1 min-w-0">
          <h4 class="text-[13px] font-medium text-arena-100">Restore from Backup</h4>
          <p class="text-xs text-arena-500 mt-0.5">
            Upload a previously downloaded <code class="text-arena-400 bg-piedra-800/80 px-1 rounded">.tar.gz</code> backup to replace all current data. This action cannot be undone.
          </p>
          <div class="flex items-center gap-3 mt-3">
            <button
              @click="triggerRestore"
              :disabled="restoreLoading"
              class="px-4 py-1.5 bg-blue-500/15 hover:bg-blue-500/25 text-blue-300 text-xs font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ restoreLoading ? 'Restoring...' : 'Upload & Restore' }}
            </button>
            <span v-if="restoreFile" class="text-[11px] text-arena-400 truncate">{{ restoreFile.name }}</span>
          </div>
          <input ref="fileInput" type="file" accept=".tar.gz,.tgz" class="hidden" @change="onFileSelected" />
        </div>
      </div>
    </Card>

    <!-- Warning -->
    <div class="flex items-start gap-3 p-3 rounded-lg bg-lava-500/5 border border-lava-500/10">
      <svg class="w-4 h-4 text-lava-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
      </svg>
      <div>
        <p class="text-xs font-medium text-lava-300">Restoring replaces everything</p>
        <p class="text-[11px] text-lava-400/70 mt-0.5">All current agents, flows, skills, conversations, secrets, and configuration will be overwritten. Consider downloading a backup first.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject } from 'vue'
import { backupApi } from '../../lib/api/index.js'
import { useDataStore } from '../../lib/stores/data.js'
import Card from '../../components/Card.vue'
import Icon from '../../components/Icon.vue'

const store = useDataStore()
const toast = inject('toast')
const requestDelete = inject('requestDelete')

const backupLoading = ref(false)
const restoreLoading = ref(false)
const restoreFile = ref(null)
const fileInput = ref(null)

async function onBackup() {
  backupLoading.value = true
  try {
    await backupApi.download()
    toast.success('Backup downloaded')
  } catch (e) {
    toast.error('Backup failed: ' + e.message)
  } finally {
    backupLoading.value = false
  }
}

function triggerRestore() {
  fileInput.value.value = ''
  fileInput.value.click()
}

function onFileSelected(e) {
  const file = e.target.files?.[0]
  if (!file) return
  restoreFile.value = file

  requestDelete(
    'This will replace ALL data (agents, flows, skills, conversations, secrets). Are you sure?',
    async () => {
      restoreLoading.value = true
      try {
        await backupApi.restore(file)
        toast.success('Backup restored — reloading data')
        store.refresh()
      } catch (err) {
        toast.error('Restore failed: ' + err.message)
      } finally {
        restoreLoading.value = false
        restoreFile.value = null
      }
    }
  )
}
</script>
