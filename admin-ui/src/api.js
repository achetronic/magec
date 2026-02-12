const BASE = '/api/v1/admin';

async function request(path, opts = {}) {
    const res = await fetch(`${BASE}${path}`, {
        headers: { 'Content-Type': 'application/json', ...opts.headers },
        ...opts,
    });
    if (res.status === 204) return null;
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
    return data;
}

export const api = {
    listBackends: () => request('/backends'),
    getBackend: (id) => request(`/backends/${id}`),
    createBackend: (b) => request('/backends', { method: 'POST', body: JSON.stringify(b) }),
    updateBackend: (id, b) => request(`/backends/${id}`, { method: 'PUT', body: JSON.stringify(b) }),
    deleteBackend: (id) => request(`/backends/${id}`, { method: 'DELETE' }),

    listMemory: () => request('/memory'),
    getMemory: (id) => request(`/memory/${id}`),
    createMemory: (m) => request('/memory', { method: 'POST', body: JSON.stringify(m) }),
    updateMemory: (id, m) => request(`/memory/${id}`, { method: 'PUT', body: JSON.stringify(m) }),
    deleteMemory: (id) => request(`/memory/${id}`, { method: 'DELETE' }),
    checkMemoryHealth: (id) => request(`/memory/${id}/health`),
    listMemoryTypes: () => request('/memory/types'),

    listMCPs: () => request('/mcps'),
    getMCP: (id) => request(`/mcps/${id}`),
    createMCP: (m) => request('/mcps', { method: 'POST', body: JSON.stringify(m) }),
    updateMCP: (id, m) => request(`/mcps/${id}`, { method: 'PUT', body: JSON.stringify(m) }),
    deleteMCP: (id) => request(`/mcps/${id}`, { method: 'DELETE' }),

    listAgents: () => request('/agents'),
    getAgent: (id) => request(`/agents/${id}`),
    createAgent: (a) => request('/agents', { method: 'POST', body: JSON.stringify(a) }),
    updateAgent: (id, a) => request(`/agents/${id}`, { method: 'PUT', body: JSON.stringify(a) }),
    deleteAgent: (id) => request(`/agents/${id}`, { method: 'DELETE' }),

    listAgentMCPs: (id) => request(`/agents/${id}/mcps`),
    linkAgentMCP: (id, mcpId) => request(`/agents/${id}/mcps/${mcpId}`, { method: 'PUT' }),
    unlinkAgentMCP: (id, mcpId) => request(`/agents/${id}/mcps/${mcpId}`, { method: 'DELETE' }),

    listCrons: () => request('/crons'),
    getCron: (id) => request(`/crons/${id}`),
    createCron: (c) => request('/crons', { method: 'POST', body: JSON.stringify(c) }),
    updateCron: (id, c) => request(`/crons/${id}`, { method: 'PUT', body: JSON.stringify(c) }),
    deleteCron: (id) => request(`/crons/${id}`, { method: 'DELETE' }),

    listClients: () => request('/clients'),
    getClient: (id) => request(`/clients/${id}`),
    createClient: (c) => request('/clients', { method: 'POST', body: JSON.stringify(c) }),
    updateClient: (id, c) => request(`/clients/${id}`, { method: 'PUT', body: JSON.stringify(c) }),
    deleteClient: (id) => request(`/clients/${id}`, { method: 'DELETE' }),
    regenerateClientToken: (id) => request(`/clients/${id}/regenerate-token`, { method: 'POST' }),
    listClientTypes: () => request('/clients/types'),
};
