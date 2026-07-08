<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Secrets</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Secret
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.secrets.length" />

    <EmptyState v-else-if="!store.secrets.length" title="No secrets configured" subtitle="Store API keys and credentials securely. Reference them as ${KEY_NAME} in any field." icon="key" color="amber" actionLabel="+ New Secret" @action="openDialog()" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="s in store.secrets" :key="s.id" color="amber">
        <div class="flex items-start justify-between gap-3">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 bg-amber-500/15">
              <Icon name="key" size="md" class="text-amber-400" />
            </div>
            <div class="min-w-0">
              <h3 class="font-medium text-arena-100 text-sm">{{ s.name }}</h3>
              <button @click="copyRef(s.key)" class="cursor-pointer hover:opacity-80 active:scale-95 mt-0.5" :title="copied === s.key ? 'Copied!' : 'Click to copy reference'">
                <Badge :variant="copied === s.key ? 'sol' : 'muted'" class="font-mono">{{ copied === s.key ? 'Copied!' : refLabel(s.key) }}</Badge>
              </button>
            </div>
          </div>
          <div class="flex gap-0.5 flex-shrink-0">
            <button @click="openDialog(s)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
              <Icon name="edit" size="sm" class="text-arena-400" />
            </button>
            <button @click="handleDelete(s)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <p v-if="s.description" class="text-[10px] text-arena-500 mt-2 truncate">{{ s.description }}</p>
      </Card>
    </div>

    <SecretDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { secretsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Icon from '../../components/Icon.vue'
import Badge from '../../components/Badge.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import SecretDialog from './SecretDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const copied = ref(null)
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function copyRef(key) {
  const text = '${' + key + '}'
  if (navigator.clipboard?.writeText) {
    navigator.clipboard.writeText(text).catch(() => fallbackCopy(text))
  } else {
    fallbackCopy(text)
  }
  copied.value = key
  setTimeout(() => { if (copied.value === key) copied.value = null }, 1500)
}

function fallbackCopy(text) {
  const ta = document.createElement('textarea')
  ta.value = text
  ta.style.position = 'fixed'
  ta.style.opacity = '0'
  document.body.appendChild(ta)
  ta.select()
  document.execCommand('copy')
  document.body.removeChild(ta)
}

function refLabel(key) {
  return '${' + key + '}'
}

function openDialog(secret = null) {
  dialog.value?.open(secret)
}

function handleDelete(s) {
  requestDelete(`Delete secret "${s.name}"? This cannot be undone.`, async () => {
    try {
      await secretsApi.delete(s.id)
      await store.refresh()
    } catch (e) {
      toast.error(e.message)
    }
  })
}
</script>
