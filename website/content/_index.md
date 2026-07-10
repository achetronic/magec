---
title: "Magec: Self-hosted Multi-Agent AI Platform"
---

<!-- HERO -->
<section class="hero">
  <canvas id="hero-orb" class="hero__orb"></canvas>
  <div class="hero__content">
    <h1 class="hero__title">Build AI agents.<br><span>Make them work together.</span></h1>
    <p class="hero__subtitle">Create agents with their own brain, memory, voice and tools. Chain them into teams. Connect any agent or entire team to Telegram, voice, webhooks, cron, or all at once. Everything runs on your server.</p>
    <div class="hero__actions">
      <a href="docs/getting-started/" class="btn btn--primary">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
        Get Started
      </a>
      <a href="docs/" class="btn btn--ghost">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
        Documentation
      </a>
    </div>
    <div class="hero__terminal">
      <div class="terminal">
        <div class="terminal__bar"><span class="terminal__dot"></span><span class="terminal__dot"></span><span class="terminal__dot"></span></div>
        <div class="terminal__body">
          <span class="terminal__comment"># One command. Fully local. No API keys.</span><br>
          <span class="terminal__prompt">$</span> curl -fsSL <span class="terminal__url">https://magec.dev/install</span> | bash<br><br>
          <span class="terminal__comment"># Admin panel → <span class="terminal__url">localhost:8081</span></span><br>
          <span class="terminal__comment"># Voice interface → <span class="terminal__url">localhost:8080</span></span>
        </div>
      </div>
    </div>
  </div>
  <a href="#agents" class="scene__hint hero__hint" aria-label="Start the tour">
    <svg class="scene__hint-ball" viewBox="0 0 64 64" fill="none"><defs><linearGradient id="hintg0" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#f59e0b"/><stop offset="100%" stop-color="#f87171"/></linearGradient></defs><circle cx="32" cy="32" r="26" fill="url(#hintg0)"/><circle cx="20" cy="24" r="2" fill="white" opacity=".9"/><circle cx="26" cy="18" r="1.5" fill="white" opacity=".7"/><circle cx="40" cy="22" r="1.8" fill="white" opacity=".8"/><circle cx="38" cy="38" r="2" fill="white" opacity=".75"/><circle cx="24" cy="40" r="1.4" fill="white" opacity=".65"/></svg>
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
  </a>
</section>

<!-- SCENE: AGENTS -->
<section id="agents" class="scene" data-scene>
  <div class="scene__sticky">
    <div class="scene__head" data-step="0">
      <span class="scene__label ec-agent">Agents</span>
      <h2 class="scene__title">One agent, everything it needs</h2>
      <p class="scene__desc">Pick its brain. Plug in tools, skills and memory. Each agent is an independent unit, assembled in minutes.</p>
    </div>
    <div class="scene__fit" style="--mh:580">
      <div class="scene__stage">
        <svg class="scene__wires" viewBox="0 0 1040 560" aria-hidden="true">
          <path class="ec-backend" data-wire="0.10,0.24" d="M 285 122 C 345 122, 345 238, 394 238"/>
          <path class="ec-mcp" data-wire="0.26,0.40" d="M 285 272 C 340 272, 345 270, 394 270"/>
          <path class="ec-skill" data-wire="0.42,0.56" d="M 285 422 C 345 422, 345 302, 394 302"/>
          <path class="ec-memory" data-wire="0.58,0.72" d="M 778 197 C 725 197, 715 246, 652 246"/>
          <path class="ec-extra" data-wire="0.74,0.88" d="M 778 357 C 725 357, 715 286, 652 286"/>
        </svg>
        <div class="scene__end ec-backend" style="left:120px; top:92px" data-step="0.08">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="5" y="5" width="14" height="14" rx="2"/><rect x="10" y="10" width="4" height="4"/><line x1="9" y1="2" x2="9" y2="5"/><line x1="15" y1="2" x2="15" y2="5"/><line x1="9" y1="19" x2="9" y2="22"/><line x1="15" y1="19" x2="15" y2="22"/><line x1="2" y1="9" x2="5" y2="9"/><line x1="2" y1="15" x2="5" y2="15"/><line x1="19" y1="9" x2="22" y2="9"/><line x1="19" y1="15" x2="22" y2="15"/></svg>
          <span class="scene__tag">all AI models</span>
        </div>
        <div class="scene__end ec-mcp" style="left:120px; top:242px" data-step="0.24">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="9" y1="2" x2="9" y2="8"/><line x1="15" y1="2" x2="15" y2="8"/><path d="M6 8h12v4a6 6 0 0 1-6 6 6 6 0 0 1-6-6z"/><line x1="12" y1="18" x2="12" y2="22"/></svg>
          <span class="scene__tag">mcps</span>
        </div>
        <div class="scene__end ec-skill" style="left:120px; top:392px" data-step="0.40">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
          <span class="scene__tag">skills</span>
        </div>
        <div class="scene__end ec-memory" style="left:760px; top:167px" data-step="0.56">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/></svg>
          <span class="scene__tag">semantic memory</span>
        </div>
        <div class="scene__end ec-extra" style="left:760px; top:327px" data-step="0.72">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span class="scene__tag">and more</span>
        </div>
        <div class="fnode fnode--agent" style="left:400px; top:210px;" data-step="0">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div>
            <span class="fnode__label">AGENT</span>
            <span class="fnode__id">writer_1</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__skel" style="width:85%"></div>
            <div class="fnode__skel" style="width:65%"></div>
            <div class="fnode__skel" style="width:74%"></div>
          </div>
          <span class="fnode__port" style="left:-6px; top:22px"></span>
          <span class="fnode__port" style="left:-6px; top:54px"></span>
          <span class="fnode__port" style="left:-6px; top:86px"></span>
          <span class="fnode__port" style="right:-6px; top:30px"></span>
          <span class="fnode__port" style="right:-6px; top:70px"></span>
        </div>
      </div>
      <div class="scene__stage scene__stage--mobile" style="height:580px"><svg class="scene__wires" viewBox="0 0 360 580" aria-hidden="true"><path class="ec-backend" data-wire="0.10,0.24" d="M 90 66 C 90 120, 130 130, 130 180"/><path class="ec-mcp" data-wire="0.26,0.40" d="M 90 356 C 90 320, 130 316, 130 298"/><path class="ec-skill" data-wire="0.42,0.56" d="M 270 356 C 270 320, 230 316, 230 298"/><path class="ec-memory" data-wire="0.58,0.72" d="M 270 66 C 270 120, 230 130, 230 180"/><path class="ec-extra" data-wire="0.74,0.88" d="M 180 470 C 180 420, 180 350, 180 298"/></svg><div class="scene__end ec-backend" style="left:10px; top:0px" data-step="0.08"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="5" y="5" width="14" height="14" rx="2"/><rect x="10" y="10" width="4" height="4"/><line x1="9" y1="2" x2="9" y2="5"/><line x1="15" y1="2" x2="15" y2="5"/><line x1="9" y1="19" x2="9" y2="22"/><line x1="15" y1="19" x2="15" y2="22"/><line x1="2" y1="9" x2="5" y2="9"/><line x1="2" y1="15" x2="5" y2="15"/><line x1="19" y1="9" x2="22" y2="9"/><line x1="19" y1="15" x2="22" y2="15"/></svg><span class="scene__tag">all AI models</span></div><div class="scene__end ec-memory" style="left:190px; top:0px" data-step="0.56"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/></svg><span class="scene__tag">semantic memory</span></div><div class="scene__end ec-mcp" style="left:10px; top:360px" data-step="0.24"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="9" y1="2" x2="9" y2="8"/><line x1="15" y1="2" x2="15" y2="8"/><path d="M6 8h12v4a6 6 0 0 1-6 6 6 6 0 0 1-6-6z"/><line x1="12" y1="18" x2="12" y2="22"/></svg><span class="scene__tag">mcps</span></div><div class="scene__end ec-skill" style="left:190px; top:360px" data-step="0.40"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg><span class="scene__tag">skills</span></div><div class="scene__end ec-extra" style="left:100px; top:474px" data-step="0.72"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg><span class="scene__tag">and more</span></div><div class="fnode fnode--agent" style="left:60px; top:180px;" data-step="0"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div><span class="fnode__label">AGENT</span><span class="fnode__id">writer_1</span></div><div class="fnode__body"><div class="fnode__skel" style="width:85%"></div><div class="fnode__skel" style="width:65%"></div><div class="fnode__skel" style="width:74%"></div></div></div></div>
    </div>
    <div class="scene__hint">
      <svg class="scene__hint-ball" viewBox="0 0 64 64" fill="none"><defs><linearGradient id="hintg1" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#f59e0b"/><stop offset="100%" stop-color="#f87171"/></linearGradient></defs><circle cx="32" cy="32" r="26" fill="url(#hintg1)"/><circle cx="20" cy="24" r="2" fill="white" opacity=".9"/><circle cx="26" cy="18" r="1.5" fill="white" opacity=".7"/><circle cx="40" cy="22" r="1.8" fill="white" opacity=".8"/><circle cx="38" cy="38" r="2" fill="white" opacity=".75"/><circle cx="24" cy="40" r="1.4" fill="white" opacity=".65"/></svg>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
  </div>
