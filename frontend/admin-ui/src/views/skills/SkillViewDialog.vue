<template>
  <AppDialog ref="dialogRef" :title="title" size="xl">
    <div v-if="loading" class="text-center text-xs text-arena-500 py-12">Loading…</div>
    <div v-else-if="!skill" class="text-center text-xs text-arena-500 py-12">Skill unavailable.</div>

    <!-- Visual hierarchy intent (read top-to-bottom):
           PRIMARY   — skill name, description, instructions body.
           SECONDARY — tab navigation, current section's content.
           TERTIARY  — slug, file paths, metadata.
                       (read on demand, never grab attention.)

         Concretely:
           - one accent colour (sol, the Magec yellow) reserved for
             the Download CTA and the active-tab underline. Everything
             else is arena/piedra greys. The brand accent earns its
             space; using it anywhere else cheapens it.
           - generous vertical rhythm (`space-y-7`) between bands; the
             modal is meant to be read, not scanned in pieces.
           - a single border-bottom under the tab nav gives a clean
             horizon line; the tabs anchor below it without their own
             background. -->
    <div v-else class="space-y-7">

      <!-- HERO — name carries the weight; everything else is muted.
           Identity chips (license, version, compatibility) are NOT
           shown here. They're frontmatter metadata and live in the
           Metadata tab; surfacing them in the header competes with
           the title and gives every skill a different visual length
           depending on how many tags its author wrote. -->
      <header ref="heroAnchor" class="space-y-4">
        <div class="flex items-center gap-4">
          <!-- Neutral medallion. With license/version/compatibility
               moved to the Metadata tab the title row is single-line,
               so the medallion is sized to match a single line of
               type rather than a stacked title+chips block. -->
          <div class="w-9 h-9 rounded-lg bg-piedra-800/70 border border-piedra-700/50 flex items-center justify-center flex-shrink-0">
            <Icon name="skill" size="sm" class="text-arena-400" />
          </div>

          <div class="min-w-0 flex-1">
            <!-- Title line. Slug rides along as a quiet appendix in
                 the same row, separated by whitespace. Less vertical
                 noise, easier to scan. -->
            <h2 class="text-lg font-semibold text-arena-100 leading-tight">
              <span class="break-words">{{ displayName }}</span>
              <span v-if="showSlug" class="ml-2 text-[11px] font-mono font-normal text-arena-600">{{ skill.slug }}</span>
            </h2>
          </div>

          <!-- Sol (Magec yellow) download button. The viewer is read-
               only; download is the ONLY action the operator can
               take here, so it earns the brand accent. Solid amber
               on piedra-950 text follows the primary-CTA recipe
               from .agents/ADMIN_UI_DESIGN_SYSTEM.md. -->
          <button
            type="button"
            @click="downloadPackage"
            class="flex items-center gap-1.5 px-3 py-1.5 text-[11px] font-medium text-piedra-950 bg-sol-500 hover:bg-sol-600 rounded-lg transition-colors flex-shrink-0"
            title="Download as .tar.gz"
          >
            <Icon name="download" size="xs" />
            <span class="hidden sm:inline">Download</span>
          </button>
        </div>

        <!-- Description: the second-most-important text in the modal,
             so it gets a real type size and generous leading. -->
        <p v-if="skill.description" class="text-[13px] text-arena-300 leading-7 whitespace-pre-line">
          {{ skill.description }}
        </p>
      </header>

      <!-- TABS — minimal nav, no chip backgrounds, no count bubbles
           when zero. Only the active tab gets a colour cue: a thin
           sol-400 underline that echoes the brand accent already used
           by the Download button above and by every primary CTA in
           the admin UI. -->
      <nav class="flex items-center gap-5 border-b border-piedra-800">
        <button
          v-for="tab in availableTabs" :key="tab.id"
          type="button"
          @click="activeTab = tab.id"
          class="relative -mb-px flex items-baseline gap-1.5 pb-2.5 text-[12px] font-medium transition-colors"
          :class="activeTab === tab.id
            ? 'text-arena-100'
            : 'text-arena-500 hover:text-arena-300'"
        >
          <span>{{ tab.label }}</span>
          <span
            v-if="tab.count && tab.count > 0"
            class="text-[10px] font-normal tabular-nums"
            :class="activeTab === tab.id ? 'text-arena-400' : 'text-arena-600'"
          >{{ tab.count }}</span>
          <span
            v-if="activeTab === tab.id"
            class="absolute left-0 right-0 -bottom-px h-px bg-sol-400 rounded-full"
          ></span>
        </button>
      </nav>

      <!-- TAB PANES — every pane has the same outer rhythm; no
           repeated section headers, no decorative dots. The content
           itself defines its own structure. -->

      <!-- Instructions: the heavy lifter. The Copy button only
           reveals on hover; it does not crowd the title or the
           markdown surface. -->
      <section v-show="activeTab === 'instructions'" class="relative">
        <div v-if="renderedHtml" class="relative group/code">
          <button
            type="button"
            @click="copyInstructions"
            class="absolute top-0 right-0 z-10 flex items-center gap-1 px-2 py-1 text-[10px] rounded-md text-arena-500 hover:text-arena-200 hover:bg-piedra-800 opacity-0 group-hover/code:opacity-100 transition-opacity"
            :title="copied ? 'Copied' : 'Copy SKILL.md instructions'"
          >
            <Icon name="copy" size="xs" />
            <span>{{ copied ? 'Copied' : 'Copy' }}</span>
          </button>
          <article class="magec-markdown" v-html="renderedHtml"></article>
        </div>
        <EmptyTab v-else icon="skill" message="No instructions in SKILL.md." />
      </section>

      <!-- Resources: each kind gets a quiet sub-header, then a list
           of rows. No heavy borders, no per-row hover dance — just
           legible rows with breathing room. -->
      <section v-show="activeTab === 'resources'" class="space-y-6">
        <template v-if="resourceKinds.length">
          <div v-for="kind in resourceKinds" :key="kind">
            <div class="flex items-baseline justify-between mb-3">
              <h4 class="text-[11px] text-arena-400">
                {{ kindTitle(kind) }}
                <span class="text-arena-600 ml-1">{{ groupedResources[kind].length }}</span>
              </h4>
              <span class="text-[10px] text-arena-600 tabular-nums">{{ formatSize(groupTotalSize(kind)) }}</span>
            </div>
            <ul class="divide-y divide-piedra-800/70">
              <li v-for="r in groupedResources[kind]" :key="r.path"
                  class="flex items-center justify-between gap-3 py-2.5">
                <div class="flex items-center gap-3 min-w-0">
                  <span class="text-[9px] font-mono uppercase text-arena-600 w-8 flex-shrink-0">{{ extOf(r.path) }}</span>
                  <p class="text-[12px] text-arena-300 font-mono truncate">{{ relPath(r) }}</p>
                </div>
                <span class="text-[10px] text-arena-600 tabular-nums flex-shrink-0">{{ formatSize(r.size) }}</span>
              </li>
            </ul>
          </div>
        </template>
        <EmptyTab v-else icon="upload" message="This skill does not bundle any references, assets or scripts." />
      </section>

      <!-- Metadata: the whole frontmatter. Every key the operator
           wrote in SKILL.md shows up here — scalar fields in a
           two-column dl, list-shaped fields (allowed-tools, tags…)
           rendered as inline chips, and the canonical `metadata`
           map gets its own grouped block at the end so YAML maps
           don't collapse into stringy noise. -->
      <section v-show="activeTab === 'metadata'">
        <div v-if="hasAnyMetadata" class="space-y-6">
          <!-- Scalar fields. license, compatibility, version, name,
               description and any custom string the operator added. -->
          <dl
            v-if="metadataScalars.length"
            class="grid grid-cols-1 sm:grid-cols-[180px_1fr] gap-x-6 gap-y-3"
          >
            <template v-for="p in metadataScalars" :key="p.key">
              <dt class="text-[11px] text-arena-500 font-mono">{{ p.key }}</dt>
              <dd class="text-[12px] text-arena-200 break-words leading-6">{{ p.value }}</dd>
            </template>
          </dl>

          <!-- List-shaped fields (allowed-tools, tags…). One row per
               key with the values laid out as muted chips. -->
          <div v-if="metadataLists.length" class="space-y-4">
            <div
              v-for="entry in metadataLists" :key="entry.key"
              class="grid grid-cols-1 sm:grid-cols-[180px_1fr] gap-x-6 gap-y-2 items-start"
            >
              <span class="text-[11px] text-arena-500 font-mono pt-0.5">{{ entry.key }}</span>
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="(item, idx) in entry.items" :key="`${entry.key}-${idx}`"
                  class="inline-flex items-center px-2 py-0.5 rounded-md bg-piedra-800/70 text-[11px] text-arena-300 font-mono"
                >{{ item }}</span>
              </div>
            </div>
          </div>

          <!-- Canonical `metadata:` map (Agent Skills spec). Same
               two-column layout but indented under a small title so
               it reads as a sub-document. -->
          <div v-if="metadataMap.length">
            <h4 class="text-[10px] uppercase tracking-wider text-arena-500 mb-3">metadata</h4>
            <dl class="grid grid-cols-1 sm:grid-cols-[180px_1fr] gap-x-6 gap-y-3 pl-3 border-l border-piedra-800">
              <template v-for="[k, v] in metadataMap" :key="k">
                <dt class="text-[11px] text-arena-500 font-mono">{{ k }}</dt>
                <dd class="text-[12px] text-arena-200 break-words leading-6">{{ v }}</dd>
              </template>
            </dl>
          </div>
        </div>
        <EmptyTab v-else icon="command" message="No frontmatter fields declared in this SKILL.md." />
      </section>

      <!-- Back-to-top floating button. Sits at the end of the content
           flow with `position: sticky; bottom: 1rem`, glued to the
           bottom of the scroll viewport. The wrapper has zero height
           and disabled pointer events so it doesn't reserve layout
           space; the button re-enables pointer events for itself.
           Visibility is bound to the IntersectionObserver watching
           the hero header. -->
      <div class="sticky bottom-8 -mt-7 h-0 pointer-events-none flex justify-end">
        <button
          type="button"
          @click="scrollToTop"
          class="pointer-events-auto flex items-center justify-center w-9 h-9 rounded-full bg-piedra-800/95 backdrop-blur-sm border border-piedra-700/60 text-arena-400 hover:text-arena-100 hover:border-piedra-600 shadow-lg transition-all duration-150"
          :class="showBackToTop ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2 pointer-events-none'"
          :tabindex="showBackToTop ? 0 : -1"
          :aria-hidden="!showBackToTop"
          title="Back to top"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
          </svg>
        </button>
      </div>
    </div>

    <template #footer>
      <button type="button" @click="close" class="px-4 py-2 text-sm text-arena-400 hover:text-arena-200 hover:bg-piedra-800 rounded-lg transition-colors">
        Close
      </button>
    </template>
  </AppDialog>
