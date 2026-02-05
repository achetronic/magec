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

/**
 * FeedbackSound - Generates pleasant audio feedback tones using Web Audio API
 * Used to provide audible confirmation when wake word is detected
 */
export class FeedbackSound {
    constructor(options = {}) {
        this.audioContext = null;
        this.volume = options.volume ?? 0.25;
        this.enabled = true;
    }

    _getContext() {
        if (!this.audioContext) {
            this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        }
        return this.audioContext;
    }

    /**
     * Celestial chime - ethereal, shimmering sound with detuned layers
     * Inspired by wind chimes and meditation bells
     */
    playWakeChime() {
        if (!this.enabled) return;

        const ctx = this._getContext();
        const now = ctx.currentTime;

        // Master chain with reverb
        const reverb = this._createLushReverb(ctx);
        const reverbGain = ctx.createGain();
        reverbGain.gain.value = 0.45;
        
        const dryGain = ctx.createGain();
        dryGain.gain.value = 0.55;

        const master = ctx.createGain();
        master.gain.value = this.volume;
        master.connect(ctx.destination);

        dryGain.connect(master);
        reverb.connect(reverbGain);
        reverbGain.connect(master);

        // Celestial chord: stacked fifths with shimmer (tighter timing)
        const notes = [
            { freq: 392.00, time: 0,    dur: 0.5 },   // G4 - root
            { freq: 587.33, time: 0.03, dur: 0.45 },  // D5 - fifth
            { freq: 783.99, time: 0.06, dur: 0.4 },   // G5 - octave
            { freq: 1174.66, time: 0.09, dur: 0.35 }, // D6 - high fifth
        ];

        notes.forEach(note => {
            this._playCelestialTone(ctx, note.freq, now + note.time, note.dur, dryGain, reverb);
        });

        // Add shimmer layer - very high, quiet, detuned tones
        this._playShimmer(ctx, 2349.32, now + 0.06, 0.35, dryGain, reverb); // D7
        this._playShimmer(ctx, 3135.96, now + 0.09, 0.3, dryGain, reverb);  // G7
    }

    /**
     * Soft descending tone - indicates recording stopped
     * Complementary to the ascending wake chime
     */
    playStopChime() {
        if (!this.enabled) return;

        const ctx = this._getContext();
        const now = ctx.currentTime;

        const reverb = this._createLushReverb(ctx);
        const reverbGain = ctx.createGain();
        reverbGain.gain.value = 0.4;
        
        const dryGain = ctx.createGain();
        dryGain.gain.value = 0.6;

        const master = ctx.createGain();
        master.gain.value = this.volume * 0.8;
        master.connect(ctx.destination);

        dryGain.connect(master);
        reverb.connect(reverbGain);
        reverbGain.connect(master);

        // Descending notes - gentle resolution
        const notes = [
            { freq: 783.99, time: 0,    dur: 0.25 },  // G5
            { freq: 587.33, time: 0.08, dur: 0.3 },   // D5
        ];

        notes.forEach(note => {
            this._playCelestialTone(ctx, note.freq, now + note.time, note.dur, dryGain, reverb);
        });
    }

    /**
     * Main tone with gentle vibrato and harmonics
     */
    _playCelestialTone(ctx, frequency, startTime, duration, dryNode, wetNode) {
        // Main tone with subtle vibrato
        const osc = ctx.createOscillator();
        osc.type = 'sine';
        osc.frequency.value = frequency;

        // LFO for gentle vibrato
        const vibrato = ctx.createOscillator();
        const vibratoGain = ctx.createGain();
        vibrato.frequency.value = 5.5; // Slightly faster vibrato for shorter sound
        vibratoGain.gain.value = frequency * 0.003;
        vibrato.connect(vibratoGain);
        vibratoGain.connect(osc.frequency);
        vibrato.start(startTime);
        vibrato.stop(startTime + duration + 0.3);

        // Soft harmonic (fifth above, very quiet)
        const harmonic = ctx.createOscillator();
        harmonic.type = 'sine';
        harmonic.frequency.value = frequency * 1.5;

        // Bell-like envelope - soft attack, long decay
        const gain = ctx.createGain();
        const harmonicGain = ctx.createGain();
        
        // Main tone envelope
        gain.gain.setValueAtTime(0, startTime);
        gain.gain.linearRampToValueAtTime(0.4, startTime + 0.04);
        gain.gain.exponentialRampToValueAtTime(0.12, startTime + duration * 0.35);
        gain.gain.exponentialRampToValueAtTime(0.001, startTime + duration);

        // Harmonic envelope (quieter, faster decay)
        harmonicGain.gain.setValueAtTime(0, startTime);
        harmonicGain.gain.linearRampToValueAtTime(0.08, startTime + 0.03);
        harmonicGain.gain.exponentialRampToValueAtTime(0.001, startTime + duration * 0.5);

        osc.connect(gain);
        harmonic.connect(harmonicGain);
        
        gain.connect(dryNode);
        gain.connect(wetNode);
        harmonicGain.connect(dryNode);
        harmonicGain.connect(wetNode);

        osc.start(startTime);
        osc.stop(startTime + duration + 0.1);
        harmonic.start(startTime);
        harmonic.stop(startTime + duration * 0.6);
    }

    /**
     * High shimmer tones - detuned pairs for ethereal effect
     */
    _playShimmer(ctx, frequency, startTime, duration, dryNode, wetNode) {
        // Two slightly detuned oscillators for chorus/shimmer effect
        [-3, 3].forEach(detuneCents => {
            const osc = ctx.createOscillator();
            osc.type = 'sine';
            osc.frequency.value = frequency;
            osc.detune.value = detuneCents;

            const gain = ctx.createGain();
            gain.gain.setValueAtTime(0, startTime);
            gain.gain.linearRampToValueAtTime(0.03, startTime + 0.05);
            gain.gain.exponentialRampToValueAtTime(0.001, startTime + duration);

            osc.connect(gain);
            gain.connect(dryNode);
            gain.connect(wetNode);

            osc.start(startTime);
            osc.stop(startTime + duration + 0.1);
        });
    }

    /**
     * Lush reverb for ethereal atmosphere (shorter tail)
     */
    _createLushReverb(ctx) {
        const convolver = ctx.createConvolver();
        const rate = ctx.sampleRate;
        const length = rate * 0.8; // 800ms tail - shorter but still lush
        const impulse = ctx.createBuffer(2, length, rate);
        
        for (let channel = 0; channel < 2; channel++) {
            const data = impulse.getChannelData(channel);
            for (let i = 0; i < length; i++) {
                const t = i / length;
                // Smooth exponential decay with modulation for richness
                const decay = Math.pow(1 - t, 2.0);
                const modulation = 1 + 0.1 * Math.sin(t * 50);
                data[i] = (Math.random() * 2 - 1) * decay * modulation * 0.5;
            }
        }
        
        convolver.buffer = impulse;
        return convolver;
    }

    setEnabled(enabled) {
        this.enabled = enabled;
    }

    setVolume(volume) {
        this.volume = Math.max(0, Math.min(1, volume));
    }
}