</section>

<!-- SCENE: TOOLS AND SKILLS -->
<section id="tools" class="scene" data-scene>
  <div class="scene__sticky">
    <div class="scene__head" data-step="0">
      <span class="scene__label ec-mcp">MCPs &amp; Skills</span>
      <h2 class="scene__title">Real tools, learned know-how</h2>
      <p class="scene__desc">Through MCP your agents operate real systems. Skills are playbooks they load to do the job the way you want.</p>
    </div>
    <div class="scene__fit" style="--mh:620">
      <div class="scene__stage">
        <svg class="scene__wires" viewBox="0 0 1040 560" aria-hidden="true">
          <path id="w-mcp" class="ec-mcp" data-wire="0.10,0.24" d="M 334 240 C 490 240, 540 160, 696 160"/>
          <path id="w-skill" class="ec-skill" data-wire="0.58,0.72" d="M 334 280 C 490 280, 540 420, 696 420"/>
        </svg>
        <div class="fnode fnode--agent" style="left:90px; top:210px;" data-step="0">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div>
            <span class="fnode__label">AGENT</span>
            <span class="fnode__id">writer_1</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__skel" style="width:85%"></div>
            <div class="fnode__skel" style="width:65%"></div>
          </div>
          <span class="fnode__port" style="right:-6px; top:24px"></span>
          <span class="fnode__port" style="right:-6px; top:64px"></span>
        </div>
        <div class="fnode fnode--mcp" style="left:700px; top:100px;" data-step="0.06">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="9" y1="2" x2="9" y2="8"/><line x1="15" y1="2" x2="15" y2="8"/><path d="M6 8h12v4a6 6 0 0 1-6 6 6 6 0 0 1-6-6z"/><line x1="12" y1="18" x2="12" y2="22"/></svg></div>
            <span class="fnode__label">MCP</span>
            <span class="fnode__id">home_assistant</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__line" data-step="0.28">turn_on_lights</div>
            <div class="fnode__line" data-step="0.34">set_temperature</div>
            <div class="fnode__line" data-step="0.40">snapshot_camera <em>+38</em></div>
          </div>
          <span class="fnode__port" style="left:-6px; top:54px"></span>
        </div>
        <div class="fnode fnode--skill" style="left:700px; top:360px;" data-step="0.54">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg></div>
            <span class="fnode__label">SKILL</span>
            <span class="fnode__id">daily_report</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__line" data-step="0.74">1. gather metrics</div>
            <div class="fnode__line" data-step="0.79">2. compare trends</div>
            <div class="fnode__line" data-step="0.84">3. draft summary</div>
          </div>
          <span class="fnode__port" style="left:-6px; top:54px"></span>
        </div>
        <span class="scene__pulse ec-mcp" data-travel="w-mcp,0.42,0.52"></span>
        <span class="scene__pulse ec-skill" data-travel="w-skill,0.86,0.96"></span>
      </div>
      <div class="scene__stage scene__stage--mobile" style="height:620px"><svg class="scene__wires" viewBox="0 0 360 620" aria-hidden="true"><path id="m-w-mcp" class="ec-mcp" data-wire="0.10,0.24" d="M 60 70 C 16 110, 16 200, 60 240"/><path id="m-w-skill" class="ec-skill" data-wire="0.58,0.72" d="M 300 70 C 346 140, 346 400, 300 520"/></svg><div class="fnode fnode--agent" style="left:60px; top:0px;" data-step="0"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div><span class="fnode__label">AGENT</span><span class="fnode__id">writer_1</span></div><div class="fnode__body"><div class="fnode__skel" style="width:85%"></div><div class="fnode__skel" style="width:65%"></div></div></div><div class="fnode fnode--mcp" style="left:60px; top:190px;" data-step="0.06"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="9" y1="2" x2="9" y2="8"/><line x1="15" y1="2" x2="15" y2="8"/><path d="M6 8h12v4a6 6 0 0 1-6 6 6 6 0 0 1-6-6z"/><line x1="12" y1="18" x2="12" y2="22"/></svg></div><span class="fnode__label">MCP</span><span class="fnode__id">home_assistant</span></div><div class="fnode__body"><div class="fnode__line" data-step="0.28">turn_on_lights</div><div class="fnode__line" data-step="0.34">set_temperature</div><div class="fnode__line" data-step="0.40">snapshot_camera <em>+38</em></div></div></div><div class="fnode fnode--skill" style="left:60px; top:472px;" data-step="0.54"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg></div><span class="fnode__label">SKILL</span><span class="fnode__id">daily_report</span></div><div class="fnode__body"><div class="fnode__line" data-step="0.74">1. gather metrics</div><div class="fnode__line" data-step="0.79">2. compare trends</div><div class="fnode__line" data-step="0.84">3. draft summary</div></div></div><span class="scene__pulse ec-mcp" data-travel="m-w-mcp,0.42,0.52"></span><span class="scene__pulse ec-skill" data-travel="m-w-skill,0.86,0.96"></span></div>
    </div>
    <div class="scene__hint">
      <svg class="scene__hint-ball" viewBox="0 0 64 64" fill="none"><defs><linearGradient id="hintg2" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#f59e0b"/><stop offset="100%" stop-color="#f87171"/></linearGradient></defs><circle cx="32" cy="32" r="26" fill="url(#hintg2)"/><circle cx="20" cy="24" r="2" fill="white" opacity=".9"/><circle cx="26" cy="18" r="1.5" fill="white" opacity=".7"/><circle cx="40" cy="22" r="1.8" fill="white" opacity=".8"/><circle cx="38" cy="38" r="2" fill="white" opacity=".75"/><circle cx="24" cy="40" r="1.4" fill="white" opacity=".65"/></svg>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
  </div>
