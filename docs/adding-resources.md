# Adding (and refactoring) a resource — the pulumi-unifi way

The internal authoring guide. It turns the general
[best practices](./pulumi-provider-best-practices.md) and the
[quality rubric](./quality-rubric.md) into the concrete conventions this codebase
already follows, plus the gotchas that have cost real iterations (see also the project
memory: `pulumi-go-provider-infer-gotchas`).

## The shape of a resource

Every resource is **three Go types** in a file under `provider/network/` or
`provider/protect/`:

```go
// 1. The marker (controller) struct — empty, carries the CRUD methods.
type Vlan struct{}

// 2. Args — the user inputs. Optional fields are POINTERS + `,optional`.
type VlanArgs struct {
    Name    string  `pulumi:"name"`                 // required: value type, no ,optional
    Purpose *string `pulumi:"purpose,optional"`     // optional: pointer + ,optional
    // ...
}

// 3. State — embeds Args, adds controller-assigned (read-only) fields.
type VlanState struct {
    VlanArgs
    NetworkId string `pulumi:"networkId"`           // the controller _id
}
```

Register it in `provider/provider.go` `WithResources(...)` **and** add its token to the
`want` list in `provider/provider_test.go`.

### Naming rules (load-bearing)
- **Property tags are camelCase**, set by the tag value, regardless of the Go field name:
  `IGMPSnooping` → `pulumi:"igmpSnooping"`, `DHCPDV6Enabled` → `pulumi:"dhcpV6Enabled"`.
  Never let upstream `go-unifi` capitalization leak into the schema.
- The `network` package is **shared** by all go-unifi resources, so they share one
  Pulumi module namespace (`unifi:network`). **Prefix nested/helper struct names** to stay
  unique: `NetworkWanDhcpOption`, `VlanDhcp`, `DevicePortOverride` — never a bare `Dhcp`.
- Name the OUT mapper resource-prefixed: `vlanStateFrom`, `deviceStateFrom`.

## The mapper pattern

Keep CRUD methods thin; put all mapping in **two pure functions** that are unit-tested
without a network:

- `func (a XArgs) toUnifi(id string) *unifi.X` — build the go-unifi struct from inputs.
  `id` is `""` on create.
- `func xStateFrom(u *unifi.X, prior XArgs) XState` — map the controller object back into
  state.

### Round-trip rules (this is where the bugs live)
- **Optional fields use "reflect controller value if set, else keep prior."** Use the
  shared helpers in `helpers.go` (`ptr`, `derefOr`) and the per-resource
  `*StrPtr`/`*IntPtr`/`*BoolPtr` helpers (e.g. `vlanStrPtr(v, prior)` returns
  `ptr(v)` when `v != ""`, else `prior`). This preserves write-only secrets the
  controller blanks and avoids spurious diffs from server defaults. **Never** redefine
  `ptr`/`derefOr`.
- **Secrets are preserved from `prior`**, because the controller never echoes them:
  `args.WanPassword = prior.WanPassword`. Tag them `provider:"secret"` (see below).
- **Always-send bools:** an upstream field whose JSON lacks `omitempty` and whose
  controller default is `true` must be sent explicitly: `n.Field = derefOr(a.Field, true)`.
  Omitting sends Go's zero `false` and silently disables the feature. Known cases:
  `internetAccessEnabled`, `dhcpdv6_dns_auto`. There's a `TestVlanDefaults`-style test for
  this — add one for any new always-send bool.
- **`*Enabled` toggle pairs:** when a value field implicitly enables a feature, set both:
  `if a.InterfaceMtu != nil { n.InterfaceMtu = *a.InterfaceMtu; n.InterfaceMtuEnabled = true }`,
  and still honor an explicit `interfaceMtuEnabled` if given.
