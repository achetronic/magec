import { request } from './client.js'

export const voiceApi = {
  listTypes: () => request('/voice/types'),
}
