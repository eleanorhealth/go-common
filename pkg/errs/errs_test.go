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

	var err error
	err = Wrap(err, "nil error")
	assert.Nil(err)

	err = errors.New("test")
	wrapped := Wrap(err, "test message")
	unwrapped := errors.Unwrap(wrapped)

	assert.Equal(err, unwrapped)
}

func TestWrapf(t *testing.T) {
	assert := assert.New(t)

	var err error
	err = Wrap(err, "nil error")
	assert.Nil(err)

	err = errors.New("test")
	wrapped := Wrapf(err, "this is a wrapped %s", "error")
	cause := Cause(wrapped)

	assert.Equal(err, cause)
	assert.Equal("this is a wrapped error: test", wrapped.Error())
}

func TestIsAny(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	err3 := errors.New("err3")

	type args struct {
		err  error
		errs []error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "none",
			args: args{
				err:  err1,
				errs: []error{err2, err3},
			},
			want: false,
		},
		{
			name: "one",
			args: args{
				err:  err1,
				errs: []error{err3, err2, err1},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAny(tt.args.err, tt.args.errs...); got != tt.want {
				t.Errorf("IsAny() = %v, want %v", got, tt.want)
			}
		})
	}
}
