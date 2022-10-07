package request

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithUserAgent(t *testing.T) {
	assert := assert.New(t)

	userAgent := "agent"
	o := WithUserAgent(userAgent)

	c := &client{}
	o(c)

	assert.Equal(userAgent, c.userAgent)
}

func TestWithHTTPClient(t *testing.T) {
	assert := assert.New(t)

	httpClient := &http.Client{}
	o := WithHTTPClient(httpClient)

	c := &client{}
	o(c)

	assert.Equal(httpClient, c.httpClient)
}

func TestWithBearerTokenAuth(t *testing.T) {
	assert := assert.New(t)

	token := "token"
	o := WithBearerTokenAuth(token)

	c := &client{}
	o(c)

	assert.Equal(token, c.bearerToken)
}

func TestWithBasicAuth(t *testing.T) {
	assert := assert.New(t)

	user := "user"
	pass := "password"
	o := WithBasicAuth(user, pass)

	c := &client{}
	o(c)

	assert.Equal(user, c.basicUser)
	assert.Equal(pass, c.basicPass)
}

func TestWithErrChecker(t *testing.T) {
	assert := assert.New(t)

	called := false
	errChecker := func(req *http.Request, res *http.Response) error {
		called = true
		return nil
	}

	o := WithErrChecker(errChecker)

	c := &client{}
	o(c)

	c.errChecker(nil, nil)

	assert.True(called)
}

func TestWithResponseUnmarshaler(t *testing.T) {
	assert := assert.New(t)

	called := false
	responseUnmarshaler := func(bytes []byte, out any) error {
		called = true
		return nil
	}

	o := WithResponseUnmarshaler(responseUnmarshaler)

	c := &client{}
	o(c)

	c.responseUnmarshaler(nil, nil)

	assert.True(called)
}
