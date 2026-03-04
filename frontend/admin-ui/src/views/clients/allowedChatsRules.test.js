import test from 'node:test'
import assert from 'node:assert/strict'

import { buildAllowedChatsPayload, ALLOWED_CHATS_RULES_ERROR } from './allowedChatsRules.js'

test('buildAllowedChatsPayload trims values and keeps valid rules', () => {
  const payload = buildAllowedChatsPayload([
    { chatId: ' -1001234567890 ', threadId: ' 12 ' },
    { chatId: ' -1001234567891 ', threadId: '' },
  ])

  assert.deepEqual(payload, [
    { chatId: -1001234567890, threadId: 12 },
    { chatId: -1001234567891 },
  ])
})

test('buildAllowedChatsPayload ignores fully empty rows', () => {
  const payload = buildAllowedChatsPayload([
    { chatId: '', threadId: '' },
    { chatId: '   ', threadId: '   ' },
  ])

  assert.deepEqual(payload, [])
})

test('buildAllowedChatsPayload rejects chatId 0', () => {
  assert.throws(
    () => buildAllowedChatsPayload([{ chatId: '0', threadId: '' }]),
    (error) => error.message.includes(ALLOWED_CHATS_RULES_ERROR) && error.message.includes('Row 1: chatId must be non-zero.'),
  )
})

test('buildAllowedChatsPayload rejects non-positive threadId when provided', () => {
  assert.throws(
    () => buildAllowedChatsPayload([{ chatId: '-1001234567890', threadId: '0' }]),
    (error) => error.message.includes(ALLOWED_CHATS_RULES_ERROR) && error.message.includes('Row 1: threadId must be greater than zero when provided.'),
  )
})

test('buildAllowedChatsPayload reports explicit error for missing chatId', () => {
  assert.throws(
    () => buildAllowedChatsPayload([{ chatId: ' ', threadId: '12' }]),
    (error) => error.message.includes('Row 1: chatId is required.'),
  )
})

test('buildAllowedChatsPayload deduplicates repeated chat rules', () => {
  const payload = buildAllowedChatsPayload([
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
