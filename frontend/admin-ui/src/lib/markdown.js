// Magec-flavoured markdown renderer.
//
// We don't use Tailwind's `prose` utility (the typography plugin isn't
// in this project) nor any heavy syntax-highlight library — both would
// fight the carefully tuned piedra/arena/sol palette. Instead we
// configure `marked` with a custom renderer that emits classes scoped
// under `.magec-markdown` (defined in src/style.css), and a tiny
// regex-based highlighter for the handful of languages the operator
// actually pastes into SKILL.md files (YAML, JSON, shell, Go, JS/TS,
// Python, plain text).
//
// The output is opinionated:
//
//   - Code spans/blocks use the project's `font-mono` weight + a
//     subtle border so they don't visually crash with the surrounding
//     prose.
//   - Inline links pick up the atlántico accent — same hue as
//     interactive elements elsewhere in the admin UI.
//   - Tables, blockquotes and lists adopt the same border/spacing
//     vocabulary as Cards, so rendered Markdown looks like it belongs
//     in the same dialog as the rest of the skill metadata.
//   - Heading anchors are stripped to keep the read-only viewer free
//     of dangling links to nowhere.
//
// All embedded HTML in the SKILL.md body is escaped to text via the
// `html` and `tag` renderer overrides below. Marked v17 has no
// built-in sanitiser; without these overrides, raw `<script>` tags in
// a malicious SKILL.md would execute when v-html'd into the viewer.

import { marked } from 'marked'

const TOKEN_CLASS = {
  comment: 'magec-tk-comment',
  keyword: 'magec-tk-keyword',
  string: 'magec-tk-string',
  number: 'magec-tk-number',
  func: 'magec-tk-func',
  builtin: 'magec-tk-builtin',
  property: 'magec-tk-property',
  symbol: 'magec-tk-symbol',
}

