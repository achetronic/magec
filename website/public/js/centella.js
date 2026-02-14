export class Centella {
  constructor(canvas) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    this._time = 0;
    this._particles = [];
    this._resizeObserver = null;
    this._raf = null;
    this._setup();
  }
  _setup() {
    const resize = () => {
      const r = this.canvas.getBoundingClientRect();
      this.canvas.width = r.width * devicePixelRatio;
      this.canvas.height = r.height * devicePixelRatio;
      this.ctx.scale(devicePixelRatio, devicePixelRatio);
    };
    resize();
    this._resizeObserver = new ResizeObserver(resize);
    this._resizeObserver.observe(this.canvas);
  }
  start() { const d = () => { this._raf = requestAnimationFrame(d); this._draw(); }; d(); }
  stop() { if (this._raf) cancelAnimationFrame(this._raf); if (this._resizeObserver) this._resizeObserver.disconnect(); }
  _draw() {
    const w = this.canvas.offsetWidth, h = this.canvas.offsetHeight;
    if (!w || !h) return;
    this._time += .006;
    const ctx = this.ctx, cx = w/2, cy = h/2, rad = Math.min(w,h)*.38;
    const c = {r:251,g:191,b:36}, ab = .6;
    ctx.clearRect(0,0,w,h);
    if (Math.random() < .2) {
      const a = Math.random()*Math.PI*2, d = Math.random()*rad*.7;
      this._particles.push({x:cx+Math.cos(a)*d,y:cy+Math.sin(a)*d,vx:(Math.random()-.5)*.05,vy:(Math.random()-.5)*.05,life:1,size:1+Math.random()*2,drift:Math.random()*Math.PI*2});
    }
    this._particles = this._particles.filter(p => {
      p.life -= .002; if (p.life <= 0) return false;
      p.drift += .003; p.x += p.vx+Math.cos(p.drift)*.05; p.y += p.vy+Math.sin(p.drift)*.05;
      const dx=p.x-cx, dy=p.y-cy, dist=Math.sqrt(dx*dx+dy*dy);
      if (dist > rad*.85) { p.x=cx+(dx/dist)*rad*.85; p.y=cy+(dy/dist)*rad*.85; }
      ctx.fillStyle = `rgba(${c.r},${c.g},${c.b},${p.life*.6*ab})`;
      ctx.beginPath(); ctx.arc(p.x,p.y,p.size*p.life,0,Math.PI*2); ctx.fill();
      return true;
    });
    for (let s=0;s<3;s++) {
      ctx.beginPath();
      const so=(s/3)*Math.PI*2, sr=rad*(.3+s*.15);
      for (let i=0;i<=60;i++) {
        const t=i/60, a=t*Math.PI*2+this._time*(1+s*.3)+so;
        const wb=Math.sin(t*Math.PI*4+this._time*2)*5, r=sr+wb;
        const x=cx+Math.cos(a)*r, y=cy+Math.sin(a)*r;
        i===0?ctx.moveTo(x,y):ctx.lineTo(x,y);
      }
      ctx.closePath(); ctx.strokeStyle=`rgba(${c.r},${c.g},${c.b},${.06*ab})`; ctx.lineWidth=1; ctx.stroke();
    }
    const cg=ctx.createRadialGradient(cx,cy,0,cx,cy,rad*.4);
    cg.addColorStop(0,`rgba(${c.r},${c.g},${c.b},${.12*ab})`); cg.addColorStop(1,`rgba(${c.r},${c.g},${c.b},0)`);
    ctx.fillStyle=cg; ctx.beginPath(); ctx.arc(cx,cy,rad*.4,0,Math.PI*2); ctx.fill();
    const og=ctx.createRadialGradient(cx,cy,rad*.9,cx,cy,rad*1.3);
    og.addColorStop(0,`rgba(${c.r},${c.g},${c.b},${.15*ab})`); og.addColorStop(1,`rgba(${c.r},${c.g},${c.b},0)`);
    ctx.fillStyle=og; ctx.beginPath(); ctx.arc(cx,cy,rad*1.3,0,Math.PI*2); ctx.fill();
    const br=Math.sin(this._time*1.5)*2;
    ctx.beginPath(); ctx.arc(cx,cy,rad+br,0,Math.PI*2);
    ctx.strokeStyle=`rgba(${c.r},${c.g},${c.b},.5)`; ctx.lineWidth=1.5; ctx.stroke();
  }
}
