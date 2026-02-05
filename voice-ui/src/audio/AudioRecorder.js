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

export class AudioRecorder {
    constructor(micStream) {
        this.micStream = micStream;
        this.mediaRecorder = null;
        this.recordedChunks = [];
        this.isRecording = false;
        this.onRecordingComplete = null;
    }

    start() {
        if (this.isRecording) return;
        
        this.isRecording = true;
        this.recordedChunks = [];
        
        this.mediaRecorder = new MediaRecorder(this.micStream);
        this.mediaRecorder.ondataavailable = (e) => {
            if (e.data.size > 0) {
                this.recordedChunks.push(e.data);
            }
        };
        this.mediaRecorder.onstop = () => this._processRecording();
        this.mediaRecorder.start(100);
    }

    stop() {
        if (!this.isRecording) return;
        
        this.isRecording = false;
        if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
            this.mediaRecorder.stop();
        }
    }

    getIsRecording() {
        return this.isRecording;
    }

    _processRecording() {
        const blob = new Blob(this.recordedChunks, { type: 'audio/webm' });
        if (blob.size >= 1000 && this.onRecordingComplete) {
            this.onRecordingComplete(blob);
        }
    }
}
