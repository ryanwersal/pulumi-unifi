# Conformance report — pulumi-unifi vs. the quality rubric

How this provider scores against [`quality-rubric.md`](./quality-rubric.md). Produced by a
multi-agent review that graded each dimension against the code with `file:line` evidence,
then **adversarially re-verified every non-pass finding** by re-reading the cited code
(28 checks, **0 corrections** — the findings below are confirmed, several empirically:
the reviewers built the binary and inspected the generated schema).

**Verdicts:** ✔ pass · ✱ partial · ✘ fail · — n/a. Severity: 🔴 gate · 🟡 expected · 🟢 polish.

## Scorecard

| Dim | Area | 🔴 gates | 🟡 expected | 🟢 polish |
|-----|------|----------|-------------|-----------|
| A | Schema & framework | ✔✔✔✔ (4/4) | ✔✔ ✱✱✱ (2 pass, 3 partial), ✘ enums | — |
| B | Resource lifecycle | ✔✔ ✱✱✱ ✘ (2 pass, 3 partial, 1 fail) | ✱✱✱ ✔ | — n/a |
| C | Secrets | ✔✔✔ ✱ (3 pass, 1 partial) | ✔✔ | — |
| D | Versioning & publishing | ✔✔✔✔✔ (5/5) | ✔✔ ✱ (2 pass, 1 partial) | — |
| E | Testing | ✔ ✱ (1 pass, 1 partial) | ✔✔ ✘ (2 pass, 1 fail) | ✱ |
| F | Docs & UX | ✘ (registry docs) | ✱✱ ✘✘ (2 partial, 2 fail) | — |
| G | CI & hygiene | ✱✱ (2 partial) | ✱✱✱ ✘ (3 partial, 1 fail) | ✱ |
| H | Errors & logging | ✱ (1 partial, gate-critical part holds) | ✱✱✱ ✘ (3 partial, 1 fail) | — |

**Headline:** the provider's *structure* is strong — clean tokens, flawless
required/optional discipline, secrets correctly tagged, exemplary versioning/distribution
(all 5 D-gates pass), and a real schema-builds guard. The gaps cluster in **lifecycle
robustness** (replacement on identity fields, drift-to-deleted, idempotent delete),
**consumer-facing polish** (field descriptions don't reach the SDK, no Registry docs, no
enums, raw error messages), **wide-resource ergonomics** (the nesting refactor), and
**CI/hygiene completeness**. Most are 🟡 "before 1.0" items; the true blockers are few.

---

## Release-blocking issues (do first)

1. **🔴 Identity/bind-key fields don't force replacement (B §2.4).** No
   `provider:"replaceOnChanges"` anywhere. Changing `Device.mac` (`device.go:177`) or
   `Camera.cameraId` (`camera.go:32`) is treated as an in-place Update that patches the
   **old** device/camera and never rebinds — a wrong-object write + perpetual diff. **Fix:**
   tag `Device.Mac` and `Camera.CameraId` (consider `User.Mac`) `provider:"replaceOnChanges"`.
2. **🔴 `Read` doesn't signal drift-to-deleted (B §2.2).** Only `alarm_automation.go:242`
   returns an empty `ReadResponse{}` on not-found; the other 13 resources return the raw
   error, so `pulumi refresh` of an out-of-band-deleted object **errors** instead of
   marking it for recreation. go-unifi exposes a typed `unifi.ErrNotFound`. **Fix:** in
   each `Read`, `errors.Is(err, unifi.ErrNotFound)` → return empty `ReadResponse{}`.
3. **🔴 Registry docs missing (F §6.1).** No `docs/_index.md` / `docs/installation-configuration.md`
   → the package cannot be listed in the Pulumi Registry. **Fix:** add both (lift from README).
4. **🔴 Field descriptions don't reach consumers (F §6.4).** Empirically, **0 of 385
   input properties** carry a description in the generated schema — this build's `gen-sdk`
   does not pick up Go doc-comments; only `a.Describe` calls surface (resource-level + 8
   nested fields). The committed `sdk/nodejs/network/vlan.ts` has zero JSDoc on its 120
   fields. **Fix:** add `a.Describe(&x.Field, …)` for every property (reuse the existing
   doc-comment text) and add a schema test asserting non-empty descriptions.

