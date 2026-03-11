# env

Package `env` wraps `os.LookupEnv` with typed parsing and provides a small
global for tracking which environment the process is running in.

## Environment identity

Call `env.Setenv` once at startup with one of the three recognised values:

| Constant | Value |
|----------|-------|
| `EnvLocal` | `"local"` |
| `EnvQA` | `"qa"` |
| `EnvProd` | `"prod"` |

Passing any other value panics immediately—this is intentional to surface
misconfiguration at startup.

```go
env.Setenv(env.EnvProd)

env.IsLocal() // false
env.IsProd()  // true
```

## Reading environment variables

### Get

```go
func Get[T bool | []byte | int | string](key string, defaultVal T) T
```

Returns the typed value of `key`, or `defaultVal` if the variable is
unset or cannot be parsed to `T`.

```go
port := env.Get("PORT", 8080)          // int
debug := env.Get("DEBUG", false)       // bool
token := env.Get("API_TOKEN", "")      // string
raw := env.Get("CERT_BYTES", []byte{}) // []byte
```

### GetExists

```go
func GetExists[T bool | []byte | int | string](key string) (T, bool)
```

Returns the value and a boolean indicating whether the variable was
present at all. Useful when you need to distinguish "not set" from the
zero value.

```go
if timeout, ok := env.GetExists[int]("DB_TIMEOUT"); ok {
    db.SetConnMaxLifetime(time.Duration(timeout) * time.Second)
}
```

### GetString

```go
func GetString(key, defaultVal string) string
```

Like `Get[string]`, but also falls back to `defaultVal` when the
variable exists but is an empty string.

## Environment variables read by `infra.DB`

The `infra` package reads the following vars automatically when
constructing a database connection:

| Variable | Type | Effect |
|----------|------|--------|
| `DB_CONN_MAX_IDLE_TIME` | `int` (minutes) | `db.SetConnMaxIdleTime` |
| `DB_CONN_MAX_LIFETIME` | `int` (minutes) | `db.SetConnMaxLifetime` |
| `DB_MAX_IDLE_CONNS` | `int` | `db.SetMaxIdleConns` |
| `DB_MAX_OPEN_CONNS` | `int` | `db.SetMaxOpenConns` (default 5) |
| `CLOUD_SQL_INSTANCE` | `string` | Enables Cloud SQL private-IP dialer |
| `ZEROLOG_CONSOLE_WRITER` | `bool` | Switches logger to human-readable console output |
