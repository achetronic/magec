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
import { stripMetadata } from '../utils/metadata.js';

export class SessionService {
    constructor() {
        this.baseUrl = CONFIG.agent.baseUrl;
        this.appName = CONFIG.agent.appName;
        this.userId = CONFIG.agent.defaultUserId;
    }

    setAgent(agentId) {
        this.appName = agentId;
    }

    async listSessions() {
        const response = await errorHandler.wrapFetch(
            fetch(`${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions`),
            ErrorContext.SESSION_LIST,
            { logPrefix: 'SessionService.listSessions' }
        );
        
        if (!response) return [];
        
        try {
            const sessions = await response.json();
            return sessions || [];
        } catch (e) {
            errorHandler.handle(e, ErrorContext.SESSION_LIST, { logPrefix: 'SessionService.listSessions' });
            return [];
        }
    }

    async getSession(sessionId) {
        let response;
        try {
            response = await fetch(
                `${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions/${sessionId}`
            );
        } catch (e) {
            errorHandler.handle(e, ErrorContext.SESSION_LOAD, { logPrefix: 'SessionService.getSession' });
            return null;
        }
        
        if (!response.ok) {
            if (response.status === 404) {
                return null;
            }
            errorHandler.handle(response, ErrorContext.SESSION_LOAD, { logPrefix: 'SessionService.getSession' });
            return null;
        }
        
        try {
            return await response.json();
        } catch (e) {
            errorHandler.handle(e, ErrorContext.SESSION_LOAD, { logPrefix: 'SessionService.getSession' });
            return null;
        }
    }

    async deleteSession(sessionId) {
        const response = await errorHandler.wrapFetch(
            fetch(
                `${this.baseUrl}/apps/${this.appName}/users/${this.userId}/sessions/${sessionId}`,
                { method: 'DELETE' }
            ),
            ErrorContext.SESSION_DELETE,
            { logPrefix: 'SessionService.deleteSession' }
        );
        
        return response !== null;
    }

    extractMessages(session) {
        if (!session?.events) return [];
        
        const messages = [];
        for (const event of session.events) {
            if (event.content?.role && event.content?.parts?.[0]?.text) {
                messages.push({
                    role: event.content.role,
                    text: stripMetadata(event.content.parts[0].text)
                });
            }
        }
        return messages;
    }

    getSessionPreview(session) {
        const messages = this.extractMessages(session);
        if (messages.length === 0) return 'Empty conversation';
        
        const firstUserMessage = messages.find(m => m.role === 'user');
        if (firstUserMessage) {
            const text = firstUserMessage.text;
            return text.length > 50 ? text.substring(0, 50) + '...' : text;
        }
        
        return 'Conversation';
    }
}
