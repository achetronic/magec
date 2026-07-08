// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { ref } from 'vue'

const adminPassword = ref('')
const authenticated = ref(false)
const authRequired = ref(null)

export function setPassword(password) {
  adminPassword.value = password
}

export function getPassword() {
  return adminPassword.value
}

export function isAuthenticated() {
  return authenticated.value
}

export function isAuthRequired() {
  return authRequired.value
}

export function logout() {
  adminPassword.value = ''
  authenticated.value = false
}

export async function checkAuth() {
  try {
    const res = await fetch('/api/v1/admin/auth/check')
    if (res.status === 401) {
      authRequired.value = true
      return false
    }
    authRequired.value = false
    authenticated.value = true
    return true
  } catch {
    authRequired.value = false
    authenticated.value = true
    return true
  }
}

export async function login(password) {
  const res = await fetch('/api/v1/admin/auth/check', {
    headers: { 'Authorization': `Bearer ${password}` },
  })
  if (res.ok) {
    adminPassword.value = password
    authenticated.value = true
    return true
  }
  return false
}

export function getAuthHeaders() {
  if (adminPassword.value) {
    return { 'Authorization': `Bearer ${adminPassword.value}` }
  }
  return {}
}
