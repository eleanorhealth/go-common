# filelog

Package `filelog` provides a minimal `Logger` interface for appending
plain-text messages to a file, plus a no-op implementation for tests and
environments where file logging is disabled.

## Interface

```go
type Logger interface {
    Log(msg string) error
}
```

## Implementations

### FileLogger

```go
type FileLogger struct {
    Path string
}
```

Opens the file at `Path` in append mode on each call to `Log`, writes
`msg + "\n"`, then closes the file. The file is created if it does not
exist. All flags include `os.ModeAppend`.

Returns a wrapped error if the file cannot be opened or written to.

### Nop

```go
type Nop struct{}
```

Implements `Logger` as a no-op. `Log` always returns `nil`. Useful in
tests or when the consumer accepts a `Logger` but logging is not needed.

## Example

```go
var logger filelog.Logger

if cfg.AuditLogPath != "" {
    logger = filelog.FileLogger{Path: cfg.AuditLogPath}
} else {
    logger = filelog.Nop{}
}

logger.Log("user 123 performed action X")
```
