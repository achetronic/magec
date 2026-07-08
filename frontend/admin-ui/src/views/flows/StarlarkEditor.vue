<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- Starlark editor without editor dependencies: a plain textarea whose text
     is transparent (the caret stays visible) over a pre that renders the same
     content tokenized by a small regex highlighter. Both layers share font
     and padding so the glyphs align exactly; the textarea scroll is mirrored
     onto the pre. -->
<template>
  <div class="stark flex-1 min-h-[4rem] relative rounded-lg bg-piedra-800 border border-piedra-700/50 focus-within:border-emerald-500/50 overflow-hidden">
    <pre ref="mirrorRef" aria-hidden="true" class="stark-mirror" v-html="highlighted"></pre>
    <textarea
      :value="modelValue"
      @input="$emit('update:modelValue', $event.target.value)"
      @scroll="syncScroll"
      @pointerdown.stop
      spellcheck="false"
      :placeholder="placeholder"
      class="stark-input"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  modelValue:  { type: String, default: '' },
  placeholder: { type: String, default: '' },
})

defineEmits(['update:modelValue'])

const mirrorRef = ref(null)

// One alternation per token class, tried in order: comments win over
// everything, strings over words, numbers and keywords over builtins.
const TOKEN_RE = /(#[^\n]*)|("(?:[^"\\\n]|\\.)*"?|'(?:[^'\\\n]|\\.)*'?)|(\b\d+(?:\.\d+)?\b)|(\b(?:and|break|continue|def|elif|else|for|if|in|lambda|load|not|or|pass|return|True|False|None)\b)|(\b(?:input|state|output|len|str|int|float|bool|list|dict|tuple|range|enumerate|sorted|reversed|zip|min|max|sum|abs|any|all|print|type|hasattr|getattr|dir|fail)\b)/g

const TOKEN_CLASSES = ['tok-comment', 'tok-str', 'tok-num', 'tok-kw', 'tok-builtin']

function escapeHtml(s) {
  return s.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;')
}

// highlight tokenizes the source in a single pass, escaping every emitted
// piece. The trailing newline keeps the mirror's last line in sync with the
// textarea when the source ends with a line break.
function highlight(src) {
  let out = ''
  let last = 0
  for (const m of src.matchAll(TOKEN_RE)) {
    out += escapeHtml(src.slice(last, m.index))
    const cls = TOKEN_CLASSES[m.slice(1).findIndex(g => g !== undefined)]
    out += `<span class="${cls}">${escapeHtml(m[0])}</span>`
    last = m.index + m[0].length
  }
  return out + escapeHtml(src.slice(last)) + '\n'
}

const highlighted = computed(() => highlight(props.modelValue))

function syncScroll(e) {
  const mirror = mirrorRef.value
  if (!mirror) return
  mirror.scrollTop = e.target.scrollTop
  mirror.scrollLeft = e.target.scrollLeft
}
</script>

<style scoped>
/* Identical metrics on both layers so the transparent text and the coloured
   mirror overlap glyph by glyph. */
.stark-mirror,
.stark-input {
  font-family: ui-monospace, monospace;
  font-size: 10px;
  line-height: 1.5;
  padding: 4px 8px;
  white-space: pre-wrap;
  word-break: break-word;
}
.stark-mirror {
  position: absolute;
  inset: 0;
  margin: 0;
  overflow: hidden;
  color: var(--color-arena-200);
  pointer-events: none;
}
.stark-input {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  background: transparent;
  color: transparent;
  caret-color: var(--color-arena-200);
  resize: none;
  outline: none;
  border: 0;
}
.stark-input::placeholder { color: var(--color-arena-600); }

/* Token colours. :deep() because the spans come from v-html and carry no
   scoped data attribute. */
.stark-mirror :deep(.tok-comment) { color: var(--color-arena-600); font-style: italic; }
.stark-mirror :deep(.tok-str)     { color: var(--color-sol-300); }
.stark-mirror :deep(.tok-num)     { color: var(--color-lava-300); }
.stark-mirror :deep(.tok-kw)      { color: var(--color-atlantico-300); }
.stark-mirror :deep(.tok-builtin) { color: var(--color-emerald-300, #6ee7b7); }
</style>
