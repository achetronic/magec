<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 py-5 space-y-8">
    <!-- Search -->
    <div class="sticky top-0 z-10 -mx-4 sm:-mx-6 px-4 sm:px-6 pt-0 pb-4 bg-gradient-to-b from-piedra-950 via-piedra-950 to-transparent">
      <div class="flex items-center gap-2.5 px-3 py-2 rounded-lg border border-piedra-700/50 bg-piedra-900/80 backdrop-blur-sm focus-within:border-arena-600/50 transition-colors">
        <svg class="w-3.5 h-3.5 text-arena-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          v-model="search"
          type="text"
          class="flex-1 bg-transparent text-sm text-arena-100 placeholder-arena-500 outline-none"
          placeholder="Search settings..."
        />
        <kbd v-if="!search" class="hidden sm:inline-flex px-1.5 py-0.5 text-[9px] font-mono text-arena-600 bg-piedra-800 border border-piedra-700/50 rounded">/</kbd>
        <button v-else @click="search = ''" class="text-arena-500 hover:text-arena-300 transition-colors">
          <Icon name="close" size="xs" />
        </button>
      </div>
    </div>

    <!-- Sections -->
    <template v-for="section in filteredSections" :key="section.id">
      <section :ref="el => sectionRefs[section.id] = el">
        <component :is="section.component" />
      </section>
    </template>

    <!-- No results -->
    <div v-if="filteredSections.length === 0" class="text-center py-16">
      <p class="text-sm text-arena-400">No settings match "<span class="text-arena-200">{{ search }}</span>"</p>
      <button @click="search = ''" class="mt-2 text-xs text-arena-500 hover:text-arena-300 transition-colors">Clear search</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted, markRaw } from 'vue'
import Icon from '../../components/Icon.vue'
import BackupSection from './BackupSection.vue'
import FlowsSection from './FlowsSection.vue'
import MaintenanceSection from './MaintenanceSection.vue'
import RuntimeSection from './RuntimeSection.vue'

const search = ref('')
const sectionRefs = ref({})

const sections = [
  {
    id: 'runtime',
    keywords: ['runtime', 'temporary', 'temp', 'directory', 'dir', 'workdir', 'filesystem', 'path'],
    component: markRaw(RuntimeSection),
  },
  {
    id: 'flows',
    keywords: ['flows','flow','starlark','code','script','library','libraries','timeout','limit','sandbox'],
    component: markRaw(FlowsSection),
  },
  {
    id: 'backup',
    keywords: ['backup', 'restore', 'export', 'import', 'download', 'upload', 'data', 'archive', 'tar'],
    component: markRaw(BackupSection),
  },
  {
    id: 'maintenance',
    keywords: ['maintenance', 'dead', 'references', 'integrity', 'clean', 'cleanup', 'orphan', 'stale', 'broken'],
    component: markRaw(MaintenanceSection),
  },
]

const filteredSections = computed(() => {
  const q = search.value.toLowerCase().trim()
  if (!q) return sections
  return sections.filter(s =>
    s.keywords.some(k => k.includes(q)) || s.id.includes(q)
  )
})

watch(filteredSections, async (visible) => {
  if (search.value && visible.length === 1) {
    await nextTick()
    sectionRefs.value[visible[0].id]?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
})

function onKeydown(e) {
  if (e.key === '/' && !e.metaKey && !e.ctrlKey && !e.altKey) {
    const tag = e.target.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
    e.preventDefault()
    document.querySelector('[placeholder="Search settings..."]')?.focus()
  }
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>
