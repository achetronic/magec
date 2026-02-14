---
title: "Magec — Self-hosted Multi-Agent AI Platform"
---

<!-- HERO -->
<section class="hero">
  <canvas id="hero-orb" class="hero__orb"></canvas>
  <div class="hero__content">
    <div class="hero__badge">✦ Self-hosted · Open Source · Apache 2.0</div>
    <h1 class="hero__title">Your AI agents,<br><span>your rules.</span></h1>
    <p class="hero__subtitle">Define multiple agents with their own LLMs, memory, and tools. Chain them into workflows. Access via voice, Telegram, webhooks, or cron. Manage it all from a visual panel.</p>
    <div class="hero__actions">
      <a href="#quickstart" class="btn btn--primary">
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
          <span class="terminal__comment"># Fully local — no API keys needed</span><br>
          <span class="terminal__prompt">$</span> git clone <span class="terminal__url">https://github.com/achetronic/magec.git</span><br>
          <span class="terminal__prompt">$</span> cd magec/docker/compose/fully-local<br>
          <span class="terminal__prompt">$</span> docker compose up -d<br><br>
          <span class="terminal__comment"># Admin UI → <span class="terminal__url">http://localhost:8081</span></span><br>
          <span class="terminal__comment"># Voice UI → <span class="terminal__url">http://localhost:8080</span></span>
        </div>
      </div>
    </div>
  </div>
</section>

<!-- FEATURES -->
<section id="features">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">Features</span>
      <h2 class="section-title">Everything you need to run AI agents</h2>
      <p class="section-desc">From single agents to complex multi-step workflows, with full control over every component.</p>
    </div>
    <div class="features-grid stagger-children">
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--sol">✦</div>
        <h3 class="feature-card__title">Multi-Agent</h3>
        <p class="feature-card__desc">Independent agents, each with its own LLM, memory, voice, and tools. Hot-reload from the Admin UI.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--pink">⛓</div>
        <h3 class="feature-card__title">Agentic Flows</h3>
        <p class="feature-card__desc">Visual drag-and-drop editor. Sequential, parallel, loop, nested. From 3 agents to 13.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--purple">⚡</div>
        <h3 class="feature-card__title">AI Backends</h3>
        <p class="feature-card__desc">OpenAI, Anthropic, Gemini, Ollama. Mix cloud and local models per agent.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--green">🔧</div>
        <h3 class="feature-card__title">MCP Tools</h3>
        <p class="feature-card__desc">Home Assistant, GitHub, databases, and hundreds more via Model Context Protocol.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--atlantico">🧠</div>
        <h3 class="feature-card__title">Memory</h3>
        <p class="feature-card__desc">Session memory (Redis) and long-term semantic search (PostgreSQL + pgvector). Per agent.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--lava">🎙</div>
        <h3 class="feature-card__title">Voice</h3>
        <p class="feature-card__desc">Wake word, VAD, STT, TTS. Server-side via ONNX Runtime. WebSocket streaming.</p>
      </div>
    </div>
  </div>
</section>

<!-- ARCHITECTURE -->
<section id="architecture" class="architecture">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">Architecture</span>
      <h2 class="section-title">How it all connects</h2>
      <p class="section-desc">Clients on one side, AI backends on the other, Magec orchestrating in the middle.</p>
    </div>
    <div class="reveal">
      <img src="img/architecture.svg" alt="Magec Architecture" class="architecture__img">
    </div>
  </div>
</section>

<!-- SCREENSHOTS -->
<section id="screenshots">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">Admin Panel</span>
      <h2 class="section-title">Manage everything visually</h2>
      <p class="section-desc">7 resource types. Visual flow editor. Dynamic forms. Live health checks. Keyboard shortcuts.</p>
    </div>
    <div class="screenshots reveal">
      <img src="img/screenshots/admin-agents.png" alt="Admin UI — 19 agents configured" class="screenshot screenshot--desktop">
      <img src="img/screenshots/admin-flows.png" alt="Admin UI — Software Factory pipeline" class="screenshot screenshot--desktop">
    </div>
    <div class="screenshots reveal" style="margin-top: 1rem;">
      <img src="img/screenshots/admin-backends.png" alt="Admin UI — AI Backends" class="screenshot screenshot--desktop">
      <img src="img/screenshots/admin-clients.png" alt="Admin UI — Clients" class="screenshot screenshot--desktop">
    </div>
    <div class="section-header reveal" style="margin-top: 4rem;">
      <span class="section-label">Voice Interface</span>
      <h2 class="section-title">Talk to your agents</h2>
      <p class="section-desc">Just say "Oye Magec" or tap to talk. Your conversations are saved. Install it on your phone like a native app.</p>
    </div>
    <div class="screenshots reveal">
      <img src="img/screenshots/voice-ui-home-idle.png" alt="Voice UI — Home" class="screenshot screenshot--phone">
      <img src="img/screenshots/voice-ui-home-recording.png" alt="Voice UI — Recording" class="screenshot screenshot--phone">
      <img src="img/screenshots/voice-ui-chat.png" alt="Voice UI — Chat" class="screenshot screenshot--phone">
      <img src="img/screenshots/voice-ui-settings.png" alt="Voice UI — Settings" class="screenshot screenshot--phone">
    </div>
  </div>
