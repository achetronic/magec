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

export class WaveformRenderer {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.analyser = null;
        this.isRecording = false;
        this._resizeObserver = null;
        this._smoothedLevel = 0;
        this._awakeness = 0;
        this._smoothedRadius = 0;
        this._time = 0;
        this._particles = [];
        this._setupCanvas();
    }

    _setupCanvas() {
        const resize = () => {
            const rect = this.canvas.getBoundingClientRect();
            this.canvas.width = rect.width * devicePixelRatio;
            this.canvas.height = rect.height * devicePixelRatio;
            this.ctx.scale(devicePixelRatio, devicePixelRatio);
        };
        
        resize();
        this._resizeObserver = new ResizeObserver(resize);
        this._resizeObserver.observe(this.canvas);
    }

    setAnalyser(analyser) {
        this.analyser = analyser;
    }

    setRecording(isRecording) {
        this.isRecording = isRecording;
    }

    start() {
        this._draw();
    }

    _draw() {
        requestAnimationFrame(() => this._draw());
        
        if (!this.analyser) return;
        
        const w = this.canvas.offsetWidth;
        const h = this.canvas.offsetHeight;
        
        // Skip drawing if canvas is not visible
        if (w === 0 || h === 0) return;
        
        this._time += 0.008;
        const centerX = w / 2;
        const centerY = h / 2;
        
        // Limit Magec size to prevent it from becoming too large
        const maxDimension = Math.min(w, h, 384);
        const baseRadius = maxDimension * 0.25;
        const maxRadius = maxDimension * 0.44;
        
        const data = new Uint8Array(this.analyser.frequencyBinCount);
        this.analyser.getByteFrequencyData(data);
        
        // Calculate overall audio level
        let sum = 0;
        const relevantBins = Math.min(48, data.length);
        for (let i = 0; i < relevantBins; i++) {
            const weight = 1 + (i < 16 ? 0.5 : 0);
            sum += data[i] * data[i] * weight;
        }
        const raw = Math.sqrt(sum / relevantBins) / 255;
        const target = Math.pow(raw, 0.7) * 1.4;
        
        this._smoothedLevel += (target - this._smoothedLevel) * 0.25;
        const level = Math.min(1, this._smoothedLevel);
        
        const awakeTarget = this.isRecording ? 1 : 0;
        const awakeSpeed = this.isRecording ? 0.15 : 0.03;
        this._awakeness += (awakeTarget - this._awakeness) * awakeSpeed;
        const awake = this._awakeness;
        
        const dormantColor = { r: 251, g: 191, b: 36 };
        const awakeColor = { r: 239, g: 68, b: 68 };
        
        const color = {
            r: Math.round(dormantColor.r + (awakeColor.r - dormantColor.r) * awake),
            g: Math.round(dormantColor.g + (awakeColor.g - dormantColor.g) * awake),
            b: Math.round(dormantColor.b + (awakeColor.b - dormantColor.b) * awake)
        };
        
        const alphaBoost = 0.6 + awake * 0.4;
        
        this.ctx.clearRect(0, 0, w, h);
        
        // Target radius based on state and audio level
        // In standby (awake < 0.5), keep stable size; when awake, react to audio
        const ambientLevel = awake < 0.5 ? 0 : level;
        const targetRadius = baseRadius + ambientLevel * (maxRadius - baseRadius);
        
        // Smooth radius transition
        if (this._smoothedRadius === 0) this._smoothedRadius = baseRadius;
        this._smoothedRadius += (targetRadius - this._smoothedRadius) * 0.08;
        const radius = this._smoothedRadius;
        
        // === PARTICLES ===
        // Spawn rate and speed based on awakeness
        const particleSpeed = 0.05 + awake * 0.15;
        const particleDrift = 0.003 + awake * 0.007;
        const particleDecay = 0.002 + awake * 0.006;
        // Base spawn rate ensures particles even in standby; more when awake/loud
        const spawnChance = 0.25 + awake * 0.4 + level * 0.3;
        
        if (Math.random() < spawnChance) {
            const angle = Math.random() * Math.PI * 2;
            const dist = Math.random() * radius * 0.7;
            this._particles.push({
                x: centerX + Math.cos(angle) * dist,
                y: centerY + Math.sin(angle) * dist,
                vx: (Math.random() - 0.5) * particleSpeed,
                vy: (Math.random() - 0.5) * particleSpeed,
                life: 1,
                size: 1 + Math.random() * 2,
                drift: Math.random() * Math.PI * 2
            });
        }
        
        this._particles = this._particles.filter(p => {
            p.life -= particleDecay;
            if (p.life <= 0) return false;
            
            p.drift += particleDrift;
            p.x += p.vx + Math.cos(p.drift) * particleSpeed;
            p.y += p.vy + Math.sin(p.drift) * particleSpeed;
            
            const dx = p.x - centerX;
            const dy = p.y - centerY;
            const dist = Math.sqrt(dx * dx + dy * dy);
            if (dist > radius * 0.85) {
                p.x = centerX + (dx / dist) * radius * 0.85;
                p.y = centerY + (dy / dist) * radius * 0.85;
            }
            
            const alpha = p.life * 0.6 * alphaBoost;
            this.ctx.fillStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${alpha})`;
            
            this.ctx.beginPath();
            this.ctx.arc(p.x, p.y, p.size * p.life, 0, Math.PI * 2);
            this.ctx.fill();
            
            return true;
        });
        
        // === SWIRLS ===
        const numSwirls = 3;
        for (let s = 0; s < numSwirls; s++) {
            this.ctx.beginPath();
            const swirlOffset = (s / numSwirls) * Math.PI * 2;
            const swirlRadius = radius * (0.3 + s * 0.15);
            
            for (let i = 0; i <= 60; i++) {
                const t = i / 60;
                const angle = t * Math.PI * 2 + this._time * (1 + s * 0.3) + swirlOffset;
                const wobble = Math.sin(t * Math.PI * 4 + this._time * 2) * (5 + level * 10);
                const r = swirlRadius + wobble;
                
                const x = centerX + Math.cos(angle) * r;
                const y = centerY + Math.sin(angle) * r;
                
                if (i === 0) {
                    this.ctx.moveTo(x, y);
                } else {
                    this.ctx.lineTo(x, y);
                }
            }
            
            this.ctx.closePath();
            const swirlAlpha = (0.06 + level * 0.08) * alphaBoost;
            this.ctx.strokeStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${swirlAlpha})`;
            this.ctx.lineWidth = 1;
            this.ctx.stroke();
        }
        
        // Core glow
        const coreGradient = this.ctx.createRadialGradient(
            centerX, centerY, 0,
            centerX, centerY, radius * 0.4
        );
        const coreAlpha = 0.12 + level * 0.08;
        coreGradient.addColorStop(0, `rgba(${color.r}, ${color.g}, ${color.b}, ${coreAlpha * alphaBoost})`);
        coreGradient.addColorStop(1, `rgba(${color.r}, ${color.g}, ${color.b}, 0)`);
        this.ctx.fillStyle = coreGradient;
        this.ctx.beginPath();
        this.ctx.arc(centerX, centerY, radius * 0.4, 0, Math.PI * 2);
        this.ctx.fill();
        
        // Outer glow
        const glowGradient = this.ctx.createRadialGradient(
            centerX, centerY, radius * 0.9,
            centerX, centerY, radius * 1.3
        );
        const glowAlpha = 0.15 * alphaBoost;
        glowGradient.addColorStop(0, `rgba(${color.r}, ${color.g}, ${color.b}, ${glowAlpha})`);
        glowGradient.addColorStop(1, `rgba(${color.r}, ${color.g}, ${color.b}, 0)`);
        this.ctx.fillStyle = glowGradient;
        this.ctx.beginPath();
        this.ctx.arc(centerX, centerY, radius * 1.3, 0, Math.PI * 2);
        this.ctx.fill();
        
        // Main circle
        const breathe = Math.sin(this._time * 1.5) * 2;
        this.ctx.beginPath();
        this.ctx.arc(centerX, centerY, radius + breathe, 0, Math.PI * 2);
        const strokeAlpha = 0.5 + awake * 0.35;
        this.ctx.strokeStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${strokeAlpha})`;
        this.ctx.lineWidth = 2;
        this.ctx.stroke();
    }
}
