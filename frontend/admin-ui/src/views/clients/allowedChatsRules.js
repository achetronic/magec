export const ALLOWED_CHATS_RULES_ERROR = 'Invalid allowed chat rules.'

export function buildAllowedChatsPayload(val) {
  const rows = normalizeAllowedChatRows(val)
  const payload = []
  const seen = new Set()
  const errors = []

  for (const [index, row] of rows.entries()) {
    const rowNum = index + 1
    const chatRaw = row.chatId?.toString().trim() ?? ''
    const threadRaw = row.threadId?.toString().trim() ?? ''

    if (chatRaw === '' && threadRaw === '') {
      continue
    }

    if (chatRaw === '') {
      errors.push(`Row ${rowNum}: chatId is required.`)
      continue
    }

    const chatID = Number(chatRaw)
    if (Number.isNaN(chatID) || !Number.isInteger(chatID)) {
      errors.push(`Row ${rowNum}: chatId must be an integer.`)
      continue
    }
    if (Math.trunc(chatID) === 0) {
      errors.push(`Row ${rowNum}: chatId must be non-zero.`)
      continue
    }

    const item = { chatId: Math.trunc(chatID) }
    if (threadRaw !== '') {
      const threadID = Number(threadRaw)
      if (Number.isNaN(threadID) || !Number.isInteger(threadID)) {
        errors.push(`Row ${rowNum}: threadId must be an integer when provided.`)
        continue
      }
      if (Math.trunc(threadID) <= 0) {
        errors.push(`Row ${rowNum}: threadId must be greater than zero when provided.`)
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

  if (errors.length > 0) {
    throw new Error(`${ALLOWED_CHATS_RULES_ERROR} ${errors.join(' ')}`)
  }

  return payload
}

function normalizeAllowedChatRows(val) {
  if (!Array.isArray(val)) return []
  return val.map(rule => {
    if (rule && typeof rule === 'object') {
      return {
        chatId: rule.chatId !== undefined && rule.chatId !== null ? String(rule.chatId) : '',
        threadId: rule.threadId !== undefined && rule.threadId !== null ? String(rule.threadId) : '',
      }
    }
    return null
  }).filter(Boolean)
}
