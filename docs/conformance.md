# Conformance report — pulumi-unifi vs. the quality rubric

How this provider scores against [`quality-rubric.md`](./quality-rubric.md). The
original baseline was produced by a multi-agent review that graded each dimension
against the code with `file:line` evidence and adversarially re-verified every
non-pass finding. This revision reflects the **post-remediation** state after the
P0/P1/P2 roadmap was worked through — most findings are now resolved; the few
remaining are listed with rationale.

**Verdicts:** ✔ pass · ✱ partial · ✘ fail · — n/a. Severity: 🔴 gate · 🟡 expected · 🟢 polish.

## Scorecard

| Dim | Area | 🔴 gates | 🟡 expected | 🟢 polish |
|-----|------|----------|-------------|-----------|
| A | Schema & framework | ✔ (4/4) | ✔ defaults, ✔ enums, ✱ builder logo | — |
| B | Resource lifecycle | ✔ (replaceOnChanges, drift-read, §2.7, Check, idempotent delete) | ✔ | — n/a |
| C | Secrets | ✔ (4/4, incl. config-secret assertions) | ✔ | — |
| D | Versioning & publishing | ✔ (5/5) | ✔ (`needs: binaries` added) | — |
| E | Testing | ✔ (golden snapshot, client seam + LifeCycleTests) | ✔ | ✱ fuzz/pulumitest |
| F | Docs & UX | ✔ (registry docs, field descriptions) | ✱ examples dir | — |
| G | CI & hygiene | ✔ (shared gate, SPDX) | ✔ (gosec, govulncheck, Dependabot) | ✔ SHA-pinned |
| H | Errors & logging | ✔ (logger silenced) | ✔ (wrapping, ProviderErrorf) | ✱ ResourceInitFailedError |

**Headline:** the structure was already strong; remediation closed the lifecycle,
docs, and hygiene gaps. Identity fields force replacement, reads signal
drift-to-deleted, deletes are idempotent, controller defaults are declared (no
post-refresh diff), every property carries a description, closed value sets are
enums (71 of them), credentials are secret and asserted, the SDK plugin pipeline
is intact with `needs: binaries`, CI runs the same gate as `mise run ci` plus
govulncheck, and a client-injection seam backs hermetic LifeCycleTests + a golden
schema snapshot. Remaining work is optional polish (a multi-language `examples/`
tree, a logo asset, `ResourceInitFailedError`).

> The remediation also surfaced and fixed a **latent runtime bug**: the provider
> config was registered by value (`infer.Config(config.Config{})`) while
> `Configure` has a pointer receiver, so infer's `any(*receiver).(CustomConfigure)`
> check never matched and `Configure` never ran — clients were never built and the
> first real resource op would nil-deref. Registering by pointer fixed it; the new
> LifeCycleTests catch the class.

---

## Resolved release blockers (were P0)

1. **🔴 Identity/bind-key fields force replacement (B §2.4).** `provider:"replaceOnChanges"`
   on `Device.mac`, `Camera.cameraId`, and `User.mac` (matching Terraform's
   `ForceNew` on user mac). Guarded by `TestProviderSchemaBuilds`.
2. **🔴 `Read` signals drift-to-deleted (B §2.2).** A `notFound` helper (go-unifi
   `ErrNotFound` **or** an HTTP 404 `ServerError`) and a `isProtectNotFound`
   helper (the unified client has no typed sentinel) make every `Read` return an
   empty `ReadResponse{}` on not-found, so `pulumi refresh` recreates instead of erroring.
3. **🔴 Registry docs added (F §6.1).** `docs/_index.md` + `docs/installation-configuration.md`.
4. **🔴 Field descriptions reach consumers (F §6.4).** `a.Describe` on every input,
   output, and nested property (0/182 → 182/182 inputs; 11/382 → 382/382 type
   props). `TestProviderSchemaBuilds` asserts 100% coverage.

