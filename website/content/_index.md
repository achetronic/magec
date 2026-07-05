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
</section>

<!-- HOW IT WORKS -->
<section id="how-it-works">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">How it works</span>
      <h2 class="section-title">Three steps to AI automation</h2>
    </div>
    <div class="steps stagger-children">
      <div class="step">
        <div class="step__number">1</div>
        <div class="step__icon step__icon--sol">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><line x1="19" y1="8" x2="19" y2="14"/><line x1="22" y1="11" x2="16" y2="11"/></svg>
        </div>
        <h3 class="step__title">Create agents</h3>
        <p class="step__desc">Give each agent its own LLM, personality, memory, voice and tools. Mix OpenAI, Anthropic, Gemini or Ollama, even in the same setup.</p>
      </div>
      <div class="step__connector"><svg viewBox="0 0 40 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity=".3"><path d="M0 12h32m0 0l-6-6m6 6l-6 6"/></svg></div>
      <div class="step">
        <div class="step__number">2</div>
        <div class="step__icon step__icon--pink">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/><line x1="4" y1="4" x2="9" y2="9"/></svg>
        </div>
        <h3 class="step__title">Chain into flows</h3>
        <p class="step__desc">Build multi-agent workflows in a visual editor. Fan out in parallel, route on conditions, loop until the work is good. One agent writes, another reviews, another fact-checks.</p>
      </div>
      <div class="step__connector"><svg viewBox="0 0 40 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity=".3"><path d="M0 12h32m0 0l-6-6m6 6l-6 6"/></svg></div>
      <div class="step">
        <div class="step__number">3</div>
        <div class="step__icon step__icon--atlantico">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        </div>
        <h3 class="step__title">Connect everywhere</h3>
        <p class="step__desc">Expose any agent or flow through voice, Telegram, webhooks or cron. Each client gets its own token and permissions. Add as many as you need.</p>
      </div>
    </div>
  </div>
</section>

<!-- PILLARS -->
<section id="what">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">The platform</span>
      <h2 class="section-title">Everything you need to run AI agents</h2>
      <p class="section-desc">Magec is a complete platform, not just an API wrapper. Agents, workflows, tools, memory, voice and integrations, managed from a single admin panel.</p>
    </div>
    <div class="pillars stagger-children">
      <div class="pillar">
        <div class="pillar__icon pillar__icon--sol">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
        </div>
        <h3 class="pillar__title">Multi-Agent System</h3>
        <p class="pillar__desc">Each agent is an independent unit with its own LLM, system prompt, tools and memory. Hot-reload from the admin: no restarts, no config files.</p>
        <ul class="pillar__list">
          <li>Per-agent LLM selection (GPT, Claude, Gemini, Ollama)</li>
          <li>Hundreds of tools via MCP (Model Context Protocol)</li>
          <li>Session + long-term semantic memory</li>
        </ul>
      </div>
      <div class="pillar">
        <div class="pillar__icon pillar__icon--pink">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
        </div>
        <h3 class="pillar__title">Agentic Flows</h3>
        <p class="pillar__desc">Chain agents into teams that work together. Build pipelines of 2 agents or 20 in a visual graph editor.</p>
        <ul class="pillar__list">
          <li>Parallel branches, condition routing, loops</li>
          <li>Deterministic blocks that cost no tokens</li>
          <li>Every run recorded and inspectable</li>
        </ul>
      </div>
      <div class="pillar">
        <div class="pillar__icon pillar__icon--atlantico">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/></svg>
        </div>
        <h3 class="pillar__title">Voice First</h3>
        <p class="pillar__desc">Wake word detection, speech-to-text, text-to-speech, all server-side via ONNX Runtime. Each agent can have its own voice.</p>
        <ul class="pillar__list">
          <li>Privacy-first: audio never leaves your server</li>
          <li>PWA installable on tablets and phones</li>
          <li>"Oye Magec" hands-free activation</li>
        </ul>
      </div>
      <div class="pillar">
        <div class="pillar__icon pillar__icon--green">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
        </div>
        <h3 class="pillar__title">Automations</h3>
        <p class="pillar__desc">Agents don't need a human to start working. Schedule them with cron, trigger them from webhooks, or chain them with external systems.</p>
        <ul class="pillar__list">
          <li>Cron jobs with standard syntax + @daily shorthands</li>
          <li>Webhooks with passthrough or fixed commands</li>
          <li>Full REST API with Swagger docs</li>
        </ul>
      </div>
    </div>
  </div>
