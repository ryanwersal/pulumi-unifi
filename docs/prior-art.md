# Prior art — reference repos and what to copy from each

Curated, fact‑checked catalog of the providers, framework sources, and docs that define
"good" for a native Go `infer` provider, plus the UniFi‑specific field‑coverage
references. Each entry says **what to learn or copy**. These back the citations in
[`pulumi-provider-best-practices.md`](./pulumi-provider-best-practices.md).

## Framework & canonical templates

- **[pulumi/pulumi-go-provider](https://github.com/pulumi/pulumi-go-provider)** — the
  framework itself. Read `infer/` for the `ProviderBuilder` fluent API, `Annotator`
  (`Describe`/`SetDefault`/`SetToken`/`AddAlias`/`Deprecate`), `Enum[T]`/`EnumValue`,
  `Config`/`GetConfig`, token derivation, and the replace‑vs‑update reflection logic in
  `infer/resource.go`. The single most authoritative source.
- **[infer API reference (pkg.go.dev)](https://pkg.go.dev/github.com/pulumi/pulumi-go-provider/infer)**
  — exact signatures for everything above plus the lifecycle hooks (`CustomCreate/Read/
  Update/Delete/Check/Diff`, `ExplicitDependencies.WireDependencies`,
  `CustomStateMigrations`) and `ProviderBuilder.With*`.
- **[pulumi/pulumi-go-provider `examples/`](https://github.com/pulumi/pulumi-go-provider/tree/main/examples)**
  — the most copyable patterns. `file` (custom `Check`/`Diff`/`Update`, idempotent
  `Delete`), **`dna-store` (implementing `Read` — the template for UniFi adoption/import
  resources like `Device`/`Camera`)**, `configurable` (provider that connects to an
  external system via a client — the shape of a controller client), `credentials`
  (`Configure` + `provider:"secret"`; note it logs the password as a deliberate
  anti‑pattern), `random-login`/`component-provider` (components + `WithMocks` tests),
  `str` (functions/invokes), `remember` (parameterized providers).
- **[pulumi/pulumi-provider-boilerplate](https://github.com/pulumi/pulumi-provider-boilerplate)**
  — reference repo layout, `Makefile` targets, `.golangci.yml`, and `.goreleaser.yml`.
  The directory/binary/SDK conventions to mirror. (Note: its `.goreleaser.yml` is
  templated by `pulumi/ci-mgmt` and sets `changelog: skip: true`; don't assume every
  config in it is hand‑authored.)
- **["The Easier Way to Create Pulumi Providers in Go" (boilerplate‑v2 blog)](https://www.pulumi.com/blog/pulumi-go-boilerplate-v2/)**
  and **["Pulumi Go Provider SDK is now GA"](https://www.pulumi.com/blog/pulumi-go-provider-v1/)**
  — the recommended `NewProviderBuilder` workflow and the built‑in testing story.

## Real‑world native (non‑bridged) providers

- **[pulumi/pulumi-command](https://github.com/pulumi/pulumi-command)** — best production
  `infer` reference. Builds via `infer.Provider(infer.Options{…})` (the older form; this
  repo uses the newer `NewProviderBuilder` — both valid in v1.x), registers resources
  **and** a function (`infer.Function`), with full multi‑language SDK metadata. Copy its
  secret handling and language‑map setup.
- **[pulumi/pulumi-docker-build](https://github.com/pulumi/pulumi-docker-build)** —
  **counterexample / when to drop below `infer`.** A production native provider generated
  from the boilerplate that wires the lower‑level `p.Provider` via `gp.RawServer(...)`
  instead of `infer`, for fine‑grained control. Useful as the boilerplate‑derived
  layout/release reference and to know the escape hatch exists.
- **[pulumiverse/pulumi-esxi-native](https://github.com/pulumiverse/pulumi-esxi-native)** —
  a native Go provider distributed **entirely outside the pulumi org**, exactly this
  repo's situation: `pulumi plugin install … --server github://api.github.com/pulumiverse`,
  npm `@pulumiverse/esxi-native`, Go `…/sdk/go/esxi`. The scoping/distribution pattern to
  mirror (this repo uses `@ryanwersal/...`).
- **[mbrav/pulumi-netbird](https://pkg.go.dev/github.com/mbrav/pulumi-netbird/provider/resource)** —
  community `infer` provider showing real `Enum[T]` `Values()` implementations.

## Distribution & release tooling

- **[Authoring an Executable Plugin Package](https://www.pulumi.com/docs/iac/guides/building-extending/packages/executable-plugin/)**
  — authoritative spec for binary/asset naming, the os/arch matrix, the
  `github://api.github.com/<org>[/<repo>]` resolver, plain‑URL hosting layout, and
  `gen-sdk`/SDK publishing.
- **[Package Schema reference](https://www.pulumi.com/docs/iac/guides/building-extending/packages/schema/)**
  — token format, casing, metadata fields, and the language map (`packageName`,
  `respectSchemaVersion`, `importBasePath`).
- **[Publishing Packages guide](https://www.pulumi.com/docs/iac/guides/building-extending/packages/publishing-packages/)**
  + **[pulumi/registry](https://github.com/pulumi/registry)** — community‑package
  requirements: the two `docs/` files, the release tag, SDK targets, and the schema
  metadata the Registry renders (keywords `category/…`, `kind/native`).
- **[npm trusted publishing](https://docs.npmjs.com/trusted-publishers)** — OIDC setup,
  `id-token: write`, automatic provenance.
- **[pulumi/pulumi-package-publisher](https://github.com/pulumi/pulumi-package-publisher)**
  + **[pulumi/publish-go-sdk-action](https://github.com/pulumi/publish-go-sdk-action)** —
  multi‑language SDK publishing and path‑prefixed Go module tagging, if/when more SDKs are
  added.
- **[golang/govulncheck-action](https://github.com/golang/govulncheck-action)** and
  **[Dependabot config](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file)**
  — supply‑chain hygiene in CI.

## Testing

- **[`integration/integration.go` @ v1.3.2](https://raw.githubusercontent.com/pulumi/pulumi-go-provider/v1.3.2/integration/integration.go)**
  — the exact harness this repo's pinned version exposes: `NewServer(ctx, pkg, version,
  …) (Server, error)`, the `Server` interface, `Operation`, `LifeCycleTest`,
  `WithProvider`/`WithMocks`. Pin to this when copying.
- **[`infer/tests/` (create/update/schema)](https://github.com/pulumi/pulumi-go-provider/tree/main/infer/tests)**
  — method‑level tests (`server.Create(p.CreateRequest{…})` — **no ctx arg**), inputs via
  `property.NewMap`/`property.New`, and a full‑schema `assert.JSONEq` snapshot.
- **[pulumi/providertest + pulumitest](https://github.com/pulumi/providertest)** —
  program‑level tests via `opttest.AttachProviderServer`, `PreviewProviderUpgrade` +
  recorded `grpc.json`. Incubating.
- **[pgregory.net/rapid](https://pkg.go.dev/pgregory.net/rapid)** — property‑based testing
  for the wide round‑trip mappers.

## Secrets, logging & diagnostics (Pulumi docs + source)

- **[Secrets](https://www.pulumi.com/docs/iac/concepts/secrets/)** (IDs are always
  plaintext), **[write‑only fields](https://www.pulumi.com/docs/iac/concepts/secrets/write-only-fields/)**,
  **[additionalSecretOutputs](https://www.pulumi.com/docs/iac/concepts/options/additionalsecretoutputs/)**,
  **[logging/credential exposure](https://www.pulumi.com/docs/support/debugging/logging/)**.
- **[`logging.go`](https://github.com/pulumi/pulumi-go-provider/blob/main/logging.go)**,
  **[`infer/errors.go`](https://raw.githubusercontent.com/pulumi/pulumi-go-provider/main/infer/errors.go)**
  (`ResourceInitFailedError`, `ProviderErrorf`), and
  **[pulumi#7156](https://github.com/pulumi/pulumi/issues/7156)** (stdout/gRPC handshake).

## UniFi‑specific field‑coverage prior art

The de‑facto specification for *which resources/fields* a UniFi IaC provider should cover.
Mirror their resource sets, field names, defaults, and validation.

- **[filipowm/terraform-provider-unifi](https://github.com/filipowm/terraform-provider-unifi)**
  — **closest parity reference**: it rides the **same `filipowm/go-unifi` client this
  provider depends on** (`go.mod`: `v1.8.1`), so its field handling maps almost 1:1. Adds
  beyond paultyng (DNS records, zone‑based firewalls, traffic management, more settings) —
  but **verify per‑resource against current docs**, since several are roadmap items, not
  all shipped.
- **[paultyng/terraform-provider-unifi](https://registry.terraform.io/providers/paultyng/unifi/latest/docs)**
  — the original (now archived) provider: `unifi_network`, `unifi_wlan`, `unifi_user`,
  `unifi_firewall_group`, `unifi_firewall_rule`, `unifi_port_forward`,
  `unifi_port_profile`, `unifi_static_route`, `unifi_user_group`, `unifi_device`,
  `unifi_dynamic_dns`, `unifi_account`, and `unifi_setting_*`. The baseline coverage map.
- **[filipowm/go-unifi](https://github.com/filipowm/go-unifi)** (used here) and the
  archived original **[paultyng/go-unifi](https://github.com/paultyng/go-unifi)** — the
  client structs are **code‑generated from JSON field definitions embedded in the UniFi
  Controller JAR**, so they are the authoritative source of UniFi field names/types and
  field coverage tracks the controller version. Maintained forks: `filipowm/go-unifi`,
  `ubiquiti-community/go-unifi`, `akerl/go-unifi`.
- **[ClifHouck/unified](https://github.com/ClifHouck/unified)** — the Protect V1
  integration client used here (pre‑1.0; pin and expect churn). No Terraform provider
  models Protect cameras, so there is no upstream naming to mirror — the surface is the
  writable fields of `CameraPatchRequest`.

## How this provider relates
`pulumi-unifi` already follows the strongest structural patterns: `infer` with
package‑derived tokens, `Args`/`State`/marker triad, `infer.Config` + `Configure`,
adoption model for `Device`/`Camera` (mirroring `examples/dna-store`), the schema‑builds
guard, `respectSchemaVersion` + `github://` distribution, OIDC npm publishing, and pinned
toolchain. The gaps measured against the rubric — `Check`/`Diff`/`SetDefault`/
not‑found handling, error wrapping, wide‑resource nesting, CI completeness, Registry docs
— are tracked in [`conformance.md`](./conformance.md).
