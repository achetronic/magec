const METADATA_REGEX = /<!--MAGEC_META:.*?:MAGEC_META-->\n?/gs

export function stripMetadata(text) {
  if (!text) return text
  return text.replace(METADATA_REGEX, '').trimStart()
}

export function extractMetadata(text) {
  if (!text) return null
  const match = text.match(/<!--MAGEC_META:(.*?):MAGEC_META-->/s)
  if (!match) return null
  try {
    return JSON.parse(match[1])
  } catch {
    return null
  }
}
