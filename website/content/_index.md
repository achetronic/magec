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
        <h3 class="feature-card__title">Multi-Agent System</h3>
        <p class="feature-card__desc">Define as many agents as you need, each with independent configuration.</p>
        <ul class="feature-card__list">
          <li>Per-agent LLM — different models and providers</li>
          <li>Per-agent memory — session (Redis) and long-term (pgvector)</li>
          <li>Per-agent voice — individual STT/TTS backends</li>
          <li>Per-agent tools — different MCP servers</li>
          <li>Hot-reload — no restart needed</li>
        </ul>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--pink">⛓</div>
        <h3 class="feature-card__title">Agentic Flows</h3>
        <p class="feature-card__desc">Chain agents into multi-step workflows with a visual drag-and-drop editor.</p>
        <ul class="feature-card__list">
          <li>Sequential — run agents one after another</li>
          <li>Parallel — run simultaneously, merge results</li>
          <li>Loop — iterate until a condition is met</li>
          <li>Nested — combine any of the above</li>
          <li>Built-in: Research, Debate, Software Factory</li>
        </ul>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--purple">⚡</div>
        <h3 class="feature-card__title">AI Backends</h3>
        <p class="feature-card__desc">Works with virtually any LLM provider — cloud or local.</p>
        <ul class="feature-card__list">
          <li>OpenAI — GPT-4.1, o3, whisper, tts</li>
          <li>Anthropic — Claude Opus, Sonnet, Haiku</li>
          <li>Google Gemini — Pro, Flash</li>
          <li>Ollama, LM Studio — run locally</li>
          <li>Any OpenAI-compatible API</li>
        </ul>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--green">🔧</div>
        <h3 class="feature-card__title">Tool Integration (MCP)</h3>
        <p class="feature-card__desc">Connect agents to external tools via the Model Context Protocol.</p>
        <ul class="feature-card__list">
          <li>HTTP and Stdio transport</li>
          <li>Per-agent tool assignment</li>
          <li>System prompt injection per MCP</li>
          <li>Home Assistant, GitHub, databases, and hundreds more</li>
        </ul>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--atlantico">🧠</div>
        <h3 class="feature-card__title">Memory</h3>
        <p class="feature-card__desc">Agents remember context across sessions — short and long term.</p>
        <ul class="feature-card__list">
          <li>Session memory — Redis with configurable TTL</li>
          <li>Long-term — PostgreSQL + pgvector semantic search</li>
          <li>Per-agent configuration</li>
          <li>Auto-save — agents learn preferences over time</li>
        </ul>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--lava">🎙</div>
        <h3 class="feature-card__title">Voice Capabilities</h3>
        <p class="feature-card__desc">Server-side voice processing powered by ONNX Runtime.</p>
        <ul class="feature-card__list">
          <li>Wake word detection — "Oye Magec"</li>
          <li>Voice Activity Detection (Silero VAD)</li>
          <li>Any Whisper-compatible STT</li>
          <li>Any OpenAI-compatible TTS</li>
          <li>WebSocket streaming for low latency</li>
        </ul>
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
      <p class="section-desc">Wake word, push-to-talk, animated orb, session management. Installable as PWA.</p>
    </div>
    <div class="screenshots reveal">
      <img src="img/screenshots/home.png" alt="Voice UI — Home" class="screenshot screenshot--phone">
      <img src="img/screenshots/chat.png" alt="Voice UI — Chat" class="screenshot screenshot--phone">
      <img src="img/screenshots/settings.png" alt="Voice UI — Settings" class="screenshot screenshot--phone">
      <img src="img/screenshots/notifications.png" alt="Voice UI — Notifications" class="screenshot screenshot--phone">
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
      <div class="client-card"><div><div class="client-card__name">Voice UI</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Web app with wake word, push-to-talk, waveform visualizer, sessions. PWA.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Admin UI</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Full management panel — agents, flows, backends, memory, tools, clients.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Telegram</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Text and voice messages. Response modes: text, voice, mirror, both.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Webhooks</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">HTTP endpoints with token auth. Passthrough or fixed commands.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Cron</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Scheduled agent invocations with cron expressions.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">REST API</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Full API with Swagger docs. SSE streaming. WebSocket for voice.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Discord</div><div class="client-card__status client-card__status--soon">Coming soon</div><p class="client-card__desc">Typed and ready, implementation pending.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Slack</div><div class="client-card__status client-card__status--soon">Coming soon</div><p class="client-card__desc">Typed and ready, implementation pending.</p></div></div>
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
      <div class="provider-card"><div class="provider-card__name">LM Studio</div><div class="provider-card__type">Local</div></div>
      <div class="provider-card"><div class="provider-card__name">OpenRouter</div><div class="provider-card__type">Cloud</div></div>
      <div class="provider-card"><div class="provider-card__name">vLLM / TGI</div><div class="provider-card__type">Self-hosted</div></div>
      <div class="provider-card"><div class="provider-card__name">Any OpenAI-compatible</div><div class="provider-card__type">Any</div></div>
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
