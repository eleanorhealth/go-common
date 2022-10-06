package request

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Client interface {
	Request(ctx context.Context, method, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
	Get(ctx context.Context, path string, query url.Values, headers http.Header, out interface{}) (*http.Response, error)
	Post(ctx context.Context, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
	Put(ctx context.Context, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
	Patch(ctx context.Context, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
	Delete(ctx context.Context, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error)
}
