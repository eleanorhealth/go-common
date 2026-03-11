# go-common

`go-common` is a shared Go library for Eleanor Health services. It provides
infrastructure setup helpers, database access patterns, HTTP utilities, and
small general-purpose packages that are useful across the fleet.

**Module:** `github.com/eleanorhealth/go-common`  
**Go version:** 1.24

---

## Package index

| Package | Import path | Purpose |
|---------|-------------|---------|
| [`bao`](./bao.md) | `.../pkg/bao` | Generic CRUD helpers on top of [bun](https://bun.uptrace.dev/) |
| [`bao/hook`](./bao.md#hooks) | `.../pkg/bao/hook` | Before/after hook types for bao operations |
| [`clock`](./clock.md) | `.../pkg/clock` | Testable clock abstraction |
| [`date`](./date.md) | `.../pkg/date` | Time/date utility functions |
| [`env`](./env.md) | `.../pkg/env` | Typed environment variable helpers |
| [`errs`](./errs.md) | `.../pkg/errs` | Error wrapping and inspection helpers |
| [`filelog`](./filelog.md) | `.../pkg/filelog` | Append-only file logger with a no-op implementation |
| [`infra`](./infra.md) | `.../pkg/infra` | Application infrastructure: DB, pgx, logging, HTTP tracing, Pub/Sub |
| [`request`](./request.md) | `.../pkg/request` | Opinionated HTTP client with auth and tracing support |

---

## Design conventions

- **Generics throughout.** Most packages use Go generics to avoid boilerplate
  and keep call sites type-safe.
- **Interfaces for everything injectable.** DB connections, clocks, loggers, and
  publishers all hide behind interfaces so callers can swap in test doubles.
- **No-op variants provided.** Where side effects matter (`filelog.Nop`,
  `infra.NopPublisher`, `infra.NopReceiver`), a no-op implementation ships
  alongside the real one.
- **Errors are wrapped, not replaced.** All internal errors pass through
  `errs.Wrap` / `errs.Wrapf`, so the full call chain is preserved and
  `errors.Is` keeps working.
