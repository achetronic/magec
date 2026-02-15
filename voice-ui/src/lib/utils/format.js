export function escapeHtml(text) {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

export function renderMarkdown(text) {
  if (!text) return ''

  let html = escapeHtml(text)

  html = html.replace(/```(\w*)\n?([\s\S]*?)```/g, (_, lang, code) =>
    `<pre class="bg-piedra-900 rounded-lg p-3 my-2 overflow-x-auto"><code class="text-xs text-arena-300">${code.trim()}</code></pre>`
  )

  html = html.replace(/`([^`]+)`/g, '<code class="bg-piedra-800 px-1.5 py-0.5 rounded text-sol-300 text-xs">$1</code>')

  html = html.replace(/^### (.+)$/gm, '<h3 class="text-base font-semibold text-arena-100 mt-3 mb-1">$1</h3>')
  html = html.replace(/^## (.+)$/gm, '<h2 class="text-lg font-semibold text-arena-100 mt-3 mb-1">$1</h2>')
  html = html.replace(/^# (.+)$/gm, '<h1 class="text-xl font-bold text-arena-100 mt-3 mb-2">$1</h1>')

  html = html.replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold text-arena-50">$1</strong>')
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')

  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener" class="text-sol-400 hover:text-sol-300 underline">$1</a>')

  html = html.replace(/^- (.+)$/gm, '<li class="ml-4 list-disc">$1</li>')
  html = html.replace(/(<li[^>]*>.*<\/li>\n?)+/g, '<ul class="my-2 space-y-1">$&</ul>')

  html = html.replace(/^\d+\. (.+)$/gm, '<li class="ml-4 list-decimal">$1</li>')
  html = html.replace(/(<li class="ml-4 list-decimal">.*<\/li>\n?)+/g, '<ol class="my-2 space-y-1">$&</ol>')

  html = html.replace(/\n\n+/g, '</p><p class="mt-2">')
  html = html.replace(/\n/g, '<br>')

  if (!html.match(/^<(h[1-6]|ul|ol|pre|p)/)) {
    html = `<p>${html}</p>`
  }

  return html
}

export function formatRelativeDate(timestamp, t) {
  const date = timestamp instanceof Date ? timestamp : new Date(timestamp)
  const diff = Date.now() - date.getTime()

  const MINUTE = 60000
  const HOUR = 3600000
  const DAY = 86400000
  const WEEK = 604800000

  if (diff < MINUTE) return t('time.justNow')
  if (diff < HOUR) return t('time.minutesAgo', { n: Math.floor(diff / MINUTE) })
  if (diff < DAY) return t('time.hoursAgo', { n: Math.floor(diff / HOUR) })
  if (diff < WEEK) return t('time.daysAgo', { n: Math.floor(diff / DAY) })

  return date.toLocaleDateString()
}
