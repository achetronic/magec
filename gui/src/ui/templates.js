// ==================== Message Templates ====================

import { t } from '../i18n/index.js';

const MESSAGE_STYLES = {
    user: {
        avatar: 'bg-piedra-800',
        avatarIcon: 'text-arena-400',
        iconPath: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z',
        bubble: 'bg-piedra-800'
    },
    ai: {
        avatar: 'bg-sol-500/20',
        avatarIcon: 'text-sol-400',
        iconPath: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6m-6 4h6',
        bubble: 'bg-sol-500/20'
    }
};

function createMessageHTML(type, text, formatFn = escapeHtml) {
    const s = MESSAGE_STYLES[type];
    return `
        <div class="flex gap-3 items-start">
            <div class="w-8 h-8 rounded-full ${s.avatar} flex items-center justify-center flex-shrink-0">
                <svg class="w-4 h-4 ${s.avatarIcon}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="${s.iconPath}"/>
                </svg>
            </div>
            <div class="flex-1 ${s.bubble} rounded-2xl rounded-tl-md px-4 py-3">
                <div class="text-sm text-arena-100">${formatFn(text)}</div>
            </div>
        </div>
    `;
}

export function createUserMessageHTML(text) {
    return createMessageHTML('user', text, escapeHtml);
}

export function createAIMessageHTML(text) {
    return createMessageHTML('ai', text, renderMarkdown);
}

// ==================== Session Templates ====================

export function createSessionItemHTML(session, isActive) {
    const preview = session.preview || t('sessions.emptyPreview');
    const date = session.createdAt ? formatRelativeDate(session.createdAt) : '';
    const activeClass = isActive 
        ? 'bg-sol-500/20' 
        : 'hover:bg-piedra-800';
    
    return `
        <div class="group relative">
            <button data-session-id="${session.id}"
                class="w-full text-left px-3 py-2.5 rounded-lg transition-colors ${activeClass}">
                <p class="text-sm text-arena-200 truncate pr-6">${escapeHtml(preview)}</p>
                <p class="text-xs text-arena-500 mt-0.5">${date}</p>
            </button>
            <button data-delete-session="${session.id}"
                class="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded-md opacity-0 group-hover:opacity-100 hover:bg-piedra-700 transition-all"
                title="${t('sessions.delete')}">
                <svg class="w-4 h-4 text-arena-500 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                </svg>
            </button>
        </div>
    `;
}

// ==================== Utilities ====================

export function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Renders Markdown to HTML. Supports:
 * - **bold**, *italic*, `code`
 * - Code blocks (```)
 * - Headers (# ## ###)
 * - Lists (- and 1.)
 * - Links [text](url)
 * - Paragraphs
 */
export function renderMarkdown(text) {
    if (!text) return '';
    
    let html = escapeHtml(text);
    
    // Code blocks (must be first to protect content)
    html = html.replace(/```(\w*)\n?([\s\S]*?)```/g, (_, lang, code) => 
        `<pre class="bg-piedra-900 rounded-lg p-3 my-2 overflow-x-auto"><code class="text-xs text-arena-300">${code.trim()}</code></pre>`
    );
    
    // Inline code
    html = html.replace(/`([^`]+)`/g, '<code class="bg-piedra-800 px-1.5 py-0.5 rounded text-sol-300 text-xs">$1</code>');
    
    // Headers
    html = html.replace(/^### (.+)$/gm, '<h3 class="text-base font-semibold text-arena-100 mt-3 mb-1">$1</h3>');
    html = html.replace(/^## (.+)$/gm, '<h2 class="text-lg font-semibold text-arena-100 mt-3 mb-1">$1</h2>');
    html = html.replace(/^# (.+)$/gm, '<h1 class="text-xl font-bold text-arena-100 mt-3 mb-2">$1</h1>');
    
    // Bold and italic
    html = html.replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold text-arena-50">$1</strong>');
    html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');
    
    // Links
    html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener" class="text-sol-400 hover:text-sol-300 underline">$1</a>');
    
    // Unordered lists
    html = html.replace(/^- (.+)$/gm, '<li class="ml-4 list-disc">$1</li>');
    html = html.replace(/(<li[^>]*>.*<\/li>\n?)+/g, '<ul class="my-2 space-y-1">$&</ul>');
    
    // Ordered lists
    html = html.replace(/^\d+\. (.+)$/gm, '<li class="ml-4 list-decimal">$1</li>');
    html = html.replace(/(<li class="ml-4 list-decimal">.*<\/li>\n?)+/g, '<ol class="my-2 space-y-1">$&</ol>');
    
    // Paragraphs (double newlines)
    html = html.replace(/\n\n+/g, '</p><p class="mt-2">');
    
    // Single newlines to <br> (except inside lists/pre)
    html = html.replace(/\n/g, '<br>');
    
    // Wrap in paragraph if not starting with block element
    if (!html.match(/^<(h[1-6]|ul|ol|pre|p)/)) {
        html = `<p>${html}</p>`;
    }
    
    return html;
}

export function formatRelativeDate(timestamp) {
    const date = timestamp instanceof Date ? timestamp : new Date(timestamp);
    const diff = Date.now() - date.getTime();
    
    const MINUTE = 60000;
    const HOUR = 3600000;
    const DAY = 86400000;
    const WEEK = 604800000;
    
    if (diff < MINUTE) return t('time.justNow');
    if (diff < HOUR) return t('time.minutesAgo', { n: Math.floor(diff / MINUTE) });
    if (diff < DAY) return t('time.hoursAgo', { n: Math.floor(diff / HOUR) });
    if (diff < WEEK) return t('time.daysAgo', { n: Math.floor(diff / DAY) });
    
    return date.toLocaleDateString();
}
