# Provider quality rubric — "what good looks like"

A checklist for a native Go (`pulumi-go-provider`/`infer`) provider. Each item links to
the rationale in [`pulumi-provider-best-practices.md`](./pulumi-provider-best-practices.md)
(§ numbers below). Use it for self‑review, PR review, and release readiness.

**Severity legend**

- 🔴 **Gate** — must hold before publishing/releasing; a violation is a correctness,
  security, or distribution defect.
- 🟡 **Expected** — the mark of a quality provider; address before 1.0.
- 🟢 **Polish** — raises the ceiling; nice to have.

A "✔ / ✱ / ✘" column is meant to be filled in per review (see
[`conformance.md`](./conformance.md) for this repo's current scoring).

---

## A. Schema & framework (§1)

| Sev | Criterion |
|-----|-----------|
| 🔴 | Tokens are stable and intentional: one Go package per Pulumi module; no accidental retoken on refactor; a test pins the token set. (§1.1) |
| 🔴 | Every property tag is camelCase; resource/type names PascalCase; acronyms normalized (`dhcpV6Enabled`, not `DHCPDV6`). (§1.2) |
| 🔴 | Required = plain tag + value type; optional = `,optional` + pointer/slice/map. State embeds Args; server‑assigned fields only on State. (§1.3) |
| 🟡 | Every resource **and field** has a description (doc‑comment and/or `Describe`); none blank in the generated schema. (§1.4, §6.4) |
| 🟡 | User‑facing defaults declared with `SetDefault` (visible in schema/SDK), distinct from API‑boundary "always‑send" values. (§1.5) |
| 🟡 | Closed value sets modeled as `Enum[T]`, not free‑form strings. (§1.6) |
| 🟡 | Nested objects are named, annotated complex types (uniquely named within the module), not `map[string]any`. (§1.7) |
| 🟡 | Wide resources are decomposed correctly: nested by facet for single‑object resources; split into separate resources **only** where the backend has a per‑facet endpoint. (§1.9) |
| 🔴 | Provider config is an `infer.Config` struct; clients built once in `Configure` on unexported fields; resources use `GetConfig`. (§1.8) |
| 🟡 | Full builder metadata: displayName, description (incl. package name), homepage, repository, publisher, license; `keywords` (`category/…`,`kind/native`) + `logoUrl` for the Registry. (§1.10, §6.3) |

## B. Resource lifecycle (§2)

| Sev | Criterion |
|-----|-----------|
| 🔴 | `Create`/`Update` branch on `DryRun`: no side effects in preview, planned output returned. (§2.1) |
| 🔴 | `Read` reconstructs `Inputs` from the fetched object (not just validation) so `pulumi import` works; refresh detects drift. (§2.2) |
| 🔴 | `Update` implemented for any resource whose fields can change in place (else every edit destroys+recreates). (§2.3) |
| 🔴 | Immutable/identity fields forced to replace via `provider:"replaceOnChanges"` or a custom `Diff`. (§2.4) |
| 🟡 | Custom `Diff` only where the structural default is wrong; `DeleteBeforeReplace` set where uniqueness constraints require it. (§2.5) |
| 🟡 | Input validation/normalization in `Check` (layered on `DefaultCheck`), returning per‑property `CheckFailure`s — not errors from `Create`. (§2.6) |
| 🔴 | Round‑trip preserves write‑only/secret and server‑defaulted fields (no spurious "always shows changes" diffs). (§2.7) |
| 🟡 | Stable IDs (never from a mutable field); `Delete` idempotent (already‑gone → success). (§2.8) |
| 🟡 | Adoption resources: `Create` verifies existence + clear error; `Update` is read‑modify‑write; `Delete` no‑op; bind key replace‑forcing. (§2.9) |
| 🟢 | State‑shape changes shipped with `CustomStateMigrations`. (§2.10) |

## C. Secrets & sensitive data (§3)

| Sev | Criterion |
|-----|-----------|
| 🔴 | Every credential/sensitive input tagged `provider:"secret"` (config **and** resources); verified by a schema test. (§3.1) |
| 🔴 | Resource IDs never contain a secret; secrets never copied into unmarked output/state fields. (§3.5) |
| 🔴 | No secret value interpolated into any log or error string. (§3.5) |
| 🟡 | Provider‑generated sensitive outputs marked `AlwaysSecret()`; secret→output flow correct (default all‑wired, or explicit `WireDependencies`). (§3.2, §3.3) |
| 🟡 | Write‑only credentials preserved from prior in `Read`/`Diff` (not clobbered by the API's nil). (§3.4) |
| 🔴 | Auth validated in `Configure`, fail‑fast, value‑free error. (§3.5) |

## D. Versioning & publishing (§4)

| Sev | Criterion |
|-----|-----------|
| 🔴 | One SemVer flows from the git tag → ldflags → reported version → schema version → asset name → SDK version. (§4.1) |
| 🔴 | `respectSchemaVersion: true`; SDKs generated from the **version‑stamped** binary. (§4.2, §4.6) |
| 🔴 | Binary `pulumi-resource-<name>`; assets `pulumi-resource-<name>-v<ver>-<os>-<arch>.tar.gz` with binary at root. (§4.3) |
| 🔴 | `pluginDownloadURL` embedded (`github://…`) so consumers auto‑install the plugin. (§4.4) |
| 🔴 | Cross‑compiled matrix (CGO disabled) covering all shipped os/arch; none missing. (§4.5) |
| 🟡 | SDKs are build output (never hand‑edited); committed SDKs regenerated at the tag. (§4.6) |
| 🟡 | npm SDK under own scope, `--access public`; OIDC trusted publishing + provenance (no long‑lived token). (§4.7, §4.8) |
| 🟡 | Tag‑driven release; binaries published before SDKs; least‑privilege per‑job tokens. (§4.9) |

## E. Testing (§5)

| Sev | Criterion |
|-----|-----------|
| 🔴 | Schema‑builds guard test (tokens present, secrets marked, field floors); ideally a full‑schema `JSONEq` snapshot. (§5.1) |
| 🟡 | In‑process `LifeCycleTest` per resource family (Create→Update→Delete at the gRPC boundary) against a faked client. (§5.2, §5.6) |
| 🟡 | Pure mapping round‑trip tests (identity + optionals survive; controller‑zero+nil‑prior stays nil). (§5.5) |
| 🟢 | Targeted Diff/Check method tests; component `WithMocks`; program tests via `pulumitest`; property‑based fuzz of the round‑trip. (§5.3, §5.7) |
| 🔴 | Cheap tiers (schema + round‑trip) gated in CI; SDK regeneration in the build gate. (§5.8) |

## F. Documentation, examples & UX (§6)

| Sev | Criterion |
|-----|-----------|
| 🔴 | `docs/_index.md` + `docs/installation-configuration.md` present (Registry requirement). (§6.1) |
| 🟡 | Multi‑language examples via the chooser / `example-program` shortcodes. (§6.2) |
| 🟡 | Full consume flow documented (released, pinned, local, ambient‑config vs explicit Provider). (§6.5) |
| 🟡 | Actionable, namespaced error messages distinguishing user errors from provider bugs. (§8.1, §8.5) |

## G. CI/CD & repo hygiene (§7)

| Sev | Criterion |
|-----|-----------|
| 🔴 | CI gates build, vet, lint, test, `go mod tidy` drift, and generated‑SDK freshness — and matches the local pre‑commit gate. (§7.1) |
| 🟡 | golangci‑lint v2 with security/quality linters (`gosec`, `revive`, `misspell`, …) + formatters; generated paths excluded. (§7.2) |
| 🔴 | Root `LICENSE` present (matches goreleaser `archives.files`); per‑file SPDX/Apache headers. (§7.5) |
| 🟡 | Toolchain + deps pinned (go directive + toolchain directive, `go.sum`, version‑managed tools). (§7.3) |
| 🟡 | `govulncheck` in CI (fails on findings); Dependabot for `gomod` + `github-actions`. (§7.4) |
| 🟢 | CHANGELOG or enforced Conventional Commits; CODEOWNERS; SHA‑pinned actions; branch protection. (§7.6, §7.7) |

## H. Errors, logging & diagnostics (§8)

| Sev | Criterion |
|-----|-----------|
| 🟡 | Vendor errors wrapped with `%w` + operation/ID/site context in every CRUD path. (§8.1) |
| 🔴 | Logging only via `p.GetLogger(ctx)`; **nothing** written to stdout; chatty vendor loggers silenced. (§8.2, §8.3) |
| 🟡 | Partial creation signaled with `ResourceInitFailedError`; provider bugs via `ProviderErrorf`/`InternalErrorf`. (§8.4, §8.5) |
| 🟡 | `Delete` downgrades not‑found to a warning and succeeds. (§8.6) |

---

### How to score
For each row mark **✔ pass / ✱ partial / ✘ fail / — N/A**, with a one‑line note and a
`file:line` pointer to the evidence. Tally 🔴 gates first: any open 🔴 blocks a release.
Then 🟡, then 🟢. The filled‑in scorecard for this repo lives in
[`conformance.md`](./conformance.md).