</section>

<!-- USE CASES -->
<section id="use-cases">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">What you can build</span>
      <h2 class="section-title">Real examples, not buzzwords</h2>
    </div>
    <div class="features-grid stagger-children">
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--sol">🏠</div>
        <h3 class="feature-card__title">Smart home with natural language</h3>
        <p class="feature-card__desc">Connect Home Assistant via MCP. Ask "turn off the living room lights" from your voice tablet, Telegram, or a cron job that dims everything at midnight.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--pink">🏗</div>
        <h3 class="feature-card__title">Software factory</h3>
        <p class="feature-card__desc">13 agents in a pipeline: product manager → architect → developers → QA → documentation. Feed it a feature request, get back a complete technical spec with code.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--atlantico">🎙</div>
        <h3 class="feature-card__title">Multi-agent platform for your business</h3>
        <p class="feature-card__desc">Put a tablet at the front desk. Staff speaks, the agent listens, checks inventory via MCP, and answers hands-free. Each agent has its own voice.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--green">📊</div>
        <h3 class="feature-card__title">Automated reports</h3>
        <p class="feature-card__desc">A cron job fires every morning. An agent queries your database, another writes the summary, a third formats it. The result lands in your inbox via webhook.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--lava">🔬</div>
        <h3 class="feature-card__title">Research pipeline</h3>
        <p class="feature-card__desc">Two researchers work in parallel, a critic reviews their output, a fact-checker verifies claims, a synthesizer produces the final report. All from a single prompt.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--purple">🔌</div>
        <h3 class="feature-card__title">Whatever you connect</h3>
        <p class="feature-card__desc">MCP gives your agents access to hundreds of tools: GitHub, databases, file systems, APIs. The more tools you connect, the more your agents can do.</p>
      </div>
    </div>
  </div>
</section>

<!-- FEATURES -->
<section id="features">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">Under the hood</span>
      <h2 class="section-title">What powers it</h2>
    </div>
    <div class="features-grid stagger-children">
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--sol">✦</div>
        <h3 class="feature-card__title">Agents</h3>
        <p class="feature-card__desc">Each agent has its own LLM, system prompt, memory, voice, and tools. Create as many as you need. Changes take effect instantly, no restarts.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--pink">⛓</div>
        <h3 class="feature-card__title">Flows</h3>
        <p class="feature-card__desc">Chain agents into workflows on a visual canvas. Parallel branches, condition routing, loops, flows inside flows. Deterministic blocks handle the plumbing so you spend tokens only where thinking happens.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--purple">⚡</div>
        <h3 class="feature-card__title">AI Backends</h3>
        <p class="feature-card__desc">OpenAI, Anthropic, Google Gemini, Ollama. Mix cloud and local models. One agent can use GPT-4, another a local Qwen, in the same flow.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--green">🔧</div>
        <h3 class="feature-card__title">MCP Tools</h3>
        <p class="feature-card__desc">Connect external tools via Model Context Protocol. Home Assistant, GitHub, databases, file systems: hundreds of integrations, growing every day.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--atlantico">🧠</div>
        <h3 class="feature-card__title">Memory</h3>
        <p class="feature-card__desc">Session memory keeps conversation history in Redis. Long-term memory stores facts about you in PostgreSQL with semantic search. Your agents remember.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--lava">🎙</div>
        <h3 class="feature-card__title">Voice</h3>
        <p class="feature-card__desc">Wake word detection, voice activity detection, speech-to-text, text-to-speech. All processed server-side via ONNX Runtime. Each agent can have its own voice.</p>
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
      <p class="section-desc">Clients on the left, AI backends on the right, Magec orchestrating in the middle. Every connection is configurable.</p>
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
      <p class="section-desc">No config files to edit. Create agents, design flows, connect tools, manage clients, all from your browser.</p>
    </div>
    <div class="screenshots reveal">
      <img src="img/screenshots/admin-agents.png" alt="Admin UI: Agents" class="screenshot screenshot--desktop">
      <img src="img/screenshots/admin-flow-editor.png" alt="Admin UI: Flow editor" class="screenshot screenshot--desktop">
    </div>
    <div class="screenshots reveal" style="margin-top: 1rem;">
      <img src="img/screenshots/admin-backends.png" alt="Admin UI: Backends" class="screenshot screenshot--desktop">
      <img src="img/screenshots/admin-clients.png" alt="Admin UI: Clients" class="screenshot screenshot--desktop">
    </div>
    <div class="screenshots reveal" style="margin-top: 1rem;">
      <img src="img/screenshots/admin-conversations.png" alt="Admin UI: Conversations" class="screenshot screenshot--desktop">
      <img src="img/screenshots/admin-conversation-detail.png" alt="Admin UI: Conversation detail" class="screenshot screenshot--desktop">
    </div>
    <div class="section-header reveal" style="margin-top: 4rem;">
      <span class="section-label">Voice Interface</span>
      <h2 class="section-title">Talk to your agents</h2>
      <p class="section-desc">Say "Oye Magec" or tap to talk. Switch between agents and flows. Choose who speaks when a team responds. Install it on your phone like a native app.</p>
    </div>
    <div class="screenshots reveal">
      <img src="img/screenshots/voice-ui-home-idle.png" alt="Voice UI: Home" class="screenshot screenshot--phone">
      <img src="img/screenshots/voice-ui-home-recording.png" alt="Voice UI: Recording" class="screenshot screenshot--phone">
      <img src="img/screenshots/voice-ui-chat.png" alt="Voice UI: Chat" class="screenshot screenshot--phone">
      <img src="img/screenshots/voice-ui-settings.png" alt="Voice UI: Settings" class="screenshot screenshot--phone">
    </div>
  </div>
