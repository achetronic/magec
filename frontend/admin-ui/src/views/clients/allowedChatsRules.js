export const ALLOWED_CHATS_RULES_ERROR = 'Invalid allowed chat rules. chatId must be a non-zero integer and threadId must be a positive integer when provided.'

export function buildAllowedChatsPayload(val) {
  const rows = normalizeAllowedChatRows(val)
  const payload = []
  const seen = new Set()
  let hasInvalidRows = false

  for (const row of rows) {
    const chatRaw = row.chatId?.toString().trim() ?? ''
    const threadRaw = row.threadId?.toString().trim() ?? ''

    if (chatRaw === '' && threadRaw === '') {
      continue
    }

    const chatID = Number(chatRaw)
    if (chatRaw === '' || Number.isNaN(chatID) || !Number.isInteger(chatID) || Math.trunc(chatID) === 0) {
      hasInvalidRows = true
      continue
    }

    const item = { chatId: Math.trunc(chatID) }
    if (threadRaw !== '') {
      const threadID = Number(threadRaw)
      if (Number.isNaN(threadID) || !Number.isInteger(threadID) || Math.trunc(threadID) <= 0) {
        hasInvalidRows = true
        continue
      }
      item.threadId = Math.trunc(threadID)
    }

    const key = `${item.chatId}:${item.threadId ?? 0}`
    if (!seen.has(key)) {
      seen.add(key)
      payload.push(item)
    }
  }

  if (hasInvalidRows) {
    throw new Error(ALLOWED_CHATS_RULES_ERROR)
  }

  return payload
}

function parseLegacyAllowedChatRule(value) {
  const match = value?.toString().match(/^\s*(-?\d+)(?:\s*-\s*(\d+))?\s*$/)
  if (!match) return null
  return { chatId: match[1], threadId: match[2] || '' }
}

function normalizeAllowedChatRows(val) {
  if (!Array.isArray(val)) return []
  return val.map(rule => {
    if (typeof rule === 'string') {
      const parsed = parseLegacyAllowedChatRule(rule)
      if (!parsed) return null
      return parsed
    }
    if (rule && typeof rule === 'object') {
      return {
        chatId: rule.chatId !== undefined && rule.chatId !== null ? String(rule.chatId) : '',
        threadId: rule.threadId !== undefined && rule.threadId !== null ? String(rule.threadId) : '',
      }
    }
    return null
  }).filter(Boolean)
}
