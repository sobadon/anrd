package errutil

import "errors"

type InternalError struct {
	err error
}

func NewInternalError(msg string) InternalError {
	return InternalError{err: errors.New(msg)}
}

func (e InternalError) Error() string {
	return e.err.Error()
}
