# Best practices for native Go Pulumi providers (`pulumi-go-provider` / `infer`)

> Scope: how to build, test, document, version, and ship a **native** (non‑bridged)
> Pulumi provider written in Go with the [`pulumi-go-provider`](https://github.com/pulumi/pulumi-go-provider)
> `infer` layer, where the schema is **derived from Go structs + `Annotate`** rather
> than hand‑written. This is the architecture of `pulumi-unifi`.

Each practice is tagged with a **Status**:

- **Framework** — backed by official Pulumi docs or `pulumi-go-provider` source; an API
  you can rely on. (Every API named below was independently fact‑checked against
  `pkg.go.dev`/source — see [Provenance & caveats](#provenance--caveats).)
- **Convention** — sound engineering practice for wrapper providers, but *not* a
  documented framework feature. Apply with judgment.

Cross‑references: [`quality-rubric.md`](./quality-rubric.md) turns these into a
pass/fail checklist; [`conformance.md`](./conformance.md) measures this repo against
them; [`prior-art.md`](./prior-art.md) catalogs the reference repos cited here;
[`adding-resources.md`](./adding-resources.md) is the day‑to‑day authoring guide.

---

## 1. Schema & framework

The `infer` layer reflects your Go types into a Pulumi Package Schema at
`GetSchema` time, then `pulumi package gen-sdk` turns that schema into per‑language
SDKs. The schema is the public contract; protect it deliberately.

### 1.1 Let tokens fall out of package layout; override only deliberately — *Framework*
`infer` derives each token as `pkg:module:Member`: `pkg` is the provider name,
`module` is the Go package the type lives in, `Member` is the struct name. Keep **one
Go package per logical Pulumi module** (`provider/network` → `unifi:network:*`,
`provider/protect` → `unifi:protect:*`). Types in the root package land in the
`index` module. To rename without moving code, use
`ProviderBuilder.WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{…})`; to
retoken one type, implement `Annotate` and call `a.SetToken("module", "Name")`; to
rename a resource without breaking existing state, add `a.AddAlias(module, name)`.
**Why:** tokens are the stable identity of every resource across all SDKs and stacks;
casual package reorgs silently rename tokens and break consumers.
Source: `pkg.go.dev/.../infer`, [schema reference](https://www.pulumi.com/docs/iac/guides/building-extending/packages/schema/).

### 1.2 Follow Pulumi casing: PascalCase types, camelCase properties — *Framework*
Resource/complex‑type names are PascalCase (`Vlan`, `FirewallZonePolicy`). Property
names in the schema are **camelCase, set by the `pulumi:"…"` struct‑tag value**
(`QuerierAddress string \`pulumi:"querierAddress"\``). The Go field stays exported
PascalCase; the tag controls the wire/SDK name. This is the one thing `infer` cannot
guess for acronym‑heavy fields — spell `igmpSnooping`, `dhcpV6Enabled`,
`ipv6RaPriority` explicitly so odd capitalization never leaks into SDKs.
Source: [schema reference](https://www.pulumi.com/docs/iac/guides/building-extending/packages/schema/).

### 1.3 Encode required/optional with pointers + `,optional`; embed Args in State — *Framework*
A field is **required** with a plain `pulumi:"name"` tag and a non‑pointer type;
**optional** with `pulumi:"name,optional"` **and** a pointer/slice/map type so absence
is representable. Embed the `Args` (input) struct inside the `State` (output) struct so
inputs round‑trip into outputs, and add controller‑assigned fields (IDs) only on
`State`. (Embedding is a recommended *Convention*, not enforced.)
**Why:** mismatched optionality (optional tag on a value type, or required on a pointer)
yields confusing diffs and lost values. Source: `pulumi-go-provider` README.

### 1.4 Document with doc‑comments **and** `Annotate` — describe fields on the owning struct — *Framework*
`infer` reads field descriptions from **Go doc‑comments automatically**, and
`Annotate(a infer.Annotator)` (pointer receiver) lets you add more:
`a.Describe(&x, "…")` for the resource/type, `a.Describe(&x.Field, "…")` for a field,
`a.Deprecate(&x.Field, "use Y")` for deprecations. **Field‑level `Describe` must live on
the struct that owns the field** (the Args/State/nested type) — never inside the empty
marker struct's `Annotate` (it has no fields and won't compile). A schema‑builds test
catches this whole class at `go test` time.
Source: `pkg.go.dev/.../infer`, README.
> Note: the two facets of this codebase's research disagreed on whether doc‑comments
> are read. They **are** — but treat doc‑comments as the primary doc source and reserve
> `Describe` for resource‑level prose and anything a comment can't express. Keeping both
> in sync is the safest default.

### 1.5 Express defaults with `SetDefault`, not just imperative `derefOr` — *Framework*
Declare user‑visible defaults in `Annotate` via
`a.SetDefault(&x.Field, value, env...)`. The value **must be a primitive** in the
Pulumi type system (no object/array defaults). The variadic `env...` sources the
default from environment variables — ideal for config/credentials
(`a.SetDefault(&c.Site, "default", "UNIFI_SITE")`). `SetDefault` values appear in the
schema/SDKs/Registry docs and are applied by `DefaultCheck`.
**Why:** a default buried only in your `Create`/`toUnifi` code makes the SDK *lie* about
behavior, and can cause perpetual diffs. Keep this **distinct** from values you inject
purely at the API boundary — e.g. "always‑send" bools whose upstream JSON lacks
`omitempty` and whose backend default is `true` (`derefOr(field, true)`); those are an
API‑write concern, not a schema default. Source: `pkg.go.dev/.../infer`.

### 1.6 Model closed value sets as real enums — *Framework*
For a fixed legal set, define a named type over a primitive (`EnumKind` is
`~string | ~float64 | ~bool | ~int`) and implement
`Values() []infer.EnumValue[T]`, giving each member a `Name`, `Value`, and
`Description`. `infer` emits a typed enum into the schema and **every** SDK, replacing
"magic strings" with IDE‑validated choices.
**Why:** turns typo‑class errors that only fail at apply time into compile/preview‑time
errors, and self‑documents allowed values. Real‑world example: `mbrav/pulumi-netbird`.
Source: `pkg.go.dev/.../infer`, [enum support blog](https://www.pulumi.com/blog/announcing-enum-support/).

### 1.7 Model nested objects as named, annotated structs — *Framework*
Represent sub‑objects and lists‑of‑objects as dedicated pulumi‑tagged Go structs
(`[]NetworkWanDhcpOption`, `*NetworkWanProviderCapabilities`), not `map[string]any`.
`infer` promotes them to named complex types; they can implement `Annotate` too. **Give
each a unique, descriptive name** — within one module, complex‑type names share a token
namespace and silently collide. In a shared package (like `network`), prefix nested
types (`Network*`/resource‑prefixed). Source: README.

### 1.8 Define provider config as an annotated struct; build clients in `Configure` — *Framework*
Register config with `.WithConfig(infer.Config(Config{}))`, using the same
`pulumi:`/`provider:"secret"` tags. The config type can implement `Annotate`,
`CustomConfigure` (a `Configure(ctx) error` that runs **once per process**, after the
receiver is hydrated), and even `CustomCheck`/`CustomDiff`. Build vendor clients in
`Configure` and stash them on **unexported (untagged)** fields so they never serialize
to state. Resources fetch config via `infer.GetConfig[Config](ctx)`.
**Why:** centralizes auth/endpoint handling and secret typing in one place; validating
here surfaces auth errors early. Source: `pkg.go.dev/.../infer`, README; examples
`configurable`, `credentials`.

### 1.9 Decompose wide resources by API boundary; nest by facet — *Convention*
When a resource grows to dozens of fields, there are two very different ways to tame it,
and choosing wrong reintroduces correctness bugs:

- **Split into separate resources — only along real API boundaries.** The Terraform AWS
  provider split `aws_s3_bucket` into `aws_s3_bucket_versioning`,
  `aws_s3_bucket_lifecycle_configuration`, etc. That works because S3 exposes an
  **independent write endpoint per facet** (`PutBucketVersioning`, …), so each child
  owns its own endpoint and attaches by `bucket` id. Apply this pattern **only** when the
  backend genuinely has a separate endpoint per facet. A separate resource that has to
  read‑modify‑write the **same** backend object as another resource creates *multiple
  owners of one mutable object* → concurrent RMW races, non‑deterministic diffs, and
  undefined delete semantics. (UniFi's `networkconf`/`device` are single objects with a
  single write endpoint, so a `Vlan` + `VlanDhcp` split would hit exactly this hazard;
  `PortProfile`↔`Device.portProfileId` is the *correct* kind of split because the profile
  has its own endpoint.)
- **Nest fields into named complex types — the default for single‑object resources.**
  Group related flat fields into sub‑objects (`dhcp`, `ipv6`, `wan`, `igmp`) per §1.7.
  This keeps *one resource = one object = one owner*, slashes top‑level field count, and
  improves SDK/docs ergonomics — all without new write‑surface races. Cost: nested objects
  can't carry schema `SetDefault` defaults (primitive‑only, §1.5), so defaults stay in the
  mapper, and optionality on nested pointers needs care.

Either change reshapes the schema. If the provider already has users, that means a major
version bump + state migrations (§2.10) + aliases (§1.1), so do it early; if it's
unreleased, it's a free change. (The AWS S3 v4 split was contentious precisely because it
landed post‑1.0 with painful migrations.) Source: convention, informed by the Terraform
AWS S3 v4 refactor and UniFi's single‑object API shape.

### 1.10 Set full display + language metadata on the builder — *Framework*
Populate `WithDisplayName`, `WithDescription` (include the package name),
`WithHomepage`, `WithRepository`, `WithPublisher`, `WithLicense`, and for Registry
discoverability `WithKeywords(…)` and `WithLogoURL(…)`. Map languages with
`WithLanguageMap` — for Node.js, `packageName` (your own scope, since you don't own
`@pulumi`) and `respectSchemaVersion: true`; for Go, `WithGoImportPath`. Pair with
`WithPluginDownloadURL` so the CLI can auto‑install the binary.
**Registry keyword form is `category/<name>` and `kind/native`** (valid categories:
`cloud, database, infrastructure, monitoring, network, utility, versioncontrol`) — bare
`network`/`native` are *not* recognized facets. Source: `pkg.go.dev/.../infer`,
[publishing‑packages](https://www.pulumi.com/docs/iac/guides/building-extending/packages/publishing-packages/).

---

## 2. Resource lifecycle correctness

`infer` resources are an empty controller struct (`type Vlan struct{}`) carrying CRUD
methods, plus `Args`/`State` structs. **Only `Create` is mandatory**; add
`Read`/`Update`/`Delete`/`Diff`/`Check` when the default is wrong.

### 2.1 Honor `DryRun` and return planned output during preview — *Framework*
Branch on `req.DryRun` at the top of `Create`/`Update`: do **no** side effects and
return a populated response whose `Output` mirrors inputs (plus any state you can fill).
`infer` encodes preview output with `AllowUnknown(req.DryRun)`, so values you can't know
yet (server IDs) surface as unknowns. Leave genuinely‑unknown fields unset — unknownness
comes from what you *don't* populate, not an automatic per‑field heuristic.
**Why:** mutating the backend during `DryRun` would have `pulumi preview` create real
resources; an empty preview output makes the plan garbage. Source: README.

### 2.2 Implement `Read` for `pulumi import` and refresh — *Framework*
`Read(ctx, ReadRequest[I,O]) (ReadResponse[I,O], error)` fetches live state by
`req.ID` and returns `{ID, Inputs, State}`. The same method backs **refresh** (inputs
/state pre‑populated from the checkpoint) and **import** (only `req.ID` is set, so `Read`
must reconstruct `Inputs` purely from the fetched object). The default `Read` only
re‑validates that stored inputs/state deserialize — insufficient for import.
**Caveat:** import passes empty prior inputs, so any field the controller blanks (e.g. a
PPPoE/WLAN password) is empty after import — document that secrets must be re‑supplied.
Source: `pkg.go.dev/.../infer`; example `dna-store`.

### 2.3 Implement `Update`, or every change forces a replace — *Framework*
`infer` decides replace‑vs‑update by reflection: `_, hasUpdate := any(*r).(CustomUpdate[I,O])`.
**No `CustomUpdate` ⇒ the default `Diff` marks every changed property as a replace.**
Implement `Update(ctx, UpdateRequest[I,O]) (UpdateResponse[O], error)` and return the
freshly‑read `Output` so the checkpoint reflects server‑applied values.
**Why:** a missing `Update` silently turns benign edits (rename, toggle a flag) into
destroy+recreate of stateful infra. Source: `infer/resource.go`.

### 2.4 Force replacement on immutable fields — *Framework*
For inputs the backend can't change in place, tag them
`provider:"replaceOnChanges"` next to the `pulumi:` tag — `infer`'s `introspect.ParseTag`
recognizes it, and the default `Diff` emits a replace for that property (in the branch
taken when `CustomUpdate` exists). Alternatively implement `Diff` and return a replacing
`DiffKind`. The user‑side `replaceOnChanges` resource option is the consumer complement
for when the *provider* can't know; prefer provider‑side tagging when it does.
**Why:** an immutable field allowed to "update" fails at apply or silently no‑ops.
Typical candidates: bind‑key/identity fields (a camera ID, a device MAC).
Source: `infer/introspect.go`. (Note: one research facet claimed `replaceOnChanges` is
*only* a user‑side option — that's wrong; the `provider:"replaceOnChanges"` tag is real
and confirmed in `introspect.go`.)

### 2.5 Use a custom `Diff` only when the structural default is wrong — *Framework*
The default diff is structural equality over **inputs** — prefer it. Implement
`Diff(ctx, DiffRequest[I,O]) (DiffResponse, error)` when you need finer control: build a
`DetailedDiff map[string]PropertyDiff` with `DiffKind` constants
(`Add, AddReplace, Delete, DeleteReplace, Update, UpdateReplace, Stable`), set
`HasChanges`, and set `DeleteBeforeReplace: true` when the backend can't tolerate two
instances during a replace (a uniqueness constraint on name/VLAN‑ID). `DeleteBeforeReplace`
is a **`DiffResponse` field**, not a `DiffKind`. `PropertyDiff.InputDiff` distinguishes
input‑vs‑input from state changes. Use it also to suppress spurious diffs on
controller‑managed/computed fields. Source: `pkg.go.dev/pulumi-go-provider`; example `file`.

### 2.6 Validate & normalize inputs in `Check`, layered on `DefaultCheck` — *Framework*
Implement `Check(ctx, CheckRequest) (CheckResponse[I], error)`: call
`infer.DefaultCheck[I](ctx, req.NewInputs)` first (decode + apply `SetDefault`), then
append `p.CheckFailure{Property, Reason}` for domain validation. **Return failures, not a
Go error**, for user‑input problems — they render as per‑property diagnostics at preview
time. `Check` is also where you canonicalize values (lowercase MACs, trim CIDRs) so they
don't oscillate against controller‑normalized forms. `CheckRequest.NewInputs` is a
`property.Map` in v1.x; `infer.CheckResponse[I]` carries typed `Inputs I` (distinct from
top‑level `p.CheckResponse`). Source: `pkg.go.dev/.../infer`; example `file`.

### 2.7 Preserve write‑only/secret & server‑defaulted fields across round‑trips — *Convention*
When mapping a fetched object back into state, reconcile each optional field against the
prior input: reflect the backend value when meaningfully set, else fall back to prior.
This (a) keeps write‑only secrets the backend never echoes (passwords) stable, and (b)
avoids surfacing server defaults as user‑set inputs that diff forever. The default diff
compares **inputs**, so a server default that lands only in output‑state generally won't
diff — the danger is copying a server default *into an input field*. The
"reflect‑if‑nonzero‑else‑prior" technique itself is project convention (validate
per‑field); the `provider:"secret"` tag and input‑based diff behavior are the framework
parts. **This is the single biggest source of "`pulumi up` always shows changes" bugs in
wrapper providers.** Source: framework parts in `infer/introspect.go`.

### 2.8 Keep operations idempotent; tolerate already‑gone on `Delete` — *Convention*
`Create` must return a **stable** `CreateResponse.ID` that `Read`/`Update`/`Delete` can
resolve forever — never derive it from a mutable field (use the controller `_id`). Make
`Delete` idempotent: if the backend reports the object already gone (404), return
`DeleteResponse{}, nil` (optionally `Warningf`) so an interrupted destroy can retry.
**Why:** non‑idempotent delete wedges stacks; unstable IDs break refresh/import.
Source: `DeleteResponse{}`/`CreateResponse.ID` are framework; idempotency is convention;
example `file` shows the 404‑to‑warning pattern.

### 2.9 Adoption model: `Create` adopts, `Delete` is a no‑op — *Convention*
For resources that represent pre‑existing hardware/objects the API can't create or
destroy: `Create` verifies the target exists (clear error if not) and applies desired
settings; `Update` does a **read‑modify‑write** (fetch live, overlay only managed fields,
merge keyed sub‑collections by key, write back); `Read` binds by ID; `Delete` is an
intentional no‑op that only releases Pulumi management. Tag the bind key
`provider:"replaceOnChanges"`. Document that the engine‑level `retainOnDelete` option is
the consumer equivalent. **Note:** `infer`'s default `Delete` is *already* a no‑op when
`CustomDelete` is unimplemented, so an explicit no‑op is documentation/intent, not a
requirement. Source: example `dna-store` (Read); [retainOnDelete docs](https://www.pulumi.com/docs/iac/concepts/resources/options/retainondelete/).

### 2.10 Evolve state shape with state migrations — *Framework*
When you change a resource's state shape across schema versions, implement
`CustomStateMigrations[O]` returning `StateMigrations(ctx) []StateMigrationFunc[O]`
(use the `infer.StateMigration` helper). This is the supported mechanism for
non‑breaking state evolution. Source: `pkg.go.dev/.../infer`.

---

## 3. Secrets & sensitive data

### 3.1 Tag secrets with `provider:"secret"` — *Framework*
Add `provider:"secret"` alongside `pulumi:` on any credential/sensitive field, on both
the `Config` struct and resource inputs. `infer` emits `"secret": true` in **both**
`properties` and `inputProperties`; the engine then encrypts it in state and masks it as
`[secret]` in CLI output — **you write no encryption code**. Without the tag the value is
plaintext in state and printed in previews/diffs. Source: README, `infer/tests/schema_test.go`,
[secrets docs](https://www.pulumi.com/docs/iac/concepts/secrets/).

### 3.2 Let secretness flow to outputs; refine with `WireDependencies` — *Framework*
Default: if a resource doesn't implement `ExplicitDependencies`, **every output depends
on every input**, so a secret input taints all outputs (the deliberate safe default).
To refine, implement `WireDependencies(f FieldSelector, args *I, state *O)` and use
`f.OutputField(&state.X).DependsOn(f.InputField(&args.Y))`, plus `.AlwaysSecret()`,
`.NeverSecret()`, `.AlwaysKnown()` and `InputField.Secret()`/`.Computed()`. Narrow only
deliberately — wrong wiring either leaks a secret or over‑marks innocuous outputs.
Source: `pkg.go.dev/.../infer`, issue #66/PR #69.

### 3.3 Mark provider‑generated secret outputs `AlwaysSecret()` — *Framework*
For an output the provider produces that is sensitive but **doesn't derive from a secret
input** (a server‑generated token/voucher/key), mark it
`f.OutputField(&state.Token).AlwaysSecret()` — secret‑input taint can't cover it. This is
the provider‑side, always‑on analog of the user‑facing `additionalSecretOutputs` option
(which only marks top‑level properties). Don't rely on consumers remembering the option.
Source: `pkg.go.dev/.../infer`, [additionalSecretOutputs](https://www.pulumi.com/docs/iac/concepts/options/additionalsecretoutputs/).

### 3.4 Handle write‑only credentials the backend never echoes — *Framework (concept) / Convention (handling)*
For secrets the backend accepts on write but never returns on read, Pulumi's
**write‑only fields** model says the value is stored as a secret in state *inputs* only,
"will never appear in state outputs," and "on subsequent Read operations, the value will
be set to null"; a companion `version`/rotation field forces re‑apply. In a hand‑written
`infer` `Read`/`Diff`, **do not overwrite the stored secret input with the API's
empty/nil value** or you get a perpetual diff / wipe the credential — preserve it from
prior (as in 2.7). *Caveat:* write‑only fields are a Pulumi framework concept; explicit
`infer` SDK support for them was **not** confirmed — treat the prior‑preservation handling
as the reliable approach. Source: [write‑only fields](https://www.pulumi.com/docs/iac/concepts/secrets/write-only-fields/).

### 3.5 Validate auth in `Configure`; keep secrets out of IDs, logs, errors — *Framework*
Validate credentials and mutually‑exclusive auth modes in `Configure` and fail fast with
a clear, **value‑free** message (field name + URL, never the key/password). For missing
required keys, `p.ConfigMissingKeys(map[string]string)` builds a standardized error.
Never compose a resource **ID** out of a secret — IDs are "always stored in plain text in
the state file and cannot be encrypted." Never interpolate a secret into a log or error
string: tagging encrypts state, but the value is still plaintext in Go memory, and
Pulumi's own verbose logging "at log level 11 … will intentionally expose some known
credentials." Source: [secrets docs](https://www.pulumi.com/docs/iac/concepts/secrets/),
[logging docs](https://www.pulumi.com/docs/support/debugging/logging/), `provider.go`.

---

## 4. Versioning, packaging & publishing

### 4.1 Drive one SemVer end‑to‑end from the git tag — *Framework*
Pick one valid SemVer per release and make it flow through **every** artifact: the
`vX.Y.Z` tag, the goreleaser ldflags stamp, the version the plugin reports (the third arg
to `p.RunProvider`/`p.Run`), the schema `version` (required, "must be valid semver"), the
release‑asset filename, and every SDK package version. Derive them all from the tag —
never hand‑edit versions in two places.
**Why:** if the reported version, asset filename, and SDK version drift, the `github://`
resolver 404s at install or consumers hit version‑mismatch failures.
Source: [schema reference](https://www.pulumi.com/docs/iac/guides/building-extending/packages/schema/).

### 4.2 Set `respectSchemaVersion: true` so SDK versions track the schema — *Framework*
In the per‑language config, `respectSchemaVersion: true` means "use the package.version
field in the generated SDK." Set it via `WithLanguageMap` (there is **no** dedicated
builder method). **Build the version‑stamped binary first, then run `gen-sdk` against it**
so the schema already carries the release version. Without it, SDKs get a placeholder/dev
version and Pulumi can't reliably auto‑acquire the matching plugin.
Source: [schema reference](https://www.pulumi.com/docs/iac/guides/building-extending/packages/schema/).

### 4.3 Honor the binary + asset naming exactly — *Framework*
The binary **must** be `pulumi-resource-<name>` (`.exe` on Windows); Pulumi discovers
plugins by this convention, so `cmd/pulumi-resource-<name>/` and goreleaser's `binary:`
are load‑bearing. Each release asset **must** be
`pulumi-resource-<name>-v<version>-<os>-<arch>.tar.gz` with the binary at the **root**
(no wrapping directory). The filename carries a leading `v`; the binary's reported
version does not. Source: [executable‑plugin docs](https://www.pulumi.com/docs/iac/guides/building-extending/packages/executable-plugin/).

### 4.4 Embed `pluginDownloadURL` with the `github://` resolver — *Framework*
`WithPluginDownloadURL("github://api.github.com/<org>/<repo>")` makes a consumer's first
`pulumi up` auto‑acquire the plugin from your GitHub Releases (omit the repo segment and
it defaults to `pulumi-<name>`). For non‑GitHub hosting, point at a base URL and serve
`${url}/pulumi-resource-<name>-v<version>-<os>-<arch>.tar.gz`. Source:
[executable‑plugin docs](https://www.pulumi.com/docs/iac/guides/building-extending/packages/executable-plugin/).
*Unconfirmed in fetched docs:* the exact `pulumi plugin install resource <name> <ver>
--server github://…` string, and that consumer‑side `GITHUB_TOKEN` dodges anonymous rate
limits — both plausible, neither doc‑verified.

### 4.5 Cross‑compile with goreleaser, CGO disabled — *Framework*
Build with `CGO_ENABLED=0` so the provider cross‑compiles freely, covering Pulumi's
matrix: `{linux,darwin,windows} × {amd64,arm64}` (six combos; Windows/arm64 sometimes
dropped). Use `ldflags -s -w` + `-X <module>/.../version.Version={{.Version}}`,
`mod_timestamp` for reproducibility, `tar.gz` archives with no wrapping dir. Run
`goreleaser release --clean` in CI to build, archive, and create the Release in one step.
**Why:** a CGO build needs per‑platform toolchains; a missing os/arch is a 404 at install
with no fallback. Source: boilerplate `.goreleaser.yml`.

### 4.6 Regenerate SDKs from the stamped binary — never hand‑edit — *Framework*
Treat SDKs as build output: at release, build the stamped binary, then
`pulumi package gen-sdk ./bin/pulumi-resource-<name> --language <lang> --out sdk`.
Use `pulumi package get-schema ./bin/…` to inspect. If you commit SDKs, regenerate and
commit them at the tag so the checked‑in SDK never lags the schema. Hand‑edits break on
the next regen. Source: README, [executable‑plugin docs](https://www.pulumi.com/docs/iac/guides/building-extending/packages/executable-plugin/).

### 4.7 Publish the npm SDK under your own scope with `--access public` — *Framework*
You can't publish under `@pulumi`. Set `packageName` ("Custom name for the NPM package")
to your own scope, keeping a `pulumi` signal (`@scope/pulumi-<name>`). The first publish
of a scoped package needs `--access public` (else it's restricted/private). Mirrors
`pulumiverse` (`@pulumiverse/<name>`). Source: [schema reference](https://www.pulumi.com/docs/iac/guides/building-extending/packages/schema/).

### 4.8 Use npm trusted publishing (OIDC) + provenance, not long‑lived tokens — *Framework*
Configure a Trusted Publisher on npmjs.com (provider = GitHub Actions; org, repo,
workflow filename, optional environment) and give the job `permissions: id-token: write`.
From a public repo, npm mints short‑lived OIDC credentials and **auto‑generates
provenance** — no `NODE_AUTH_TOKEN`, no `--provenance` flag. (Token fallback needs
`--provenance`, npm ≥ 9.5.0, and a case‑sensitive `repository` match in package.json.)
**Why:** long‑lived tokens are the top supply‑chain leak vector; OIDC removes the secret
and proves the package was built from your repo. Source: [trusted‑publishers](https://docs.npmjs.com/trusted-publishers).

### 4.9 Make releases tag‑driven, binaries before SDKs — *Framework*
Trigger on a `v*` tag push. Publish **plugin binaries/Release assets first** (goreleaser),
**then** SDKs, because consumers (and smoke tests) resolve the plugin from the
just‑created Release; SDK‑first risks a missing‑plugin race. Scope tokens per job
(`contents: write` for the Release job; `contents: read` + `id-token: write` for npm).
For multi‑language fan‑out, `pulumi/pulumi-package-publisher` publishes prebuilt
`<lang>-sdk.tar.gz` to npm/PyPI/NuGet/Maven via an `sdk:` input; `pulumi/publish-go-sdk-action`
pushes a path‑prefixed Go module tag (`sdk/v<version>`). Source:
[package‑publisher](https://github.com/pulumi/pulumi-package-publisher), [publish‑go‑sdk‑action](https://github.com/pulumi/publish-go-sdk-action).

---

## 5. Testing

A right‑sized pyramid for a community wrapper provider (each tier wired into CI as noted):

### 5.1 Schema‑builds guard — cheapest, highest leverage — *Framework*
Construct the provider, `integration.NewServer(ctx, Name, version, WithProvider(prov))`,
call `server.GetSchema(p.GetSchemaRequest{})`, and assert: non‑empty schema, every
expected token present, credential fields marked `secret`, and field‑count floors for
wide resources. The strongest form adds a full‑schema `assert.JSONEq` **snapshot** (as
`infer/tests/schema_test.go` does) so any unintended schema drift fails CI and forces a
deliberate SDK regen. Source: `integration/integration.go`, `infer/tests/schema_test.go`.

### 5.2 Drive full CRUD in‑process with `integration.LifeCycleTest` — *Framework*
`LifeCycleTest{Resource, Create, Updates}.Run(t, server)` performs Check → preview
Create → real Create → per‑Update Check+Diff (honoring `DetailedDiff`/`DeleteBeforeReplace`)
→ Delete, entirely at the gRPC boundary without a real engine. `Operation` supports
`Inputs`, `ExpectedOutput`, `Hook`, `ExpectFailure`, `CheckFailures` (for negative tests).
This exercises the pipeline pure mapping tests don't: Check, default application, secret
round‑tripping, diff/replace computation. Source: `integration/integration.go`.

### 5.3 Call gRPC methods directly for targeted Diff/Check assertions — *Framework*
For sharp assertions on one operation, call the integration `Server`'s methods —
`server.Create(p.CreateRequest{…})`, `server.Update(…)`, `server.Diff(…)`,
`server.Check(…)` — with testify. **The `Server` methods do not take a `ctx`** (it's bound
at `NewServer`). Ideal for replace‑vs‑update logic, preview unknowns, `Check`
default‑filling, and `CheckFailures`. Source: `infer/tests/create_test.go`.

### 5.4 Build inputs with the `property` package — *Framework*
In v1.x, `Operation.Inputs` is `property.Map`
(`github.com/pulumi/pulumi/sdk/v3/go/property`). Build with
`property.NewMap(map[string]property.Value{"vlan": property.New(30.0)})`, numbers as
**float64**, secrets via `property.New(…).WithSecret(true)`. `resource.PropertyMap` is
stale for v1. Source: example `random-login/main_test.go`.

### 5.5 Keep mapping pure and round‑trip test it — *Convention*
Factor vendor mapping into pure functions (`args.toUnifi(id)`, `stateFrom(obj, prior)`)
and table‑test them: identity fields and optionals survive the round‑trip, and a
controller‑echoed zero with a nil prior **stays nil** (no spurious diff). Pure functions
need no network and run in milliseconds, so cover wide structs exhaustively. This is the
foundation layer beneath 5.1–5.2. Source: general Go practice.

### 5.6 Design for client injection so lifecycle tests are hermetic — *Convention*
Build the vendor client behind an interface and allow a fake to be substituted in tests
(a constructor seam or unexported field the test package sets). **Important:**
`integration.WithMocks` mocks only the Pulumi resource monitor (child/component
resources) — it does **not** intercept your outbound vendor HTTP calls. So a client fake
is the *only* way to keep custom‑resource `LifeCycleTest`s hermetic. Source: design
recommendation; `random-login` confirms the `WithMocks` scope.

### 5.7 Optional tiers: component mocks, program tests, property tests — *Mixed*
- **Components:** test infer components with `integration.WithMocks(mocks)` implementing
  `NewResourceF`/`CallF` (example `random-login`). *Framework.*
- **Program/E2E:** keep runnable example programs and drive them with
  `pulumi/providertest`'s `pulumitest` + `opttest.AttachProviderServer` (in‑process via
  `PULUMI_DEBUG_PROVIDERS`); `PreviewProviderUpgrade` + recorded `grpc.json` catches
  upgrade/spurious‑diff regressions. Gate behind a build tag / env var. *Framework
  (incubating).*
- **Property‑based:** fuzz the pure round‑trip with `pgregory.net/rapid` (auto‑shrinks)
  — strong fit for 80+‑field structs. *Convention.*

### 5.8 Right‑size and gate the pyramid — *Convention*
Tier 1 (schema guard + pure round‑trip) and Tier 2 (in‑process lifecycle against a fake)
in CI on every PR; Tier 3 (program/live) behind a build tag, run on demand/nightly. Keep
schema/SDK regeneration in the build gate so a schema change can't merge without
regenerated SDKs. A solo provider can't maintain cloud‑grade live matrices — concentrate
on the deterministic tiers.

---

## 6. Documentation, examples & UX

### 6.1 Ship the Registry‑required docs files — *Framework*
A community package needs a `docs/` directory with **`docs/_index.md`** (overview +
purpose + a usage code sample per language, rendered as the package index) and
**`docs/installation-configuration.md`** (links to published SDKs + provider
configuration/credentials). Per‑resource API pages are generated from schema
descriptions. *Conventional (not doc‑verified) front matter:* `title`, `meta_desc`,
`layout: package` — present in real provider `_index.md` files. Source:
[registry README](https://github.com/pulumi/registry), [publishing‑packages](https://www.pulumi.com/docs/iac/guides/building-extending/packages/publishing-packages/),
real example `pulumi-azuredevops/docs/_index.md`.

### 6.2 Author multi‑language examples with the chooser — *Framework*
In `_index.md`, wrap the headline example in `{{< chooser language "typescript,python,go,csharp,yaml" >}}` …
`{{% choosable language typescript %}}` blocks. For reusable programs, use
`{{< example-program path="name" >}}` (auto‑builds the chooser). Directory naming is
`<program-name>-<language>`. *Unverified:* that example programs must live under
`examples/` **in the provider repo** (conventionally a separate examples repo). Source:
[CODE-EXAMPLES.md](https://github.com/pulumi/docs/blob/master/CODE-EXAMPLES.md).

### 6.3 Populate Registry browse/search metadata — *Framework*
Beyond the display metadata (1.9), ensure `keywords` (`category/<name>`, `kind/native`)
and a web‑accessible `logoUrl`; include the package name in the description. These drive
the browse page and search filters. Source: [publishing‑packages](https://www.pulumi.com/docs/iac/guides/building-extending/packages/publishing-packages/).

### 6.4 Field‑level docs are the biggest doc‑quality lever — *Framework*
Every field a consumer sees in SDK hover/IntelliSense and Registry API pages comes from
its description. Whether via doc‑comments (which `infer` reads) or `Describe`, **no field
should be undocumented**; for enum‑like strings, restate allowed values. Keep
descriptions self‑contained sentences. Source: `pkg.go.dev/.../infer`.

### 6.5 Document the full consume flow — *Framework*
Cover: released use (`npm add @scope/pulumi-<name>`, first `up` auto‑downloads the
plugin), explicit pin (`pulumi plugin install resource <name> <ver> --server github://…`),
local/unpublished use (`pulumi package add <path>/bin/pulumi-resource-<name>`), and
configuring credentials both as `new unifi.Provider(...)` and ambient
`pulumi config set unifi:apiKey --secret`. Add a contributor `get-schema`/`gen-sdk`
pointer. Source: [pulumi package add](https://www.pulumi.com/docs/iac/cli/commands/pulumi_package_add/).

---

## 7. CI/CD & repository hygiene

### 7.1 Run a complete CI gate on every PR — *Framework*
Gate: `go build ./...`, `go vet ./...`, `golangci-lint run`, `go test ./...`, a
**`go mod tidy` drift check** (`go mod tidy` then `git diff --exit-code go.mod go.sum`),
and a **generated‑SDK freshness check** (regenerate, then `git diff --exit-code` the SDK
dir). Make lint + test required status checks on the protected branch. Keep the local
pre‑commit gate identical to CI (one task that runs them all). Source: boilerplate
`Makefile`; the tidy/SDK‑freshness diff checks are standard, added explicitly (not in the
boilerplate's `ensure` target). Pin the Go version used for the tidy check — different Go
versions produce different `go.mod`/`go.sum`.

### 7.2 Curated golangci‑lint v2 config with security/quality linters — *Framework*
Pin golangci‑lint and ship a `version: "2"` `.golangci.yml`. Beyond the Go defaults
(errcheck, govet, staticcheck, ineffassign, unused), the boilerplate enables `goconst,
gosec, lll, misspell, nakedret, revive, unconvert` plus formatters `gci`/`gofmt`, and
excludes generated paths (`schema.go`, `pulumiManifest.go`, vendored/SDK/examples).
**`gosec` is especially valuable for a provider handling credentials.** Run `--fix`
locally; run without `--fix` in CI so it fails on findings. Source: boilerplate `.golangci.yml`.

### 7.3 Lock the toolchain and pin dependencies — *Framework*
Make builds reproducible: (1) Go — `go 1.x` directive (minimum) + a separate `toolchain
go1.x.y` directive (exact); with `GOTOOLCHAIN=auto` the go command fetches exactly that.
(2) Commit `go.sum`. (3) Pin all external tools (go, pulumi, golangci‑lint, goreleaser,
node) to exact versions via a version manager and wire the same versions into CI.
Source: [Go toolchain docs](https://go.dev/doc/toolchain).

### 7.4 Add vulnerability scanning + Dependabot — *Framework*
Run `govulncheck ./...` in CI (call‑graph reachability against `vuln.go.dev`; the Go team
recommends it in CI). Use the binary or the action with **`output-format text`** so the
job actually fails on findings (json/sarif return success and only feed code scanning).
Add `.github/dependabot.yml` (v2) with `gomod` and `github-actions` ecosystems on a weekly
schedule. High value here given the crypto/HTTP‑client dependency graph. Source:
[govulncheck-action](https://github.com/golang/govulncheck-action), [Dependabot config](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file).

### 7.5 LICENSE file + per‑file SPDX headers — *Framework*
Ship a root `LICENSE` (without it the repo is all‑rights‑reserved and not legally
reusable, blocking adoption and Registry listing). Add a per‑file header — full Apache
block or the terser `// SPDX-License-Identifier: Apache-2.0` — consistently across
hand‑written source. **A goreleaser `archives.files: [LICENSE]` entry will fail the
release if `LICENSE` is missing.** Source: Pulumi source headers, boilerplate.

### 7.6 Changelog / release notes tied to the tag — *Framework*
Either generate notes from commit history at release (goreleaser `changelog` block with
Conventional‑Commit‑friendly filters) and/or keep a curated `CHANGELOG.md` (as
pulumi‑kubernetes does). For a solo repo, goreleaser/GitHub Release notes usually suffice
— but then **enforce Conventional Commits** so the filters and any `svu`‑style version
bump work. Source: [goreleaser changelog](https://goreleaser.com/customization/changelog/).

### 7.7 Follow the boilerplate layout, one task runner — *Convention*
Layout: `provider/` (impl) with entrypoint `cmd/pulumi-resource-<name>/main.go`, `sdk/`
(generated), `examples/` (runnable programs / CI smoke tests), `.github/`, plus root
`.goreleaser.yml`, `.golangci.yml`, and a tool‑version file. Centralize dev/CI commands in
one runner (Makefile in the boilerplate; a `mise.toml` task table is an equivalent
reproducible substitute) so local and CI never drift. Consider CODEOWNERS, a PR template,
SHA‑pinned actions, and branch protection. Source: boilerplate `Makefile`.

---

## 8. Errors, logging & diagnostics

### 8.1 Wrap vendor errors with `%w` + actionable context — *Framework*
Every CRUD/`Check`/`Diff`/`Configure` method returns a Go error; wrap vendor‑client errors
with `fmt.Errorf("…: %w", err)` (so `errors.Is/As` still work) and prefix with the
operation + identifying inputs (resource name/ID, resolved site, URL). A bare `401` or
`unexpected end of JSON` forces users to guess what failed. Returning non‑nil fails that
step; returning nil from Create/Update commits the returned `Output` to state.
Source: example `file`.

### 8.2 Log through `p.GetLogger(ctx)`, never `fmt.Print` to stdout — *Framework*
`logger := p.GetLogger(ctx)` gives `Debug/Info/Warning/Error` (+ `…f` and `…Status`/`…Statusf`
transient‑progress variants), routed to the engine and attached to the resource URN.
`GetLogger` falls back to an slog sink with no engine, so it's safe in tests. Use the
`Status` variants for progress text. Source: `logging.go`.

### 8.3 Keep stdout pristine; silence chatty vendor loggers — *Framework*
A plugin writes its listening port to **stdout** for the engine handshake; any other
bytes there (a vendor lib's default logger, gRPC's logger, a stray `Println`) get parsed
as the port and crash startup. Redirect noisy loggers to stderr or `io.Discard` (e.g.
`grpclog.SetLoggerV2`, or a discarding logrus). Prefer `io.Discard` for truly‑internal
chatter. Source: [pulumi#7156](https://github.com/pulumi/pulumi/issues/7156).

### 8.4 Signal partial creation with `ResourceInitFailedError` — *Framework*
If the backend object is created but a follow‑up step fails, return your best‑effort
`Output` **together with** `infer.ResourceInitFailedError{Reasons: […]}` (maps to
top‑level `p.InitializationFailed`). The framework commits the returned state, records the
resource as created (so it isn't orphaned), and makes the next op an `Update`. `Read`
honors it too. Its `.Error()` is generic, so put the real detail in `Reasons`. Most
relevant to read‑modify‑write adoption resources where the entity exists but a settings
write fails. Source: `infer/errors.go`.

### 8.5 Distinguish provider bugs from user errors — *Framework*
For impossible states / invariant violations (a successful create that returns no ID, an
unhandled enum), return `infer.ProviderErrorf(…)` (or top‑level `p.InternalErrorf(…)`),
which renders "please report this to the provider author." Use these in OUT mappers and
read‑modify‑write merges; reserve plain wrapped errors for "fix your inputs/credentials."
Source: `infer/errors.go`, `provider.go`.

### 8.6 Make `Delete` idempotent (404 → warning → nil) — *Framework (pattern)*
On an already‑gone object at delete time, `p.GetLogger(ctx).Warningf("%q already deleted")`
and return success rather than erroring (the `file` example does exactly this). For
adoption resources, `Delete` is a no‑op. Source: example `file`.

---

## Provenance & caveats

This guide was synthesized from a fan‑out of web research, then **adversarially
fact‑checked**: each claimed API/flag was re‑verified against `pkg.go.dev` and
`pulumi-go-provider` source, and unverifiable claims were demoted or flagged. Key
corrections folded in above:

- **`provider:"replaceOnChanges"` is a real provider‑side struct tag** (confirmed in
  `internal/introspect/introspect.go`), in addition to the user‑side resource option of
  the same name. (One research pass wrongly called it user‑side only.)
- **`DeleteBeforeReplace` is a `DiffResponse` field**, not a `PropertyDiff.Kind` value.
- **`SetDefault`'s default value must be a primitive** Pulumi type (no object/array
  defaults); the variadic trailing args are env‑var names.
- **`respectSchemaVersion` has no builder method** — set it inside `WithLanguageMap`.
- **Write‑only fields** are a Pulumi *concept*; explicit `infer` support was not confirmed
  — rely on prior‑value preservation in `Read`.
- **`integration.Server` methods don't take a `ctx`** argument (bound at `NewServer`).
- Unverified‑in‑docs items are flagged inline (the exact `plugin install --server`
  string; consumer `GITHUB_TOKEN` rate‑limit behavior; in‑repo `examples/` requirement;
  `_index.md` front‑matter keys).

All APIs named without a caveat were confirmed present in `pulumi-go-provider` (the repo
pins **v1.3.2**; pin version in any copied example, since `integration.NewServer` and
`ProviderBuilder.Build()` signatures have shifted across releases).