// Per-language token tables. Each table is an array of
// `{ pattern, type }` rules tried in order. The first match at any
// position wins, the matched span is wrapped in a span with the
// corresponding class, and the cursor advances past the match. Order
// matters: longest/most-specific patterns must come first so e.g. the
// number rule does not eat parts of an identifier.
const RULES = {
  yaml: [
    { pattern: /^#[^\n]*/, type: 'comment' },
    { pattern: /^"(?:\\.|[^"\\])*"|^'(?:[^'\\]|\\.)*'/, type: 'string' },
    { pattern: /^(true|false|null|~)\b/, type: 'keyword' },
    { pattern: /^-?\d+(\.\d+)?/, type: 'number' },
    { pattern: /^[A-Za-z_][\w-]*(?=\s*:)/, type: 'property' },
    { pattern: /^[-:|>]/, type: 'symbol' },
  ],
  json: [
    { pattern: /^"(?:\\.|[^"\\])*"(?=\s*:)/, type: 'property' },
    { pattern: /^"(?:\\.|[^"\\])*"/, type: 'string' },
    { pattern: /^(true|false|null)\b/, type: 'keyword' },
    { pattern: /^-?\d+(\.\d+)?([eE][+-]?\d+)?/, type: 'number' },
    { pattern: /^[{}[\],:]/, type: 'symbol' },
  ],
  bash: [
    { pattern: /^#[^\n]*/, type: 'comment' },
    { pattern: /^"(?:\\.|[^"\\])*"|^'(?:[^'\\]|\\.)*'/, type: 'string' },
    { pattern: /^\$[A-Za-z_][\w]*/, type: 'symbol' },
    { pattern: /^\$\{[^}]*\}/, type: 'symbol' },
    { pattern: /^(if|then|else|elif|fi|for|while|do|done|case|esac|in|function|return|export|local|read|echo|printf|cd|pwd|mkdir|rm|cp|mv|cat|grep|sed|awk|find|set|unset)\b/, type: 'keyword' },
    { pattern: /^-{1,2}[A-Za-z][\w-]*/, type: 'builtin' },
    { pattern: /^\b\d+\b/, type: 'number' },
  ],
  shell: 'bash',
  sh: 'bash',
  go: [
    { pattern: /^\/\/[^\n]*/, type: 'comment' },
    { pattern: /^\/\*[\s\S]*?\*\//, type: 'comment' },
    { pattern: /^"(?:\\.|[^"\\])*"|^`[^`]*`|^'(?:\\.|[^'\\])'/, type: 'string' },
    { pattern: /^\b(package|import|func|var|const|type|struct|interface|map|chan|range|for|if|else|switch|case|default|break|continue|return|defer|go|select|fallthrough|nil|true|false|iota)\b/, type: 'keyword' },
    { pattern: /^\b(string|int|int8|int16|int32|int64|uint|uint8|uint16|uint32|uint64|byte|rune|float32|float64|bool|error|any)\b/, type: 'builtin' },
    { pattern: /^\b[A-Za-z_]\w*(?=\s*\()/, type: 'func' },
    { pattern: /^\b\d+(\.\d+)?([eE][+-]?\d+)?\b/, type: 'number' },
  ],
  js: [
    { pattern: /^\/\/[^\n]*/, type: 'comment' },
    { pattern: /^\/\*[\s\S]*?\*\//, type: 'comment' },
    { pattern: /^"(?:\\.|[^"\\])*"|^'(?:\\.|[^'\\])*'|^`(?:\\.|[^`\\])*`/, type: 'string' },
    { pattern: /^\b(const|let|var|function|class|extends|new|return|if|else|for|while|do|switch|case|break|continue|try|catch|finally|throw|async|await|import|export|from|as|of|in|typeof|instanceof|null|undefined|true|false|this|super)\b/, type: 'keyword' },
    { pattern: /^\b[A-Za-z_$][\w$]*(?=\s*\()/, type: 'func' },
    { pattern: /^\b\d+(\.\d+)?\b/, type: 'number' },
  ],
  ts: 'js',
  typescript: 'js',
  javascript: 'js',
  python: [
    { pattern: /^#[^\n]*/, type: 'comment' },
    { pattern: /^"""[\s\S]*?"""|^'''[\s\S]*?'''/, type: 'string' },
    { pattern: /^"(?:\\.|[^"\\])*"|^'(?:\\.|[^'\\])*'/, type: 'string' },
    { pattern: /^\b(def|class|return|if|elif|else|for|while|in|and|or|not|is|None|True|False|import|from|as|with|try|except|finally|raise|lambda|yield|pass|break|continue|global|nonlocal|async|await)\b/, type: 'keyword' },
    { pattern: /^\b[A-Za-z_]\w*(?=\s*\()/, type: 'func' },
    { pattern: /^\b\d+(\.\d+)?\b/, type: 'number' },
  ],
  py: 'python',
  markdown: [
    { pattern: /^#{1,6}\s+[^\n]+/, type: 'keyword' },
    { pattern: /^\*\*[^*\n]+\*\*|^__[^_\n]+__/, type: 'string' },
    { pattern: /^`[^`\n]+`/, type: 'symbol' },
    { pattern: /^!?\[[^\]]+\]\([^)]+\)/, type: 'func' },
  ],
  md: 'markdown',
}

function rulesFor(lang) {
  if (!lang) return null
  let r = RULES[lang.toLowerCase()]
  if (typeof r === 'string') r = RULES[r]
  return r || null
}

const HTML_ESCAPES = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;',
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) => HTML_ESCAPES[c])
}

// highlight tokenises `code` using the rules for `lang` and returns a
// HTML string with token spans. Unknown languages return the escaped
// code untouched so the caller can still display it cleanly.
function highlight(code, lang) {
  const rules = rulesFor(lang)
  if (!rules) return escapeHtml(code)

  let out = ''
  let i = 0
  while (i < code.length) {
    const remaining = code.slice(i)
    let matched = false
    for (const { pattern, type } of rules) {
      const m = remaining.match(pattern)
      if (!m || m.index !== 0) continue
      const cls = TOKEN_CLASS[type]
      out += `<span class="${cls}">${escapeHtml(m[0])}</span>`
      i += m[0].length
      matched = true
      break
    }
    if (!matched) {
      out += escapeHtml(code[i])
      i += 1
    }
  }
  return out
}

// Configure a fresh marked instance so we never collide with another
// caller's options (every renderer is module-local here, but
// future-proof).
const renderer = new marked.Renderer()

// HTML-token overrides: marked v17 no longer ships a `sanitize` option,
// and by default raw `<script>`/`<img onerror=...>` tags inside the
// markdown body pass through untouched. Skill bodies come from operator
// uploads (and may include third-party SKILL.md packages downloaded
// from elsewhere), so we treat any embedded HTML as text. The `html`
// renderer fires for block-level HTML; the `tag` rule fires for inline
// tags like `<span>` mid-paragraph. Both render escaped so the user
// sees the raw markup, not its execution.
renderer.html = function ({ text }) {
  return escapeHtml(text)
}
renderer.tag = function ({ text }) {
  return escapeHtml(text)
}

renderer.heading = function ({ tokens, depth }) {
  const text = this.parser.parseInline(tokens)
  return `<h${depth} class="magec-md-h magec-md-h${depth}">${text}</h${depth}>\n`
}

renderer.paragraph = function ({ tokens }) {
  return `<p class="magec-md-p">${this.parser.parseInline(tokens)}</p>\n`
}

renderer.list = function ({ ordered, items }) {
  const tag = ordered ? 'ol' : 'ul'
  const cls = ordered ? 'magec-md-ol' : 'magec-md-ul'
  let body = ''
  for (const item of items) body += this.listitem(item)
  return `<${tag} class="${cls}">${body}</${tag}>\n`
}

renderer.listitem = function (item) {
  const body = this.parser.parse(item.tokens)
  return `<li class="magec-md-li">${body}</li>\n`
}

renderer.blockquote = function ({ tokens }) {
  return `<blockquote class="magec-md-quote">${this.parser.parse(tokens)}</blockquote>\n`
}

renderer.code = function ({ text, lang }) {
  const langAttr = lang ? ` data-lang="${escapeHtml(lang)}"` : ''
  const langLabel = lang
    ? `<span class="magec-md-code-lang">${escapeHtml(lang)}</span>`
    : ''
  const body = highlight(text, lang)
  return `<pre class="magec-md-pre"${langAttr}>${langLabel}<code class="magec-md-code">${body}</code></pre>\n`
}

renderer.codespan = function ({ text }) {
  return `<code class="magec-md-icode">${escapeHtml(text)}</code>`
}

renderer.link = function ({ href, title, tokens }) {
  const inner = this.parser.parseInline(tokens)
  const titleAttr = title ? ` title="${escapeHtml(title)}"` : ''
  const safeHref = escapeHtml(href || '')
  return `<a href="${safeHref}" class="magec-md-link" target="_blank" rel="noopener"${titleAttr}>${inner}</a>`
}

renderer.image = function ({ href, title, text }) {
  const safeHref = escapeHtml(href || '')
  const titleAttr = title ? ` title="${escapeHtml(title)}"` : ''
  return `<img src="${safeHref}" alt="${escapeHtml(text || '')}" class="magec-md-img"${titleAttr} />`
}

renderer.table = function ({ header, rows }) {
  let head = '<tr>'
  for (const cell of header) {
    const align = cell.align ? ` style="text-align:${cell.align}"` : ''
    head += `<th class="magec-md-th"${align}>${this.parser.parseInline(cell.tokens)}</th>`
  }
  head += '</tr>'

  let body = ''
  for (const row of rows) {
    body += '<tr class="magec-md-tr">'
    for (const cell of row) {
      const align = cell.align ? ` style="text-align:${cell.align}"` : ''
      body += `<td class="magec-md-td"${align}>${this.parser.parseInline(cell.tokens)}</td>`
    }
    body += '</tr>'
  }

  return `<div class="magec-md-table-wrap"><table class="magec-md-table"><thead>${head}</thead><tbody>${body}</tbody></table></div>\n`
}

renderer.hr = function () {
  return '<hr class="magec-md-hr" />\n'
}

renderer.strong = function ({ tokens }) {
  return `<strong class="magec-md-strong">${this.parser.parseInline(tokens)}</strong>`
}

renderer.em = function ({ tokens }) {
  return `<em class="magec-md-em">${this.parser.parseInline(tokens)}</em>`
}

const options = {
  renderer,
  breaks: true,
  gfm: true,
}

// renderMarkdown is the only export. Returns the empty string on
// invalid input so callers can `v-html` the result safely without
// extra defensive checks.
export function renderMarkdown(text) {
  if (!text || typeof text !== 'string') return ''
  try {
    return marked.parse(text, options)
  } catch {
    return ''
  }
}