</section>

<!-- SCENE: MEMORY -->
<section id="memory" class="scene" data-scene>
  <div class="scene__sticky">
    <div class="scene__head" data-step="0">
      <span class="scene__label ec-memory">Memory</span>
      <h2 class="scene__title">It learns you</h2>
      <p class="scene__desc">Things you mention in passing become long-term facts. Months later, your agents still remember, and it shows.</p>
    </div>
    <div class="scene__fit" style="--mh:660">
      <div class="scene__stage">
        <svg class="scene__wires" viewBox="0 0 1040 560" aria-hidden="true">
          <path id="w-m1" class="ec-memory" data-wire="0.10,0.20" d="M 320 112 C 380 112, 380 240, 424 240"/>
          <path id="w-m2" class="ec-memory" data-wire="0.26,0.36" d="M 320 268 C 380 268, 380 268, 424 268"/>
          <path id="w-m3" class="ec-memory" data-wire="0.42,0.52" d="M 320 422 C 380 422, 380 296, 424 296"/>
          <path id="w-m4" class="ec-memory" data-wire="0.62,0.72" d="M 674 268 C 720 268, 720 250, 764 250"/>
        </svg>
        <div class="mem-quote" style="left:70px; top:80px" data-step="0.06">
          <em>march</em>
          "I'm vegetarian"
        </div>
        <div class="mem-quote" style="left:70px; top:236px" data-step="0.22">
          <em>april</em>
          "my partner is Ana"
        </div>
        <div class="mem-quote" style="left:70px; top:390px" data-step="0.38">
          <em>may</em>
          "we moved to Tenerife"
        </div>
        <div class="fnode fnode--memory" style="left:430px; top:210px;" data-step="0">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/></svg></div>
            <span class="fnode__label">MEMORY</span>
            <span class="fnode__id">facts</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__line" data-step="0.20">diet: vegetarian</div>
            <div class="fnode__line" data-step="0.36">partner: Ana</div>
            <div class="fnode__line" data-step="0.52">home: Tenerife</div>
          </div>
          <span class="fnode__port" style="left:-6px; top:30px"></span>
          <span class="fnode__port" style="left:-6px; top:58px"></span>
          <span class="fnode__port" style="left:-6px; top:86px"></span>
          <span class="fnode__port" style="right:-6px; top:58px"></span>
        </div>
        <div class="mem-quote" style="left:770px; top:180px; width:250px" data-step="0.58">
          <em>june</em>
          "book us a dinner saturday"
        </div>
        <div class="mem-quote" style="left:770px; top:290px; width:250px" data-step="0.76">
          <em>the agent, without asking</em>
          table for two, <mark>veggie</mark> tasting menu, <mark>Santa Cruz de Tenerife</mark>
        </div>
        <span class="scene__pulse ec-memory" data-travel="w-m4,0.68,0.76"></span>
      </div>
      <div class="scene__stage scene__stage--mobile" style="height:660px"><svg class="scene__wires" viewBox="0 0 360 660" aria-hidden="true"><path class="ec-memory" data-wire="0.10,0.20" d="M 150 58 C 215 80, 262 100, 262 190 C 262 250, 258 275, 248 302"/><path class="ec-memory" data-wire="0.26,0.36" d="M 150 148 C 200 168, 228 180, 228 240 C 228 272, 224 288, 214 302"/><path class="ec-memory" data-wire="0.42,0.52" d="M 150 238 C 175 255, 194 262, 194 278 C 194 290, 192 296, 186 302"/><path id="m-w-m4" class="ec-memory" data-wire="0.62,0.72" d="M 340 390 C 356 440, 356 500, 300 562"/></svg><div class="mem-quote" style="left:0; top:0; width:200px" data-step="0.06"><em>march</em>"I'm vegetarian"</div><div class="mem-quote" style="left:0; top:90px; width:200px" data-step="0.22"><em>april</em>"my partner is Ana"</div><div class="mem-quote" style="left:0; top:180px; width:200px" data-step="0.38"><em>may</em>"we moved to Tenerife"</div><div class="fnode fnode--memory" style="left:100px; top:300px;" data-step="0"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/></svg></div><span class="fnode__label">MEMORY</span><span class="fnode__id">facts</span></div><div class="fnode__body"><div class="fnode__line" data-step="0.20">diet: vegetarian</div><div class="fnode__line" data-step="0.36">partner: Ana</div><div class="fnode__line" data-step="0.52">home: Tenerife</div></div></div><div class="mem-quote" style="left:60px; top:470px; width:240px" data-step="0.58"><em>june</em>"book us a dinner saturday"</div><div class="mem-quote" style="left:60px; top:556px; width:240px" data-step="0.76"><em>the agent, without asking</em>table for two, <mark>veggie</mark> tasting menu, <mark>Santa Cruz de Tenerife</mark></div><span class="scene__pulse ec-memory" data-travel="m-w-m4,0.68,0.78"></span></div>
    </div>
    <div class="scene__hint">
      <svg class="scene__hint-ball" viewBox="0 0 64 64" fill="none"><defs><linearGradient id="hintg8" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#f59e0b"/><stop offset="100%" stop-color="#f87171"/></linearGradient></defs><circle cx="32" cy="32" r="26" fill="url(#hintg8)"/><circle cx="20" cy="24" r="2" fill="white" opacity=".9"/><circle cx="26" cy="18" r="1.5" fill="white" opacity=".7"/><circle cx="40" cy="22" r="1.8" fill="white" opacity=".8"/><circle cx="38" cy="38" r="2" fill="white" opacity=".75"/><circle cx="24" cy="40" r="1.4" fill="white" opacity=".65"/></svg>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
  </div>
