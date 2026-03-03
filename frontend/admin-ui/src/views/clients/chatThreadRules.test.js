import test from 'node:test'
import assert from 'node:assert/strict'

import { buildAllowedChatThreadsPayload, CHAT_THREAD_RULES_ERROR } from './chatThreadRules.js'

test('buildAllowedChatThreadsPayload trims values and keeps valid rules', () => {
  const payload = buildAllowedChatThreadsPayload([
    { chatId: ' -1001234567890 ', threadId: ' 12 ' },
    { chatId: ' -1001234567891 ', threadId: '' },
  ])

  assert.deepEqual(payload, [
    { chatId: -1001234567890, threadId: 12 },
    { chatId: -1001234567891 },
  ])
})

test('buildAllowedChatThreadsPayload ignores fully empty rows', () => {
  const payload = buildAllowedChatThreadsPayload([
    { chatId: '', threadId: '' },
    { chatId: '   ', threadId: '   ' },
  ])

  assert.deepEqual(payload, [])
})

test('buildAllowedChatThreadsPayload rejects chatId 0', () => {
  assert.throws(
    () => buildAllowedChatThreadsPayload([{ chatId: '0', threadId: '' }]),
    new Error(CHAT_THREAD_RULES_ERROR),
  )
})

test('buildAllowedChatThreadsPayload rejects non-positive threadId when provided', () => {
  assert.throws(
    () => buildAllowedChatThreadsPayload([{ chatId: '-1001234567890', threadId: '0' }]),
    new Error(CHAT_THREAD_RULES_ERROR),
  )
})

test('buildAllowedChatThreadsPayload deduplicates repeated chat-thread rules', () => {
  const payload = buildAllowedChatThreadsPayload([
    { chatId: '-1001234567890', threadId: '12' },
    { chatId: '-1001234567890', threadId: '12' },
    { chatId: '-1001234567890', threadId: '' },
    { chatId: '-1001234567890', threadId: '' },
  ])

  assert.deepEqual(payload, [
    { chatId: -1001234567890, threadId: 12 },
    { chatId: -1001234567890 },
  ])
})
