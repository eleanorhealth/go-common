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

func IsAny(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", msg, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	return Wrap(err, fmt.Sprintf(format, args...))
}