</section>

<!-- SCENE: FLOWS -->
<section id="flows" class="scene" data-scene>
  <div class="scene__sticky">
    <div class="scene__head" data-step="0">
      <span class="scene__label ec-flow">Flows</span>
      <h2 class="scene__title">Chain agents into teams</h2>
      <p class="scene__desc">Draw the pipeline on a canvas: parallel branches, condition routing, loops. One agent writes, another decides, code does the plumbing.</p>
    </div>
    <div class="scene__fit" style="--mh:760">
      <div class="scene__stage">
        <svg class="scene__wires" viewBox="0 0 1040 560" aria-hidden="true">
          <path id="w-f1" class="ec-flow" data-wire="0.06,0.16" d="M 76 274 L 84 274"/>
          <path id="w-f2" class="ec-flow" data-wire="0.22,0.34" d="M 334 274 C 385 274, 380 274, 424 274"/>
          <path id="w-f3" class="ec-flow" data-wire="0.40,0.52" d="M 674 250 C 730 250, 715 164, 764 164"/>
          <path id="w-f4" class="ec-flow" data-wire="0.54,0.66" d="M 674 298 C 730 298, 715 424, 764 424"/>
          <path class="ec-flow wire-loop" data-step="0.72" d="M 550 336 C 550 470, 210 470, 210 336"/>
        </svg>
        <div class="scene__start" style="left:-14px; top:254px" data-step="0.02">
          <div class="scene__start-card">
            <span class="scene__start-ic"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg></span>
            <span class="scene__start-label">Start</span>
          </div>
          <span class="scene__start-port"></span>
        </div>
        <div class="fnode fnode--agent" style="left:90px; top:220px;" data-step="0.16">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div>
            <span class="fnode__label">AGENT</span>
            <span class="fnode__id">researcher</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__skel" style="width:80%"></div>
            <div class="fnode__skel" style="width:60%"></div>
          </div>
          <span class="fnode__port" style="left:-6px; top:48px"></span>
          <span class="fnode__port" style="right:-6px; top:48px"></span>
        </div>
        <div class="fnode fnode--router" style="left:430px; top:220px;" data-step="0.34">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/><line x1="4" y1="4" x2="9" y2="9"/></svg></div>
            <span class="fnode__label">ROUTER</span>
            <span class="fnode__id">quality_gate</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__line">pass: publisher</div>
            <div class="fnode__line">fix: format_md</div>
            <div class="fnode__line">retry: researcher</div>
          </div>
          <span class="fnode__port" style="left:-6px; top:48px"></span>
          <span class="fnode__port" style="right:-6px; top:26px"></span>
          <span class="fnode__port" style="right:-6px; top:74px"></span>
        </div>
        <div class="fnode fnode--agent" style="left:770px; top:110px;" data-step="0.52">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div>
            <span class="fnode__label">AGENT</span>
            <span class="fnode__id">publisher</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__skel" style="width:75%"></div>
            <div class="fnode__skel" style="width:55%"></div>
          </div>
          <span class="fnode__port" style="left:-6px; top:48px"></span>
        </div>
        <div class="fnode fnode--code" style="left:770px; top:360px; width:280px;" data-step="0.66">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg></div>
            <span class="fnode__label">CODE</span>
            <span class="fnode__id">format_md</span>
          </div>
          <div class="fnode__code"><span class="tk-c"># deterministic step: no tokens</span>
