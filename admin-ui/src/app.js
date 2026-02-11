import { api } from './api.js';

function $(id) { return document.getElementById(id); }
function esc(s) { const d = document.createElement('div'); d.textContent = s; return d.innerHTML; }

class AdminApp {
    constructor() {
        this.backends = [];
        this.agents = [];
        this.mcps = [];
        this.crons = [];
        this.devices = [];
        this.memory = [];
        this.memoryTypes = [];
        this.expandedAgentId = null;
        this._setupTabs();
        this._setupMCPTypeToggle();
    }

    async init() {
        try { this.memoryTypes = await api.listMemoryTypes(); } catch { this.memoryTypes = []; }
        await this.refresh();
    }

    async refresh() {
        try {
            [this.backends, this.agents, this.mcps, this.crons, this.devices, this.memory] = await Promise.all([
                api.listBackends(), api.listAgents(), api.listMCPs(), api.listCrons(), api.listDevices(), api.listMemory()
            ]);
        } catch (e) {
            console.error('Failed to load data:', e);
            this.backends = []; this.agents = []; this.mcps = []; this.crons = []; this.devices = []; this.memory = [];
        }
        this._renderOverview();
        this._renderBackends();
        this._renderMemory();
        this._renderMCPs();
        this._renderAgents();
        this._renderDevices();
        this._renderCrons();
    }

    // ==================== Tabs ====================

