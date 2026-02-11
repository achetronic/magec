/*
 * Copyright 2025 Alby Hernández
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { CONFIG } from '../config.js';
import { errorHandler, ErrorContext } from '../errors/index.js';

export class AgentClient {
    constructor() {
        this.baseUrl = CONFIG.agent.baseUrl;
        this.appName = CONFIG.agent.appName;
        this.userId = CONFIG.agent.defaultUserId;
    }

    setAgent(agentId) {
        this.appName = agentId;
    }

    async createSession(sessionId) {
        const response = await errorHandler.wrapFetch(
            fetch(`${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions/${sessionId}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({})
            }),
            ErrorContext.SESSION_CREATE,
            { logPrefix: 'AgentClient.createSession' }
        );
        return response !== null;
    }

    async ensureSession(sessionId) {
        const exists = await this.sessionExists(sessionId);
        if (!exists) {
            return this.createSession(sessionId);
        }
        return true;
    }

    async sessionExists(sessionId) {
        try {
            const response = await fetch(`${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions/${sessionId}`);
            return response.ok;
        } catch (e) {
            return false;
        }
    }

    async sendMessage(sessionId, message) {
        await this.ensureSession(sessionId);
        
        let response;
        try {
            response = await fetch(`${this.baseUrl}/run`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    appName: this.appName,
                    userId: this.userId,
                    sessionId: sessionId,
                    newMessage: {
                        role: 'user',
                        parts: [{ text: message }]
                    }
                })
            });
        } catch (e) {
            errorHandler.handle(e, ErrorContext.AGENT_SEND, { logPrefix: 'AgentClient.sendMessage' });
            throw e;
        }

        if (!response.ok) {
            const errorText = await response.text().catch(() => '');
            const error = new Error(errorText || `Server error: ${response.status}`);
            error.status = response.status;
            errorHandler.handle(error, ErrorContext.AGENT_RESPONSE, { logPrefix: 'AgentClient.sendMessage' });
            throw error;
        }

        try {
            return this._extractResponses(await response.json());
        } catch (e) {
            errorHandler.handle(e, ErrorContext.AGENT_RESPONSE, { logPrefix: 'AgentClient.sendMessage' });
            throw e;
        }
    }

    _extractResponses(result) {
        const responses = [];
        
        if (Array.isArray(result)) {
            for (const event of result) {
                if (event.content?.parts?.[0]?.text) {
                    responses.push(event.content.parts[0].text);
                }
            }
        } else if (result.content?.parts?.[0]?.text) {
            responses.push(result.content.parts[0].text);
        }
        
        return responses;
    }
}