best = <span class="tk-b">sorted</span>(drafts, key=score)[-1]
out = <span class="tk-b">render</span>(best, format=<span class="tk-b">"md"</span>)</div>
          <span class="fnode__port" style="left:-6px; top:48px"></span>
        </div>
        <span class="scene__pulse ec-flow" data-travel="w-f2,0.78,0.88"></span>
      </div>
      <div class="scene__stage scene__stage--mobile" style="height:760px"><svg class="scene__wires" viewBox="0 0 360 760" aria-hidden="true"><path class="ec-flow" data-wire="0.06,0.16" d="M 180 40 C 180 50, 180 60, 180 72"/><path id="m-w-f2" class="ec-flow" data-wire="0.22,0.34" d="M 180 174 C 180 192, 180 210, 180 228"/><path class="ec-flow" data-wire="0.40,0.52" d="M 120 352 C 100 380, 100 410, 120 444"/><path class="ec-flow" data-wire="0.54,0.66" d="M 240 352 C 300 420, 300 520, 220 592"/><path class="ec-flow wire-loop" data-step="0.72" d="M 305 300 C 345 240, 345 160, 300 110"/></svg><div class="scene__start" style="left:135px; top:0" data-step="0.02"><div class="scene__start-card"><span class="scene__start-ic"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg></span><span class="scene__start-label">Start</span></div></div><div class="fnode fnode--agent" style="left:60px; top:72px;" data-step="0.16"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div><span class="fnode__label">AGENT</span><span class="fnode__id">researcher</span></div><div class="fnode__body"><div class="fnode__skel" style="width:80%"></div><div class="fnode__skel" style="width:60%"></div></div></div><div class="fnode fnode--router" style="left:60px; top:228px;" data-step="0.34"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/><line x1="4" y1="4" x2="9" y2="9"/></svg></div><span class="fnode__label">ROUTER</span><span class="fnode__id">quality_gate</span></div><div class="fnode__body"><div class="fnode__line">pass: publisher</div><div class="fnode__line">fix: format_md</div><div class="fnode__line">retry: researcher</div></div></div><div class="fnode fnode--agent" style="left:10px; top:444px;" data-step="0.52"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div><span class="fnode__label">AGENT</span><span class="fnode__id">publisher</span></div><div class="fnode__body"><div class="fnode__skel" style="width:75%"></div><div class="fnode__skel" style="width:55%"></div></div></div><div class="fnode fnode--code" style="left:40px; top:592px; width:280px;" data-step="0.66"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg></div><span class="fnode__label">CODE</span><span class="fnode__id">format_md</span></div><div class="fnode__code"><span class="tk-c"># deterministic step: no tokens</span>
best = <span class="tk-b">sorted</span>(drafts, key=score)[-1]
out = <span class="tk-b">render</span>(best, format=<span class="tk-b">"md"</span>)</div></div><span class="scene__pulse ec-flow" data-travel="m-w-f2,0.78,0.88"></span></div>
    </div>
    <div class="scene__hint">
      <svg class="scene__hint-ball" viewBox="0 0 64 64" fill="none"><defs><linearGradient id="hintg3" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#f59e0b"/><stop offset="100%" stop-color="#f87171"/></linearGradient></defs><circle cx="32" cy="32" r="26" fill="url(#hintg3)"/><circle cx="20" cy="24" r="2" fill="white" opacity=".9"/><circle cx="26" cy="18" r="1.5" fill="white" opacity=".7"/><circle cx="40" cy="22" r="1.8" fill="white" opacity=".8"/><circle cx="38" cy="38" r="2" fill="white" opacity=".75"/><circle cx="24" cy="40" r="1.4" fill="white" opacity=".65"/></svg>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
  </div>
</section>