</template>

<script setup>
import { ref, computed, inject, onBeforeUnmount, nextTick, watch } from 'vue'
import { renderMarkdown } from '../../lib/markdown.js'
import { skillsApi } from '../../lib/api/index.js'
import { getAuthHeaders } from '../../lib/auth.js'
import AppDialog from '../../components/AppDialog.vue'
import Icon from '../../components/Icon.vue'
import EmptyTab from './EmptyTab.vue'

const toast = inject('toast')
const dialogRef = ref(null)
const skill = ref(null)
const loading = ref(false)
const activeTab = ref('instructions')
const copied = ref(false)
const heroAnchor = ref(null)
const showBackToTop = ref(false)
let heroObserver = null

// --- Identity (header) -----------------------------------------------------

const displayName = computed(() => {
  if (!skill.value) return 'Skill'
  const name = (skill.value.name || '').trim()
  return name || skill.value.slug
})

const showSlug = computed(() => {
  if (!skill.value?.slug) return false
  const name = (skill.value.name || '').trim().toLowerCase()
  return name !== skill.value.slug.toLowerCase()
})

const title = computed(() => displayName.value)
const frontmatter = computed(() => skill.value?.frontmatter || {})

// --- Tab content -----------------------------------------------------------

const renderedHtml = computed(() => renderMarkdown(skill.value?.instructions))

