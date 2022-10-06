package request

import "net/http"

type option func(c *client)

func WithUserAgent(userAgent string) option {
	return func(c *client) {
		c.userAgent = userAgent
	}
}

// WithContentType the content type of the request body
func WithContentType(contentType string) option {
	return func(c *client) {
		c.contentType = contentType
	}
}

func WithHTTPClient(httpClient *http.Client) option {
	return func(c *client) {
		c.httpClient = httpClient
	}
}

func WithTokenAuth(token string) option {
	return func(c *client) {
		c.token = token
		c.authType = authTypeToken
	}
}

func WithBasicAuth(user, pass string) option {
	return func(c *client) {
		c.basicUser = user
		c.basicPass = pass
		c.authType = authTypeBasic
	}
}

func WithErrChecker(errChecker HTTPErrChecker) option {
	return func(c *client) {
		c.errChecker = errChecker
	}
}

func WithResponseUnmarshaler(responseUnmarshaler ResponseUnmarshaler) option {
	return func(c *client) {
		c.responseUnmarshaler = responseUnmarshaler
	}
}
