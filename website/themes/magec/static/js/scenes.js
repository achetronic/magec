// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

// Scroll-driven feature scenes. Each [data-scene] section is a tall block with
// a sticky viewport inside; the scroll progress through the section drives
// SVG wire drawing ([data-wire="start,end"]), element reveals
// ([data-step="at"]) and pulses travelling along paths ([data-travel]).
//
// Lazy-viewer fallback: once a scene's sticky viewport is engaged, an
// internal clock also advances its progress (AUTOPLAY_MS for the full
// story). Whichever is further ahead, scroll or clock, wins.
//
// Dead-scroll skip: once a scene is fully played, wheeling down jumps
// straight to the next scene instead of grinding through the leftover
// scroll budget (and wheeling up jumps back to the scene start).
// Clicking the bobbing hint ball also advances to the next scene.
// No libraries: one rAF, passive listeners except the wheel hijack.

const AUTOPLAY_MS = 6000;
const DONE_AT = 0.85;

const clamp = (v, a, b) => Math.min(b, Math.max(a, v));
const seg = (p, a, b) => clamp((p - a) / (b - a), 0, 1);
const ease = t => (t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2);

export function initScenes() {
  const sections = [...document.querySelectorAll('[data-scene]')];
  if (!sections.length) return;

  const scenes = sections.map(scene => {
    const wires = [...scene.querySelectorAll('[data-wire]')].map(path => {
      // Faint ghost of the full route: hints that the circuit is incomplete
      // and invites scrolling to draw it.
      const ghost = path.cloneNode();
      ghost.removeAttribute('data-wire');
      ghost.classList.add('wire-ghost');
      path.parentNode.insertBefore(ghost, path);

      const len = path.getTotalLength() || 1;
      path.style.strokeDasharray = `${len}`;
      path.style.strokeDashoffset = `${len}`;
      const [a, b] = path.dataset.wire.split(',').map(Number);
      return { path, len, a, b };
    });
    const steps = [...scene.querySelectorAll('[data-step]')].map(el => ({
      el,
      at: Number(el.dataset.step),
    }));
    const travellers = [...scene.querySelectorAll('[data-travel]')].map(el => {
      const [id, a, b] = el.dataset.travel.split(',');
      const path = scene.querySelector(`#${id}`);
      return { el, path, len: path.getTotalLength() || 1, a: Number(a), b: Number(b) };
    });
    return { scene, wires, steps, travellers, auto: 0, done: false };
  });

  const finish = ({ scene, wires, steps, travellers }) => {
    wires.forEach(w => { w.path.style.strokeDashoffset = '0'; });
    steps.forEach(s => s.el.classList.add('on'));
    travellers.forEach(t => { t.el.style.opacity = '0'; });
    scene.classList.add('scene--done', 'scene--started');
  };

  if (matchMedia('(prefers-reduced-motion: reduce)').matches) {
    scenes.forEach(finish);
    return;
  }

  let rafId = 0;
  let lastT = 0;
  let jumping = false;

  const render = (now) => {
    rafId = 0;
    const dt = lastT ? Math.min(now - lastT, 100) : 0;
    lastT = now;
    const vh = innerHeight;
    let animating = false;

    scenes.forEach(sc => {
      const { scene, wires, steps, travellers } = sc;
      const rect = scene.getBoundingClientRect();
      if (rect.bottom < 0 || rect.top > vh) {
        if (rect.top > vh) { sc.auto = 0; sc.done = false; } // below viewport: rearm
        return;
      }
      const total = rect.height - vh;
      const scrollP = total > 0 ? clamp(-rect.top / total, 0, 1) : 1;

      // Sticky engaged: the clock starts ticking for lazy viewers.
      if (rect.top <= 1 && sc.auto < 1) {
        sc.auto = clamp(sc.auto + dt / AUTOPLAY_MS, 0, 1);
        if (sc.auto < 1) animating = true;
      }
      const p = Math.max(scrollP, sc.auto);
      sc.done = p >= DONE_AT;

      wires.forEach(w => {
        w.path.style.strokeDashoffset = `${w.len * (1 - ease(seg(p, w.a, w.b)))}`;
      });
      travellers.forEach(t => {
        const tp = seg(p, t.a, t.b);
        const active = tp > 0 && tp < 1;
        t.el.style.opacity = active ? '1' : '0';
        if (active) {
          const pt = t.path.getPointAtLength(t.len * tp);
          t.el.style.transform = `translate(${pt.x}px, ${pt.y}px)`;
        }
      });
      steps.forEach(s => s.el.classList.toggle('on', p >= s.at));
      scene.classList.toggle('scene--started', p > 0.04);
      scene.classList.toggle('scene--done', sc.done);
    });

    if (animating) schedule();
  };

  const schedule = () => {
    if (!rafId) rafId = requestAnimationFrame(render);
  };

  const jumpTo = (top) => {
    jumping = true;
    scrollTo({ top, behavior: 'smooth' });
    setTimeout(() => { jumping = false; }, 700);
  };

  // The scroll offset where a scene's sticky releases (= end of its budget).
  const sceneExit = sc => sc.scene.offsetTop + sc.scene.offsetHeight - innerHeight;

  // Landing spot for "next": the next scene's own top, where its sticky is
  // already engaged and the autoplay clock starts on its own. sceneExit would
  // land one viewport too early (next scene still below the fold).
  const nextSceneTop = sc => {
    const i = scenes.indexOf(sc);
    const next = scenes[i + 1];
    return next ? next.scene.offsetTop + 2 : sceneExit(sc);
  };

  // Wheel hijack: skip dead scroll once the story has been told.
  addEventListener('wheel', e => {
    if (jumping) { e.preventDefault(); return; }
    const vh = innerHeight;
    const sc = scenes.find(s => {
      const r = s.scene.getBoundingClientRect();
      return r.top <= 0 && r.bottom >= vh;
    });
    if (!sc || !sc.done) return;
    if (e.deltaY > 0) {
      const remaining = sceneExit(sc) - scrollY;
      if (remaining > 40) {
        e.preventDefault();
        jumpTo(nextSceneTop(sc));
      }
    } else if (e.deltaY < 0) {
      const back = scrollY - sc.scene.offsetTop;
      if (back > 40) {
        e.preventDefault();
        jumpTo(sc.scene.offsetTop + 2);
      }
    }
  }, { passive: false });

  // Hint ball click: advance to the next scene in one go.
  scenes.forEach(sc => {
    const hint = sc.scene.querySelector('.scene__hint');
    if (hint) hint.addEventListener('click', () => jumpTo(nextSceneTop(sc)));
  });

  addEventListener('scroll', schedule, { passive: true });
  addEventListener('resize', schedule, { passive: true });
  schedule();
}
