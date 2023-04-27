package connectors

import (
	"context"
	"errors"
	"fmt"
)

// ErrorCode describes the error's category.
type ErrorCode int

const (
	// ErrorCodeOK is returned by the Code function on a nil error. It is not a valid
	// code for an error.
	ErrorCodeOK ErrorCode = iota

	// ErrorCodeUnknown means that the error could not be categorized.
	ErrorCodeUnknown

	// ErrorCodeInternal means that something unexpected happened. Internal errors always indicate
	// bugs in the code (or possibly the underlying service).
	ErrorCodeInternal

	// ErrorCodePermissionDenied means that the caller does not have permission to execute the specified operation.
	ErrorCodePermissionDenied

	// ErrorCodeCanceled means that the operation was canceled.
	ErrorCodeCanceled

	// ErrorCodeDeadlineExceeded means that the operation timed out.
	ErrorCodeDeadlineExceeded
)

// run `stringer --type ErrorCode --trimprefix ErrorCode` whenever changing the above list of error codes.

// Error describes connector error
type Error struct {
	// Code is the error code.
	Code ErrorCode
	msg  string
	err  error
}

// Error returns the error as a string.
func (e *Error) Error() string {
	if e.msg == "" {
		return fmt.Sprintf("code=%v", e.Code)
	}
	return fmt.Sprintf("%s (code=%v)", e.msg, e.Code)
}

// Unwrap returns the error underlying the receiver, which may be nil.
func (e *Error) Unwrap() error {
	return e.err
}

// NewError returns a new error with the given code, underlying error and message.
func NewError(c ErrorCode, err error, msg string) *Error {
	return &Error{
		Code: c,
		msg:  msg,
		err:  err,
	}
}

func Code(err error) ErrorCode {
	if err == nil {
		return ErrorCodeOK
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	if errors.Is(err, context.Canceled) {
		return ErrorCodeCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorCodeDeadlineExceeded
	}
	return ErrorCodeUnknown
}