---

## By dimension

### A — Schema & framework — *strong*
- ✔ 🔴 Tokens stable, one package per module, token set pinned. Tag casing fixed
  (`macaclFormat`→`macAclFormat`; the old `dpigroupId` was already resolved by the
  facet-nesting refactor).
- ✔ 🔴 Config is `infer.Config(&config.Config{})` — registered by **pointer** so the
  pointer-receiver `Configure` runs (see the latent-bug note above).
- ✔ 🟡 **`SetDefault`** declares every non-zero controller default (24 fields) so the
  SDK/Registry report them and `DefaultCheck` fills unset inputs (fixes §2.7). Config
  adds env fallbacks (`UNIFI_URL`/`UNIFI_API_KEY`/…); `url` is optional + guarded.
- ✔ 🟡 **Enums** — 71 closed value sets modeled as `Enum[T]`. Genuinely open sets
  (channel, txPower, speed, FZP protocol, network/interface group names, camera
  modes) stay free-form to avoid rejecting valid controller values.
- ✱ 🟡 Builder metadata: `WithKeywords` + Go import path + sharpened description added;
  **`WithLogoURL` still missing** (needs a hosted logo asset).
- ✔ 🟡 Wide resources nested by facet (the refactor is in place).

### B — Resource lifecycle — *solid*
- ✔ 🔴 DryRun guards on all Create/Update; Update implemented everywhere.
- ✔ 🔴 replaceOnChanges on bind keys; ✔ 🔴 drift-to-deleted Read.
- ✔ 🔴 **Server defaults no longer leak into inputs (§2.7)** — declared via `SetDefault`.
- ✔ 🟡 **`Check`** — `AlarmAutomation.Check` validates required collections at
  preview time as per-property `CheckFailure`s (via `infer.DefaultCheck`); enum/format
  validation is handled by the schema-level enums.
- ✔ 🟡 **Delete idempotent** on all 11 deletable network resources (swallow not-found);
  Device/Camera deletes stay no-ops (adoption model).
- — 🟢 `CustomStateMigrations` n/a — unreleased, no shipped state.

### C — Secrets — *strong*
- ✔ 🔴 Every credential tagged `provider:"secret"`; IDs/logs/errors secret-free; auth
  validated fail-fast.
- ✔ 🔴 **Test coverage** — `TestProviderSchemaBuilds` asserts secrecy on the config
  `apiKey`/`password` plus the nested secret types (PPPoE/WEP/IAPP/PPSK/SAE).

### D — Versioning & publishing — *exemplary*
- ✔ all 5 gates (SemVer tag → ldflags → schema → asset → SDK; `respectSchemaVersion`;
  correct naming/archive root; `github://` round-trip; CGO-disabled matrix).
- ✔ 🟡 **`publish-npm` `needs: binaries`** — SDK can't publish before the Release assets exist.
- Note: no `toolchain` directive — it is redundant here (the `go 1.26.4` directive is
  already patch-pinned, and `go mod tidy` strips a matching toolchain line).

### E — Testing — *strong*
- ✔ 🟡 Pure round-trip tests across all resources (unchanged, still comprehensive).
- ✔ 🔴 Cheap tiers + SDK-freshness gated in CI (now via the shared `ci` task).
- ✔ 🔴 **Schema guard** — full-schema **golden snapshot** (`UPDATE_SCHEMA=1` to
  regenerate) + targeted assertions (descriptions, secrets, replaceOnChanges, enum floor).
- ✔ 🟡 **`LifeCycleTest` + client seam** — `config.InjectClientsForTest` swaps in fake
  clients; LifeCycleTests cover a representative set (FirewallGroup, DnsRecord,
  UserGroup, Camera) across both clients and both lifecycle models.
- ✱ 🟢 No property-fuzz / pulumitest / component-mock tests (optional).

