<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <div class="min-h-screen flex items-center justify-center bg-piedra-950">
    <div class="w-full max-w-sm mx-auto px-6">
      <div class="flex flex-col items-center mb-8">
        <img src="/assets/logo.svg" alt="Magec" class="w-12 h-12 mb-4" />
        <h1 class="text-lg font-bold text-arena-50 tracking-tight">Magec</h1>
        <p class="text-xs text-arena-500 mt-1">Admin Console</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="block text-xs text-arena-400 mb-1.5">Password</label>
          <input
            ref="inputRef"
            v-model="password"
            type="password"
            autocomplete="current-password"
            class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2.5 text-sm text-arena-100 focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none placeholder:text-arena-600"
            placeholder="Enter admin password"
            required
          />
        </div>

        <p v-if="error" class="text-xs text-lava-400">{{ error }}</p>

        <button
          type="submit"
          :disabled="loading"
          class="w-full py-2.5 bg-sol-500 hover:bg-sol-600 disabled:opacity-50 text-piedra-950 text-sm font-medium rounded-lg transition-colors"
        >
          {{ loading ? 'Checking...' : 'Sign In' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { login as authLogin } from '../lib/auth.js'

const emit = defineEmits(['authenticated'])

const password = ref('')
const error = ref('')
const loading = ref(false)
const inputRef = ref(null)

onMounted(() => {
  inputRef.value?.focus()
})

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    const ok = await authLogin(password.value)
    if (ok) {
      emit('authenticated')
    } else {
      error.value = 'Invalid password'
      password.value = ''
      inputRef.value?.focus()
    }
  } catch {
    error.value = 'Connection error'
  } finally {
    loading.value = false
  }
}
</script>