const resourceKinds = computed(() => {
  if (!skill.value) return []
  const kinds = new Set((skill.value.resources || []).map((r) => r.kind))
  return ['references', 'assets', 'scripts'].filter((k) => kinds.has(k))
})

const groupedResources = computed(() => {
  const out = { references: [], assets: [], scripts: [] }
  for (const r of skill.value?.resources || []) {
    if (out[r.kind]) out[r.kind].push(r)
  }
  for (const k of Object.keys(out)) {
    out[k].sort((a, b) => a.path.localeCompare(b.path))
  }
  return out
})

const totalResources = computed(() => (skill.value?.resources || []).length)

// --- Metadata tab ----------------------------------------------------------
//
// The Metadata tab surfaces the FULL frontmatter the operator wrote
// in SKILL.md. We split it into three buckets so the layout adapts to
// each value's shape:
//
//   • metadataScalars — strings/numbers/booleans. Shown in a
//     two-column dl. Includes license/version/compatibility (which
//     used to be pinned near the title — now they live here, where
//     the rest of the metadata lives, so every skill's header has
//     the same visual length regardless of how decorated its
//     frontmatter is).
//   • metadataLists   — array fields like allowed-tools or tags.
//     Each list gets its own row with the items rendered as muted
//     mono chips.
//   • metadataMap     — the canonical `metadata:` map from the
//     Agent Skills spec, rendered as an indented sub-document so
//     YAML maps don't squash into stringy noise.
//
// `name` and `description` are deliberately omitted — they're
// already prominent in the hero header; repeating them here would
// be noise.

