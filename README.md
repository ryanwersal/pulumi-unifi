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

Managed by [mise](https://mise.jdx.dev) (Go + Pulumi pinned in `mise.toml`).

```sh
mise install            # Go + Pulumi
mise run build          # → bin/pulumi-resource-unifi
mise run schema         # print the derived Pulumi schema
mise run sdk:nodejs     # generate the TypeScript SDK into sdk/nodejs
mise run check          # vet + build (pre-commit gate)
```

## Consuming from a TypeScript / Bun Pulumi program

A native provider is **not** a dynamic provider, so the
[`runtime: bun`](https://www.pulumi.com/blog/introducing-bun-as-a-runtime-for-pulumi/)
limitation on dynamic providers does **not** apply — a Bun program consumes the
generated SDK like any other package.

1. Build the binary and generate the SDK (`mise run build && mise run sdk:nodejs`).
2. Make the plugin discoverable, then add the SDK package:

   ```sh
   # from the consuming Pulumi project (e.g. atlas/pulumi)
   pulumi package add /path/to/pulumi-unifi/bin/pulumi-resource-unifi
   bun add @pulumi/unifi@../../pulumi-unifi/sdk/nodejs   # local path dep
   ```

3. Use it:

   ```ts
   import * as unifi from "@pulumi/unifi";

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

## Caveats / risks

- The Network client targets largely **undocumented** controller endpoints;
  UniFi OS firmware upgrades can change them. Mitigation: the upstream client is
  actively maintained and tracks controller versions; bump it deliberately.
- The Protect client (`ClifHouck/unified`) is pre-1.0; pin it and expect churn.
- `filipowm/go-unifi` is a fork of an archived project. It's active today, but
  worth watching.

## License

Apache-2.0.