    _setupTabs() {
        document.querySelectorAll('[data-tab]').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('[data-tab]').forEach(b => {
                    b.classList.remove('tab-active');
                    b.classList.add('tab-inactive');
                });
                btn.classList.remove('tab-inactive');
                btn.classList.add('tab-active');

                $('panelBackends').classList.toggle('hidden', btn.dataset.tab !== 'backends');
                $('panelMemory').classList.toggle('hidden', btn.dataset.tab !== 'memory');
                $('panelMcps').classList.toggle('hidden', btn.dataset.tab !== 'mcps');
                $('panelAgents').classList.toggle('hidden', btn.dataset.tab !== 'agents');
                $('panelDevices').classList.toggle('hidden', btn.dataset.tab !== 'devices');
                $('panelCrons').classList.toggle('hidden', btn.dataset.tab !== 'crons');
            });
        });
    }

    _setupMCPTypeToggle() {
        const sel = $('mcpType');
        if (sel) {
            sel.addEventListener('change', () => {
                $('mcpHttpFields').classList.toggle('hidden', sel.value !== 'http');
                $('mcpStdioFields').classList.toggle('hidden', sel.value !== 'stdio');
            });
        }
    }

    // ==================== Overview ====================

    _renderOverview() {
        $('overviewBadges').innerHTML = [
            `<span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">${this.backends.length} backend${this.backends.length !== 1 ? 's' : ''}</span>`,
            `<span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">${this.mcps.length} MCP${this.mcps.length !== 1 ? 's' : ''}</span>`,
            `<span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">${this.agents.length} agent${this.agents.length !== 1 ? 's' : ''}</span>`,
            `<span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">${this.devices.length} device${this.devices.length !== 1 ? 's' : ''}</span>`,
            `<span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">${this.crons.length} cron${this.crons.length !== 1 ? 's' : ''}</span>`,
        ].join('');
    }

    // ==================== Agents ====================

    toggleAgent(id) {
        this.expandedAgentId = this.expandedAgentId === id ? null : id;
        this._renderAgents();
    }

    _renderAgents() {
        const el = $('agentsList');
        if (!this.agents.length) {
            el.innerHTML = `<div class="text-center py-12 text-arena-500"><p class="text-sm">No agents configured</p><p class="text-xs mt-1">Create your first agent to get started</p></div>`;
            return;
        }
        el.innerHTML = this.agents.map(a => {
            const expanded = this.expandedAgentId === a.id;
            const mcpCount = (a.mcpServers || []).length;
            return `
            <div class="bg-piedra-900 border ${expanded ? 'border-sol-500/30' : 'border-piedra-700/50 hover:border-piedra-600/50'} rounded-xl transition-colors">
                <!-- Summary row -->
                <div class="flex items-center gap-3 p-4 cursor-pointer" onclick="app.toggleAgent('${esc(a.id)}')">
                    <div class="w-9 h-9 rounded-lg bg-sol-500/15 flex items-center justify-center flex-shrink-0">
                        <span class="text-sm font-semibold text-sol-400">${esc((a.name || a.id).charAt(0).toUpperCase())}</span>
                    </div>
                    <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-2">
                            <h3 class="font-medium text-arena-100 text-sm">${esc(a.name || a.id)}</h3>
                            <span class="text-[10px] text-arena-500 font-mono">${esc(a.id)}</span>
                        </div>
                        <div class="flex items-center gap-1.5 mt-1 flex-wrap">
                            <span class="px-1.5 py-0.5 bg-piedra-800 text-arena-300 text-[10px] rounded">${esc(a.llm?.backend || '?')} / ${esc(a.llm?.model || '?')}</span>
                            ${a.transcription?.backend ? '<span class="px-1.5 py-0.5 bg-piedra-800 text-arena-400 text-[10px] rounded">STT</span>' : ''}
                            ${a.tts?.backend ? '<span class="px-1.5 py-0.5 bg-lava-500/15 text-lava-300 text-[10px] rounded">TTS</span>' : ''}
                            ${mcpCount ? `<span class="px-1.5 py-0.5 bg-atlantico-500/15 text-atlantico-300 text-[10px] rounded">${mcpCount} MCP${mcpCount > 1 ? 's' : ''}</span>` : ''}
                            ${a.telegram?.enabled ? '<span class="px-1.5 py-0.5 bg-sol-500/15 text-sol-300 text-[10px] rounded">TG</span>' : ''}
                        </div>
                    </div>
                    <div class="flex items-center gap-1 flex-shrink-0">
                        <button onclick="event.stopPropagation(); app.editAgent('${esc(a.id)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
                            <svg class="w-4 h-4 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                        </button>
                        <button onclick="event.stopPropagation(); app.confirmDelete('agent', '${esc(a.id)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
                            <svg class="w-4 h-4 text-arena-400 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                        </button>
                        <svg class="w-4 h-4 text-arena-500 transition-transform ${expanded ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 9l-7 7-7-7"/></svg>
                    </div>
                </div>
                ${expanded ? this._renderAgentDetail(a) : ''}
            </div>`;
        }).join('');
    }

    _renderAgentDetail(a) {
        const row = (label, value) => value ? `<div class="detail-row"><span class="label">${label}</span><span class="value">${esc(value)}</span></div>` : '';
        const mcpNames = a.mcpServers || [];
        const tg = a.telegram || {};
        const mem = a.memory || {};

        return `
        <div class="border-t border-piedra-700/30 p-4 grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Left -->
            <div class="space-y-4">
                ${a.description ? `<p class="text-xs text-arena-400">${esc(a.description)}</p>` : ''}

                <div class="space-y-1.5">
                    <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">LLM</h4>
                    ${row('Backend', a.llm?.backend)}
                    ${row('Model', a.llm?.model)}
                </div>

                <div class="space-y-1.5">
                    <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Transcription (STT)</h4>
                    ${a.transcription?.backend
                        ? row('Backend', a.transcription.backend) + row('Model', a.transcription.model)
                        : '<p class="text-[11px] text-arena-600">Disabled</p>'}
                </div>

                <div class="space-y-1.5">
                    <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">TTS</h4>
                    ${a.tts?.backend
                        ? row('Backend', a.tts.backend) + row('Model', a.tts.model) + row('Voice', a.tts.voice) + row('Speed', a.tts.speed ? a.tts.speed + 'x' : '')
                        : '<p class="text-[11px] text-arena-600">Disabled</p>'}
                </div>

                ${a.systemPrompt ? `
                <div class="space-y-1">
                    <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">System Prompt</h4>
                    <p class="text-[11px] text-arena-300 whitespace-pre-wrap bg-piedra-800/50 rounded-lg p-2 max-h-28 overflow-y-auto">${esc(a.systemPrompt)}</p>
                </div>` : ''}
            </div>

            <!-- Right -->
            <div class="space-y-4">
                <div class="space-y-1.5">
                    <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">MCP Servers</h4>
                    ${mcpNames.length ? mcpNames.map(name => {
                        const mcp = this.mcps.find(m => m.name === name);
                        return `<div class="flex items-center gap-2 p-1.5 bg-piedra-800/50 rounded-lg">
                            <svg class="w-3 h-3 text-atlantico-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>
                            <span class="text-[11px] text-arena-200">${esc(name)}</span>
                            <span class="text-[10px] text-arena-500">${esc(mcp?.type || 'http')}</span>
                        </div>`;
                    }).join('') : '<p class="text-[11px] text-arena-600">None linked</p>'}
                </div>

                <div class="space-y-1.5">
                    <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Telegram</h4>
                    <div class="flex items-center gap-1.5">
                        <span class="w-1.5 h-1.5 rounded-full ${tg.enabled ? 'bg-green-400' : 'bg-arena-600'}"></span>
                        <span class="text-[11px] ${tg.enabled ? 'text-green-400' : 'text-arena-600'}">${tg.enabled ? 'Enabled' : 'Disabled'}</span>
                        ${tg.enabled && tg.responseMode ? `<span class="text-[10px] text-arena-500 ml-1">(${esc(tg.responseMode)})</span>` : ''}
                    </div>
                </div>

                <div class="space-y-1.5">
                    <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Memory</h4>
                    ${row('Session', mem.session || 'Not configured')}
                    ${row('Long-term', mem.longTerm || 'Not configured')}
                </div>
            </div>
        </div>`;
    }

    showAgentDialog(agent = null) {
        const isEdit = !!agent;
        $('agentDialogTitle').textContent = isEdit ? 'Edit Agent' : 'New Agent';
        $('agentEditId').value = isEdit ? agent.id : '';
        $('agentId').value = agent?.id || '';

        $('agentName').value = agent?.name || '';
        $('agentDescription').value = agent?.description || '';
        $('agentLlmModel').value = agent?.llm?.model || '';
        $('agentSystemPrompt').value = agent?.systemPrompt || '';
        $('agentSystemPromptSuffix').value = agent?.systemPromptSuffix || '';
        $('agentTelegramEnabled').checked = agent?.telegram?.enabled || false;
        $('agentTelegramToken').value = agent?.telegram?.token || '';
        $('agentTelegramResponseMode').value = agent?.telegram?.responseMode || 'text';

        const noneOpt = '<option value="">(none)</option>';
        const backendOpts = this.backends.map(b =>
            `<option value="${esc(b.name)}">${esc(b.name)} (${esc(b.type)})</option>`
        ).join('');

        $('agentLlmBackend').innerHTML = backendOpts;
        $('agentLlmBackend').value = agent?.llm?.backend || '';

        $('agentTranscriptionBackend').innerHTML = noneOpt + backendOpts;
        $('agentTranscriptionBackend').value = agent?.transcription?.backend || '';
        $('agentTranscriptionModel').value = agent?.transcription?.model || '';

        $('agentTtsBackend').innerHTML = noneOpt + backendOpts;
        $('agentTtsBackend').value = agent?.tts?.backend || '';
        $('agentTtsModel').value = agent?.tts?.model || '';
        $('agentTtsVoice').value = agent?.tts?.voice || '';
        $('agentTtsSpeed').value = agent?.tts?.speed || '';

        const mcpContainer = $('agentMcpCheckboxes');
        const agentMcpNames = agent?.mcpServers || [];
        if (this.mcps.length) {
            $('agentMcpEmpty').classList.add('hidden');
            mcpContainer.innerHTML = this.mcps.map(m => `
                <label class="flex items-center gap-1.5 px-2.5 py-1 bg-piedra-800 rounded-lg cursor-pointer hover:bg-piedra-700 transition-colors">
                    <input type="checkbox" value="${esc(m.name)}" ${agentMcpNames.includes(m.name) ? 'checked' : ''} class="rounded border-piedra-600 bg-piedra-800 text-sol-500 focus:ring-sol-500">
                    <span class="text-xs text-arena-300">${esc(m.name)}</span>
                </label>
            `).join('');
        } else {
            mcpContainer.innerHTML = '';
            $('agentMcpEmpty').classList.remove('hidden');
        }

        const noneMemOpt = '<option value="">(none)</option>';
        const sessionOpts = this.memory.filter(m => m.category === 'session').map(m =>
            `<option value="${esc(m.name)}">${esc(m.name)}</option>`
        ).join('');
        const longTermOpts = this.memory.filter(m => m.category === 'longterm').map(m =>
            `<option value="${esc(m.name)}">${esc(m.name)}</option>`
        ).join('');
        $('agentMemorySession').innerHTML = noneMemOpt + sessionOpts;
        $('agentMemorySession').value = agent?.memory?.session || '';
        $('agentMemoryLongTerm').innerHTML = noneMemOpt + longTermOpts;
        $('agentMemoryLongTerm').value = agent?.memory?.longTerm || '';

        $('agentDialog').showModal();
    }

    async editAgent(id) {
        const agent = this.agents.find(a => a.id === id);
        if (agent) this.showAgentDialog(agent);
    }

    async saveAgent() {
        const editId = $('agentEditId').value;
        const isEdit = !!editId;

        const selectedMcps = Array.from($('agentMcpCheckboxes').querySelectorAll('input[type=checkbox]:checked'))
            .map(cb => cb.value);

        const agent = {
            id: $('agentId').value.trim(),
            name: $('agentName').value.trim() || $('agentId').value.trim(),
            description: $('agentDescription').value.trim(),
            systemPrompt: $('agentSystemPrompt').value.trim(),
            systemPromptSuffix: $('agentSystemPromptSuffix').value.trim(),
            llm: { backend: $('agentLlmBackend').value, model: $('agentLlmModel').value.trim() },
            transcription: {
                backend: $('agentTranscriptionBackend').value,
                model: $('agentTranscriptionModel').value.trim(),
            },
            tts: {
                backend: $('agentTtsBackend').value,
                model: $('agentTtsModel').value.trim(),
                voice: $('agentTtsVoice').value.trim(),
                speed: parseFloat($('agentTtsSpeed').value) || 0,
            },
            mcpServers: selectedMcps,
            memory: {
                session: $('agentMemorySession').value,
                longTerm: $('agentMemoryLongTerm').value,
            },
            telegram: {
                enabled: $('agentTelegramEnabled').checked,
                token: $('agentTelegramToken').value.trim(),
                responseMode: $('agentTelegramResponseMode').value,
            },
        };

        try {
            if (isEdit) {
                await api.updateAgent(editId, agent);
            } else {
                await api.createAgent(agent);
            }
            $('agentDialog').close();
            await this.refresh();
        } catch (e) {
            alert('Error: ' + e.message);
        }
    }

    // ==================== Backends ====================

    _renderBackends() {
        const el = $('backendsList');
        if (!this.backends.length) {
            el.innerHTML = `<div class="col-span-full text-center py-12 text-arena-500"><p class="text-sm">No backends configured</p></div>`;
            return;
        }

        const usedBy = {};
        for (const a of this.agents) {
            const label = a.name || a.id;
            if (a.llm?.backend) (usedBy[a.llm.backend] ??= new Set()).add(label);
            if (a.transcription?.backend) (usedBy[a.transcription.backend] ??= new Set()).add(label);
            if (a.tts?.backend) (usedBy[a.tts.backend] ??= new Set()).add(label);
            if (a.memory?.longTerm?.embedding?.backend) (usedBy[a.memory.longTerm.embedding.backend] ??= new Set()).add(label);
        }

        el.innerHTML = this.backends.map(b => {
            const agents = [...(usedBy[b.name] || [])];
            return `
            <div class="bg-piedra-900 border border-piedra-700/50 rounded-xl p-4 hover:border-piedra-600/50 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-2">
                    <div class="flex items-center gap-3 min-w-0">
                        <div class="w-8 h-8 rounded-lg bg-piedra-800 flex items-center justify-center flex-shrink-0">
                            <span class="text-[10px] font-mono font-bold text-arena-400">${esc(b.type?.substring(0, 3).toUpperCase())}</span>
                        </div>
                        <div class="min-w-0">
                            <h3 class="font-medium text-arena-100 text-sm">${esc(b.name)}</h3>
                            <p class="text-[10px] text-arena-500 truncate">${esc(b.url || b.type)}</p>
                        </div>
                    </div>
                    <div class="flex gap-0.5 flex-shrink-0">
                        <button onclick="app.editBackend('${esc(b.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
                            <svg class="w-3.5 h-3.5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                        </button>
                        <button onclick="app.confirmDelete('backend', '${esc(b.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
                            <svg class="w-3.5 h-3.5 text-arena-400 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                        </button>
                    </div>
                </div>
                ${agents.length ? `<div class="flex flex-wrap gap-1">${agents.map(n => `<span class="px-1.5 py-0.5 bg-sol-500/10 text-sol-300 text-[10px] rounded">${esc(n)}</span>`).join('')}</div>` : `<p class="text-[10px] text-arena-600">Not used by any agent</p>`}
            </div>`;
        }).join('');
    }

    showBackendDialog(backend = null) {
        const isEdit = !!backend;
        $('backendDialogTitle').textContent = isEdit ? 'Edit Backend' : 'New Backend';
        $('backendEditName').value = isEdit ? backend.name : '';
        $('backendName').value = backend?.name || '';

        $('backendType').value = backend?.type || 'openai';
        $('backendUrl').value = backend?.url || '';
        $('backendApiKey').value = backend?.apiKey || '';
        $('backendDialog').showModal();
    }

    async editBackend(name) {
        const b = this.backends.find(b => b.name === name);
        if (b) this.showBackendDialog(b);
    }

    async saveBackend() {
        const editName = $('backendEditName').value;
        const isEdit = !!editName;
        const backend = {
            name: $('backendName').value.trim(),
            type: $('backendType').value,
            url: $('backendUrl').value.trim(),
            apiKey: $('backendApiKey').value.trim(),
        };
        try {
            if (isEdit) {
                await api.updateBackend(editName, backend);
            } else {
                await api.createBackend(backend);
            }
            $('backendDialog').close();
            await this.refresh();
        } catch (e) {
            alert('Error: ' + e.message);
        }
    }

    // ==================== Memory Providers ====================

    _renderMemory() {
        const sessionEl = $('memorySessionList');
        const longTermEl = $('memoryLongTermList');

        const sessionTypesList = this.memoryTypes.filter(t => t.categories?.includes('session'));
        const longTermTypesList = this.memoryTypes.filter(t => t.categories?.includes('longterm'));

        $('memorySessionBadges').innerHTML = sessionTypesList.map(t =>
            `<span class="text-[10px] px-1.5 py-0.5 rounded bg-lava-500/15 text-lava-300 font-medium">${esc(t.displayName)}</span>`
        ).join('');
        $('memoryLongTermBadges').innerHTML = longTermTypesList.map(t =>
            `<span class="text-[10px] px-1.5 py-0.5 rounded bg-atlantico-500/15 text-atlantico-300 font-medium">${esc(t.displayName)}</span>`
        ).join('');

        const usedBy = {};
        for (const a of this.agents) {
            const label = a.name || a.id;
            if (a.memory?.session) (usedBy[a.memory.session] ??= new Set()).add(label);
            if (a.memory?.longTerm) (usedBy[a.memory.longTerm] ??= new Set()).add(label);
        }

        const sessionProviders = this.memory.filter(m => m.category === 'session');
        const longTermProviders = this.memory.filter(m => m.category === 'longterm');

        const renderCard = (m) => {
            const agents = [...(usedBy[m.name] || [])];
            const typeInfo = this.memoryTypes.find(t => t.type === m.type);
            const isSession = m.category === 'session';
            const displayName = typeInfo?.displayName || m.type;
            const abbr = displayName.substring(0, 3).toUpperCase();
            const cfg = m.config || {};
            const subtitle = cfg.connectionString || 'not configured';
            const safeId = m.name.replace(/[^a-zA-Z0-9_-]/g, '_');
            return `
            <div class="bg-piedra-900 border border-piedra-700/50 rounded-xl p-4 hover:border-piedra-600/50 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-2">
                    <div class="flex items-center gap-3 min-w-0">
                        <div class="w-8 h-8 rounded-lg ${isSession ? 'bg-lava-500/15' : 'bg-atlantico-500/15'} flex items-center justify-center flex-shrink-0 relative">
                            <span class="text-[10px] font-mono font-bold ${isSession ? 'text-lava-300' : 'text-atlantico-300'}">${abbr}</span>
                            <span id="memHealth_${safeId}" class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900 bg-piedra-600" title="Checking..."></span>
                        </div>
                        <div class="min-w-0">
                            <div class="flex items-center gap-1.5">
                                <h3 class="font-medium text-arena-100 text-sm">${esc(m.name)}</h3>
                                <span class="text-[10px] px-1.5 py-0.5 rounded bg-piedra-800 text-arena-500">${esc(displayName)}</span>
                            </div>
                            <p class="text-[10px] text-arena-500 truncate">${esc(subtitle)}</p>
                        </div>
                    </div>
                    <div class="flex gap-0.5 flex-shrink-0">
                        <button onclick="app.testMemoryHealth('${esc(m.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Test Connection">
                            <svg class="w-3.5 h-3.5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>
                        </button>
                        <button onclick="app.editMemory('${esc(m.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
                            <svg class="w-3.5 h-3.5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                        </button>
                        <button onclick="app.confirmDelete('memory', '${esc(m.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
                            <svg class="w-3.5 h-3.5 text-arena-400 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                        </button>
                    </div>
                </div>
                ${m.embedding?.backend ? `<p class="text-[10px] text-arena-400 mb-2">Embedding: ${esc(m.embedding.backend)} / ${esc(m.embedding.model || '?')}</p>` : ''}
                ${cfg.ttl ? `<p class="text-[10px] text-arena-400 mb-2">TTL: ${esc(cfg.ttl)}</p>` : ''}
                ${agents.length ? `<div class="flex flex-wrap gap-1">${agents.map(n => `<span class="px-1.5 py-0.5 bg-sol-500/10 text-sol-300 text-[10px] rounded">${esc(n)}</span>`).join('')}</div>` : `<p class="text-[10px] text-arena-600">Not used by any agent</p>`}
            </div>`;
        };

        const emptyMsg = (text) => `<div class="col-span-full text-center py-6 text-arena-600"><p class="text-xs">${text}</p></div>`;

        sessionEl.innerHTML = sessionProviders.length
            ? sessionProviders.map(renderCard).join('')
            : emptyMsg('No session providers configured');

        longTermEl.innerHTML = longTermProviders.length
            ? longTermProviders.map(renderCard).join('')
            : emptyMsg('No long-term providers configured');

        for (const m of this.memory) {
            const safeId = m.name.replace(/[^a-zA-Z0-9_-]/g, '_');
            api.checkMemoryHealth(m.name).then(res => {
                const dot = document.getElementById(`memHealth_${safeId}`);
                if (!dot) return;
                dot.className = `absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900 ${res.healthy ? 'bg-green-500' : 'bg-lava-500'}`;
                dot.title = res.detail || (res.healthy ? 'Connected' : 'Unreachable');
            }).catch(() => {
                const dot = document.getElementById(`memHealth_${safeId}`);
                if (!dot) return;
                dot.className = 'absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900 bg-lava-500';
                dot.title = 'Health check failed';
            });
        }
    }

    async testMemoryHealth(name) {
        const safeId = name.replace(/[^a-zA-Z0-9_-]/g, '_');
        const dot = document.getElementById(`memHealth_${safeId}`);
        if (dot) {
            dot.className = 'absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900 bg-piedra-600 animate-pulse';
            dot.title = 'Testing...';
        }
        try {
            const res = await api.checkMemoryHealth(name);
            if (dot) {
                dot.className = `absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900 ${res.healthy ? 'bg-green-500' : 'bg-lava-500'}`;
                dot.title = res.detail || (res.healthy ? 'Connected' : 'Unreachable');
            }
        } catch {
            if (dot) {
                dot.className = 'absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900 bg-lava-500';
                dot.title = 'Health check failed';
            }
        }
    }

    newSessionProvider() {
        this.showMemoryDialog(null, 'session');
    }

    newLongTermProvider() {
        this.showMemoryDialog(null, 'longterm');
    }

    showMemoryDialog(mem = null, forceCategory = null) {
        const isEdit = !!mem;
        const category = mem?.category || forceCategory || 'session';
        const typesInCategory = this.memoryTypes.filter(t => t.categories?.includes(category));
        const type = mem?.type || typesInCategory[0]?.type || 'redis';
        const categoryLabel = category === 'session' ? 'Session Provider' : 'Long-Term Provider';
        $('memoryDialogTitle').textContent = isEdit ? `Edit ${categoryLabel}` : `New ${categoryLabel}`;
        $('memoryEditName').value = isEdit ? mem.name : '';
        $('memoryCategory').value = category;
        $('memoryName').value = mem?.name || '';


        const typeSelect = $('memoryType');
        typeSelect.innerHTML = typesInCategory.map(t =>
            `<option value="${esc(t.type)}">${esc(t.displayName)}</option>`
        ).join('');
        typeSelect.value = type;

        typeSelect.onchange = () => this._renderMemoryConfigFields(typeSelect.value, category, {});

        $('memoryEmbeddingModel').value = mem?.embedding?.model || '';
        const backendOpts = '<option value="">(none)</option>' + this.backends.map(b =>
            `<option value="${esc(b.name)}">${esc(b.name)}</option>`
        ).join('');
        $('memoryEmbeddingBackend').innerHTML = backendOpts;
        $('memoryEmbeddingBackend').value = mem?.embedding?.backend || '';

        this._renderMemoryConfigFields(type, category, mem?.config || {});

        $('memoryTestLabel').textContent = 'Test Connection';
        $('memoryTestBtn').disabled = false;
        $('memoryTestBtn').className = 'flex items-center gap-1.5 px-3 py-2 text-xs text-arena-400 hover:text-arena-200 hover:bg-piedra-800 rounded-lg border border-piedra-700 transition-colors';

        $('memoryDialog').showModal();
    }

    _renderMemoryConfigFields(type, category, cfg) {
        const typeInfo = this.memoryTypes.find(t => t.type === type);
        const fields = typeInfo?.fields || [];
        const container = $('memoryConfigFields');
        container.innerHTML = fields.map(f => {
            const val = cfg[f.key] ?? f.default ?? '';
            const inputType = f.type === 'password' ? 'password' : 'text';
            const mono = f.key === 'connectionString' ? ' font-mono text-xs' : '';
            return `<div>
                <label class="block text-xs text-arena-400 mb-1">${esc(f.label)}${f.required ? ' <span class="text-lava-400">*</span>' : ''}</label>
                <input id="memCfg_${f.key}" type="${inputType}" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm${mono} focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none" placeholder="${esc(f.placeholder || '')}" value="${esc(String(val))}">
            </div>`;
        }).join('');
        $('memoryEmbeddingFields').classList.toggle('hidden', category !== 'longterm');
    }

    async editMemory(name) {
        const m = this.memory.find(m => m.name === name);
        if (m) this.showMemoryDialog(m);
    }

    async testMemoryConnection() {
        const editName = $('memoryEditName').value;
        const isEdit = !!editName;

        if (!isEdit) {
            $('memoryTestLabel').textContent = 'Save first to test';
            return;
        }

        const btn = $('memoryTestBtn');
        const label = $('memoryTestLabel');
        btn.disabled = true;
        label.textContent = 'Testing...';

        try {
            const res = await api.checkMemoryHealth(editName);
            if (res.healthy) {
                label.textContent = '✓ Connected';
                btn.classList.add('text-green-400', 'border-green-500/30');
                btn.classList.remove('text-arena-400', 'border-piedra-700', 'text-lava-400', 'border-lava-500/30');
            } else {
                label.textContent = `✗ ${res.detail}`;
                btn.classList.add('text-lava-400', 'border-lava-500/30');
                btn.classList.remove('text-arena-400', 'border-piedra-700', 'text-green-400', 'border-green-500/30');
            }
        } catch {
            label.textContent = '✗ Check failed';
            btn.classList.add('text-lava-400', 'border-lava-500/30');
            btn.classList.remove('text-arena-400', 'border-piedra-700', 'text-green-400', 'border-green-500/30');
        }
        btn.disabled = false;
    }

    async saveMemory() {
        const editName = $('memoryEditName').value;
        const isEdit = !!editName;
        const type = $('memoryType').value;
        const category = $('memoryCategory').value;
        const mem = { name: $('memoryName').value.trim(), type, category };
        const typeInfo = this.memoryTypes.find(t => t.type === type);
        mem.config = {};
        for (const f of (typeInfo?.fields || [])) {
            const el = document.getElementById(`memCfg_${f.key}`);
            if (el && el.value.trim()) mem.config[f.key] = el.value.trim();
        }
        if (category === 'longterm') {
            const embBackend = $('memoryEmbeddingBackend').value;
            if (embBackend) {
                mem.embedding = { backend: embBackend, model: $('memoryEmbeddingModel').value.trim() };
            }
        }
        try {
            if (isEdit) {
                await api.updateMemory(editName, mem);
            } else {
                await api.createMemory(mem);
            }
            $('memoryDialog').close();
            await this.refresh();
        } catch (e) {
            alert('Error: ' + e.message);
        }
    }

    // ==================== MCPs ====================

    _renderMCPs() {
        const el = $('mcpsList');
        if (!this.mcps.length) {
            el.innerHTML = `<div class="col-span-full text-center py-12 text-arena-500"><p class="text-sm">No MCP servers configured</p></div>`;
            return;
        }

        const agentsByMcp = {};
        for (const a of this.agents) {
            for (const name of (a.mcpServers || [])) {
                (agentsByMcp[name] ??= []).push(a.name || a.id);
            }
        }

        el.innerHTML = this.mcps.map(m => {
            const agents = agentsByMcp[m.name] || [];
            return `
            <div class="bg-piedra-900 border border-piedra-700/50 rounded-xl p-4 hover:border-piedra-600/50 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-2">
                    <div class="flex items-center gap-3 min-w-0">
                        <div class="w-8 h-8 rounded-lg bg-atlantico-500/15 flex items-center justify-center flex-shrink-0">
                            <svg class="w-4 h-4 text-atlantico-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>
                        </div>
                        <div class="min-w-0">
                            <div class="flex items-center gap-1.5">
                                <h3 class="font-medium text-arena-100 text-sm">${esc(m.name)}</h3>
                                <span class="text-[10px] px-1.5 py-0.5 rounded bg-piedra-800 text-arena-500">${esc(m.type || 'http')}</span>
                            </div>
                            <p class="text-[10px] text-arena-500 truncate">${esc(m.endpoint || m.command || '')}</p>
                        </div>
                    </div>
                    <div class="flex gap-0.5 flex-shrink-0">
                        <button onclick="app.editMCP('${esc(m.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
                            <svg class="w-3.5 h-3.5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                        </button>
                        <button onclick="app.confirmDelete('mcp', '${esc(m.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
                            <svg class="w-3.5 h-3.5 text-arena-400 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                        </button>
                    </div>
                </div>
                ${m.systemPrompt ? `<p class="text-[10px] text-arena-400 mb-2 line-clamp-2">${esc(m.systemPrompt)}</p>` : ''}
                ${agents.length ? `<div class="flex flex-wrap gap-1">${agents.map(n => `<span class="px-1.5 py-0.5 bg-sol-500/10 text-sol-300 text-[10px] rounded">${esc(n)}</span>`).join('')}</div>` : `<p class="text-[10px] text-arena-600">Not linked to any agent</p>`}
            </div>`;
        }).join('');
    }

    showMCPDialog(mcp = null) {
        const isEdit = !!mcp;
        $('mcpDialogTitle').textContent = isEdit ? 'Edit MCP Server' : 'New MCP Server';
        $('mcpEditName').value = isEdit ? mcp.name : '';
        $('mcpName').value = mcp?.name || '';

        $('mcpType').value = mcp?.type || 'http';
        $('mcpEndpoint').value = mcp?.endpoint || '';
        $('mcpCommand').value = mcp?.command || '';
        $('mcpArgs').value = (mcp?.args || []).join(', ');
        $('mcpSystemPrompt').value = mcp?.systemPrompt || '';

        $('mcpHttpFields').classList.toggle('hidden', (mcp?.type || 'http') !== 'http');
        $('mcpStdioFields').classList.toggle('hidden', (mcp?.type || 'http') !== 'stdio');

        $('mcpDialog').showModal();
    }

    async editMCP(name) {
        const m = this.mcps.find(m => m.name === name);
        if (m) this.showMCPDialog(m);
    }

    async saveMCP() {
        const editName = $('mcpEditName').value;
        const isEdit = !!editName;
        const type = $('mcpType').value;
        const mcp = {
            name: $('mcpName').value.trim(),
            type,
            systemPrompt: $('mcpSystemPrompt').value.trim(),
        };
        if (type === 'http') {
            mcp.endpoint = $('mcpEndpoint').value.trim();
        } else {
            mcp.command = $('mcpCommand').value.trim();
            const argsStr = $('mcpArgs').value.trim();
            mcp.args = argsStr ? argsStr.split(',').map(s => s.trim()).filter(Boolean) : [];
        }
        try {
            if (isEdit) {
                await api.updateMCP(editName, mcp);
            } else {
                await api.createMCP(mcp);
            }
            $('mcpDialog').close();
            await this.refresh();
        } catch (e) {
            alert('Error: ' + e.message);
        }
    }

    // ==================== Crons ====================

    _renderCrons() {
        const el = $('cronsList');
        if (!this.crons.length) {
            el.innerHTML = `<div class="col-span-full text-center py-12 text-arena-500">
                <svg class="w-10 h-10 mx-auto mb-3 text-arena-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                <p class="text-sm">No cron jobs configured</p>
                <p class="text-xs mt-1">Schedule prompts to run automatically on agents</p>
            </div>`;
            return;
        }
        el.innerHTML = this.crons.map(c => {
            const agent = this.agents.find(a => a.id === c.agentId);
            const agentLabel = agent?.name || c.agentId || '?';
            return `
            <div class="bg-piedra-900 border border-piedra-700/50 rounded-xl p-4 hover:border-piedra-600/50 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-2">
                    <div class="flex items-center gap-3 min-w-0">
                        <div class="w-8 h-8 rounded-lg ${c.enabled ? 'bg-sol-500/15' : 'bg-piedra-800'} flex items-center justify-center flex-shrink-0">
                            <svg class="w-4 h-4 ${c.enabled ? 'text-sol-400' : 'text-arena-500'}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                        </div>
                        <div class="min-w-0">
                            <div class="flex items-center gap-1.5">
                                <h3 class="font-medium text-arena-100 text-sm">${esc(c.name)}</h3>
                                ${c.enabled ? '' : '<span class="text-[10px] px-1.5 py-0.5 rounded bg-piedra-800 text-arena-500">paused</span>'}
                            </div>
                            <p class="text-[10px] text-arena-500 font-mono">${esc(c.schedule)}</p>
                        </div>
                    </div>
                    <div class="flex gap-0.5 flex-shrink-0">
                        <button onclick="app.editCron('${esc(c.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
                            <svg class="w-3.5 h-3.5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                        </button>
                        <button onclick="app.confirmDelete('cron', '${esc(c.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
                            <svg class="w-3.5 h-3.5 text-arena-400 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                        </button>
                    </div>
                </div>
                ${c.description ? `<p class="text-[10px] text-arena-400 mb-2">${esc(c.description)}</p>` : ''}
                <div class="flex items-center gap-2">
                    <span class="px-1.5 py-0.5 bg-sol-500/10 text-sol-300 text-[10px] rounded">${esc(agentLabel)}</span>
                </div>
                <p class="text-[10px] text-arena-500 mt-2 line-clamp-2 italic">"${esc(c.prompt)}"</p>
            </div>`;
        }).join('');
    }

    showCronDialog(cron = null) {
        const isEdit = !!cron;
        $('cronDialogTitle').textContent = isEdit ? 'Edit Cron Job' : 'New Cron Job';
        $('cronEditName').value = isEdit ? cron.name : '';
        $('cronName').value = cron?.name || '';

        $('cronDescription').value = cron?.description || '';
        $('cronSchedule').value = cron?.schedule || '';
        $('cronPrompt').value = cron?.prompt || '';
        $('cronEnabled').checked = cron?.enabled ?? true;

        $('cronAgentId').innerHTML = this.agents.map(a =>
            `<option value="${esc(a.id)}" ${a.id === cron?.agentId ? 'selected' : ''}>${esc(a.name || a.id)}</option>`
        ).join('');

        $('cronDialog').showModal();
    }

    async editCron(name) {
        const c = this.crons.find(c => c.name === name);
        if (c) this.showCronDialog(c);
    }

    async saveCron() {
        const editName = $('cronEditName').value;
        const isEdit = !!editName;
        const cron = {
            name: $('cronName').value.trim(),
            description: $('cronDescription').value.trim(),
            schedule: $('cronSchedule').value.trim(),
            agentId: $('cronAgentId').value,
            prompt: $('cronPrompt').value.trim(),
            enabled: $('cronEnabled').checked,
        };
        try {
            if (isEdit) {
                await api.updateCron(editName, cron);
            } else {
                await api.createCron(cron);
            }
            $('cronDialog').close();
            await this.refresh();
        } catch (e) {
            alert('Error: ' + e.message);
        }
    }

    // ==================== Delete Confirmation ====================

    // ==================== Devices ====================

    _renderDevices() {
        const el = $('devicesList');
        if (!this.devices.length) {
            el.innerHTML = `<div class="col-span-full text-center py-12 text-arena-500">
                <svg class="w-10 h-10 mx-auto mb-3 text-arena-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M12 18h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"/></svg>
                <p class="text-sm">No devices configured</p>
                <p class="text-xs mt-1">Create a device to pair voice-UI access points</p>
            </div>`;
            return;
        }
        el.innerHTML = this.devices.map(d => {
            const agentCount = (d.allowedAgents || []).length;
            const defaultAgent = this.agents.find(a => a.id === d.defaultAgent);
            const defaultLabel = defaultAgent?.name || d.defaultAgent || '?';
            return `
            <div class="bg-piedra-900 border border-piedra-700/50 rounded-xl p-4 hover:border-piedra-600/50 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-2">
                    <div class="flex items-center gap-3 min-w-0">
                        <div class="w-8 h-8 rounded-lg ${d.enabled ? 'bg-sol-500/15' : 'bg-piedra-800'} flex items-center justify-center flex-shrink-0">
                            <svg class="w-4 h-4 ${d.enabled ? 'text-sol-400' : 'text-arena-500'}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 18h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"/></svg>
                        </div>
                        <div class="min-w-0">
                            <div class="flex items-center gap-1.5">
                                <h3 class="font-medium text-arena-100 text-sm">${esc(d.name)}</h3>
                                ${d.enabled ? '' : '<span class="text-[10px] px-1.5 py-0.5 rounded bg-piedra-800 text-arena-500">disabled</span>'}
                            </div>
                            <p class="text-[10px] text-arena-500 font-mono truncate">${esc(d.token)}</p>
                        </div>
                    </div>
                    <div class="flex gap-0.5 flex-shrink-0">
                        <button onclick="app.editDevice('${esc(d.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
                            <svg class="w-3.5 h-3.5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                        </button>
                        <button onclick="app.confirmDelete('device', '${esc(d.name)}')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
                            <svg class="w-3.5 h-3.5 text-arena-400 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                        </button>
                    </div>
                </div>
                <div class="flex items-center gap-2 flex-wrap">
                    <span class="px-1.5 py-0.5 bg-sol-500/10 text-sol-300 text-[10px] rounded">${esc(defaultLabel)}</span>
                    ${agentCount > 1 ? `<span class="text-[10px] text-arena-500">+ ${agentCount - 1} more</span>` : ''}
                </div>
            </div>`;
        }).join('');
    }

    showDeviceDialog(device = null) {
        const isEdit = !!device;
        $('deviceDialogTitle').textContent = isEdit ? 'Edit Device' : 'New Device';
        $('deviceEditName').value = isEdit ? device.name : '';
        $('deviceName').value = device?.name || '';

        $('deviceEnabled').checked = device?.enabled ?? true;

        const agentOpts = this.agents.map(a =>
            `<option value="${esc(a.id)}" ${a.id === device?.defaultAgent ? 'selected' : ''}>${esc(a.name || a.id)}</option>`
        ).join('');
        $('deviceDefaultAgent').innerHTML = agentOpts;

        const allowedAgents = device?.allowedAgents || [];
        if (this.agents.length) {
            $('deviceAgentEmpty').classList.add('hidden');
            $('deviceAgentCheckboxes').innerHTML = this.agents.map(a => `
                <label class="flex items-center gap-1.5 px-2.5 py-1 bg-piedra-800 rounded-lg cursor-pointer hover:bg-piedra-700 transition-colors">
                    <input type="checkbox" value="${esc(a.id)}" ${allowedAgents.includes(a.id) ? 'checked' : ''} class="rounded border-piedra-600 bg-piedra-800 text-sol-500 focus:ring-sol-500">
                    <span class="text-xs text-arena-300">${esc(a.name || a.id)}</span>
                </label>
            `).join('');
        } else {
            $('deviceAgentCheckboxes').innerHTML = '';
            $('deviceAgentEmpty').classList.remove('hidden');
        }

        if (isEdit && device.token) {
            $('deviceTokenSection').classList.remove('hidden');
            $('deviceTokenDisplay').value = device.token;
        } else {
            $('deviceTokenSection').classList.add('hidden');
            $('deviceTokenDisplay').value = '';
        }

        $('deviceDialog').showModal();
    }

    async editDevice(name) {
        const d = this.devices.find(d => d.name === name);
        if (d) this.showDeviceDialog(d);
    }

    async saveDevice() {
        const editName = $('deviceEditName').value;
        const isEdit = !!editName;

        const selectedAgents = Array.from($('deviceAgentCheckboxes').querySelectorAll('input[type=checkbox]:checked'))
            .map(cb => cb.value);

        const device = {
            name: $('deviceName').value.trim(),
            defaultAgent: $('deviceDefaultAgent').value,
            allowedAgents: selectedAgents,
            enabled: $('deviceEnabled').checked,
        };
        try {
            if (isEdit) {
                await api.updateDevice(editName, device);
            } else {
                await api.createDevice(device);
            }
            $('deviceDialog').close();
            await this.refresh();
        } catch (e) {
            alert('Error: ' + e.message);
        }
    }

    copyDeviceToken() {
        const token = $('deviceTokenDisplay').value;
        navigator.clipboard.writeText(token).then(() => {
            const btn = $('deviceTokenDisplay').nextElementSibling;
            const orig = btn.innerHTML;
            btn.innerHTML = '<span class="text-xs text-sol-400">Copied!</span>';
            setTimeout(() => { btn.innerHTML = orig; }, 1500);
        });
    }

    async regenerateToken() {
        const name = $('deviceEditName').value;
        if (!name) return;
        if (!confirm('Regenerate token? The old token will stop working immediately.')) return;
        try {
            const updated = await api.regenerateDeviceToken(name);
            $('deviceTokenDisplay').value = updated.token;
            await this.refresh();
        } catch (e) {
            alert('Error: ' + e.message);
        }
    }

    // ==================== Delete Confirmation ====================

    confirmDelete(type, id) {
        const labels = { agent: 'agent', backend: 'backend', memory: 'memory provider', mcp: 'MCP server', cron: 'cron job', device: 'device' };
        $('confirmMessage').textContent = `Delete ${labels[type]} "${id}"? This cannot be undone.`;
        const btn = $('confirmBtn');
        btn.onclick = async () => {
            try {
                if (type === 'agent') await api.deleteAgent(id);
                else if (type === 'backend') await api.deleteBackend(id);
                else if (type === 'memory') await api.deleteMemory(id);
                else if (type === 'mcp') await api.deleteMCP(id);
                else if (type === 'cron') await api.deleteCron(id);
                else if (type === 'device') await api.deleteDevice(id);
                $('confirmDialog').close();
                await this.refresh();
            } catch (e) {
                alert('Error: ' + e.message);
            }
        };
        $('confirmDialog').showModal();
    }
}

const app = new AdminApp();
window.app = app;
app.init();
