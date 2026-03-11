# infra

Package `infra` contains opinionated setup helpers for the infrastructure
components that every Eleanor Health service needs: PostgreSQL (via `sql`
and `pgx`), structured logging (zerolog), HTTP tracing (DataDog), and
Google Cloud Pub/Sub.

---

## Database — `database/sql` + pgx driver

### DB

```go
func DB(ctx context.Context, connString string, traceServiceName string) (*sql.DB, error)
```

Opens a `database/sql` pool backed by the pgx driver with DataDog APM
tracing enabled. Steps taken:

1. Parses `connString` as a pgx connection config.
2. If `CLOUD_SQL_INSTANCE` is set, replaces the dial function with the
   Cloud SQL private-IP dialer.
3. Registers the pgx driver with `dd-trace-go` under `traceServiceName`.
4. Applies pool tuning from environment variables (see [`env`](./env.md)).
5. Pings with retry until the database is reachable.

Default `MaxOpenConns` is **5** when `DB_MAX_OPEN_CONNS` is not set.

### SQLExecutorQuerier

Wraps a `*sql.DB` and implements `DBExecutorQuerier` using
[scany](https://github.com/georgysavva/scany) for row scanning.

```go
type DBExecutor interface {
    Execute(ctx context.Context, query string, args ...any) (int64, error)
}

type DBQuerier interface {
    Query(ctx context.Context, dst any, query string, args ...any) error
    QueryRow(ctx context.Context, dst any, query string, args ...any) error
}
```

`Execute` returns the number of rows affected. `Query` scans multiple
rows into `dst` (a slice pointer). `QueryRow` scans a single row.

---

## Database — pgx pool

### PgxPool

```go
func PgxPool(ctx context.Context, connString string, traceServiceName string) (*pgxpool.Pool, error)
```

Creates a `pgxpool.Pool` with the same Cloud SQL dialer support as `DB`.
The pgx pool default is the greater of 4 or `runtime.NumCPU()`.

> **Note:** DataDog tracing support for pgx is pending upstream; the
> `traceServiceName` parameter is accepted but not yet wired.

### PgxPoolExecutorQuerier

Same `DBExecutorQuerier` interface as `SQLExecutorQuerier`, backed by
`*pgxpool.Pool` and scany's `pgxscan` adapter.

---

## Logging

### Logger

```go
func Logger(envKey, fallback string) zerolog.Logger
```

Creates a zerolog logger configured for Google Cloud Logging:

- `timestamp` field name, RFC3339Nano format.
- `severity` level field name (matches GCP structured log format).
- Level is read from the environment variable named by `envKey`; if
  absent or unparseable, falls back to `fallback` (typically `"info"`).
- If `ZEROLOG_CONSOLE_WRITER=true`, switches to `ConsoleLogger()` for
  human-readable local output.

### ConsoleLogger

```go
func ConsoleLogger() zerolog.Logger
```

Returns a zerolog logger that writes coloured, human-readable output to
stdout. Intended for local development.

---

## HTTP tracing

### HTTPTracedTransport

```go
func HTTPTracedTransport(rt http.RoundTripper, serviceName string, optionFns ...HTTPTracedTransportOptionFn) http.RoundTripper
```

Wraps an `http.RoundTripper` with DataDog APM tracing. Each outbound
request automatically tags the span with service name, HTTP method, URL
path, target host, and user agent.

Use `WithHTTPTracedTransportBefore` to attach additional tags from the
request or span:

```go
transport := infra.HTTPTracedTransport(
    http.DefaultTransport,
    "my-service",
    infra.WithHTTPTracedTransportBefore(func(req *http.Request, span ddtrace.Span) {
        span.SetTag("tenant_id", req.Header.Get("X-Tenant-ID"))
    }),
)
httpClient := &http.Client{Transport: transport}
```

---

## Pub/Sub

### Publishing

```go
type PubsubMessagePublisher interface {
    Publish(ctx context.Context, msg *pubsub.Message) error
}
```

| Type | Behaviour |
|------|-----------|
| `PubsubPublisher` (via `NewPubsubPublisher(topic)`) | Publishes to a GCP Pub/Sub topic, blocks until the server acknowledges |
| `NopPublisher` | No-op; always returns `nil` |

### Receiving

```go
type PubsubMessageReceiver interface {
    Receive(ctx context.Context, f func(context.Context, *PubsubMessage)) error
}
```

`PubsubMessage` wraps the raw GCP message and exposes `Ack()` and
`Nack()` methods. The receiver calls `f` for each message; the caller is
responsible for calling `msg.Ack()` or `msg.Nack()`.

| Type | Behaviour |
|------|-----------|
| `PubsubReceiver` (via `NewPubsubReceiver(subscription)`) | Delegates to `subscription.Receive` |
| `NopReceiver` | No-op; returns `nil` immediately |
