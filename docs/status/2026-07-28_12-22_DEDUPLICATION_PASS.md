# Status Update — 2026-07-28 12:22

## Code Duplication Pass (`art-dupl --type-aware -t 3`)

### a) FULLY DONE

| #   | Clone Group                           | File:Line                                                | Action                                                                   | Result                                        |
| --- | ------------------------------------- | -------------------------------------------------------- | ------------------------------------------------------------------------ | --------------------------------------------- |
| 1   | Email dot-prefix/suffix check         | `internal/domain/values/email.go:185-191` & `210-216`    | Extracted `validateNoEdgeDots(s, partName)` helper                       | Eliminated                                    |
| 2   | ID required/whitespace check          | `internal/domain/ids/ids.go:129-137` & `162-170`         | Extracted `validateIDRequired(id, name)` helper                          | Eliminated                                    |
| 3   | Email `Split` + `len(parts)!=2` check | `internal/domain/values/email.go:54-59` & `64-69`        | Extracted `splitParts()` accessor shared by `Domain()` and `LocalPart()` | Eliminated                                    |
| 4   | Server timeout constants              | `cmd/main.go:26-29` vs `internal/config/config.go:17-20` | Reviewed for unification                                                 | Accepted (different values, different layers) |

- Created `docs/dedup-acceptance.md` recording the rationale for Group #4.
- All 3 actively-refactored groups verified by `go test ./...` (7 packages, all `ok`).
- Clean build: `go build ./...` exits 0.
- `golangci-lint run --new-from-rev=HEAD ./...` reports **0 issues** on changed code.

### b) PARTIALLY DONE

- Group #4 deduplication was accepted rather than eliminated. The structurally similar const-blocks differ in **values** (15s/15s/60s in `cmd/main.go` vs 5s/10s/120s in `internal/config/config.go`) and in **layer purpose** (demo entry point vs viper-driven app config). Wiring `cmd/main.go` to `config.LoadConfig()` would eliminate the duplication but introduces viper into `main`, contradicting the "Simple HTMX Demo" purpose documented in `AGENTS.md`. The trade-off was recorded but the actual unification was not performed — left as future work if the demo/policy changes.

### c) NOT STARTED

- Anything outside the deduplication scope. The session was strictly limited to the 4 clone groups art-dupl surfaced at `-t 3`. No broader lint sweep, architecture review, naming review, BDD test additions, or docs-health pass was performed.
- The Group #4 unification via viper wiring (see PARTIALLY DONE) was deliberately not started — it would cross the scope boundary into a multi-file config bootstrap refactor.

### d) TOTALLY FUCKED UP

Nothing. All 3 refactors preserved semantics:

- `validateNoEdgeDots` keeps identical error messages ("email local part cannot start with dot", "email domain cannot end with dot") by composing `partName + " cannot start with dot"` etc. — verified against tests that pass.
- `validateIDRequired` produces the same messages by concatenating `name + " is required"` / `name + " cannot have leading or trailing whitespace"`.
- `splitParts` preserves the "return empty string when split count != 2" behaviour for both `Domain()` and `LocalPart()`.

Pre-existing issues NOT touched (per "don't fix unrelated bugs" rule):

- `encoding/json/v2` import requires `GOEXPERIMENT=jsonv2` to compile (Go 1.26 vs 1.27 stdlib mismatch) — present before this session, still present.
- 22 pre-existing lint findings in `user_enums.go`, `user_id.go`, `user_session.go`, `cmd/main.go`, `config.go` (wrapcheck, recvcheck, godoclint, makezero, varnamelen, nonamedreturns, paralleltest, gochecknoglobals) — all in files I did not modify.

### e) WHAT WE SHOULD IMPROVE