</section>

<!-- CLIENTS -->
<section id="clients">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">Clients</span>
      <h2 class="section-title">Multiple ways to reach your agents</h2>
    </div>
    <div class="clients-grid stagger-children">
      <div class="client-card"><div><div class="client-card__name">Voice UI</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Talk to your agents from the browser. Installable on your phone.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Admin UI</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Manage everything visually. No config files needed.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Telegram</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Chat with your agents on Telegram. Text or voice.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Webhooks</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Trigger agents from any external system via HTTP.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Cron</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Run agents on a schedule. Daily reports, checks, whatever you need.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">REST API</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Build your own integration. Full Swagger docs included.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Discord</div><div class="client-card__status client-card__status--soon">Coming soon</div><p class="client-card__desc">On the way.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Slack</div><div class="client-card__status client-card__status--soon">Coming soon</div><p class="client-card__desc">On the way.</p></div></div>
    </div>
  </div>
</section>

<!-- PROVIDERS -->
<section id="providers">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">AI Backends</span>
      <h2 class="section-title">Bring any model</h2>
      <p class="section-desc">Cloud APIs, local inference, or a mix. Each agent can use a different backend.</p>
    </div>
    <div class="providers-grid stagger-children">
      <div class="provider-card"><div class="provider-card__name">OpenAI</div><div class="provider-card__type">Cloud</div></div>
      <div class="provider-card"><div class="provider-card__name">Anthropic</div><div class="provider-card__type">Cloud</div></div>
      <div class="provider-card"><div class="provider-card__name">Google Gemini</div><div class="provider-card__type">Cloud</div></div>
      <div class="provider-card"><div class="provider-card__name">Ollama</div><div class="provider-card__type">Local</div></div>
    </div>
  </div>
</section>

<!-- QUICK START -->
<section id="quickstart">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">Quick Start</span>
      <h2 class="section-title">Running in 3 commands</h2>
    </div>
    <div class="quickstart-options reveal">
      <div class="quickstart-card">
        <span class="quickstart-card__tag quickstart-card__tag--local">Fully Local</span>
        <h3 class="quickstart-card__title">No API keys needed</h3>
        <p class="quickstart-card__desc">Everything runs on your machine — LLM, speech-to-text, text-to-speech, embeddings.</p>
        <pre><code><span class="terminal__prompt">$</span> git clone https://github.com/achetronic/magec.git
<span class="terminal__prompt">$</span> cd magec/docker/compose/fully-local
<span class="terminal__prompt">$</span> docker compose up -d</code></pre>
        <p class="quickstart-card__note">First start downloads ~5GB of models. For NVIDIA GPU, uncomment the <code>deploy</code> section.</p>
      </div>
      <div class="quickstart-card">
        <span class="quickstart-card__tag quickstart-card__tag--cloud">Cloud APIs</span>
        <h3 class="quickstart-card__title">Just an API key</h3>
        <p class="quickstart-card__desc">Only Redis and PostgreSQL locally. LLM, STT, TTS, and embeddings via cloud.</p>
        <pre><code><span class="terminal__prompt">$</span> git clone https://github.com/achetronic/magec.git
<span class="terminal__prompt">$</span> cd magec/docker/compose/remote-openai
<span class="terminal__prompt">$</span> export OPENAI_API_KEY=sk-...
<span class="terminal__prompt">$</span> docker compose up -d</code></pre>
        <p class="quickstart-card__note">Configure agents at <code>http://localhost:8081</code>, chat at <code>http://localhost:8080</code>.</p>
      </div>
    </div>
  </div>
</section>

<!-- ABOUT -->
<section id="about">
  <div class="container container--narrow">
    <div class="section-header reveal">
      <span class="section-label">About</span>
      <h2 class="section-title">Why "Magec"?</h2>
      <p class="section-desc"><strong>Magec</strong> (/maˈxek/) was the god of the Sun worshipped by the Guanches, the aboriginal Berber inhabitants of Tenerife in the Canary Islands. The name honors this Canarian heritage while reflecting the project's purpose: to illuminate and assist.</p>
    </div>
  </div>
</section>
