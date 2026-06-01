# pulumi-unifi

> A native [Pulumi](https://www.pulumi.com) provider for managing a Ubiquiti
> **UniFi Dream Machine** — its **Network** and **Protect** applications — as
> code.

This is a **native** provider built with
[`pulumi-go-provider`](https://github.com/pulumi/pulumi-go-provider) (the
`infer` layer). It talks directly to the **local UniFi controller API** and is
**not** a Terraform bridge — no Terraform, no `pulumi-terraform-bridge`.

It wraps two maintained Go clients:

- **Network** → [`filipowm/go-unifi`](https://github.com/filipowm/go-unifi) — a
  maintained fork of the (archived) `paultyng/go-unifi`, with types generated
  from the controller and support for API-key auth. Used under MPL-2.0.
- **Protect** → [`ClifHouck/unified`](https://github.com/ClifHouck/unified) —
  an unofficial Go client for the official UniFi Protect V1 integration API.

## Status

Early. The following resources exist:

| Token | What | Lifecycle |
|---|---|---|
| `unifi:network:Vlan` | A network / VLAN | full CRUD + `Read` (import) |
| `unifi:network:Wlan` | A wireless network (SSID) | full CRUD + `Read` (import) |
| `unifi:protect:Camera` | Settings of an **existing** Protect camera | adoption: Create/Update patch, Read, Delete is a no-op |

Protect cameras are physical hardware — the API has no create/delete, only
settings patches — so the `Camera` resource manages a camera you've already
adopted, identified by its Protect `cameraId`. Deleting it from Pulumi releases
it from management without touching the device.

## Configuration

| Key | Notes |
|---|---|
| `unifi:url` | Controller base URL, e.g. `https://192.168.1.1` (no `/api` suffix) |
| `unifi:apiKey` | **secret.** UniFi OS local API key. Preferred. **Required for Protect.** |
| `unifi:username` / `unifi:password` | **secret.** Alternative to `apiKey` (Network only) |
| `unifi:site` | Site name, defaults to `default` |
| `unifi:insecureTls` | Skip TLS verification (self-signed controller certs) |

Authenticate with **either** an API key **or** username/password. Protect
resources require an API key (the integration API is API-key only).

## Toolchain

Managed by [mise](https://mise.jdx.dev) (Go, Pulumi, and the lint/release tools
pinned in `mise.toml`).

```sh
mise install            # Go + Pulumi + golangci-lint + goreleaser + svu + node
mise run build          # → bin/pulumi-resource-unifi
mise run schema         # print the derived Pulumi schema
mise run sdk:nodejs     # generate the TypeScript SDK into sdk/nodejs
mise run sdk:check      # fail if the committed SDK is stale vs. the schema
mise run lint           # golangci-lint
mise run test           # go test ./...
mise run check          # lint + test + build (pre-commit gate)
```

CI (`.github/workflows/ci.yml`) runs `go mod tidy` verification, `lint`, `test`,
and `sdk:check` on every push and PR to `main`.

## Consuming from another Pulumi program

A native provider is **not** a dynamic provider, so the
[`runtime: bun`](https://www.pulumi.com/blog/introducing-bun-as-a-runtime-for-pulumi/)
limitation on dynamic providers does **not** apply — a Bun program consumes the
generated SDK like any other package.

The SDK embeds a `pluginDownloadURL` pointing at this repo's GitHub Releases, so
Pulumi auto-installs the matching plugin binary — consumers don't install the
plugin separately.

### From a released version (recommended)

Once a `v*` tag is released, the SDK is published to npm as
[`@ryanwersal/pulumi-unifi`](https://www.npmjs.com/package/@ryanwersal/pulumi-unifi):

```sh
# from the consuming Pulumi project
npm add @ryanwersal/pulumi-unifi          # or: bun add / yarn add
```

The first `pulumi up` downloads the plugin from the GitHub Release automatically.
To pre-install (e.g. in CI) or pin a specific version:

```sh
pulumi plugin install resource unifi 0.1.0 \
  --server github://api.github.com/ryanwersal/pulumi-unifi
```

### From a local checkout (development)

To consume an unreleased build, point at the local binary and SDK path:

```sh
mise run build && mise run sdk:nodejs
# from the consuming Pulumi project (e.g. atlas/pulumi)
pulumi package add /path/to/pulumi-unifi/bin/pulumi-resource-unifi
bun add @ryanwersal/pulumi-unifi@../../pulumi-unifi/sdk/nodejs   # local path dep
```

### Use it

   ```ts
   import * as unifi from "@ryanwersal/pulumi-unifi";

   const lab = new unifi.network.Vlan("lab", {
     name: "lab",
     purpose: "corporate",
     vlan: 20,
     subnet: "192.168.20.1/24",
     dhcpEnabled: true,
     dhcpStart: "192.168.20.10",
     dhcpStop: "192.168.20.254",
   });

   new unifi.network.Wlan("lab-wifi", {
     name: "lab",
     networkId: lab.networkId,
     passphrase: cfg.requireSecret("wifiPassphrase"),
   });
   ```

   Provider credentials come from config/secrets, e.g. via Doppler-injected env
   wired into a `unifi.Provider`.

## Releasing

Releases are tag-driven. `mise run release <major|minor|patch>` bumps the
version with [`svu`](https://github.com/caarlos0/svu), tags, and pushes — which
triggers `.github/workflows/release.yml`:

- **goreleaser** cross-compiles the plugin (`darwin`/`linux`/`windows` ×
  `amd64`/`arm64`, CGO-free) and uploads the
  `pulumi-resource-unifi-v<ver>-<os>-<arch>.tar.gz` archives + checksums to the
  GitHub Release. These are what the `github://` `pluginDownloadURL` resolves.
- The **Node.js SDK** is regenerated at the tagged version (the binary is
  version-stamped via `-ldflags`, and `respectSchemaVersion` carries it into the
  package) and published to npm as `@ryanwersal/pulumi-unifi`.

The SDK is published via npm
[trusted publishing](https://docs.npmjs.com/trusted-publishers) (OIDC) — no
`NPM_TOKEN`, and provenance is attached automatically. The goreleaser job uses
the default `GITHUB_TOKEN`.

### First-time setup (one-time)

Trusted publishing can only be configured on a package that already exists, so
bootstrap the package once, then grant the workflow trust:

```sh
npm login                 # auth, only needed for this bootstrap publish
mise run bootstrap        # publishes a throwaway 0.0.1 to register @ryanwersal/pulumi-unifi, then deprecates it
mise run trust            # npm trust github … — point trusted publishing at release.yml
```

After that every release runs token-free via `mise run release`.

### Useful tasks

```sh
mise run snapshot         # local cross-platform release build (no publish) — sanity check before tagging
mise run sdk:build        # compile the SDK into sdk/nodejs/bin
mise run sdk:publish      # publish the SDK from a local checkout (manual escape hatch)
```

## Caveats / risks

- The Network client targets largely **undocumented** controller endpoints;
  UniFi OS firmware upgrades can change them. Mitigation: the upstream client is
  actively maintained and tracks controller versions; bump it deliberately.
- The Protect client (`ClifHouck/unified`) is pre-1.0; pin it and expect churn.
- `filipowm/go-unifi` is a fork of an archived project. It's active today, but
  worth watching.

## License

Apache-2.0.