<!-- SCENE: CLIENTS -->
<section id="clients-scene" class="scene" data-scene>
  <div class="scene__sticky">
    <div class="scene__head" data-step="0">
      <span class="scene__label ec-client">Clients</span>
      <h2 class="scene__title">Anything can wake them</h2>
      <p class="scene__desc">Voice, Telegram, webhooks, cron, plain HTTP, even other agents over A2A. Every client has its own token and reaches only the agents and flows you allow.</p>
    </div>
    <div class="scene__fit" style="--mh:680">
      <div class="scene__stage">
        <svg class="scene__wires" viewBox="0 0 1040 560" aria-hidden="true">
          <path id="w-c1" class="ec-client" data-wire="0.07,0.17" d="M 260 102 C 480 102, 540 246, 696 246"/>
          <path id="w-c2" class="ec-client" data-wire="0.21,0.31" d="M 260 198 C 480 198, 500 264, 696 264"/>
          <path id="w-c3" class="ec-trigger" data-wire="0.35,0.45" d="M 260 294 C 480 294, 500 282, 696 282"/>
          <path id="w-c4" class="ec-trigger" data-wire="0.49,0.59" d="M 260 390 C 480 390, 500 300, 696 300"/>
          <path id="w-c5" class="ec-agent" data-wire="0.63,0.73" d="M 260 486 C 480 486, 540 318, 696 318"/>
        </svg>
        <div class="scene__end ec-client" style="left:100px; top:72px" data-step="0.05">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="23"/></svg>
          <span class="scene__tag">"Oye Magec"</span>
        </div>
        <div class="scene__end ec-client" style="left:100px; top:168px" data-step="0.19">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
          <span class="scene__tag">telegram, slack, discord</span>
        </div>
        <div class="scene__end ec-trigger" style="left:100px; top:264px" data-step="0.33">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
          <span class="scene__tag">POST /webhook/alert</span>
        </div>
        <div class="scene__end ec-trigger" style="left:100px; top:360px" data-step="0.47">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><polyline points="12 7 12 12 15.5 14"/></svg>
          <span class="scene__tag">cron 0 7 * * *</span>
        </div>
        <div class="scene__end ec-agent" style="left:100px; top:456px" data-step="0.61">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="6" cy="7" r="3.5"/><circle cx="18" cy="17" r="3.5"/><path d="M9.5 7h5a3 3 0 0 1 3 3v3.5"/><path d="M14.5 17h-5a3 3 0 0 1-3-3V10.5"/></svg>
          <span class="scene__tag">other agents (a2a)</span>
        </div>
        <div class="fnode fnode--flow" style="left:700px; top:220px;" data-step="0">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="5" cy="6" r="3"/><circle cx="19" cy="18" r="3"/><path d="M8 6h8a3 3 0 0 1 3 3v6"/></svg></div>
            <span class="fnode__label">FLOW</span>
            <span class="fnode__id">morning_brief</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__skel" style="width:82%"></div>
            <div class="fnode__skel" style="width:64%"></div>
            <div class="fnode__skel" style="width:72%"></div>
          </div>
          <span class="fnode__port" style="left:-6px; top:26px"></span>
          <span class="fnode__port" style="left:-6px; top:44px"></span>
          <span class="fnode__port" style="left:-6px; top:62px"></span>
          <span class="fnode__port" style="left:-6px; top:80px"></span>
          <span class="fnode__port" style="left:-6px; top:98px"></span>
        </div>
        <span class="scene__pulse ec-client" data-travel="w-c1,0.77,0.87"></span>
        <span class="scene__pulse ec-agent" data-travel="w-c5,0.83,0.93"></span>
      </div>
      <div class="scene__stage scene__stage--mobile" style="height:680px"><svg class="scene__wires" viewBox="0 0 360 680" aria-hidden="true"><path id="m-w-c1" class="ec-client" data-wire="0.07,0.17" d="M 100 62 C 55 85, 12 100, 12 220 C 12 420, 16 500, 60 566"/><path class="ec-client" data-wire="0.21,0.31" d="M 260 158 C 310 182, 348 200, 348 300 C 348 460, 342 520, 300 566"/><path class="ec-trigger" data-wire="0.35,0.45" d="M 100 254 C 160 300, 212 380, 210 530"/><path id="m-w-c5" class="ec-agent" data-wire="0.49,0.59" d="M 260 350 C 262 420, 258 480, 252 530"/><path class="ec-trigger" data-wire="0.63,0.73" d="M 100 446 C 105 480, 112 505, 120 530"/></svg><div class="scene__end ec-client" style="left:20px; top:0px" data-step="0.05"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="23"/></svg><span class="scene__tag">"Oye Magec"</span></div><div class="scene__end ec-client" style="left:180px; top:96px" data-step="0.19"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg><span class="scene__tag">telegram, slack, discord</span></div><div class="scene__end ec-trigger" style="left:20px; top:192px" data-step="0.33"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg><span class="scene__tag">POST /webhook/alert</span></div><div class="scene__end ec-agent" style="left:180px; top:288px" data-step="0.47"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="6" cy="7" r="3.5"/><circle cx="18" cy="17" r="3.5"/><path d="M9.5 7h5a3 3 0 0 1 3 3v3.5"/><path d="M14.5 17h-5a3 3 0 0 1-3-3V10.5"/></svg><span class="scene__tag">other agents (a2a)</span></div><div class="scene__end ec-trigger" style="left:20px; top:384px" data-step="0.61"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><polyline points="12 7 12 12 15.5 14"/></svg><span class="scene__tag">cron 0 7 * * *</span></div><div class="fnode fnode--flow" style="left:60px; top:530px;" data-step="0"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="5" cy="6" r="3"/><circle cx="19" cy="18" r="3"/><path d="M8 6h8a3 3 0 0 1 3 3v6"/></svg></div><span class="fnode__label">FLOW</span><span class="fnode__id">morning_brief</span></div><div class="fnode__body"><div class="fnode__skel" style="width:82%"></div><div class="fnode__skel" style="width:64%"></div></div></div><span class="scene__pulse ec-client" data-travel="m-w-c1,0.77,0.87"></span><span class="scene__pulse ec-agent" data-travel="m-w-c5,0.83,0.93"></span></div>
    </div>
    <div class="scene__hint">
      <svg class="scene__hint-ball" viewBox="0 0 64 64" fill="none"><defs><linearGradient id="hintg4" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#f59e0b"/><stop offset="100%" stop-color="#f87171"/></linearGradient></defs><circle cx="32" cy="32" r="26" fill="url(#hintg4)"/><circle cx="20" cy="24" r="2" fill="white" opacity=".9"/><circle cx="26" cy="18" r="1.5" fill="white" opacity=".7"/><circle cx="40" cy="22" r="1.8" fill="white" opacity=".8"/><circle cx="38" cy="38" r="2" fill="white" opacity=".75"/><circle cx="24" cy="40" r="1.4" fill="white" opacity=".65"/></svg>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
  </div>
</section>

