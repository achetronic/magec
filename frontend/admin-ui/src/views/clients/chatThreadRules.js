export const CHAT_THREAD_RULES_ERROR = 'Invalid chat/thread rules. chatId must be a non-zero integer and threadId must be a positive integer when provided.'

export function buildAllowedChatThreadsPayload(val) {
  const rows = normalizeAllowedChatThreadRows(val)
  const payload = []
  let hasInvalidRows = false

  for (const row of rows) {
    const chatRaw = row.chatId?.toString().trim() ?? ''
    const threadRaw = row.threadId?.toString().trim() ?? ''

    if (chatRaw === '' && threadRaw === '') {
      continue
    }

    const chatId = Number(chatRaw)
    if (chatRaw === '' || Number.isNaN(chatId) || !Number.isInteger(chatId) || Math.trunc(chatId) === 0) {
      hasInvalidRows = true
      continue
    }

    const item = { chatId: Math.trunc(chatId) }
    if (threadRaw !== '') {
      const threadID = Number(threadRaw)
      if (Number.isNaN(threadID) || !Number.isInteger(threadID) || Math.trunc(threadID) <= 0) {
        hasInvalidRows = true
        continue
      }
      item.threadId = Math.trunc(threadID)
    }
    payload.push(item)
  }

  if (hasInvalidRows) {
    throw new Error(CHAT_THREAD_RULES_ERROR)
  }

  return payload
}

function parseLegacyChatThreadRule(value) {
  const match = value?.toString().match(/^\s*(-?\d+)(?:\s*-\s*(\d+))?\s*$/)
  if (!match) return null
  return { chatId: match[1], threadId: match[2] || '' }
}

function normalizeAllowedChatThreadRows(val) {
  if (!Array.isArray(val)) return []
  return val.map(rule => {
    if (typeof rule === 'string') {
      const parsed = parseLegacyChatThreadRule(rule)
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