</section>

<!-- CLIENTS -->
<section id="clients">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">Clients</span>
      <h2 class="section-title">Reach your agents from anywhere</h2>
      <p class="section-desc">Every client gets its own token, its own set of allowed agents, and its own way of connecting. Add as many as you need.</p>
    </div>
    <div class="clients-grid stagger-children">
      <div class="client-card"><div><div class="client-card__name">Voice UI</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Browser-based voice interface with wake word, push-to-talk, agent switching, conversation history. Installable as PWA.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Admin UI</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Visual management panel. Create agents, design flows, connect tools, manage clients. Keyboard shortcuts, search palette, live health checks.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Telegram</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Text or voice messages. Multiple response modes (text, voice, mirror, both). Per-chat agent switching.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Webhooks</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">HTTP endpoints for external integrations. Fixed command or passthrough mode. Wire them to CI, forms, alerts, or any system.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Cron</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Scheduled tasks. Daily summaries, periodic checks, automated maintenance. Standard cron syntax plus shorthands like @daily.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">REST API</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Full API with Swagger docs on both ports. Build any custom integration you can imagine.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Slack</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Socket Mode, no public URL needed. DMs and @mentions in channels. Per-channel agent switching.</p></div></div>
      <div class="client-card"><div><div class="client-card__name">Discord</div><div class="client-card__status client-card__status--ready">✓ Available</div><p class="client-card__desc">Gateway WebSocket, no public URL needed. DMs, @mentions, voice messages, reactions, and commands.</p></div></div>
    </div>
  </div>
</section>

<!-- PROVIDERS -->
<section id="providers">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">AI Backends</span>
      <h2 class="section-title">Bring any model</h2>
      <p class="section-desc">Each agent picks its own backend. Mix a cloud model for complex reasoning with a local model for fast tasks, in the same flow.</p>
    </div>
    <div class="providers-grid stagger-children">
      <div class="provider-card"><div class="provider-card__name">OpenAI</div><div class="provider-card__type">Cloud</div></div>
      <div class="provider-card"><div class="provider-card__name">Anthropic</div><div class="provider-card__type">Cloud</div></div>
      <div class="provider-card"><div class="provider-card__name">Google Gemini</div><div class="provider-card__type">Cloud</div></div>
      <div class="provider-card"><div class="provider-card__name">Ollama</div><div class="provider-card__type">Local</div></div>
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
