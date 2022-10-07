package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	baseURL := "http://scheduling"
	userAgent := "test-agent"

	client := NewClient(
		baseURL,
		WithUserAgent(userAgent),
	)
	assert.NotNil(client)

	assert.Equal(userAgent, client.userAgent)
}

func TestClient_bearerTokenAuth(t *testing.T) {
	assert := assert.New(t)

	token := "foobar"

	authOK := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		s := strings.Split(authHeader, "Bearer ")
		if len(s) < 2 {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if s[1] == token {
			authOK = true
		}
	}))

	httpClient := server.Client()
	baseURL := server.URL

	client := NewClient(baseURL, WithHTTPClient(httpClient), WithBearerTokenAuth(token))

	res, err := client.Get(context.Background(), "/", nil, nil, nil)
	assert.NotNil(res)
	assert.NoError(err)
	assert.Equal(200, res.StatusCode)
	assert.True(authOK)
}

func TestClient_basicAuth(t *testing.T) {
	assert := assert.New(t)

	user := "user"
	pass := "password"

	authOK := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, _ := r.BasicAuth()

		if u == user && p == pass {
			authOK = true
		}
	}))

	httpClient := server.Client()
	baseURL := server.URL

	client := NewClient(baseURL, WithHTTPClient(httpClient), WithBasicAuth(user, pass))

	res, err := client.Get(context.Background(), "/", nil, nil, nil)
	assert.NotNil(res)
	assert.NoError(err)
	assert.Equal(200, res.StatusCode)
	assert.True(authOK)
}

func TestClient_Get(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodGet, r.Method)

		w.Write([]byte(`{"foo": "bar"}`))
	}))

	baseURL := server.URL

	client := NewClient(baseURL, WithHTTPClient(server.Client()))

	out := map[string]string{}

	res, err := client.Get(context.Background(), "/test-get", nil, nil, &out)
	assert.NotNil(res)
	assert.NoError(err)

	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal("bar", out["foo"])
}

func TestClient_Post(t *testing.T) {
	assert := assert.New(t)

	body := []byte(`{"foo": "bar"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodPost, r.Method)

		reqBytes, _ := io.ReadAll(r.Body)
		assert.Equal(body, reqBytes)

		w.Write([]byte(`{"cat": "baz"}`))
	}))

	baseURL := server.URL

	client := NewClient(baseURL, WithHTTPClient(server.Client()))

	out := map[string]string{}

	res, err := client.Post(context.Background(), "/test-post", bytes.NewReader(body), nil, &out)
	assert.NotNil(res)
	assert.NoError(err)

	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal("baz", out["cat"])
}

func TestClient_Put(t *testing.T) {
	assert := assert.New(t)

	body := []byte(`{"foo": "bar"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodPut, r.Method)

		reqBytes, _ := io.ReadAll(r.Body)
		assert.Equal(body, reqBytes)

		w.Write([]byte(`{"cat": "baz"}`))
	}))

	baseURL := server.URL

	client := NewClient(baseURL, WithHTTPClient(server.Client()))

	out := map[string]string{}

	res, err := client.Put(context.Background(), "/test-put", bytes.NewReader(body), nil, &out)
	assert.NotNil(res)
	assert.NoError(err)

	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal("baz", out["cat"])
}

func TestClient_Patch(t *testing.T) {
	assert := assert.New(t)

	body := []byte(`{"foo": "bar"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodPatch, r.Method)

		reqBytes, _ := io.ReadAll(r.Body)
		assert.Equal(body, reqBytes)

		w.Write([]byte(`{"cat": "baz"}`))
	}))

	baseURL := server.URL

	client := NewClient(baseURL, WithHTTPClient(server.Client()))

	out := map[string]string{}

	res, err := client.Patch(context.Background(), "/test-patch", bytes.NewReader(body), nil, &out)
	assert.NotNil(res)
	assert.NoError(err)

	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal("baz", out["cat"])
}

func TestClient_Delete(t *testing.T) {
	assert := assert.New(t)

	body := []byte(`{"foo": "bar"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodDelete, r.Method)

		reqBytes, _ := io.ReadAll(r.Body)
		assert.Equal(body, reqBytes)

		w.Write([]byte(`{"cat": "baz"}`))
	}))

	baseURL := server.URL

	client := NewClient(baseURL, WithHTTPClient(server.Client()))

	out := map[string]string{}

	res, err := client.Delete(context.Background(), "/test-delete", bytes.NewReader(body), nil, &out)
	assert.NotNil(res)
	assert.NoError(err)

	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal("baz", out["cat"])
}