> Note on the logging "gate": H §8.2/§8.3 is **not** a blocker — nothing writes to stdout
> (verified), so the plugin handshake is safe. The residual issue is that the go-unifi
> Network client's default logger isn't silenced (only the Protect one is); it writes to
> stderr, so it's a cleanliness 🟡, not a crash 🔴.

---

## By dimension

### A — Schema & framework — *strong*
- ✔ 🔴 Tokens stable, one package per module, token set pinned (`provider.go:51-64`,
  `provider_test.go:52-72`).
- ✔ 🔴 Casing/optionality flawless (0 mismatches found); State embeds Args. Minor: two
  tags imperfectly segmented (`dpigroupId`→`dpiGroupId`, `radiusMacaclFormat`→`radiusMacAclFormat`,
  `wlan.go:179,202`) — rename pre-1.0.
- ✔ 🔴 Config is `infer.Config`; clients built once in `Configure` on unexported fields.
- ✱ 🟡 **No `SetDefault`** — user-visible defaults live only in `derefOr` (`vlan.go:328-331`
  etc.), so the SDK/Registry under-report them. Promote primitive defaults to `SetDefault`
  (+ env fallbacks for config: `UNIFI_URL`/`UNIFI_API_KEY`/…). Keep always-send bools in
  the mapper.
- ✘ 🟡 **No enums** — dozens of closed sets are free-form strings (`purpose`, `action`,
  `ruleset`, `recordType`, `security`, `ipVersion`, …). Model the truly-closed ones as
  `Enum[T]`.
- ✱ 🟡 Builder metadata missing `WithKeywords("category/network","kind/native")`,
  `WithLogoURL`, a Go import path, and the package name in the description (`provider.go:36-48`).
- ✱ 🟡 Wide resources not nested → the refactor (see below).

### B — Resource lifecycle — *solid skeleton, robustness gaps*
- ✔ 🔴 DryRun guards on all 28 Create/Update branches; ✔ 🔴 Update implemented everywhere
  (no accidental replace-on-change).
- ✘ 🔴 replaceOnChanges on bind keys — blocker #1.
- ✱ 🔴 drift-to-deleted Read — blocker #2.
- ✱ 🔴 **Server defaults reflected into inputs (§2.7).** `stateFrom` copies controller
  defaults into input fields even when the user didn't set them
  (`firewall_group.go:47→60-64`, `vlan.go:328-331→756-778`, `user_group.go:45-46→55-56`,
  `port_profile.go:157-160`, `port_forward.go:81-82`), so post-refresh/import shows a
  permanent diff. **Fix:** declare those defaults via `SetDefault` (so `DefaultCheck` fills
  the program input to match), or only reflect a default into the input when prior was set.
  Write-only secrets are handled correctly (`vlan.go:897`, `wlan.go:620,542-585`).
- ✱ 🟡 **No `Check` methods** — enum/format validation is left to the controller
  (apply-time errors, not preview-time per-property `CheckFailure`s). `alarm_automation`
  validates inside Create/Update as a plain error — move to `Check`.
- ✱ 🟡 **Delete not idempotent** on 11 network resources (raw `DeleteX` error, no 404
  swallow). `alarm_automation` + the no-op adoption Deletes are correct. **Fix:** swallow
  `unifi.ErrNotFound`.
- ✱ 🟡 Adoption mechanics correct except the missing replaceOnChanges (above).
- — 🟢 `CustomStateMigrations` n/a — no shipped state exists, so the nesting refactor
  doesn't need it.

### C — Secrets — *strong*
- ✔ 🔴 Every credential tagged `provider:"secret"` — config `apiKey`/`password`, `vlan
  wanPassword`, `wlan passphrase`/`xWep`/`xIappKey`, and nested `WlanPrivatePsk.password`/
  `WlanSaePsk.psk` (runtime nested enforcement verified in `infer/apply_secrets.go`).
- ✱ 🔴 **Test coverage thin** — `provider_test.go:81-94` asserts secrecy on only 2 of ~9
  secret properties and **not** the config `apiKey`/`password`. **Fix:** extend the secret
  assertions (incl. nested types + config vars).
- ✔ 🔴 IDs never contain secrets; ✔ 🔴 no secret in logs/errors; ✔ 🔴 auth validated
  fail-fast & value-free (`config.go:60-72`); ✔ 🟡 write-only secrets preserved from prior.

