---
title: "Voice UI"
---

The `direct` client type is used for browser-based access. It powers the Voice UI and any direct API calls.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/home.png" alt="Voice UI — Home" class="screenshot screenshot--phone" >}}
{{< screenshot src="img/screenshots/chat.png" alt="Voice UI — Chat" class="screenshot screenshot--phone" >}}
{{< screenshot src="img/screenshots/settings.png" alt="Voice UI — Settings" class="screenshot screenshot--phone" >}}
</div>

## Features

- **Waveform visualizer ("Centella")** — Animated orb that responds to your voice
- **Push-to-talk & wake word** — Record manually or say "Oye Magec"
- **Agent switcher** — Switch between allowed agents
- **Session management** — Create, switch, delete conversation sessions
- **PWA** — Installable as a standalone app on Android and iOS
- **i18n** — Spanish and English

## Mobile installation (PWA)

- **Android:** Chrome → Menu (⋮) → "Install app"
- **iOS:** Safari → Share (□↑) → "Add to Home Screen"

{{< callout type="info" >}}
**HTTP on local network:** For HTTP without HTTPS, add your server URL to `chrome://flags/#unsafely-treat-insecure-origin-as-secure`.
{{< /callout >}}
