# date

Package `date` contains utility functions for common date/time operations
not covered by the standard library.

## Functions

### BOD / EOD

```go
func BOD(t time.Time, loc *time.Location) time.Time
func EOD(t time.Time, loc *time.Location) time.Time
```

Return the beginning of the day (`00:00:00`) and end of the day
(`23:59:59`) for `t` in the given timezone. When `loc` is `nil`, the
timezone of `t` itself is used.

### Days

```go
func Days(duration time.Duration) int
```

Converts a `time.Duration` to a whole number of days, floored (e.g.
`25h` → `1`).

### WithinDuration

```go
func WithinDuration(expected, actual time.Time, delta time.Duration) bool
```

Returns `true` when `|expected − actual| < delta`. Useful in tests to
assert that a timestamp is approximately correct without requiring an
exact match.

### ParseAny

```go
func ParseAny(layouts []string, dateString string) (time.Time, error)
```

Tries each layout in order via `time.Parse` and returns the first
successful result. Returns `ErrNoLayoutMatched` if none of the layouts
match.

## Errors

| Sentinel | Meaning |
|----------|---------|
| `ErrNoLayoutMatched` | No layout in the slice matched the input string |

## Example

```go
layouts := []string{"2006-01-02", "01/02/2006"}
t, err := date.ParseAny(layouts, "2024-03-15")

bod := date.BOD(t, time.UTC)   // 2024-03-15 00:00:00 UTC
eod := date.EOD(t, time.UTC)   // 2024-03-15 23:59:59 UTC
```
