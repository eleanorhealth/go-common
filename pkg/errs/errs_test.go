package errs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCause(t *testing.T) {
	assert := assert.New(t)

	cause := Cause(nil)
	assert.Nil(cause)

	err := errors.New("test")
	cause = Cause(err)
	assert.Equal(err, cause)

	wrapped := Wrap(err, "test message")
	cause = Cause(wrapped)
	assert.Equal(err, cause)

	doubleWrapped := Wrap(wrapped, "double wrapped")
	cause = Cause(doubleWrapped)
	assert.Equal(err, cause)
}

func TestWrap(t *testing.T) {
	assert := assert.New(t)

	err := errors.New("test")
	wrapped := Wrap(err, "test message")
	unwrapped := errors.Unwrap(wrapped)

	assert.Equal(err, unwrapped)
}

func TestWrapf(t *testing.T) {
	assert := assert.New(t)

	err := errors.New("test")
	wrapped := Wrapf(err, "this is a wrapped %s", "error")
	cause := Cause(wrapped)
	assert.Equal(err, cause)
	assert.Equal("this is a wrapped error: test", wrapped.Error())
}
