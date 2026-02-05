/*
 * Copyright 2025 Alby Hernández
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import es from './es.js';
import en from './en.js';

const STORAGE_KEY = 'magec_language';
const DEFAULT_LANGUAGE = 'es';

const languages = { es, en };

let currentLanguage = DEFAULT_LANGUAGE;
let currentTranslations = languages[DEFAULT_LANGUAGE];
let changeListeners = [];

// Obtiene un valor anidado de un objeto usando notación de puntos: 'settings.wakeWord.title'
function getNestedValue(obj, path) {
    return path.split('.').reduce((acc, key) => acc?.[key], obj);
}

// Reemplaza placeholders tipo {name} con valores del objeto params
function interpolate(text, params) {
    if (!params || typeof text !== 'string') return text;
    return text.replace(/\{(\w+)\}/g, (_, key) => params[key] ?? `{${key}}`);
}

// Función principal de traducción
export function t(key, params = null) {
    const value = getNestedValue(currentTranslations, key);
    if (value === undefined) {
        console.warn(`[i18n] Missing translation: ${key}`);
        return key;
    }
    return interpolate(value, params);
}

// Cambia el idioma actual
export function setLanguage(lang) {
    if (!languages[lang]) {
        console.warn(`[i18n] Unknown language: ${lang}`);
        return false;
    }
    
    currentLanguage = lang;
    currentTranslations = languages[lang];
    localStorage.setItem(STORAGE_KEY, lang);
    
    // Actualiza el DOM
    applyTranslations();
    
    // Notifica a los listeners
    changeListeners.forEach(fn => fn(lang));
    
    return true;
}

// Obtiene el idioma actual
export function getLanguage() {
    return currentLanguage;
}

// Obtiene la lista de idiomas disponibles
export function getAvailableLanguages() {
    return Object.keys(languages);
}

// Registra un listener para cambios de idioma
export function onLanguageChange(callback) {
    changeListeners.push(callback);
    return () => {
        changeListeners = changeListeners.filter(fn => fn !== callback);
    };
}

// Inicializa el idioma desde localStorage
export function initLanguage() {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved && languages[saved]) {
        currentLanguage = saved;
        currentTranslations = languages[saved];
    }
    applyTranslations();
    return currentLanguage;
}

// Aplica las traducciones a todos los elementos con data-i18n en el DOM
export function applyTranslations() {
    // Texto contenido: data-i18n="key"
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        const translation = t(key);
        if (translation !== key) {
            el.textContent = translation;
        }
    });
    
    // Atributo title: data-i18n-title="key"
    document.querySelectorAll('[data-i18n-title]').forEach(el => {
        const key = el.getAttribute('data-i18n-title');
        const translation = t(key);
        if (translation !== key) {
            el.setAttribute('title', translation);
        }
    });
    
    // Atributo placeholder: data-i18n-placeholder="key"
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        const translation = t(key);
        if (translation !== key) {
            el.setAttribute('placeholder', translation);
        }
    });
}
