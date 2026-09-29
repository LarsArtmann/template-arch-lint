# Template Architecture Lint

**A working Go codebase with Clean Architecture boundaries enforced by [go-arch-lint](https://github.com/fe3dback/go-arch-lint), plus a code-quality layer generated and kept current by [golangci-lint-auto-configure](https://github.com/LarsArtmann/golangci-lint-auto-configure).**

This is a template to copy from, not a library to install.

## Why this exists

Keeping a Go architecture honest takes two mechanisms:

1. **Layer boundaries.** The domain must stay pure; dependencies flow infrastructure → application → domain. Humans fail at enforcing this in code review. `go-arch-lint` enforces it mechanically, on every commit.
2. **Code quality.** 100+ linters, security scanning, complexity and function-size limits. A hand-maintained `.golangci.yml` rots: linters get deprecated, releases move on. `golangci-lint-auto-configure` generates the config for your project and keeps repairing it as golangci-lint evolves.

This repository shows both running against a real (deliberately simple) HTTP application and hands you the configuration files.

## What to take

| Asset | Action | What it gives you |
| --- | --- | --- |
| [`.go-arch-lint.yml`](.go-arch-lint.yml) | Copy | Clean Architecture components and forbidden-dependency rules |
| [`.go-arch-lint-strict.yml`](.go-arch-lint-strict.yml) | Copy | Stricter variant for maximum boundary enforcement |
| [`.golangci.yml`](.golangci.yml) | Regenerate, don't copy | The kitchen-sink example this repo runs; generate a fresh one for your project with the tool below |
| [`.custom-gcl.yml`](.custom-gcl.yml) + [`pkg/linter-plugins/template-arch-lint/`](pkg/linter-plugins/template-arch-lint) | Copy if needed | Custom golangci-lint plugin with project-shape checks standard linters lack |
| [`internal/`](internal) | Read | The reference implementation the rules run against |

## Quick start in your project

```bash
# 1. Architecture boundaries
curl -fsSL -o .go-arch-lint.yml \
  https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.go-arch-lint.yml
go install github.com/fe3dback/go-arch-lint@latest
go-arch-lint check

# 2. Code-quality config, generated for your project
go install github.com/larsartmann/golangci-lint-auto-configure/cmd/golangci-lint-auto-configure@latest
golangci-lint-auto-configure configure --priority optional
```

Adapt the component paths in `.go-arch-lint.yml` to your layout, then rerun `go-arch-lint check` until it passes. From then on, `golangci-lint-auto-configure audit` keeps the lint config current across golangci-lint releases.

## What the reference code demonstrates

Layer structure (dependency flow: infrastructure → application → domain):

```
cmd/                          single entry point (enforced by the plugin's cmd-single rule)
internal/domain/              pure business logic, zero external imports
  entities/  values/          rich entities; validated value objects (Email, UserName, UserID)
  services/                   domain services in samber/lo functional style
  repositories/               interfaces only
  shared/                     Result pattern for functional error handling
internal/application/         HTTP handlers, DTOs, middleware, response helpers
internal/infrastructure/      repository implementations, external concerns
internal/db/                  sqlc-generated type-safe queries (SQLite)
sql/                          schema and query sources
```

Patterns worth stealing:

- **Value objects** — `Email`, `UserName`, `UserID` validate at construction; invalid states cannot exist
- **Result pattern** — `internal/domain/shared/result.go` chains success/failure paths without panics
- **Repository pattern** — domain interfaces, infrastructure implementations, in-memory fakes for tests
- **Functional style** — `lo.Map` / `lo.Filter` / `lo.Reduce` in domain services

## The custom golangci-lint plugin

[`pkg/linter-plugins/template-arch-lint`](pkg/linter-plugins/template-arch-lint) is a plugin module wired through [`.custom-gcl.yml`](.custom-gcl.yml), adding checks standard linters do not cover:

| Rule | Enforces |
| --- | --- |
| `cmd-single-main` | exactly one `main.go` under `cmd/` |
| `import-cycle-detector` | no import cycles, direct or indirect |
| `code-duplication-detector` | copy-paste blocks above a token threshold |
| `filename-validator` | file/package naming alignment |

Build it with `golangci-lint custom` (reads `.custom-gcl.yml`, produces a custom golangci-lint binary) and run that binary instead of plain `golangci-lint`.

## Development in this repo

```bash
nix develop            # Go toolchain and friends
nix build              # demo binary → ./result/bin/template-arch-lint
go test ./... -race    # BDD-style suites (Ginkgo/Gomega)
go-arch-lint check     # architecture boundaries
```

Generated architecture graphs live in [`docs/graphs/`](docs/graphs).

## Scope

The demo app (user CRUD over SQLite) is intentionally boring. The point is the enforcement, not the features: copy the configs, don't copy the app.

## Contributing

- Follow the existing code style
- Add tests for new features
- Run `go-arch-lint check` and the full lint suite before submitting

## License

MIT — see [LICENSE](LICENSE).

## Contact

LarsArtmann

Project Link: [https://github.com/LarsArtmann/template-arch-lint](https://github.com/LarsArtmann/template-arch-lint)
