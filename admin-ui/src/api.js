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
    getBackend: (name) => request(`/backends/${name}`),
    createBackend: (b) => request('/backends', { method: 'POST', body: JSON.stringify(b) }),
    updateBackend: (name, b) => request(`/backends/${name}`, { method: 'PUT', body: JSON.stringify(b) }),
    deleteBackend: (name) => request(`/backends/${name}`, { method: 'DELETE' }),

    listMemory: () => request('/memory'),
    getMemory: (name) => request(`/memory/${name}`),
    createMemory: (m) => request('/memory', { method: 'POST', body: JSON.stringify(m) }),
    updateMemory: (name, m) => request(`/memory/${name}`, { method: 'PUT', body: JSON.stringify(m) }),
    deleteMemory: (name) => request(`/memory/${name}`, { method: 'DELETE' }),
    checkMemoryHealth: (name) => request(`/memory/${name}/health`),
    listMemoryTypes: () => request('/memory/types'),

    listMCPs: () => request('/mcps'),
    getMCP: (name) => request(`/mcps/${name}`),
    createMCP: (m) => request('/mcps', { method: 'POST', body: JSON.stringify(m) }),
    updateMCP: (name, m) => request(`/mcps/${name}`, { method: 'PUT', body: JSON.stringify(m) }),
    deleteMCP: (name) => request(`/mcps/${name}`, { method: 'DELETE' }),

    listAgents: () => request('/agents'),
    getAgent: (id) => request(`/agents/${id}`),
    createAgent: (a) => request('/agents', { method: 'POST', body: JSON.stringify(a) }),
    updateAgent: (id, a) => request(`/agents/${id}`, { method: 'PUT', body: JSON.stringify(a) }),
    deleteAgent: (id) => request(`/agents/${id}`, { method: 'DELETE' }),

    listAgentMCPs: (id) => request(`/agents/${id}/mcps`),
    linkAgentMCP: (id, name) => request(`/agents/${id}/mcps/${name}`, { method: 'PUT' }),
    unlinkAgentMCP: (id, name) => request(`/agents/${id}/mcps/${name}`, { method: 'DELETE' }),

    listCrons: () => request('/crons'),
    getCron: (name) => request(`/crons/${name}`),
    createCron: (c) => request('/crons', { method: 'POST', body: JSON.stringify(c) }),
    updateCron: (name, c) => request(`/crons/${name}`, { method: 'PUT', body: JSON.stringify(c) }),
    deleteCron: (name) => request(`/crons/${name}`, { method: 'DELETE' }),

    listClients: () => request('/clients'),
    getClient: (name) => request(`/clients/${name}`),
    createClient: (c) => request('/clients', { method: 'POST', body: JSON.stringify(c) }),
    updateClient: (name, c) => request(`/clients/${name}`, { method: 'PUT', body: JSON.stringify(c) }),
    deleteClient: (name) => request(`/clients/${name}`, { method: 'DELETE' }),
    regenerateClientToken: (name) => request(`/clients/${name}/regenerate-token`, { method: 'POST' }),
    listClientTypes: () => request('/clients/types'),
};