<!-- SCENE: VOICE -->
<section id="voice" class="scene" data-scene>
  <div class="scene__sticky">
    <div class="scene__head" data-step="0">
      <span class="scene__label ec-client">Voice</span>
      <h2 class="scene__title">Talk to it like a person</h2>
      <p class="scene__desc">Wake it with your voice and just talk, like Alexa, if Alexa were actually smart and knew you.</p>
    </div>
    <div class="scene__fit" style="--mh:640">
      <div class="scene__stage">
        <svg class="scene__wires" viewBox="0 0 1040 560" aria-hidden="true">
          <path id="w-v1" class="ec-client" style="stroke-width:2" data-wire="0.06,0.24" d="M 60 270 Q 74 210, 88 270 Q 102 330, 116 270 Q 130 180, 144 270 Q 158 350, 172 270 Q 186 225, 200 270 Q 214 315, 228 270 Q 242 245, 256 270"/>
          <path id="w-v2" class="ec-client" data-wire="0.28,0.38" d="M 256 270 L 394 270"/>
          <path id="w-v3" class="ec-agent" data-wire="0.52,0.62" d="M 646 270 L 782 270"/>
          <path id="w-v4" class="ec-agent" style="stroke-width:2" data-wire="0.66,0.86" d="M 782 270 Q 796 210, 810 270 Q 824 335, 838 270 Q 852 175, 866 270 Q 880 355, 894 270 Q 908 220, 922 270 Q 936 320, 950 270 Q 964 248, 978 270"/>
        </svg>
        <div class="scene__end ec-client" style="left:80px; top:330px" data-step="0.04">
          <span class="scene__tag">"Oye Magec, dim the lights"</span>
        </div>
        <div class="fnode fnode--agent" style="left:400px; top:210px;" data-step="0.38">
          <div class="fnode__head">
            <div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div>
            <span class="fnode__label">AGENT</span>
            <span class="fnode__id">home_butler</span>
          </div>
          <div class="fnode__body">
            <div class="fnode__line" data-step="0.44"><em>stt:</em> dim the lights</div>
            <div class="fnode__line" data-step="0.48"><em>mcp:</em> lights.dim(30%)</div>
            <div class="fnode__line" data-step="0.52"><em>tts:</em> "done"</div>
          </div>
          <span class="fnode__port" style="left:-6px; top:54px"></span>
          <span class="fnode__port" style="right:-6px; top:54px"></span>
        </div>
        <div class="scene__end ec-agent" style="left:800px; top:330px" data-step="0.80">
          <span class="scene__tag">"Done. Anything else?"</span>
        </div>
        <span class="scene__pulse ec-client" data-travel="w-v2,0.40,0.48"></span>
        <span class="scene__pulse ec-agent" data-travel="w-v3,0.56,0.64"></span>
      </div>
      <div class="scene__stage scene__stage--mobile" style="height:640px"><svg class="scene__wires" viewBox="0 0 360 640" aria-hidden="true"><path class="ec-client" style="stroke-width:2" data-wire="0.06,0.24" d="M 90 50 Q 104 10, 118 50 Q 132 90, 146 50 Q 160 4, 174 50 Q 188 96, 202 50 Q 216 22, 230 50 Q 244 78, 258 50 Q 268 36, 278 50"/><path id="m-w-v2" class="ec-client" data-wire="0.28,0.38" d="M 180 150 C 180 182, 180 214, 180 246"/><path id="m-w-v3" class="ec-agent" data-wire="0.52,0.62" d="M 180 372 C 180 404, 180 436, 180 468"/><path class="ec-agent" style="stroke-width:2" data-wire="0.66,0.86" d="M 90 520 Q 104 480, 118 520 Q 132 560, 146 520 Q 160 474, 174 520 Q 188 566, 202 520 Q 216 492, 230 520 Q 244 548, 258 520 Q 268 506, 278 520"/></svg><div class="scene__end ec-client" style="left:60px; top:96px; width:240px" data-step="0.04"><span class="scene__tag">"Oye Magec, dim the lights"</span></div><div class="fnode fnode--agent" style="left:60px; top:246px;" data-step="0.38"><div class="fnode__head"><div class="fnode__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></div><span class="fnode__label">AGENT</span><span class="fnode__id">home_butler</span></div><div class="fnode__body"><div class="fnode__line" data-step="0.44"><em>stt:</em> dim the lights</div><div class="fnode__line" data-step="0.48"><em>mcp:</em> lights.dim(30%)</div><div class="fnode__line" data-step="0.52"><em>tts:</em> "done"</div></div></div><div class="scene__end ec-agent" style="left:60px; top:566px; width:240px" data-step="0.80"><span class="scene__tag">"Done. Anything else?"</span></div><span class="scene__pulse ec-client" data-travel="m-w-v2,0.40,0.48"></span><span class="scene__pulse ec-agent" data-travel="m-w-v3,0.56,0.64"></span></div>
    </div>
    <div class="scene__hint">
      <svg class="scene__hint-ball" viewBox="0 0 64 64" fill="none"><defs><linearGradient id="hintg5" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#f59e0b"/><stop offset="100%" stop-color="#f87171"/></linearGradient></defs><circle cx="32" cy="32" r="26" fill="url(#hintg5)"/><circle cx="20" cy="24" r="2" fill="white" opacity=".9"/><circle cx="26" cy="18" r="1.5" fill="white" opacity=".7"/><circle cx="40" cy="22" r="1.8" fill="white" opacity=".8"/><circle cx="38" cy="38" r="2" fill="white" opacity=".75"/><circle cx="24" cy="40" r="1.4" fill="white" opacity=".65"/></svg>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
  </div>
</section>

