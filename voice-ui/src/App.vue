<template>
  <PairingScreen v-if="store.showPairing" />
  <div v-else class="h-full flex">
    <Transition name="fade">
      <div
        v-if="store.sidebarOpen && isMobile"
        class="fixed inset-0 bg-piedra-950/80 backdrop-blur-sm z-20"
        @click="store.sidebarOpen = false"
      />
    </Transition>

    <AppSidebar />

    <div class="flex-1 flex flex-col min-w-0 h-full">
      <AppHeader />

      <main class="flex-1 flex flex-col overflow-hidden">
        <div class="flex-1 flex flex-col max-w-3xl mx-auto w-full px-4 py-6 gap-4 overflow-hidden">
          <div class="flex-1 bg-piedra-900/80 border border-piedra-700/50 rounded-2xl flex flex-col min-h-0 overflow-hidden">
            <Transition name="panel" mode="out-in">
              <PanelAssistant v-if="store.activePanel === 'assistant'" key="assistant" />
              <PanelHistory v-else-if="store.activePanel === 'history'" key="history" />
              <PanelNotifications v-else-if="store.activePanel === 'notifications'" key="notifications" />
              <PanelSettings v-else-if="store.activePanel === 'settings'" key="settings" />
            </Transition>
          </div>
        </div>
      </main>

      <FooterClock />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useAppStore } from './lib/stores/app.js'
import PairingScreen from './components/PairingScreen.vue'
import AppHeader from './components/AppHeader.vue'
import AppSidebar from './components/AppSidebar.vue'
import FooterClock from './components/FooterClock.vue'
import PanelAssistant from './components/PanelAssistant.vue'
import PanelHistory from './components/PanelHistory.vue'
import PanelNotifications from './components/PanelNotifications.vue'
import PanelSettings from './components/PanelSettings.vue'

const store = useAppStore()
const isMobile = ref(window.innerWidth < 1024)

function onResize() {
  isMobile.value = window.innerWidth < 1024
}

onMounted(() => {
  store.init()
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
})
</script>
