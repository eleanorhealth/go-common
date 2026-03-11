# errs

Package `errs` provides thin helpers for wrapping and inspecting errors.
All helpers preserve the error chain so `errors.Is` and `errors.As`
continue to work correctly.

## Functions

### Wrap

```go
func Wrap(err error, msg string) error
```

Returns `fmt.Errorf("%s: %w", msg, err)`. Returns `nil` when `err` is
`nil`, so callers do not need to guard against it:

```go
return errs.Wrap(err, "opening file")
// → "opening file: <original error>"
```

### Wrapf

```go
func Wrapf(err error, format string, args ...any) error
```

Like `Wrap` but accepts a format string for the message:

```go
return errs.Wrapf(err, "reading record %s", id)
```

### Cause

```go
func Cause(err error) error
```

Recursively unwraps `err` until it reaches the root (an error with no
`Unwrap`). Equivalent to `pkg/errors`'s `Cause`.

### IsAny

```go
func IsAny(err error, targets ...error) bool
```

Returns `true` if `errors.Is(err, target)` is true for **any** of the
provided targets. Useful when a function can fail with several distinct
sentinel errors:

```go
if errs.IsAny(err, bao.ErrIDNotUUID, bao.ErrOnePrimaryKey) {
    http.Error(w, "bad request", http.StatusBadRequest)
}
```
