# clock

Package `clock` provides a small abstraction over `time.Now()` so that
code depending on the current time can be tested without real-time
coupling.

## Interface

```go
type Clocker interface {
    Now() time.Time
}
```

## Implementations

| Type | Behaviour |
|------|-----------|
| `defaultClock` (via `NewDefaultClock()`) | Calls `time.Now()` |

In tests, inject any `Clocker` implementation that returns a fixed or
controllable `time.Time`.

## Example

```go
type Service struct {
    clock clock.Clocker
}

func (s *Service) IsExpired(expiry time.Time) bool {
    return s.clock.Now().After(expiry)
}

// Production
svc := &Service{clock: clock.NewDefaultClock()}

// Test
svc := &Service{clock: &fixedClock{t: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}}
```