<!-- SCENE: SECRETS -->
<section id="secrets" class="scene" data-scene>
  <div class="scene__sticky">
    <div class="scene__head" data-step="0">
      <span class="scene__label ec-secret">Secrets</span>
      <h2 class="scene__title">Models never see your keys</h2>
      <p class="scene__desc">Secrets live encrypted at rest and are swapped in only at call time. The model works with placeholders; your tools get the real value.</p>
    </div>
    <div class="scene__fit" style="--mh:640">
      <div class="scene__stage">
        <svg class="scene__wires" viewBox="0 0 1040 560" aria-hidden="true">
          <path id="w-s1" class="ec-secret" data-wire="0.20,0.70" d="M 520 96 L 520 440"/>
        </svg>
        <div class="sec-col" style="left:395px; top:20px; width:250px" data-step="0.08">
          <span class="sec-value sec-value--masked">{{secret:API_KEY}}</span>
          <span class="sec-caption">what the model sees</span>
        </div>
        <span class="sec-ball" style="left:507px; top:252px" data-step="0.42"></span>
        <div class="sec-col" style="left:395px; top:452px; width:250px" data-step="0.68">
          <span class="sec-value sec-value--real">sk-live-4f8a91c2</span>
          <span class="sec-caption">what your tools get</span>
        </div>
        <div class="sec-rest" style="left:110px; top:180px" data-step="0.30">
          <span>gAAAAABmJx9v3q</span><br>
          <span>Zk1uYzR0aXlXcU</span><br>
          <span>x2b3JQZXJhbmRv</span><br>
          <span class="sec-caption">at rest: encrypted</span>
        </div>
        <span class="scene__pulse ec-secret" data-travel="w-s1,0.74,0.90"></span>
      </div>
      <div class="scene__stage scene__stage--mobile" style="height:640px"><svg class="scene__wires" viewBox="0 0 360 640" aria-hidden="true"><path id="m-w-s1" class="ec-secret" data-wire="0.20,0.70" d="M 180 64 L 180 420"/></svg><div class="sec-col" style="left:60px; top:0; width:240px" data-step="0.08"><span class="sec-value sec-value--masked">{{secret:API_KEY}}</span><span class="sec-caption">what the model sees</span></div><span class="sec-ball" style="left:167px; top:228px" data-step="0.42"></span><div class="sec-col" style="left:60px; top:432px; width:240px" data-step="0.68"><span class="sec-value sec-value--real">sk-live-4f8a91c2</span><span class="sec-caption">what your tools get</span></div><div class="sec-rest" style="left:60px; top:540px; width:240px; text-align:center" data-step="0.30"><span>gAAAAABmJx9v3q</span><br><span>Zk1uYzR0aXlXcU</span><br><span>x2b3JQZXJhbmRv</span><br><span class="sec-caption">at rest: encrypted</span></div><span class="scene__pulse ec-secret" data-travel="m-w-s1,0.74,0.90"></span></div>
    </div>
    <div class="scene__hint">
      <svg class="scene__hint-ball" viewBox="0 0 64 64" fill="none"><defs><linearGradient id="hintg6" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#f59e0b"/><stop offset="100%" stop-color="#f87171"/></linearGradient></defs><circle cx="32" cy="32" r="26" fill="url(#hintg6)"/><circle cx="20" cy="24" r="2" fill="white" opacity=".9"/><circle cx="26" cy="18" r="1.5" fill="white" opacity=".7"/><circle cx="40" cy="22" r="1.8" fill="white" opacity=".8"/><circle cx="38" cy="38" r="2" fill="white" opacity=".75"/><circle cx="24" cy="40" r="1.4" fill="white" opacity=".65"/></svg>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
  </div>
</section>

<!-- SCENE: RUNS -->
<section id="runs" class="scene" data-scene>
  <div class="scene__sticky">
    <div class="scene__head" data-step="0">
      <span class="scene__label" style="color:#2dd4bf">Runs</span>
      <h2 class="scene__title">Every run, recorded</h2>
      <p class="scene__desc">Each step, each decision, each failure: timestamped and inspectable from the admin panel.</p>
    </div>
    <div class="scene__fit" style="--mh:560">
      <div class="scene__stage">
        <svg class="scene__wires" viewBox="0 0 1040 560" aria-hidden="true">
          <path id="w-r1" style="color:var(--piedra-600)" data-wire="0.06,0.80" d="M 370 60 L 370 500"/>
        </svg>
        <span class="run-dot run-dot--ok" style="left:364px; top:94px" data-step="0.16"></span>
        <div class="run-row" style="left:410px; top:88px" data-step="0.16">
          <span class="run-row__name">researcher</span><span class="run-row__meta">2.1s</span>
        </div>
        <span class="run-dot run-dot--ok" style="left:364px; top:194px" data-step="0.32"></span>
        <div class="run-row" style="left:410px; top:188px" data-step="0.32">
          <span class="run-row__name">quality_gate</span><span class="run-row__meta">0.1s</span>
        </div>
        <span class="run-dot run-dot--ok" style="left:364px; top:294px" data-step="0.48"></span>
        <div class="run-row" style="left:410px; top:288px" data-step="0.48">
          <span class="run-row__name">format_md</span><span class="run-row__meta">0.2s</span>
        </div>
        <span class="run-dot run-dot--fail" style="left:364px; top:394px" data-step="0.64"></span>
        <div class="run-row run-row--fail" style="left:410px; top:388px" data-step="0.64">
          <span class="run-row__name">publisher</span><span class="run-row__meta">failed</span>
        </div>
        <span class="run-dot run-dot--ok" style="left:364px; top:494px" data-step="0.80"></span>
        <div class="run-row" style="left:410px; top:488px" data-step="0.80">
          <span class="run-row__name">publisher (retry)</span><span class="run-row__meta">1.4s</span>
        </div>
      </div>
      <div class="scene__stage scene__stage--mobile" style="height:560px"><svg class="scene__wires" viewBox="0 0 360 560" aria-hidden="true"><path style="color:var(--piedra-600)" data-wire="0.06,0.80" d="M 36 24 L 36 500"/></svg><span class="run-dot run-dot--ok" style="left:30px; top:54px" data-step="0.16"></span><div class="run-row" style="left:60px; top:48px" data-step="0.16"><span class="run-row__name">researcher</span><span class="run-row__meta">2.1s</span></div><span class="run-dot run-dot--ok" style="left:30px; top:154px" data-step="0.32"></span><div class="run-row" style="left:60px; top:148px" data-step="0.32"><span class="run-row__name">quality_gate</span><span class="run-row__meta">0.1s</span></div><span class="run-dot run-dot--ok" style="left:30px; top:254px" data-step="0.48"></span><div class="run-row" style="left:60px; top:248px" data-step="0.48"><span class="run-row__name">format_md</span><span class="run-row__meta">0.2s</span></div><span class="run-dot run-dot--fail" style="left:30px; top:354px" data-step="0.64"></span><div class="run-row run-row--fail" style="left:60px; top:348px" data-step="0.64"><span class="run-row__name">publisher</span><span class="run-row__meta">failed</span></div><span class="run-dot run-dot--ok" style="left:30px; top:454px" data-step="0.80"></span><div class="run-row" style="left:60px; top:448px" data-step="0.80"><span class="run-row__name">publisher (retry)</span><span class="run-row__meta">1.4s</span></div></div>
    </div>
  </div>
</section>
