package database

import "errors"

// ErrNotFound is returned for single row queries that return no values.
var ErrNotFound = errors.New("database: not found")

// NewNotFoundError returns a new error that wraps ErrNotFound so checks with errors.Is(...) work.
func NewNotFoundError(msg string) error {
	return &wrappedError{msg: msg, err: ErrNotFound}
}

// ErrNotUnique is returned when a unique constraint is violated.
var ErrNotUnique = errors.New("database: violates unique constraint")

// NewNotUniqueError returns a new error that wraps ErrNotUnique so checks with errors.Is(...) work.
func NewNotUniqueError(msg string) error {
	return &wrappedError{msg: msg, err: ErrNotUnique}
}

// ErrValidation is returned when a validation check fails.
var ErrValidation = errors.New("database: validation failed")

// NewValidationError returns a new error that wraps ErrValidation so checks with errors.Is(...) work.
func NewValidationError(msg string) error {
	return &wrappedError{msg: msg, err: ErrValidation}
}

type wrappedError struct {
	msg string
	err error
}

func (e *wrappedError) Error() string {
	return e.msg
}

func (e *wrappedError) Unwrap() error {
	return e.err
}
