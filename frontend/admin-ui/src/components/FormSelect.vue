<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <div ref="wrapperRef" class="relative">
    <button
      ref="buttonRef"
      type="button"
      @click="toggle"
      class="w-full flex items-center justify-between bg-piedra-800 border rounded-lg px-3 py-2 text-sm text-left outline-none transition-colors cursor-pointer"
      :class="open
        ? 'border-sol-500 ring-1 ring-sol-500'
        : 'border-piedra-700 hover:border-piedra-600'"
    >
      <span :class="selectedLabel ? 'text-arena-300' : 'text-arena-600'">
        {{ selectedLabel || placeholder }}
      </span>
      <svg
        class="h-3.5 w-3.5 text-arena-500 shrink-0 ml-2 transition-transform duration-200"
        :class="open ? 'rotate-180' : ''"
        viewBox="0 0 20 20" fill="currentColor"
      >
        <path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
      </svg>
    </button>

    <Transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <ul
        v-if="open"
        ref="listRef"
        :style="floatingStyle"
        class="fixed z-[9999] max-h-48 overflow-auto rounded-lg border border-piedra-700 bg-piedra-800 py-1 shadow-lg shadow-black/30 text-sm"
      >
        <li
          v-for="opt in normalizedOptions"
          :key="opt.value"
          @click="select(opt.value)"
          class="flex items-center gap-2 px-3 py-1.5 cursor-pointer transition-colors"
          :class="opt.value === modelValue
            ? 'bg-sol-500/10 text-sol-300'
            : 'text-arena-400 hover:bg-piedra-700/60 hover:text-arena-300'"
        >
          <span class="truncate">{{ opt.label }}</span>
        </li>
      </ul>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount, useSlots } from 'vue'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, default: () => [] },
  placeholder: { type: String, default: 'Select...' },
})

const emit = defineEmits(['update:modelValue'])
const slots = useSlots()
const open = ref(false)
const wrapperRef = ref(null)
const buttonRef = ref(null)
const listRef = ref(null)
const floatingStyle = ref({})

function updatePosition() {
  if (!buttonRef.value) return
  const rect = buttonRef.value.getBoundingClientRect()
  floatingStyle.value = {
    top: `${rect.bottom + 4}px`,
    left: `${rect.left}px`,
    width: `${rect.width}px`,
  }
}

function toggle() {
  open.value = !open.value
}

watch(open, async (val) => {
  if (val) {
    updatePosition()
    await nextTick()
    updatePosition()
  }
})

const normalizedOptions = computed(() => {
  if (props.options.length) {
    return props.options.map(o =>
      typeof o === 'object' ? o : { value: o, label: String(o) }
    )
  }
  const slotContent = slots.default?.()
  if (!slotContent) return []
  const opts = []
  for (const vnode of slotContent) {
    if (vnode.type === 'option' || vnode.type?.name === 'option') {
      const val = vnode.props?.value ?? ''
      const label = typeof vnode.children === 'string'
        ? vnode.children
        : vnode.children?.default?.()?.map(c => c.children || '').join('') || String(val)
      opts.push({ value: val, label })
    }
    if (Array.isArray(vnode.children)) {
      for (const child of vnode.children) {
        if (child.type === 'option' || child.type?.name === 'option') {
          const val = child.props?.value ?? ''
          const label = typeof child.children === 'string'
            ? child.children
            : child.children?.default?.()?.map(c => c.children || '').join('') || String(val)
          opts.push({ value: val, label })
        }
      }
    }
  }
  return opts
})

const selectedLabel = computed(() => {
  const opt = normalizedOptions.value.find(o => String(o.value) === String(props.modelValue))
  return opt?.label || ''
})

function select(val) {
  emit('update:modelValue', val)
  open.value = false
}

function onClickOutside(e) {
  if (wrapperRef.value?.contains(e.target)) return
  if (listRef.value?.contains(e.target)) return
  open.value = false
}

function onScroll() {
  if (open.value) updatePosition()
}

onMounted(() => {
  document.addEventListener('click', onClickOutside)
  window.addEventListener('scroll', onScroll, true)
  window.addEventListener('resize', onScroll)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onClickOutside)
  window.removeEventListener('scroll', onScroll, true)
  window.removeEventListener('resize', onScroll)
})
</script>
