<!-- Small info-icon trigger that reveals a help panel for flow nodes. The panel
     is a native popover: the flow editor lives inside a <dialog> opened with
     showModal(), which sits in the browser top layer, so a plain teleported
     element is hidden behind it regardless of z-index. showPopover() promotes
     the panel into the same top layer, above the dialog. -->
<template>
  <button
    ref="triggerRef"
    type="button"
    class="inline-flex items-center gap-1 text-arena-600 hover:text-arena-400 transition-colors"
    @pointerdown.stop
    @click.stop="toggle"
    @mouseenter="show"
    @mouseleave="scheduleClose"
  >
    <span v-if="label" class="text-[9px] font-medium leading-none normal-case tracking-normal">{{ label }}</span>
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" class="w-3 h-3 relative -top-[0.5px]">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="1.5"
        d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z"
      />
    </svg>
  </button>

  <Teleport to="body">
    <div
      ref="popRef"
      popover="manual"
      class="nodehelp-pop w-[280px] bg-piedra-950 border border-piedra-700/60 rounded-lg shadow-xl p-2.5 space-y-1.5 text-arena-400"
      :style="{ left: left + 'px', top: top + 'px' }"
      @mouseenter="cancelClose"
      @mouseleave="scheduleClose"
    >
      <p v-if="title" class="text-arena-200 text-[10px] font-semibold">{{ title }}</p>
      <div v-for="(section, idx) in sections" :key="idx" class="space-y-1">
        <p v-if="section.heading" class="text-arena-300 text-[9px] font-medium">{{ section.heading }}</p>
        <p v-if="section.body" class="text-arena-400 text-[9px] leading-relaxed">{{ section.body }}</p>
        <ul v-if="section.items && section.items.length" class="space-y-0.5">
          <li v-for="(item, itemIdx) in section.items" :key="itemIdx" class="flex gap-1.5 text-[9px] leading-relaxed">
            <span class="text-arena-600">•</span>
            <span><code class="font-mono text-atlantico-300">{{ item.name }}</code><span class="text-arena-400 ml-1.5">{{ item.desc }}</span></span>
          </li>
        </ul>
        <div v-if="section.code && section.code.length" class="flex flex-wrap gap-1">
          <span
            v-for="(codeLine, codeIdx) in section.code"
            :key="codeIdx"
            class="inline-block font-mono text-[9px] text-atlantico-300 bg-piedra-800 rounded px-1 py-0.5"
          >{{ codeLine }}</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, nextTick, onBeforeUnmount } from 'vue'

const POPOVER_WIDTH = 280
const MARGIN = 8
const CLOSE_DELAY_MS = 120

const props = defineProps({
  title:    { type: String, default: '' },
  label:    { type: String, default: '' },
  sections: { type: Array, default: () => [] },
})

const triggerRef = ref(null)
const popRef = ref(null)
const visible = ref(false)
const left = ref(0)
const top = ref(0)

let closeTimer = null

// place positions the panel next to the trigger using viewport coordinates.
// Top-layer elements use the viewport as containing block, so these values
// stay correct even when the canvas is panned, zoomed or fullscreen.
function place() {
  const trigger = triggerRef.value
  if (!trigger) return

  const rect = trigger.getBoundingClientRect()
  let nextLeft = rect.right + MARGIN
  if (nextLeft + POPOVER_WIDTH > window.innerWidth - MARGIN) {
    nextLeft = rect.left - POPOVER_WIDTH - MARGIN
  }
  left.value = Math.max(MARGIN, nextLeft)
  top.value = rect.top
}

// clampVertical pulls the panel up when it overflows the bottom edge. It runs
// after the popover is shown, once the real height is measurable.
function clampVertical() {
  const pop = popRef.value
  if (!pop) return
  const popRect = pop.getBoundingClientRect()
  if (popRect.bottom > window.innerHeight - MARGIN) {
    top.value = Math.max(MARGIN, top.value - (popRect.bottom - (window.innerHeight - MARGIN)))
  }
}

function show() {
  cancelClose()
  if (visible.value) return
  const pop = popRef.value
  if (!pop) return
  place()
  if (typeof pop.showPopover === 'function') {
    pop.showPopover()
  } else {
    // Fallback for browsers without the Popover API: force display and hope
    // no top-layer element is covering the page.
    pop.style.display = 'block'
    pop.style.zIndex = '9999'
  }
  visible.value = true
  nextTick(clampVertical)
  window.addEventListener('scroll', reposition, true)
  window.addEventListener('resize', reposition)
}

function hide() {
  cancelClose()
  if (!visible.value) return
  const pop = popRef.value
  if (pop) {
    if (typeof pop.hidePopover === 'function') {
      pop.hidePopover()
    } else {
      pop.style.display = ''
    }
  }
  visible.value = false
  window.removeEventListener('scroll', reposition, true)
  window.removeEventListener('resize', reposition)
}

function toggle() {
  if (visible.value) {
    hide()
    return
  }
  show()
}

function reposition() {
  place()
  nextTick(clampVertical)
}

// scheduleClose gives the pointer a short grace period to travel across the gap
// between the icon and the panel without the panel closing underneath it.
function scheduleClose() {
  cancelClose()
  closeTimer = setTimeout(hide, CLOSE_DELAY_MS)
}

function cancelClose() {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
}

onBeforeUnmount(() => {
  hide()
})
</script>

<style scoped>
/* Override the UA popover stylesheet (inset:0 + margin:auto centers it) so the
   inline left/top coordinates take effect. */
.nodehelp-pop {
  position: fixed;
  inset: auto;
  margin: 0;
}
</style>