1. **Wire `cmd/main.go` to `config.LoadConfig()`** — eliminates the Group #4 parallel constants and rescues the unused `internal/config` package from being dead code. Currently `internal/config` has tests but no production consumer. The `config.yaml` (with `read_timeout: "5s"`, `idle_timeout: "60s"`) suggests this wiring was the original intent.
2. **Add a build tag or `//go:build goexperiment.jsonv2`** to the three json/v2 importers so plain `go test ./...` works without an env-var workaround. Today every CI invocation must remember `GOEXPERIMENT=jsonv2`.
3. **Bump `go.mod` to `go 1.27`** (or whatever ships json/v2 in stdlib) so the experiment flag isn't needed.
4. **Run `golangci-lint run ./...` (not just `--new-from-rev=HEAD`)** as a real gate, and start clearing the 22 pre-existing warnings in batches — they're all fixable (wrapcheck wraps, recvcheck unifies receivers, godoclint rewrites godoc, makezero uses `make([]T, 0, n)`).
5. **Keep `dedup-acceptance.md` linked from `AGENTS.md`** so the next session finds it without re-deriving the rationale. Currently it's in `docs/` but not referenced anywhere.
6. **Add an art-dupl CI gate** that fails on any new clone group at `-t 5` (currently 0 groups) so future regressions are caught immediately.
7. **Capture the pre-dedup baseline report** (`art-dupl --type-aware -t 3 --html` snapshot) in `docs/deduplication/` for before/after evidence. Right now we have the after-state only.
8. **Document `GOEXPERIMENT=jsonv2` in `AGENTS.md`** under "Essential Commands" / test commands — discoverable today only by hitting the build failure first.

### f) Up to 50 Things To Get Done Next

Ordered by impact (Pareto: high impact first):

