// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

const METADATA_REGEX = /<!--MAGEC_META:.*?:MAGEC_META-->\n?/gs
const THREAD_HISTORY_REGEX = /<!--MAGEC_THREAD_HISTORY:.*?:MAGEC_THREAD_HISTORY-->\n?/gs

export function stripMetadata(text) {
  if (!text) return text
  return text.replace(METADATA_REGEX, '').replace(THREAD_HISTORY_REGEX, '').trimStart()
}
