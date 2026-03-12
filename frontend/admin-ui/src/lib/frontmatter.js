import yaml from 'js-yaml'

const FENCE = '---'

export function parseFrontmatter(text) {
  if (!text || typeof text !== 'string') return { meta: null, body: text || '' }

  const trimmed = text.trimStart()
  if (!trimmed.startsWith(FENCE)) return { meta: null, body: text }

  const afterFirst = trimmed.indexOf('\n')
  if (afterFirst === -1) return { meta: null, body: text }

  const rest = trimmed.slice(afterFirst + 1)
  const closing = rest.indexOf('\n' + FENCE)
  if (closing === -1) return { meta: null, body: text }

  const yamlBlock = rest.slice(0, closing)
  const body = rest.slice(closing + FENCE.length + 1).replace(/^\n/, '')

  let meta
  try {
    meta = yaml.load(yamlBlock)
  } catch {
    return { meta: null, body: text }
  }

  if (!meta || typeof meta !== 'object' || !meta.name) return { meta: null, body: text }

  return { meta, body }
}

export function skillCardData(skill) {
  const { meta } = parseFrontmatter(skill.instructions)

  const hasStoreName = skill.name && skill.name.trim()
  const hasStoreDesc = skill.description && skill.description.trim()

  if (meta) {
    const badges = []
    if (meta.license) badges.push(meta.license)
    if (meta.compatibility) badges.push(String(meta.compatibility).length > 40 ? String(meta.compatibility).slice(0, 37) + '...' : String(meta.compatibility))

    return {
      canonical: true,
      name: hasStoreName ? skill.name : (meta.name || skill.name),
      description: hasStoreDesc ? skill.description : (meta.description || skill.description || ''),
      badges,
      metadata: typeof meta.metadata === 'object' ? meta.metadata : null,
    }
  }

  const firstLine = (skill.instructions || '').trim().split('\n').find(l => l.trim())
  const fallbackDesc = firstLine ? firstLine.replace(/^#+\s*/, '').trim() : ''

  return {
    canonical: false,
    name: skill.name,
    description: skill.description || fallbackDesc || 'No description',
    badges: [],
    metadata: null,
  }
}
