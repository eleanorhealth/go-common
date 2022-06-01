package errs

import (
	"errors"
	"fmt"
)

func Cause(err error) error {
	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		return err
	}

	return Cause(unwrapped)
}

func Wrap(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func Wrapf(err error, format string, args ...any) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}
