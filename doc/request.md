# request

Package `request` provides an opinionated HTTP client with built-in
support for authentication, custom error checking, and pluggable response
unmarshalling.

## Client interface

```go
type Client interface {
    Request(ctx context.Context, method, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
    Get(ctx context.Context, path string, query url.Values, headers http.Header, out interface{}) (*http.Response, error)
    Post(ctx context.Context, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
    Put(ctx context.Context, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
    Patch(ctx context.Context, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
    Delete(ctx context.Context, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
}
```

- `path` is appended to the `baseURL` set at construction. A leading `/`
  is added automatically if missing.
- `out` is unmarshalled from the response body when non-nil (JSON by
  default). Pass `nil` to ignore the body.
- The response body is always fully read and buffered before returning, so
  callers can re-read `resp.Body` if needed.

## Construction

```go
func NewClient(baseURL string, options ...option) *client
```

```go
c := request.NewClient("https://api.example.com",
    request.WithBearerTokenAuth(token),
)
```

## Options

| Option | Effect |
|--------|--------|
| `WithUserAgent(ua)` | Overrides the default `go-common/request@v0.0.0` User-Agent |
| `WithHTTPClient(hc)` | Replaces the underlying `*http.Client` (e.g. to add tracing transport) |
| `WithBearerTokenAuth(token)` | Adds `Authorization: Bearer <token>` to every request |
| `WithBasicAuth(user, pass)` | Adds HTTP Basic authentication to every request |
| `WithErrChecker(fn)` | Replaces the default error checker (see below) |
| `WithResponseUnmarshaler(fn)` | Replaces `json.Unmarshal` for response body decoding |

## Error handling

By default, any response outside the 2xx range produces an `*HTTPError`:

```go
type HTTPError struct {
    HTTPResponse *http.Response
    StatusCode   int
}
```

Inspect the status code or the full response:

```go
res, err := c.Get(ctx, "/users/123", nil, nil, &user)
var httpErr *request.HTTPError
if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
    // handle 404
}
```

Swap the checker entirely with `WithErrChecker` to implement custom
logic (e.g. treat 404 as a non-error for idempotent deletes).

## Example

```go
type Appointment struct {
    ID        string `json:"id"`
    PatientID string `json:"patient_id"`
}

c := request.NewClient(
    "https://scheduling-api.internal",
    request.WithBearerTokenAuth(os.Getenv("API_TOKEN")),
    request.WithHTTPClient(&http.Client{
        Transport: infra.HTTPTracedTransport(http.DefaultTransport, "scheduling-client"),
        Timeout:   10 * time.Second,
    }),
)

var appt Appointment
_, err := c.Get(ctx, "/appointments/"+id, nil, nil, &appt)
```