### D — Versioning & publishing — *exemplary*
- ✔ all 5 gates: one SemVer tag→ldflags→`RunProvider`→schema→asset→SDK; `respectSchemaVersion`
  + stamped-binary gen-sdk; correct binary/asset naming + archive root; `github://`
  `pluginDownloadURL` embedded and round-tripped into the SDK; CGO-disabled matrix.
- ✱ 🟡 **`publish-npm` lacks `needs: binaries`** (`release.yml:27`) → SDK can publish before
  the Release assets exist (missing-plugin race). **Fix:** add `needs: binaries`.
- 🟢 windows/arm64 intentionally dropped (5/6 combos) — acceptable.

### E — Testing — *strong deterministic tier, missing lifecycle tier*
- ✔ 🟡 Pure round-trip tests are comprehensive (all 14 resources + protectapi), including
  the "controller-zero + nil-prior stays nil" and write-only-secret cases.
- ✔ 🔴 Cheap tiers + SDK-freshness gated in CI.
- ✱ 🔴 **Schema guard thin** (`provider_test.go`) — 2 secret flags, 1 field floor, no
  config-secret check, no full-schema `JSONEq` snapshot. **Fix:** add a golden snapshot +
  broaden assertions.
- ✘ 🟡 **No `LifeCycleTest`** and **no client-injection seam** — `config.Configure` builds
  real clients on unexported fields, so hermetic gRPC-level CRUD tests are impossible
  without a seam. **Fix:** introduce a Network/Protect client interface + test injector,
  then add `LifeCycleTest` per resource family.
- ✱ 🟢 No targeted Diff/Check, component-mock, pulumitest, or property-fuzz tests.

### F — Docs & UX — *README strong, Registry-facing weak*
- ✘ 🔴 Registry docs — blocker #3. ✘ 🟡 field descriptions — blocker #4.
- ✘ 🟡 No multi-language examples / `examples/` dir / chooser shortcodes.
- ✱ 🟡 Consume flow: plugin/install paths well documented; **credential config missing**
  (no `pulumi config set unifi:apiKey --secret`, no `new unifi.Provider({...})` example).
- ✱ 🟡 Error messages: config + adoption paths exemplary; ~13 network CRUD paths return raw
  vendor errors → see H.

### G — CI & hygiene — *good gates, incomplete*
- ✱ 🔴 **CI ≠ local gate** (`ci.yml` runs tidy/lint/test/sdk:check but **not** `go build`/
  `go vet`/`fmt-check`; `mise check` runs build but not tidy/sdk/vet). **Fix:** one shared
  aggregate task run by both.
- ✱ 🔴 LICENSE present & wired; **no SPDX headers** on 37 Go files. **Fix:** add
  `// SPDX-License-Identifier: Apache-2.0`.
- ✱ 🟡 golangci-lint v2 base set OK but missing `gosec`/`revive`/`misspell`, no formatters
  block, no `sdk/` exclusion.
- ✱ 🟡 go.mod has no separate `toolchain` directive (tool pinning via mise is exemplary).
- ✘ 🟡 No `govulncheck`, no `.github/dependabot.yml`. **Fix:** add both.
- ✱ 🟢 No CHANGELOG/CODEOWNERS/PR template; actions pinned by floating tag, not SHA.

### H — Errors & logging — *two-tier; network module lags*
- ✱ 🔴 stdout pristine (gate-critical part holds); **go-unifi Network logger not silenced**
  (`config.go` never sets `cc.Logger`; Protect logrus → `io.Discard` is correct). **Fix:**
  `cc.Logger = unifi.NewDefaultLogger(unifi.DisabledLevel)`. Also `Configure(_ context.Context)`
  drops `ctx` — thread it to enable `p.GetLogger`.
- ✱ 🟡 **~46 network/Camera CRUD sites return raw vendor errors** (no op/resource/site
  context); protectapi + config are the gold standard. **Fix:** wrap with
  `fmt.Errorf("<op> <resource> %q (site %q): %w", …)` (a `network`-package `wrap` helper
  mirroring `protectapi.wrap`).
- ✘ 🟡 No `ResourceInitFailedError` (relevant to Device/Camera RMW) and no
  `ProviderErrorf`/`InternalErrorf` (a successful create returning no ID surfaces as a
  plain error; `alarm_automation.go:147` discards a `json.Marshal` error).
