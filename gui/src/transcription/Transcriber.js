import { AudioConverter } from '../audio/AudioConverter.js';
import { errorHandler, ErrorContext } from '../errors/index.js';

export class TranscriberInterface {
    async transcribe(blob) {
        throw new Error('transcribe() must be implemented');
    }
}

export class LocalTranscriber extends TranscriberInterface {
    constructor(config) {
        super();
        this.config = config;
        this.transcriber = null;
        this.isLoaded = false;
        this.onProgress = null;
    }

    async load() {
        if (this.isLoaded) return true;
        
        const { pipeline } = await import('https://cdn.jsdelivr.net/npm/@xenova/transformers@2.17.2');
        
        this.transcriber = await pipeline(
            'automatic-speech-recognition',
            this.config.model,
            {
                progress_callback: (p) => {
                    if (p.status === 'progress' && p.progress && this.onProgress) {
                        this.onProgress(Math.round(p.progress));
                    }
                }
            }
        );
        
        this.isLoaded = true;
        return true;
    }

    async transcribe(blob, audioContext) {
        if (!this.isLoaded) {
            await this.load();
        }
        
        const audio = await AudioConverter.blobToFloat32Array(blob, audioContext);
        const result = await this.transcriber(audio, {
            language: this.config.language,
            task: this.config.task
        });
        
        return result?.text?.trim() || '';
    }
}

export class RemoteTranscriber extends TranscriberInterface {
    constructor(config) {
        super();
        this.config = config;
        this.saveAudio = false;
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
            response = await fetch(this.config.url, {
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
