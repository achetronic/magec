import { request } from './client.js'

export const integrityApi = {
  deadReferences: () => request('/integrity/dead-references'),
  cleanDeadReferences: () => request('/integrity/dead-references/clean', { method: 'POST' }),
}