- ✱ 🟡 Delete not-found → success only in protectapi/adoption (mirrors B §2.8).

---

## Wide-resource decomposition (the refactor)

Per best-practices §1.9, the wide resources are being **nested by facet** (not split into
sub-resources — they're single controller objects with one write endpoint). Verified
top-level input counts and the agreed grouping:

| Resource | Inputs | Nested groups (new complex types) |
|----------|--------|-----------------------------------|
| `Vlan` | 120 | `dhcp`(28) `dhcpV6`(10) `ipv6`(15) `igmp`(11) `wan`(35) `nat`(2) → 19 stay top-level |
| `Wlan` | 75 | `wpa` `wpa3` `sae` `wep` `privatePresharedKeys` `radius` `vlanTagging` `bandSteering` `dtim` `minRate` `macFilter` `multicast` `schedule` `apGroups` `dpi` `iot` `p2p` `roaming` → 13 top-level |
| `PortProfile` | 39 | `vlan` `link` `stormControl`(10) `portSecurity` `dot1x` `lldpMed` `egressRateLimit` `priorityQueues` → 7 top-level |
| `Device` | 31 | `led` `snmp` `stp` `switching` `dot1x` `outlet` `vrrp` `lcm` → 7 top-level (3 override lists stay) |
| `FirewallRule` | 30 | `protocolMatch` `source` `destination` `connectionState` → 8 top-level |
| `FirewallZonePolicy` | 17 | `matching`(7) → already has nested `source`/`destination`/`schedule` |

Per-resource hazards the mapper must preserve (from the grouping analysis):
- **Always-send bools** that must still serialize when their group pointer is nil:
  `Vlan.dhcpV6.dnsAuto` (default true), `PortProfile` ~14 `derefOr` defaults
  (`lldpMed.enabled` true, `dot1x.ctrl`, storm-control enables, `vlan.forward` "native"),
  `FirewallZonePolicy.matching.matchIpSec`/`matchOppositeProtocol`.
- **Auto-enable toggles** moving into groups: `Vlan.wan.vlan`→`wanVlanEnabled`,
  `Device.lcm.brightness/idleTimeout`→`Lcm*Override`.
- **Secrets** keeping their tag inside the nested struct: `Vlan.wan.password`,
  `Wlan.wep.key`, `Wlan.roaming.iappKey` (+ existing nested PSK element secrets).
- **Optionality:** each group is a `*Struct ,optional`; `stateFrom` allocates a group only
  when ≥1 member is set, else leaves it nil (avoid spurious diffs).
- The nesting changes property shapes / adds complex-type tokens. The provider is
  unreleased with no dependents, so this is just a schema change — no version dance, no
  migrations; the nested schema is simply the starting shape.

---

## Prioritized remediation roadmap

**P0 — before any wider release**
1. `provider:"replaceOnChanges"` on `Device.mac`, `Camera.cameraId` (B §2.4).
2. drift-to-deleted `Read` across all resources (B §2.2).
3. `docs/_index.md` + `docs/installation-configuration.md` (F §6.1).
4. Field `Describe` for all properties + a schema test guarding it (F §6.4).

**P1 — quality before 1.0**
5. The **wide-resource nesting refactor** (+ `SetDefault`, which fixes the §2.7 input-default
   diffs at the same time, + enums where natural).
6. Idempotent Delete + error wrapping across the network module (B §2.8, H §8.1); silence
   the go-unifi logger (H §8.3).
7. CI = local gate (add build/vet/fmt-check), `gosec`/formatters, SPDX headers,
   `govulncheck`, Dependabot, `needs: binaries`, `toolchain` directive (G, D §4.9).
8. Client-injection seam + `LifeCycleTest` per resource; broaden the schema guard +
   golden snapshot; extend secret assertions (E, C §3.1).

**P2 — polish**
9. `Check` methods (preview-time validation), `ResourceInitFailedError`/`ProviderErrorf`,
   `WithKeywords`/`WithLogoURL`, examples dir + chooser, CHANGELOG/CODEOWNERS, SHA-pinned
   actions, tag-casing tweaks.

See [`adding-resources.md`](./adding-resources.md) for how to implement these against the
codebase conventions.
