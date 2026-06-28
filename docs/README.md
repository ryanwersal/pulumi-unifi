# pulumi-unifi docs

Engineering documentation for this provider: what "good" looks like for a native Go
Pulumi provider, how this codebase measures up, and how to extend it.

These were produced by researching official Pulumi sources + `pulumi-go-provider`
internals (every cited framework API independently fact-checked against source) and a
multi-agent conformance review of this repo (every non-pass finding adversarially
re-verified against the code).

## Read in this order

1. **[pulumi-provider-best-practices.md](./pulumi-provider-best-practices.md)** — the
   reference: 8 categories of best practices for native Go (`infer`) providers, each cited
   and tagged **Framework** (a real API) vs **Convention**. The standard.
2. **[quality-rubric.md](./quality-rubric.md)** — the same practices as a gated pass/fail
   checklist (🔴 gate / 🟡 expected / 🟢 polish). Use for PR review & release readiness.
3. **[conformance.md](./conformance.md)** — this repo scored against the rubric, with
   `file:line` evidence, release blockers, and a prioritized remediation roadmap.
4. **[prior-art.md](./prior-art.md)** — the reference repos/docs to copy from (framework
   examples, real native providers, UniFi field-coverage prior art).
5. **[adding-resources.md](./adding-resources.md)** — the day-to-day authoring guide: the
   resource triad, the mapper round-trip rules, the gotchas, the nesting recipe, and the
   pre-commit/SDK-regen flow.

## TL;DR of the assessment

The provider is **structurally strong** — clean tokens, flawless optional/required
discipline, correctly tagged secrets, exemplary versioning/distribution (all gates pass),
and a schema-builds guard. The work to do, in priority order:

- **P0 (release blockers):** `replaceOnChanges` on adoption bind keys; `Read`
  drift-to-deleted; Registry docs (`_index.md` + `installation-configuration.md`); field
  `Describe`s (the SDK currently ships **0/385** field descriptions).
- **P1 (before 1.0):** the **wide-resource nesting refactor** (Vlan/Wlan/PortProfile/
  Device/FirewallRule/FirewallZonePolicy) + `SetDefault` + enums; idempotent Delete +
  error wrapping; CI/local gate parity, SPDX headers, `govulncheck`/Dependabot; a
  client-injection seam + `LifeCycleTest`s.
- **P2 (polish):** `Check` validation, richer diagnostics, examples, Registry metadata.

See [conformance.md](./conformance.md) for the full scorecard and evidence.
