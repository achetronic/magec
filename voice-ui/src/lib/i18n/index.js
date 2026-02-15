import { ref, computed } from 'vue'
import es from './es.js'
import en from './en.js'

const STORAGE_KEY = 'magec_language'
const DEFAULT_LANGUAGE = 'en'

const languages = { es, en }
const currentLanguage = ref(DEFAULT_LANGUAGE)

function getNestedValue(obj, path) {
  return path.split('.').reduce((acc, key) => acc?.[key], obj)
}

function interpolate(text, params) {
  if (!params || typeof text !== 'string') return text
  return text.replace(/\{(\w+)\}/g, (_, key) => params[key] ?? `{${key}}`)
}

export function t(key, params = null) {
  const value = getNestedValue(languages[currentLanguage.value], key)
  if (value === undefined) {
    console.warn(`[i18n] Missing translation: ${key}`)
    return key
  }
  return interpolate(value, params)
}

export function setLanguage(lang) {
  if (!languages[lang]) {
    console.warn(`[i18n] Unknown language: ${lang}`)
    return false
  }
  currentLanguage.value = lang
  localStorage.setItem(STORAGE_KEY, lang)
  return true
}

export function getLanguage() {
  return currentLanguage.value
}

export function getAvailableLanguages() {
  return Object.keys(languages)
}

export function initLanguage() {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved && languages[saved]) {
    currentLanguage.value = saved
  }
  return currentLanguage.value
}

export const language = computed(() => currentLanguage.value)
