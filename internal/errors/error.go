package errors

import (
	stderrors "errors"
	"fmt"

	"golang.org/x/xerrors"
)

type Error struct {
	err error
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Errorf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", e, fmt.Sprintf(format, args...))
}

var (
	ErrNotFound = &Error{err: xerrors.New("not found")}
)

func Is(err error, target error) bool {
	return stderrors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}
