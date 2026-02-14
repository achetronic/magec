---
title: "Voice UI"
---

Magec ships with a built-in voice interface that runs in any browser. Talk to your agents, read their responses, switch between conversations — all from your phone or desktop. It's designed to feel like a native app, not a web page.

## Pairing

The first time you open the Voice UI, it needs to connect to your Magec server. Enter the token from your client configuration and the UI pairs with the API. Once paired, the token is stored locally and you won't need to enter it again.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/voice-ui-pairing.png" alt="Voice UI — Pairing" class="screenshot screenshot--phone" >}}
</div>

## Home

The home screen is where you start every conversation. Front and center is **Magec** — an animated orb inspired by the god of the Sun worshipped by the ancient people of the Canary Islands. Tap the microphone or just say *"Oye Magec"* to start talking.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/voice-ui-home-idle.png" alt="Voice UI — Home (idle)" class="screenshot screenshot--phone" >}}
{{< screenshot src="img/screenshots/voice-ui-home-recording.png" alt="Voice UI — Home (recording)" class="screenshot screenshot--phone" >}}
</div>

When idle, Magec breathes gently in gold — like the sun at rest. When you're recording, it pulses red and its particles follow your voice waveform. It's not just decoration — it gives you immediate visual feedback that the system is listening.

## Chat

Every conversation is stored and browsable. You can read back what was said, see both your messages and the agent's responses, and pick up where you left off.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/voice-ui-chat.png" alt="Voice UI — Chat" class="screenshot screenshot--phone" >}}
</div>

## Conversation history

The sidebar lets you manage multiple conversations. Create new sessions, switch between them, or delete old ones. Each session keeps its own context, so you can have parallel conversations with different agents about different topics.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/voice-ui-conversation-history.png" alt="Voice UI — Conversation history" class="screenshot screenshot--phone" >}}
</div>

## Settings

Switch between agents, change the UI language (Spanish and English), and configure your experience.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/voice-ui-settings.png" alt="Voice UI — Settings" class="screenshot screenshot--phone" >}}
{{< screenshot src="img/screenshots/voice-ui-agent-selector.png" alt="Voice UI — Agent selector" class="screenshot screenshot--phone" >}}
</div>

The agent selector only shows agents that the current client is allowed to use. Pick one and every new message goes to that agent.

## Notifications

System events and status updates appear here — connection changes, errors, and other things worth knowing about.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/voice-ui-notifications.png" alt="Voice UI — Notifications" class="screenshot screenshot--phone" >}}
</div>

## Install as an app (PWA)

The Voice UI is a Progressive Web App. Install it on your phone and it looks and feels like a native app — full screen, its own icon, no browser chrome.

- **Android:** Chrome → Menu (⋮) → "Install app"
- **iOS:** Safari → Share (□↑) → "Add to Home Screen"

{{< callout type="info" >}}
**HTTP on local network:** If you're running Magec over HTTP (no HTTPS), add your server URL to `chrome://flags/#unsafely-treat-insecure-origin-as-secure` for microphone access to work.
{{< /callout >}}
