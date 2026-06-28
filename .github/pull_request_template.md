## What & why

<!-- What does this change do, and why? Link any issue. -->

## Checklist

- [ ] `mise run ci` passes (tidy, fmt, vet, lint, test, build, SDK freshness)
- [ ] Schema/SDK changes are intentional and the golden snapshot is regenerated
      (`UPDATE_SCHEMA=1 go test ./provider/`)
- [ ] New/changed resource fields have `a.Describe` and, for closed sets, an enum
- [ ] `CHANGELOG.md` updated if user-facing