const HERO_OWNED_KEYS = new Set(['name', 'description'])
const RESERVED_LIST_KEYS = new Set(['allowed-tools', 'tags'])

const metadataScalars = computed(() => {
  const fm = frontmatter.value
  const out = []
  for (const [k, v] of Object.entries(fm)) {
    if (HERO_OWNED_KEYS.has(k)) continue
    if (k === 'metadata') continue
    if (v === null || v === undefined || v === '') continue
    if (Array.isArray(v)) continue
    if (typeof v === 'object') continue
    out.push({ key: k, value: stringy(v) })
  }
  return out.sort(metadataKeyOrder)
})

const metadataLists = computed(() => {
  const fm = frontmatter.value
  const out = []
  for (const [k, v] of Object.entries(fm)) {
    if (HERO_OWNED_KEYS.has(k)) continue
    if (!Array.isArray(v) || v.length === 0) continue
    out.push({ key: k, items: v.map((x) => stringy(x)) })
  }
  // Push canonical lists first (allowed-tools, tags), then anything
  // else alphabetically — same intent as the scalar order helper.
  return out.sort((a, b) => {
    const ax = RESERVED_LIST_KEYS.has(a.key) ? 0 : 1
    const bx = RESERVED_LIST_KEYS.has(b.key) ? 0 : 1
    if (ax !== bx) return ax - bx
    return a.key.localeCompare(b.key)
  })
})

const metadataMap = computed(() => {
  const meta = frontmatter.value.metadata
  if (!meta || typeof meta !== 'object' || Array.isArray(meta)) return []
  return Object.entries(meta)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([k, v]) => [k, stringy(v)])
})

const hasAnyMetadata = computed(
  () => metadataScalars.value.length + metadataLists.value.length + metadataMap.value.length > 0,
)

// metadataKeyOrder pushes the most-recognisable spec fields to the
// top of the scalar list so license/compatibility/version sit
// where the operator instinctively expects them, regardless of
// alphabetical order.
const PINNED_SCALAR_KEYS = ['license', 'compatibility', 'version']
function metadataKeyOrder(a, b) {
  const ai = PINNED_SCALAR_KEYS.indexOf(a.key)
  const bi = PINNED_SCALAR_KEYS.indexOf(b.key)
  if (ai !== -1 || bi !== -1) {
    if (ai === -1) return 1
    if (bi === -1) return -1
    return ai - bi
  }
  return a.key.localeCompare(b.key)
}

const totalMetadataCount = computed(
  () => metadataScalars.value.length + metadataLists.value.length + metadataMap.value.length,
)

const availableTabs = computed(() => [
  { id: 'instructions', label: 'Instructions' },
  { id: 'resources',    label: 'Resources', count: totalResources.value },
  { id: 'metadata',     label: 'Metadata',  count: totalMetadataCount.value },
])

// --- Helpers ---------------------------------------------------------------

function relPath(r) {
  const prefix = r.kind + '/'
  return r.path.startsWith(prefix) ? r.path.slice(prefix.length) : r.path
}

function kindTitle(kind) {
  return { references: 'References', assets: 'Assets', scripts: 'Scripts' }[kind] || kind
}

function groupTotalSize(kind) {
  return groupedResources.value[kind].reduce((sum, r) => sum + (r.size || 0), 0)
}

