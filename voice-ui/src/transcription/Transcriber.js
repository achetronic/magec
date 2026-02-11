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

import { AudioConverter } from '../audio/AudioConverter.js';
import { errorHandler, ErrorContext } from '../errors/index.js';

export class RemoteTranscriber {
    constructor(config) {
        this.config = config;
        this.saveAudio = false;
        this._agentId = 'default';
    }

    setAgent(agentId) {
        this._agentId = agentId;
    }

    _transcriptionUrl() {
        return `/api/v1/voice/${this._agentId}/transcription`;
    }

    setSaveAudio(save) {
        this.saveAudio = save;
    }

    async transcribe(blob) {
        const wavBlob = await AudioConverter.blobToWav(blob);
        
        if (this.saveAudio) {
            AudioConverter.downloadWav(wavBlob);
        }
        
        const formData = new FormData();
        formData.append('file', wavBlob, 'audio.wav');
        formData.append('model', this.config.model);
        formData.append('language', 'es');

        let response;
        try {
            response = await fetch(this._transcriptionUrl(), {
                method: 'POST',
                body: formData
            });
        } catch (e) {
            errorHandler.handle(e, ErrorContext.TRANSCRIPTION_SERVER, { logPrefix: 'RemoteTranscriber.transcribe' });
            throw e;
        }

        if (!response.ok) {
            const error = new Error(`Server error: ${response.status}`);
            error.status = response.status;
            errorHandler.handle(error, ErrorContext.TRANSCRIPTION_SERVER, { logPrefix: 'RemoteTranscriber.transcribe' });
            throw error;
        }

        try {
            const result = await response.json();
            return result.text?.trim() || '';
        } catch (e) {
            errorHandler.handle(e, ErrorContext.TRANSCRIPTION, { logPrefix: 'RemoteTranscriber.transcribe' });
            throw e;
        }
    }
}
