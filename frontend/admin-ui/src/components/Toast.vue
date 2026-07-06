<template>
  <Teleport :to="teleportTarget">
    <!--
      Two tricks combined so toasts paint above AND stay clickable over modals:
      1. popover="manual" promotes the stack to the browser top layer, above
         any <dialog>.showModal() veil; plain z-index never wins against it.
      2. showModal() makes everything outside the dialog's DOM subtree inert
         (visible or not), so the stack teleports INTO the topmost open modal
         while one exists and back to body when it closes — same lesson as
         NodeHelp. Position is unaffected: the popover top layer + fixed
         positioning resolve against the viewport from any DOM spot.
    -->
    <div ref="popRef" popover="manual" class="toast-pop fixed z-[9999] flex flex-col gap-2 pointer-events-none">
      <TransitionGroup
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="translate-y-2 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-x-4 opacity-0"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto px-4 py-3 rounded-xl bg-piedra-800 shadow-lg shadow-black/30 min-w-[240px] max-w-[360px]"
        >
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full flex-shrink-0" :class="dotClass(toast.type)" />
            <span class="text-[11px] font-bold text-arena-300 flex-1">{{ headerLabel(toast.type) }}</span>
            <button
              @click="dismiss(toast.id)"
              class="w-5 h-5 flex-shrink-0 flex items-center justify-center rounded-md hover:bg-piedra-700 transition-colors"
            >
              <svg class="w-3 h-3 text-arena-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <p class="mt-1.5 text-xs font-medium leading-snug text-arena-400">{{ toast.message }}</p>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'

const toasts = ref([])
const popRef = ref(null)
const teleportTarget = ref(document.body)
let nextId = 0
let hideTimer = null

// topmostModal returns the open modal dialog the stack should live in, or
// null when none is open. DOM order approximates top-layer order well enough
// here (stacked modals are rare in the admin UI).
function topmostModal() {
  const modals = document.querySelectorAll('dialog:modal')
  return modals.length ? modals[modals.length - 1] : null
}

// retarget moves the stack into the topmost open modal (so it escapes the
// inertness showModal() imposes on the rest of the document) or back to body.
function retarget() {
  const next = topmostModal() || document.body
  if (next !== teleportTarget.value && next.isConnected) {
    teleportTarget.value = next
  }
}

// Dialogs announce open/close via ToggleEvent ('toggle') and close via
// 'close'; neither bubbles, but capture listeners on the document still see
// them. Retargeting on both keeps live toasts interactive when a modal opens
// mid-toast and returns them to body when it closes.
function onDialogToggle(e) {
  if (e.target instanceof HTMLDialogElement) retarget()
}
onMounted(() => {
  document.addEventListener('toggle', onDialogToggle, true)
  document.addEventListener('close', onDialogToggle, true)
})
onBeforeUnmount(() => {
  document.removeEventListener('toggle', onDialogToggle, true)
  document.removeEventListener('close', onDialogToggle, true)
})

// Keep the popover in sync with the toast list: shown while there are toasts,
// hidden when empty. Re-showing on every addition re-promotes the stack to the
// top of the top layer (above any modal dialog currently open). Hiding waits
// for the leave transition (200ms) so the last toast animates out. Browsers
// without the Popover API fall back to the fixed/z-index positioning.
watch(() => toasts.value.length, (len) => {
  const el = popRef.value
  if (!el || typeof el.showPopover !== 'function') return
  clearTimeout(hideTimer)
  if (len > 0) {
    if (el.matches(':popover-open')) el.hidePopover()
    el.showPopover()
    return
  }
  hideTimer = setTimeout(() => {
    if (el.matches(':popover-open')) el.hidePopover()
  }, 250)
})

// Moving a shown popover in the DOM auto-hides it, so after every teleport
// re-show it once Vue has finished moving the nodes.
watch(teleportTarget, async () => {
  await nextTick()
  const el = popRef.value
  if (!el || typeof el.showPopover !== 'function') return
  if (toasts.value.length > 0 && !el.matches(':popover-open')) el.showPopover()
})

// dotClass picks the status dot color: Magec yellow for success/info, lava
// for errors (an error announced in cheerful yellow would be lying).
function dotClass(type) {
  if (type === 'error') return 'bg-lava-400'
  return 'bg-sol-400'
}

// headerLabel gives the toast header its short greeting per type.
function headerLabel(type) {
  if (type === 'error') return 'Ouch!'
  return 'Hey!'
}

function show(message, type = 'info', duration = 3000) {
  retarget()
  const id = ++nextId
  toasts.value.push({ id, message, type })
  if (duration > 0) {
    setTimeout(() => dismiss(id), duration)
  }
}

function dismiss(id) {
  const i = toasts.value.findIndex(t => t.id === id)
  if (i !== -1) toasts.value.splice(i, 1)
}

function success(message) { show(message, 'success') }
function error(message) { show(message, 'error', 5000) }
function info(message) { show(message, 'info') }

defineExpose({ show, success, error, info })
</script>

<style scoped>
/* Reset the UA popover styles (inset: 0, margin: auto, border, canvas
   background) and pin the stack to the bottom-right corner. Also applies in
   browsers without the Popover API, where the element is always rendered. */
.toast-pop {
  inset: auto 1rem 1rem auto;
  margin: 0;
  padding: 0;
  border: 0;
  background: transparent;
  overflow: visible;
}
</style>
