---
title: "Magec — Self-hosted Multi-Agent AI Platform"
---

<!-- HERO -->
<section class="hero">
  <canvas id="hero-orb" class="hero__orb"></canvas>
  <div class="hero__content">
    <div class="hero__badge">✦ Self-hosted · Open Source · Apache 2.0</div>
    <h1 class="hero__title">Your AI agents,<br><span>your rules.</span></h1>
    <p class="hero__subtitle">Create AI agents that think, remember, speak, and use tools. Chain them into teams that work together. Talk to them, text them, automate them. Everything runs on your server.</p>
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
          <span class="terminal__comment"># Or bring your cloud provider</span><br>
          <span class="terminal__prompt">$</span> ... | bash -s -- --openai<br>
          <span class="terminal__prompt">$</span> ... | bash -s -- --anthropic<br>
          <span class="terminal__prompt">$</span> ... | bash -s -- --gemini<br><br>
          <span class="terminal__comment"># Admin panel → <span class="terminal__url">localhost:8081</span></span><br>
          <span class="terminal__comment"># Voice interface → <span class="terminal__url">localhost:8080</span></span>
        </div>
      </div>
    </div>
  </div>
</section>

<!-- WHAT IT IS -->
<section id="what">
  <div class="container container--narrow">
    <div class="section-header reveal">
      <span class="section-label">What is Magec</span>
      <h2 class="section-title">An AI platform that lives on your server</h2>
    </div>
    <div class="reveal" style="color: var(--arena-300); line-height: 1.8; font-size: 1.0625rem;">
      <p style="margin-bottom: 1rem;">Magec lets you create <strong style="color: var(--arena-100);">AI agents</strong> — each with its own brain (LLM), personality (system prompt), memory, voice, and tools. You decide which model powers each agent, what tools it can access, and how it behaves.</p>
      <p style="margin-bottom: 1rem;">Agents can work alone or in <strong style="color: var(--arena-100);">teams</strong>. Chain them into workflows where one writes, another reviews, another fact-checks. Build pipelines of 2 agents or 20 — the visual editor handles both.</p>
      <p style="margin-bottom: 1rem;">Then connect them to the real world. Talk to them with your voice. Chat on Telegram. Trigger them from webhooks. Schedule them with cron. Every agent is reachable from any channel you configure.</p>
      <p>Everything runs on your hardware. Your data never leaves your network unless you choose a cloud LLM provider. Even then, the rest of the platform — memory, tools, voice processing — stays local.</p>
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
        <h3 class="feature-card__title">Voice assistant for your business</h3>
        <p class="feature-card__desc">Put a tablet at the front desk. Staff speaks, the agent listens, checks inventory via MCP, and answers — hands-free. Each agent has its own voice.</p>
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
        <p class="feature-card__desc">MCP gives your agents access to hundreds of tools — GitHub, databases, file systems, APIs. The more tools you connect, the more your agents can do.</p>
      </div>
    </div>
  </div>
</section>

<!-- FEATURES -->
<section id="features">
  <div class="container">
    <div class="section-header reveal">
      <span class="section-label">Platform</span>
      <h2 class="section-title">What's inside</h2>
    </div>
    <div class="features-grid stagger-children">
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--sol">✦</div>
        <h3 class="feature-card__title">Agents</h3>
        <p class="feature-card__desc">Each agent has its own LLM, system prompt, memory, voice, and tools. Create as many as you need. Changes take effect instantly — no restarts.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--pink">⛓</div>
        <h3 class="feature-card__title">Flows</h3>
        <p class="feature-card__desc">Chain agents into workflows. Sequential, parallel, loops, nested. Build them visually with drag-and-drop or define them as JSON.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--purple">⚡</div>
        <h3 class="feature-card__title">AI Backends</h3>
        <p class="feature-card__desc">OpenAI, Anthropic, Google Gemini, Ollama. Mix cloud and local models. One agent can use GPT-4, another can use a local Qwen — in the same flow.</p>
      </div>
      <div class="feature-card">
        <div class="feature-card__icon feature-card__icon--green">🔧</div>
        <h3 class="feature-card__title">MCP Tools</h3>
        <p class="feature-card__desc">Connect external tools via Model Context Protocol. Home Assistant, GitHub, databases, file systems — hundreds of integrations, growing every day.</p>
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
      <p class="section-desc">No config files to edit. Create agents, design flows, connect tools, manage clients — all from your browser.</p>
    </div>
    <div class="screenshots reveal">
      <img src="img/screenshots/admin-agents.png" alt="Admin UI — Agents" class="screenshot screenshot--desktop">
      <img src="img/screenshots/admin-flows.png" alt="Admin UI — Flow editor" class="screenshot screenshot--desktop">
    </div>
    <div class="screenshots reveal" style="margin-top: 1rem;">
      <img src="img/screenshots/admin-backends.png" alt="Admin UI — Backends" class="screenshot screenshot--desktop">
      <img src="img/screenshots/admin-clients.png" alt="Admin UI — Clients" class="screenshot screenshot--desktop">
    </div>
    <div class="screenshots reveal" style="margin-top: 1rem;">
      <img src="img/screenshots/admin-conversations.png" alt="Admin UI — Conversations" class="screenshot screenshot--desktop">
      <img src="img/screenshots/admin-conversation-detail.png" alt="Admin UI — Conversation detail" class="screenshot screenshot--desktop">
    </div>
    <div class="section-header reveal" style="margin-top: 4rem;">
      <span class="section-label">Voice Interface</span>
      <h2 class="section-title">Talk to your agents</h2>
      <p class="section-desc">Say "Oye Magec" or tap to talk. Switch between agents and flows. Choose who speaks when a team responds. Install it on your phone like a native app.</p>
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
      <p class="section-desc">Each agent picks its own backend. Mix a cloud model for complex reasoning with a local model for fast tasks — in the same flow.</p>
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