1. Wire `cmd/main.go` to `config.LoadConfig()` and delete the duplicate const block.
2. Decide policy: keep `cmd/main.go` standalone-demo vs make it a real config-driven entry point. Update `AGENTS.md` accordingly.
3. Add `//go:build goexperiment.jsonv2` to `user_enums.go`, `user_session.go`, `values_test.go` so tests run without the env-var.
4. Bump `go.mod` `go` directive to 1.27 and drop the experiment tags.
5. Add an `art-dupl` step to `.github/workflows/` (or the project's pre-commit / nix flake check) at `-t 5` that fails on new clone groups.
6. Snapshot the baseline (`-t 3` HTML) into `docs/deduplication/2026-07-28_BASELINE.html`.
7. Reference `docs/dedup-acceptance.md` from `AGENTS.md` "Project Documentation Files" table.
8. Add a "How to run tests" section to `AGENTS.md` that mentions `GOEXPERIMENT=jsonv2`.
9. Fix the 7 `wrapcheck` warnings in `internal/domain/values/{user_enums,user_id,user_session}.go` (wrap every external-package return with `fmt.Errorf("...: %w", err)` or the project's error helpers).
10. Fix the 6 `recvcheck` warnings in `user_enums.go` (choose pointer or value receiver consistently per type).
11. Fix the 4 `makezero` warnings in `ids.go` (`bytes` slices: `make([]byte, 0, idByteLength)` or use `bytes = make([]byte, n)` per makezero's rule).
12. Fix the `godoclint` warning on `IsGeneratedUserID` (start godoc with the symbol name).
13. Fix the `varnamelen` warnings on `validateUserID` / `validateSessionID` `id` parameters (rename to `value`).
14. Fix the `nonamedreturns` warning (one in project).
15. Fix the `paralleltest` warning on `TestIDs` (call `t.Parallel()`).
16. Fix the `gochecknoglobals` warning (one in project, likely the `emailRegex`).
17. Fix the `exhaustruct` warnings on `log.Options`, `httputil.ServerConfig`, and `config.Config` literal initialisations.
18. Fix the `gocritic: exitAfterDefer` warning in `cmd/main.go` (`os.Exit` after `defer cancel()`).
19. Fix the `wrapcheck` warning in `internal/config/config.go:210` (validator.Struct return).
20. Re-run `golangci-lint run ./...` after the above to confirm zero project-wide warnings.
21. Run `go-arch-lint` (the project's own architecture linter demo) and confirm zero violations.
22. Run `govulncheck ./...` and review any findings.
23. Run `nilaway ./...` (mentioned as 80% nil-panic reduction in `AGENTS.md`) and confirm.
24. Generate fresh architecture graphs via `just graph-all` and commit the SVGs into `docs/graphs/`.
25. Add BDD coverage for `internal/domain/ids/ids.go` edge cases (whitespace, unicode, boundary lengths) — currently only 16 specs, all happy-path-ish.
26. Add table-driven tests for `validateEmailLocalPart` and `validateEmailDomain` exercising the new `validateNoEdgeDots` helper.
27. Add table-driven tests for `validateIDRequired` exercising empty, whitespace-only, internal-whitespace inputs.
28. Add a `Domain() == ""` test for an `Email{}` zero value to confirm the `splitParts` guard still triggers the empty-string path.
29. Audit the `internal/domain/values` package for further duplication at threshold 4 (current threshold-3 sweep may have missed sub-token clones).
30. Audit `internal/domain/services` for duplication (samber/lo patterns often hide clones).
31. Audit `internal/application/handlers` for handler boilerplate duplication.
32. Audit `web/templates/` for templ component duplication (probably out of art-dupl scope but worth a manual look).
33. Convert remaining `//nolint:legacyerrors` directives (golangci-lint warned "unknown linter") to current directive names.
34. Migrate `justfile` (if any legacy one exists) to `flake.nix` per `AGENTS.md` policy ("never create new justfiles").
35. Audit `pkg/errors` for duplicated error-construction patterns now that we have three new helpers (`validateNoEdgeDots`, `validateIDRequired`, `splitParts`) — consider exposing them in `pkg/errors` if cross-package value emerges.
36. Add `docs/architecture-decisions/` ADR for the dedup-acceptance policy (when to refactor vs when to record + accept).
37. Add a "Deduplication" section to `AGENTS.md` linking to the skill, the acceptance doc, and the CI gate plan.
38. Decide and document whether `internal/config/config.go` const-block should also be moved to `var` or kept as `const` (current is `const` for immutability, which is correct — but verify).
39. Profile the impact of `splitParts` vs inline `strings.Split` — the new helper allocates the same slice once per call, no change in allocation behaviour, but worth a `go test -bench` to confirm.
40. Run `go test -race ./...` to ensure no data race was introduced by the helper extraction.
41. Run `go test -cover ./...` and confirm the refactored packages (`internal/domain/values`, `internal/domain/ids`) maintain their coverage ratio.
42. Add `t.Parallel()` to the `ids_test.go` Describe blocks per `paralleltest` finding.
43. Review `IsGeneratedUserID` for correct godoc (current `// IsGenerated reports whether...` should be `// IsGeneratedUserID reports whether...`).
44. Consider extracting a generic `validateNonEmptyNoEdgeWhitespace(s, name)` helper in `pkg/errors` so future value objects can reuse the pattern from `validateIDRequired`.
45. Update `docs/dedup-acceptance.md` next-review date (currently 2026-07-28) to trigger re-evaluation in 6 months or on policy change.
46. Add a `just dedup-check` (or flake equivalent) recipe that runs `art-dupl --type-aware -t 5` and exits non-zero on any clone.
47. Document the dedup workflow (run art-dupl → read HTML → for each group: extract / accept / exclude → re-run → commit) in `AGENTS.md` or a new `docs/deduplication/README.md`.
48. Consider bumping `art-dupl` threshold from default 5 to 7 in CI gates (since at threshold 5 we already have 0 groups, raising the bar gives earlier signal on regressions).
49. Investigate whether `golangci-lint`'s `dupl` linter catches anything `art-dupl` misses (they have different algorithms).
50. Once `cmd/main.go` is config-wired, delete the now-redundant const block in `internal/config/config.go` and confirm `art-dupl` reports **0 clone groups at threshold 3**.

### g) Questions I Cannot Figure Out Myself

1. **Should `cmd/main.go` be a true config-driven entry point, or stay a zero-dependency demo?** Wiring it to `config.LoadConfig()` is the "correct" cleanup but contradicts the project's self-described purpose as a "Simple HTMX Demo" / "Educational Resource" (per `AGENTS.md`). I can't read the maintainer's intent on whether the demo simplicity is a feature to preserve or a temporary scaffold to retire.

2. **Is the `GOEXPERIMENT=jsonv2` workaround acceptable long-term, or should we bump `go.mod` to 1.27?** The project's `go.mod` says `go 1.26.5` and the imports use `encoding/json/v2` (which needs 1.27). Bumping is a one-line fix but may break downstream consumers / Nix builds pinned to 1.26. I don't know the project's release-stability policy.

3. **Should `docs/dedup-acceptance.md` be linked from `AGENTS.md` "Project Documentation Files" table, or stay as a stand-alone reference?** I added it under `docs/` but did not wire it into the documented memory hierarchy. I don't know whether the maintainer prefers every doc cross-referenced from `AGENTS.md` or whether `docs/` is intentionally self-discovering.
