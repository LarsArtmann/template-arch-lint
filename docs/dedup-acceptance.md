# Duplication Acceptance Notes

These clone groups were intentionally kept after deduplication review.
Each entry records the location and a one-line rationale so future
sessions do not re-evaluate them.

## cmd/main.go vs internal/config/config.go (server timeout defaults)

**Locations:**
- `cmd/main.go:26-29` — `defaultServerReadTimeout=15s`, `defaultServerWriteTimeout=15s`,
  `defaultServerIdleTimeout=60s`, `defaultGracefulTimeout=30s`
- `internal/config/config.go:17-20` — `defaultServerReadTimeout=5s`,
  `defaultServerWriteTimeout=10s`, `defaultServerIdleTimeout=120s`,
  `defaultGracefulShutdownTimeout=30s`

**Rationale:** Different values for different layers.
`cmd/main.go` is the "Simple HTMX Demo" entry point (per `AGENTS.md`) that
intentionally avoids the viper dependency so the template boots without
configuration. `internal/config/config.go` provides viper-driven defaults
for production-grade deployments. The structural similarity is incidental;
the values are independently chosen for their respective contexts. Wiring
`cmd/main.go` to `config.LoadConfig()` would require introducing viper and
a config-translation layer in `main`, defeating the demo-first purpose of
the template.

Reviewed: 2026-07-28.