### F — Docs & UX — *registry-ready*
- ✔ 🔴 Registry docs; ✔ 🟡 field descriptions reach the SDK (JSDoc verified).
- ✱ 🟡 **No multi-language `examples/` dir / chooser shortcodes** (deferred — content
  work; `_index.md` ships a runnable TS example).
- ✔ 🟡 Consume flow + credential config documented (installation-configuration.md:
  `pulumi config set unifi:apiKey --secret`, explicit-provider example).
- ✔ 🟡 Error messages wrapped with op/resource/site context (see H).

### G — CI & hygiene — *complete*
- ✔ 🔴 **CI == local gate** — one `mise run ci` aggregate (tidy/fmt/vet/lint/test/
  build/sdk-freshness) run by both; `mise run check` aliases it.
- ✔ 🔴 SPDX headers on all 37 Go files.
- ✔ 🟡 golangci-lint adds `gosec` + `misspell` + a formatters block (`revive` left out
  to avoid churn).
- ✔ 🟡 **`govulncheck`** (CI step + `mise run vulncheck`; kept out of the local gate
  since new disclosures can fail it without a code change) and **Dependabot**
  (gomod / github-actions / npm).
- ✔ 🟢 Actions pinned by commit SHA; CHANGELOG / CODEOWNERS / PR template added.

### H — Errors & logging — *two-tier, both solid*
- ✔ 🔴 stdout pristine; **go-unifi Network logger silenced**
  (`cc.Logger = unifi.NewDefaultLogger(unifi.DisabledLevel)`). The Protect client keeps
  `context.Background()` deliberately — its `doRequest` uses the stored ctx for every
  REST call, so the request-scoped Configure ctx would break Camera ops.
- ✔ 🟡 Network/Camera CRUD wrap vendor errors with op/resource/site context (network +
  protect `wrap` helpers); not-found early returns and adoption messages preserved.
- ✔ 🟡 **`ProviderErrorf`** — every Create errors clearly if the controller returns no
  ID; the discarded `json.Marshal` error in alarm automation is now handled.
- ✱ 🟡 **No `ResourceInitFailedError`** for Device/Camera RMW (deferred — the API exposes
  no clean "init complete" signal; the adoption check already covers not-ready devices).

---

## Wide-resource decomposition (the refactor)

Done. The wide resources are nested by facet (single controller objects with one write
endpoint), with the mapper preserving always-send bools, auto-enable toggles, nested
secrets, and group-allocated-only-when-set optionality. Top-level input counts and
groupings:

| Resource | Nested groups |
|----------|---------------|
| `Vlan` | `dhcp` `dhcpV6` `ipv6` `igmp` `wan` `nat` |
| `Wlan` | `wpa` `wpa3` `sae` `wep` `privatePresharedKeys` `radius` `vlanTagging` `bandSteering` `dtim` `minRate` `macFilter` `multicast` `schedule` `apGroups` `dpi` `iot` `p2p` `roaming` |
| `PortProfile` | `vlan` `link` `stormControl` `portSecurity` `dot1x` `lldpMed` `egressRateLimit` `priorityQueues` |
| `Device` | `led` `snmp` `stp` `switching` `dot1x` `outlet` `vrrp` `lcm` (+ override lists) |
| `FirewallRule` | `protocolMatch` `source` `destination` `connectionState` |
| `FirewallZonePolicy` | `matching` `source` `destination` `schedule` |

---

## Remaining (optional polish)

- **`examples/` dir + chooser shortcodes** (F §6.2) — multi-language examples; content work.
- **`WithLogoURL`** (A §1.2) — needs a hosted logo asset.
- **`ResourceInitFailedError`** (H §8.4) — for Device/Camera RMW once an init-complete
  signal is identified.
- **`revive`**, property-fuzz / pulumitest / component-mock tests (E §5.x) — nice-to-have.

See [`adding-resources.md`](./adding-resources.md) for implementing against codebase
conventions.