- **Lists/maps:** reflect the controller list when non-empty, else fall back to `prior`
  (so an unset optional list doesn't churn). Sort map-derived lists for deterministic
  diffs (see `headerList` in `alarm_automation.go`).

## CRUD methods

```go
func (Vlan) Create(ctx context.Context, req infer.CreateRequest[VlanArgs]) (infer.CreateResponse[VlanState], error) {
    if req.DryRun {                                   // 1. preview: no side effects
        return infer.CreateResponse[VlanState]{Output: VlanState{VlanArgs: req.Inputs}}, nil
    }
    cfg := infer.GetConfig[config.Config](ctx)        // 2. get the configured client
    created, err := cfg.Network().CreateNetwork(ctx, cfg.ResolvedSite(), req.Inputs.toUnifi(""))
    if err != nil {
        return infer.CreateResponse[VlanState]{}, fmt.Errorf("unifi:network:Vlan create %q: %w", req.Inputs.Name, err) // 3. wrap!
    }
    return infer.CreateResponse[VlanState]{ID: created.ID, Output: vlanStateFrom(created, req.Inputs)}, nil
}
```

Checklist for every CRUD method:
- [ ] `Create`/`Update` branch on `req.DryRun` and return planned output.
- [ ] `Read` reconstructs `Inputs` from the fetched object (powers `pulumi import`) and
      **signals deletion** on not-found by returning an empty `ReadResponse{}` (see
      `alarm_automation.go` for the pattern — most resources don't do this yet; new ones
      should).
- [ ] `Update` exists (no `Update` ⇒ every change is a destroy+recreate).
- [ ] Errors wrapped with `fmt.Errorf("unifi:<module>:<Resource> <op> %q: %w", id, err)`.
- [ ] `Delete` tolerates already-gone (return success on 404).
- [ ] Identity/immutable inputs tagged `provider:"replaceOnChanges"` (e.g. a bind key).

### Adoption-model resources (`Device`, `Camera`)
Hardware is adopted, not created. `Create` verifies the target exists (clear error if
not), `Update` does a **read-modify-write** (fetch live, overlay only managed fields,
merge keyed sub-collections by their key, write back), `Read` binds by ID, and `Delete`
is a **no-op**. Tag the bind key (`mac`, `cameraId`) `provider:"replaceOnChanges"`.

## Secrets

Tag every credential/sensitive field `provider:"secret"` on **both** the config and the
resource (and inside nested structs — e.g. `WlanPrivatePsk.Password`). It's verified by
`provider_test.go`'s secret checks — **extend that map** when you add a secret field.
Never put a secret in a resource ID, log line, or error string.

## Nesting wide resources (the decomposition convention)

UniFi writes each object via **one endpoint**, so do **not** split a wide resource into
multiple Pulumi resources that RMW the same object (that races). Instead **nest related
fields into named complex types** (best-practices §1.9):

```go
type VlanArgs struct {
    Name string `pulumi:"name"`            // identity/core stays top-level
    Vlan *int   `pulumi:"vlan,optional"`
    Dhcp *VlanDhcp `pulumi:"dhcp,optional"` // grouped facet
    Ipv6 *VlanIpv6 `pulumi:"ipv6,optional"`
    Wan  *VlanWan  `pulumi:"wan,optional"`
}
type VlanDhcp struct { // unique, module-prefixed name
    Enabled *bool   `pulumi:"enabled,optional"`
    Start   *string `pulumi:"start,optional"`
    // ...
}
```

Rules when nesting:
- Each nested struct gets a **unique, module-prefixed Go name** and may implement
  `Annotate` for field docs.
- The nested object property is itself a `*Struct` `,optional` so absence is
  representable; in `toUnifi`, guard `if a.Dhcp != nil { ... }`.
- Schema defaults (`SetDefault`) are **primitive-only** — keep nested-field defaults in
  the mapper via `derefOr`.
- Preserve the always-send-bool and `*Enabled`-pair handling inside the nested mapper.
- Nesting changes property shapes. Because this provider is **unreleased with no
  dependents**, the nested shape is simply the schema — no migrations, no version dance.
  (Only once a provider has users does reshaping state need `CustomStateMigrations`/`AddAlias`.)

## Before you commit

```sh
mise run fmt          # gofmt
mise run vet          # go vet
mise run lint         # golangci-lint
mise run test         # go test ./...  (includes the schema-builds guard)
mise run build        # compiles the plugin
mise run sdk:nodejs   # regenerate the TS SDK from the new schema
mise run sdk:check    # fail if the committed SDK is now stale
```

A schema change **must** be followed by `sdk:nodejs` + committing the regenerated SDK —
CI's `sdk:check` will fail otherwise. Add/extend tests: a round-trip test (`*_test.go`)
covering the new fields, and update the field-floor/secret/token assertions in
`provider_test.go` as needed.

## Pointers
- General rationale & citations: [`pulumi-provider-best-practices.md`](./pulumi-provider-best-practices.md)
- Pass/fail checklist: [`quality-rubric.md`](./quality-rubric.md)
- Reference implementations to copy: [`prior-art.md`](./prior-art.md) (esp.
  `examples/dna-store` for adoption `Read`, `examples/file` for `Check`/`Diff`)
- This repo's current gaps: [`conformance.md`](./conformance.md)