// extOf returns a 1-3 letter extension for the resource list icon.
// Falls back to "FILE" for paths without an extension. We don't try
// to map MIME types here — the extension is what the operator sees in
// their editor and what the skill author chose, so it's the most
// faithful identifier.
function extOf(path) {
  const m = /\.([A-Za-z0-9]{1,5})$/.exec(path)
  if (!m) return 'FILE'
  return m[1].toUpperCase().slice(0, 3)
}

function formatSize(bytes) {
  if (!bytes && bytes !== 0) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function stringy(v) {
  if (v === null || v === undefined) return ''
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

// --- Lifecycle -------------------------------------------------------------

async function open(idOrSummary) {
  loading.value = true
  skill.value = null
  activeTab.value = 'instructions'
  copied.value = false
  showBackToTop.value = false
  dialogRef.value?.open()
  try {
    const id = typeof idOrSummary === 'string' ? idOrSummary : idOrSummary?.id
    skill.value = await skillsApi.get(id)
  } catch (e) {
    toast.error(e.message)
    dialogRef.value?.close()
    return
  } finally {
    loading.value = false
  }
  // Set up the observer AFTER loading flips to false. While loading
  // is true the v-if branch shows the placeholder and the v-else
  // (which contains the hero header) is NOT in the DOM, so
  // heroAnchor.value would be null. Two ticks of waiting guarantee
  // the new template tree is committed before we query it.
  await nextTick()
  await nextTick()
  setupHeroObserver()
}

function close() {
  teardownHeroObserver()
  dialogRef.value?.close()
}

// setupHeroObserver wires an IntersectionObserver against the hero
// header so we can flip showBackToTop based on whether the header is
// currently in the scroll viewport. The observer's `root` is the
// nearest `.overflow-y-auto` ancestor — that's the AppDialog's
// internal scroll area, which is where the user actually scrolls.
//
// We don't bother if the browser doesn't support IntersectionObserver
// (none of the modern targets we care about lack it), and we tolerate
// the observer misfiring on dialog teardown by guarding teardown in
// teardownHeroObserver.
function setupHeroObserver() {
  teardownHeroObserver()
  if (typeof IntersectionObserver === 'undefined') return
  if (!heroAnchor.value) return
  const root = findScrollRoot(heroAnchor.value)
  heroObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0]
      if (!entry) return
      // Show the back-to-top button as soon as the hero is even
      // partially out of view. Using isIntersecting alone (with
      // default threshold = 0) means: "show the button the moment
      // the hero crosses the top edge of the scroll viewport".
      showBackToTop.value = !entry.isIntersecting
    },
    { root: root || null, threshold: 0 },
  )
  heroObserver.observe(heroAnchor.value)
}

function teardownHeroObserver() {
  if (heroObserver) {
    heroObserver.disconnect()
    heroObserver = null
  }
}

// findScrollRoot walks up the DOM from a starting element looking
// for the nearest scrollable ancestor. The AppDialog's scroll area
// has `overflow-y: auto` which produces `overflow-y: auto` in the
// computed style — that's the marker we use. Returns null when no
// scroll ancestor exists; the IntersectionObserver then falls back
// to the document viewport.
function findScrollRoot(el) {
  let node = el?.parentElement
  while (node) {
    const overflow = window.getComputedStyle(node).overflowY
    if (overflow === 'auto' || overflow === 'scroll') return node
    node = node.parentElement
  }
  return null
}

function scrollToTop() {
  const root = heroAnchor.value && findScrollRoot(heroAnchor.value)
  const target = root || window
  if (target.scrollTo) {
    target.scrollTo({ top: 0, behavior: 'smooth' })
  } else if (root) {
    root.scrollTop = 0
  }
}

// Re-attach the observer if the operator opens a different skill
// without closing the dialog first (rare, but possible from the
// list view's keyboard shortcuts).
watch(skill, (val) => {
  if (val) nextTick(setupHeroObserver)
  else teardownHeroObserver()
})

onBeforeUnmount(teardownHeroObserver)

async function copyInstructions() {
  if (!skill.value?.instructions) return
  try {
    await navigator.clipboard.writeText(skill.value.instructions)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1500)
  } catch (e) {
    toast.error('Could not copy: ' + e.message)
  }
}

async function downloadPackage() {
  if (!skill.value) return
  try {
    const res = await fetch(skillsApi.downloadUrl(skill.value.id), { headers: { ...getAuthHeaders() } })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const blob = await res.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `${skill.value.slug}.tar.gz`
    a.click()
    URL.revokeObjectURL(a.href)
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
