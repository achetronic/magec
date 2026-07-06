<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Skills</h2>
      <button @click="openUpload()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + Upload Skill
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.skills.length" />

    <EmptyState v-else-if="!store.skills.length" title="No skills configured" subtitle="Upload a SKILL.md or a packaged skill (.zip / .tar.gz)" icon="skill" color="cyan" actionLabel="+ Upload Skill" @action="openUpload()" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="sk in store.skills" :key="sk.id" color="cyan" class="cursor-pointer" @click="openView(sk)">
        <div class="flex items-start justify-between gap-3 mb-2">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 bg-cyan-500/15">
              <Icon name="skill" size="md" class="text-cyan-400" />
            </div>
            <div class="min-w-0">
              <h3 class="font-medium text-arena-100 text-sm truncate">{{ displayName(sk) }}</h3>
              <p v-if="showSlug(sk)" class="text-[10px] text-arena-500 font-mono truncate">{{ sk.slug }}</p>
            </div>
          </div>
          <div class="flex gap-0.5 flex-shrink-0">
            <button @click.stop="openView(sk)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="View">
              <Icon name="eye" size="sm" class="text-arena-400" />
            </button>
            <button @click.stop="handleDelete(sk)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <p v-if="sk.description" class="text-[10px] text-arena-400 mb-2 line-clamp-2">{{ sk.description }}</p>
        <p v-else class="text-[10px] text-arena-600 italic mb-2">No description in SKILL.md frontmatter.</p>

        <div v-if="usedBy(sk.id).length" class="flex flex-wrap gap-1">
          <Tooltip v-for="ref in usedBy(sk.id)" :key="ref.name" :text="ref.tooltip">
            <Badge variant="muted">{{ ref.name }}</Badge>
          </Tooltip>
        </div>
        <p v-else class="text-[10px] text-arena-600">Not linked to any agent</p>
      </Card>
    </div>

    <SkillDialog ref="uploadDialog" @saved="onSaved" />
    <SkillViewDialog ref="viewDialog" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { skillsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Tooltip from '../../components/Tooltip.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import SkillDialog from './SkillDialog.vue'
import SkillViewDialog from './SkillViewDialog.vue'

const store = useDataStore()
const uploadDialog = ref(null)
const viewDialog = ref(null)
const deleteEntity = inject('deleteEntity')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openUpload()))
onUnmounted(() => registerNew(null))

function openUpload() {
  uploadDialog.value?.open()
}

// displayName picks the best human-readable label for a skill card.
// Order of precedence: SKILL.md frontmatter `name`, then the on-disk
// slug. We never show "no name available" because every skill must
// at least have a slug.
function displayName(sk) {
  return (sk.name && sk.name.trim()) || sk.slug
}

// showSlug hides the slug-line when it would just repeat the visible
// name. The slug is informational metadata; rendering it twice for
// skills whose canonical name IS the slug (e.g. `humanizer`) is just
// noise.
function showSlug(sk) {
  const name = (sk.name || '').trim().toLowerCase()
  return !!sk.slug && name !== sk.slug.toLowerCase()
}

function openView(sk) {
  viewDialog.value?.open(sk.id)
}

function onSaved() {
  store.refresh()
}

function usedBy(id) {
  const refs = []
  for (const a of store.agents) {
    if ((a.skills || []).includes(id)) {
      const name = a.name || a.id
      const prompt = a.systemPrompt ? a.systemPrompt.slice(0, 80) + (a.systemPrompt.length > 80 ? '...' : '') : ''
      refs.push({ name, tooltip: a.description || prompt })
    }
  }
  return refs
}

function handleDelete(sk) {
  const label = sk.name || sk.slug
  deleteEntity({
    message: `Delete skill "${label}"? This removes the on-disk package and cannot be undone.`,
    label,
    doDelete: (force) => skillsApi.delete(sk.id, force),
    after: () => store.refresh(),
  })
}
</script>
