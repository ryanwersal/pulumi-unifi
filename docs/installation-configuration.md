---
title: UniFi Installation & Configuration
meta_desc: How to install and configure the UniFi provider for Pulumi.
layout: package
---

## Installation

The UniFi provider is distributed as the [`@ryanwersal/pulumi-unifi`](https://www.npmjs.com/package/@ryanwersal/pulumi-unifi)
npm package. Add it to your project:

```bash
npm add @ryanwersal/pulumi-unifi   # or: yarn add / bun add / pnpm add
```

The SDK embeds a `pluginDownloadURL` pointing at this repo's GitHub Releases, so
the first `pulumi up` downloads the matching plugin binary automatically — you do
not install the plugin separately. To pre-install (e.g. in CI) or pin a version:

```bash
pulumi plugin install resource unifi 0.1.0 \
  --server github://api.github.com/ryanwersal/pulumi-unifi
```

## Configuration

Set the controller endpoint and credentials with `pulumi config`:

```bash
pulumi config set unifi:url https://192.168.1.1
pulumi config set --secret unifi:apiKey <local-api-key>
```

| Key | Notes |
|---|---|
| `unifi:url` | Controller base URL, e.g. `https://192.168.1.1` (no `/api` suffix). |
| `unifi:apiKey` | **Secret.** UniFi OS local API key. Preferred for automation. **Required for `protect:Camera`.** |
| `unifi:username` / `unifi:password` | **Secret.** Alternative to `apiKey`. **May be required for `protect:AlarmAutomation`.** |
| `unifi:site` | Site name. Defaults to `default`. |
| `unifi:insecureTls` | Skip TLS verification (self-signed controller certs). |

Authenticate with **either** an API key **or** username/password. The
`protect:Camera` resource requires an API key (the official integration API is
API-key only); `protect:AlarmAutomation` rides the controller session and may
require username/password.

### Explicit provider

To target a specific controller (or several) from one program, construct a
provider instead of relying on stack config:

```typescript
import * as unifi from "@ryanwersal/pulumi-unifi";

const provider = new unifi.Provider("home", {
    url: "https://192.168.1.1",
    apiKey: cfg.requireSecret("unifiApiKey"),
    insecureTls: true,
});

new unifi.network.Vlan("lab", { /* ... */ }, { provider });
```
