<template>
  <footer class="border-t border-piedra-700/50 py-3 flex-shrink-0">
    <p class="text-center text-xs text-arena-500">{{ clockText }}</p>
  </footer>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const clockText = ref('')
let timer = null

function update() {
  const now = new Date()
  const time = now.toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit' })
  const date = now.toLocaleDateString('es-ES', { weekday: 'long', day: 'numeric', month: 'long' })
  clockText.value = `${time} · ${date.charAt(0).toUpperCase() + date.slice(1)}`
}

onMounted(() => {
  update()
  timer = setInterval(update, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
