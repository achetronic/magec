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

export class AudioConverter {
    static async blobToWav(blob, targetSampleRate = 16000) {
        const arrayBuffer = await blob.arrayBuffer();
        const audioCtx = new AudioContext({ sampleRate: targetSampleRate });
        const audioBuffer = await audioCtx.decodeAudioData(arrayBuffer);
        
        let samples = audioBuffer.getChannelData(0);
        if (audioBuffer.sampleRate !== targetSampleRate) {
            samples = this._resample(samples, audioBuffer.sampleRate, targetSampleRate);
        }
        
        const pcm = this._toPCM16(samples);
        const wavBlob = this._createWavBlob(pcm, targetSampleRate);
        
        audioCtx.close();
        return wavBlob;
    }

    static _resample(samples, fromRate, toRate) {
        const ratio = fromRate / toRate;
        const newLen = Math.round(samples.length / ratio);
        const resampled = new Float32Array(newLen);
        
        for (let i = 0; i < newLen; i++) {
            const idx = i * ratio;
            const lo = Math.floor(idx);
            const hi = Math.min(lo + 1, samples.length - 1);
            resampled[i] = samples[lo] * (1 - (idx - lo)) + samples[hi] * (idx - lo);
        }
        
        return resampled;
    }

    static _toPCM16(samples) {
        const pcm = new Int16Array(samples.length);
        for (let i = 0; i < samples.length; i++) {
            const s = Math.max(-1, Math.min(1, samples[i]));
            pcm[i] = s < 0 ? s * 0x8000 : s * 0x7FFF;
        }
        return pcm;
    }

    static _createWavBlob(pcm, sampleRate) {
        const wavBuffer = new ArrayBuffer(44 + pcm.length * 2);
        const view = new DataView(wavBuffer);
        
        const writeString = (offset, str) => {
            for (let i = 0; i < str.length; i++) {
                view.setUint8(offset + i, str.charCodeAt(i));
            }
        };
        
        writeString(0, 'RIFF');
        view.setUint32(4, 36 + pcm.length * 2, true);
        writeString(8, 'WAVE');
        writeString(12, 'fmt ');
        view.setUint32(16, 16, true);
        view.setUint16(20, 1, true);
        view.setUint16(22, 1, true);
        view.setUint32(24, sampleRate, true);
        view.setUint32(28, sampleRate * 2, true);
        view.setUint16(32, 2, true);
        view.setUint16(34, 16, true);
        writeString(36, 'data');
        view.setUint32(40, pcm.length * 2, true);
        
        const wavBytes = new Uint8Array(wavBuffer);
        wavBytes.set(new Uint8Array(pcm.buffer), 44);
        
        return new Blob([wavBytes], { type: 'audio/wav' });
    }

    static downloadWav(wavBlob, prefix = 'audio') {
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const filename = `test/${prefix}_${timestamp}.wav`;
        const url = URL.createObjectURL(wavBlob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        a.click();
        URL.revokeObjectURL(url);
        console.log(`Saved: ${filename}`);
    }
}
