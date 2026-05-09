<template>
  <AppDialog ref="dialogRef" :title="title" size="2xl">
    <div v-if="loading" class="text-center text-xs text-arena-500 py-12">Loading…</div>
    <div v-else-if="!skill" class="text-center text-xs text-arena-500 py-12">Skill unavailable.</div>

    <!-- The viewer is structured as three vertical bands: hero (identity),
         status strip (at-a-glance metrics), and a tabbed content area.
         The bands sit directly on the dialog background — no nested
         cards-inside-a-card — so the whole modal reads as a single
         composed sheet, not a stack of fragments.

         Cyan is the Skills accent (decision: ENTITY_COLORS.md). We use
         it in three carefully chosen places: the hero icon ring, the
         active tab pill, and the section labels' tiny accent dot. Any
         more cyan would feel like screaming. -->
    <div v-else class="space-y-5">

      <!-- HERO -->
      <header ref="heroAnchor" class="relative">
        <div class="flex items-start gap-4">
          <!-- Icon medallion. The double-ring (15% + 30%) gives depth
               without saturating; same recipe Cards use for the entity
               accent. -->
          <div class="relative flex-shrink-0">
            <div class="w-14 h-14 rounded-2xl bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center">
              <Icon name="skill" size="xl" class="text-cyan-300" />
            </div>
            <div class="absolute inset-0 rounded-2xl ring-1 ring-cyan-500/10 pointer-events-none"></div>
          </div>

          <div class="min-w-0 flex-1">
            <h2 class="text-xl font-semibold text-arena-100 truncate">{{ displayName }}</h2>
            <p v-if="showSlug" class="text-[11px] text-arena-500 font-mono mt-0.5">{{ skill.slug }}</p>

            <!-- Compatibility & license chips travel with the title —
                 they are identity, not generic metadata. -->
            <div v-if="identityChips.length" class="flex flex-wrap gap-1.5 mt-2">
              <span
                v-for="c in identityChips" :key="c.label + c.value"
                class="inline-flex items-center gap-1 px-2 py-0.5 text-[10px] rounded-md border"
                :class="c.accent
                  ? 'bg-cyan-500/10 border-cyan-500/25 text-cyan-300'
                  : 'bg-piedra-800 border-piedra-700/50 text-arena-400'"
              >
                <span v-if="c.label" class="text-arena-600">{{ c.label }}</span>
                <span class="font-medium">{{ c.value }}</span>
              </span>
            </div>
          </div>

          <!-- Quick action stays out of the title flow so the eye
               always lands on the name first. -->
          <button
            type="button"
            @click="downloadPackage"
            class="flex items-center gap-1.5 px-2.5 py-1.5 text-[11px] font-medium rounded-lg bg-piedra-800 hover:bg-piedra-700 border border-piedra-700/50 text-arena-300 hover:text-arena-100 transition-colors flex-shrink-0"
            title="Download as .tar.gz"
          >
            <Icon name="download" size="xs" />
            <span class="hidden sm:inline">Download</span>
          </button>
        </div>

        <!-- Description gets its own band below the title row so long
             paragraphs don't squish next to the icon medallion. The
             whitespace-pre-line lets multi-line YAML descriptions
             (folded scalars) read naturally. -->
        <p v-if="skill.description" class="mt-4 text-sm text-arena-300 leading-relaxed whitespace-pre-line">
          {{ skill.description }}
        </p>
      </header>

      <!-- TABS — clickable section pills. They scroll with the rest
           of the content (not sticky); a small "back to top" floating
           button below appears when the header has scrolled out of
           the viewport so the operator can jump back without losing
           the tab they were on. -->
      <div ref="tabsAnchor" class="flex items-center gap-1.5 border-b border-piedra-700/40">
        <button
          v-for="tab in availableTabs" :key="tab.id"
          type="button"
          @click="activeTab = tab.id"
          class="relative flex items-center gap-1.5 px-3 py-2 text-[11px] font-medium transition-colors"
          :class="activeTab === tab.id
            ? 'text-cyan-300'
            : 'text-arena-500 hover:text-arena-300'"
        >
          <Icon :name="tab.icon" size="xs" />
          <span>{{ tab.label }}</span>
          <span v-if="tab.count !== undefined && tab.count > 0"
                class="ml-0.5 px-1.5 py-0 text-[9px] rounded-full"
                :class="activeTab === tab.id
                  ? 'bg-cyan-500/20 text-cyan-300'
                  : 'bg-piedra-800 text-arena-500'">
            {{ tab.count }}
          </span>
          <span
            v-if="activeTab === tab.id"
            class="absolute left-0 right-0 -bottom-px h-0.5 bg-cyan-400 rounded-full"
          ></span>
        </button>
      </div>

      <!-- TAB PANES -->

      <!-- Instructions: the main event. Markdown rendered with the
           dedicated .magec-markdown stylesheet. A floating Copy button
           sits over the panel so the operator can yank the SKILL.md
           body without selecting it manually. -->
      <section v-show="activeTab === 'instructions'" class="relative">
        <div v-if="renderedHtml" class="relative group/code">
          <button
            type="button"
            @click="copyInstructions"
            class="absolute top-0 right-0 z-10 flex items-center gap-1 px-2 py-1 text-[10px] rounded-md bg-piedra-800/90 backdrop-blur-sm border border-piedra-700/60 text-arena-400 hover:text-arena-100 hover:bg-piedra-700 opacity-0 group-hover/code:opacity-100 transition-opacity"
            :title="copied ? 'Copied' : 'Copy SKILL.md instructions'"
          >
            <Icon name="copy" size="xs" />
            <span>{{ copied ? 'Copied' : 'Copy' }}</span>
          </button>
          <article class="magec-markdown" v-html="renderedHtml"></article>
        </div>
        <EmptyTab v-else icon="skill" message="No instructions in SKILL.md." />
      </section>

      <!-- Tools: pill grid. The lock icon doubles as a hint that these
           are scoped permissions (granted by the SKILL.md frontmatter,
           not by the agent). -->
      <section v-show="activeTab === 'tools'">
        <div v-if="allowedTools.length" class="flex flex-wrap gap-1.5">
          <span
            v-for="t in allowedTools" :key="t"
            class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-piedra-800/60 border border-piedra-700/50 hover:border-cyan-500/30 transition-colors"
          >
            <Icon name="key" size="xs" class="text-cyan-400/70" />
            <span class="text-[11px] text-arena-200 font-mono">{{ t }}</span>
          </span>
        </div>
        <EmptyTab v-else icon="key" message="This skill does not declare an allowed-tools list." />
      </section>

      <!-- Resources: grouped per kind with a small per-group header
           (count + total size). Rows are visually layered (icon, path,
           size) and use mono for paths since they're file system
           identifiers, not prose. -->
      <section v-show="activeTab === 'resources'" class="space-y-4">
        <template v-if="resourceKinds.length">
          <div v-for="kind in resourceKinds" :key="kind">
            <div class="flex items-baseline justify-between mb-2">
              <h4 class="flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-widest text-arena-400">
                <span class="w-1 h-1 rounded-full bg-cyan-400/70"></span>
                {{ kindTitle(kind) }}
                <span class="text-arena-600">·</span>
                <span class="text-arena-500 normal-case tracking-normal">{{ groupedResources[kind].length }} {{ groupedResources[kind].length === 1 ? 'file' : 'files' }}</span>
              </h4>
              <span class="text-[10px] text-arena-600 tabular-nums">{{ formatSize(groupTotalSize(kind)) }}</span>
            </div>
            <ul class="space-y-1">
              <li v-for="r in groupedResources[kind]" :key="r.path"
                  class="flex items-center justify-between gap-3 px-3 py-2 rounded-lg border border-piedra-700/30 bg-piedra-800/30 hover:border-cyan-500/25 hover:bg-piedra-800/60 transition-colors group/file">
                <div class="flex items-center gap-2.5 min-w-0">
                  <div class="w-7 h-7 rounded-md flex items-center justify-center flex-shrink-0 bg-piedra-900 border border-piedra-700/50">
                    <span class="text-[8px] font-bold uppercase tracking-tight text-arena-500 group-hover/file:text-cyan-300">
                      {{ extOf(r.path) }}
                    </span>
                  </div>
                  <div class="min-w-0">
                    <p class="text-[11px] text-arena-200 font-mono truncate">{{ relPath(r) }}</p>
                  </div>
                </div>
                <span class="text-[10px] text-arena-500 tabular-nums flex-shrink-0">{{ formatSize(r.size) }}</span>
              </li>
            </ul>
          </div>
        </template>
        <EmptyTab v-else icon="upload" message="This skill does not bundle any references, assets or scripts." />
      </section>

      <!-- Metadata: the catch-all surface for everything else in the
           frontmatter (custom keys + the canonical metadata map). Two-
           column dl layout. -->
      <section v-show="activeTab === 'metadata'">
        <div v-if="metadataPairs.length || metadataMap.length" class="space-y-5">
          <div v-if="metadataPairs.length">
            <h4 class="flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-widest text-arena-400 mb-2">
              <span class="w-1 h-1 rounded-full bg-cyan-400/70"></span>
              Frontmatter fields
            </h4>
            <dl class="grid grid-cols-1 sm:grid-cols-3 gap-x-4 gap-y-1 border border-piedra-700/30 rounded-lg p-3 bg-piedra-800/20">
              <template v-for="p in metadataPairs" :key="p.key">
                <dt class="sm:col-span-1 text-[10px] text-arena-500 font-mono uppercase tracking-wider self-center">{{ p.key }}</dt>
                <dd class="sm:col-span-2 text-[11px] text-arena-200 break-words">{{ p.value }}</dd>
              </template>
            </dl>
          </div>

          <div v-if="metadataMap.length">
            <h4 class="flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-widest text-arena-400 mb-2">
              <span class="w-1 h-1 rounded-full bg-cyan-400/70"></span>
              metadata
            </h4>
            <dl class="grid grid-cols-1 sm:grid-cols-3 gap-x-4 gap-y-1 border border-piedra-700/30 rounded-lg p-3 bg-piedra-800/20">
              <template v-for="[k, v] in metadataMap" :key="k">
                <dt class="sm:col-span-1 text-[10px] text-arena-500 font-mono uppercase tracking-wider self-center">{{ k }}</dt>
                <dd class="sm:col-span-2 text-[11px] text-arena-200 break-words">{{ v }}</dd>
              </template>
            </dl>
          </div>
        </div>
        <EmptyTab v-else icon="command" message="No additional metadata declared in this SKILL.md." />
      </section>

      <!-- Back-to-top floating button. Sits at the end of the content
           flow with `position: sticky; bottom: 1rem`, which glues it
           to the bottom of the scroll viewport while still being part
           of the document. We wrap it in a zero-height div with
           pointer-events disabled so the button doesn't reserve
           layout space (no awkward gap below the last section). The
           button itself re-enables pointer events. Visibility is
           bound to whether the hero header is in the scroll viewport
           (IntersectionObserver, set up on dialog open). -->
      <div class="sticky bottom-4 -mt-5 h-0 pointer-events-none flex justify-end">
        <button
          type="button"
          @click="scrollToTop"
          class="pointer-events-auto flex items-center justify-center w-9 h-9 rounded-full bg-piedra-800/95 backdrop-blur-sm border border-piedra-700/60 text-arena-300 hover:text-cyan-300 hover:border-cyan-500/40 hover:bg-piedra-800 shadow-lg transition-all duration-150"
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
import { ref, computed, inject, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
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
const tabsAnchor = ref(null)
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

// identityChips collects the few frontmatter fields that act as
// identity markers (license, compatibility, version) so we can pin
// them next to the title. We keep it small on purpose — the
// Metadata tab carries the full picture.
const identityChips = computed(() => {
  const fm = frontmatter.value
  const out = []
  if (fm.license) out.push({ label: '', value: fm.license, accent: false })
  if (fm.compatibility) out.push({ label: '', value: stringy(fm.compatibility), accent: true })
  if (fm.version) out.push({ label: 'v', value: stringy(fm.version), accent: false })
  return out
})

// --- Tabs / status ---------------------------------------------------------

const allowedTools = computed(() => {
  const arr = frontmatter.value['allowed-tools']
  return Array.isArray(arr) ? arr.map((t) => String(t)) : []
})

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

// metadataPairs = top-level frontmatter entries that aren't already
// consumed by another section. Lists/maps drop into `metadataMap`.
const SECTION_OWNED_KEYS = new Set(['name', 'description', 'metadata', 'allowed-tools'])
const HEADER_OWNED_KEYS = new Set(['license', 'compatibility', 'version'])

const metadataPairs = computed(() => {
  const fm = frontmatter.value
  const out = []
  for (const [k, v] of Object.entries(fm)) {
    if (SECTION_OWNED_KEYS.has(k)) continue
    if (HEADER_OWNED_KEYS.has(k)) continue
    if (v === null || v === undefined || v === '') continue
    if (typeof v === 'object') continue
    out.push({ key: k, value: stringy(v) })
  }
  return out.sort((a, b) => a.key.localeCompare(b.key))
})

const metadataMap = computed(() => {
  const meta = frontmatter.value.metadata
  if (!meta || typeof meta !== 'object') return []
  return Object.entries(meta).sort(([a], [b]) => a.localeCompare(b)).map(([k, v]) => [k, stringy(v)])
})

const availableTabs = computed(() => [
  { id: 'instructions', label: 'Instructions', icon: 'skill' },
  { id: 'tools',        label: 'Tools',        icon: 'key',     count: allowedTools.value.length },
  { id: 'resources',    label: 'Resources',    icon: 'upload',  count: totalResources.value },
  { id: 'metadata',     label: 'Metadata',     icon: 'command', count: metadataPairs.value.length + metadataMap.value.length },
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
    // Wait for the skill content to render before wiring the
    // IntersectionObserver — heroAnchor.value is null while
    // skill.value is null because the v-else renders nothing.
    await nextTick()
    setupHeroObserver()
  } catch (e) {
    toast.error(e.message)
    dialogRef.value?.close()
  } finally {
    loading.value = false
  }
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

onMounted(() => {
  // No-op on mount; the observer is created on dialog open. Kept as
  // an explicit hook so the lifecycle is obvious to the next reader.
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
