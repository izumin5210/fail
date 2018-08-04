package fail

import (
	"errors"
	"fmt"
)

// Error is an error that has contextual metadata
type Error struct {
	// Err is the original error (you might call it the root cause)
	Err error
	// Message is an annotated description of the error
	Message string
	// StatusCode is a status code that is desired to be used for a HTTP response
	StatusCode interface{}
	// Report represents whether the error should be reported to administrators
	Report bool
	// Tags represents tags of the error which is classified errors.
	Tags []string
	// Params is an annotated parameters of the error.
	Params H
	// StackTrace is a stack trace of the original error
	// from the point where it was created
	StackTrace StackTrace
}

// New returns an error that formats as the given text.
// It also annotates the error with a stack trace from the point it was called
func New(text string) error {
	return &Error{
		Err:        errors.New(text),
		StackTrace: newStackTrace(0),
	}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// It also annotates the error with a stack trace from the point it was called
func Errorf(format string, args ...interface{}) error {
	return &Error{
		Err:        fmt.Errorf(format, args...),
		StackTrace: newStackTrace(0),
	}
}

// Error implements error interface
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return e.Err.Error()
}

// Copy creates a copy of the current object
func (e *Error) Copy() *Error {
	return &Error{
		Err:        e.Err,
		Message:    e.Message,
		StatusCode: e.StatusCode,
		Report:     e.Report,
		Tags:       e.Tags,
		Params:     e.Params,
		StackTrace: e.StackTrace,
	}
}

// Wrap returns an error annotated with a stack trace from the point it was called.
// It returns nil if err is nil
func Wrap(err error, opts ...Option) error {
	if err == nil {
		return nil
	}

	appErr := wrap(err)

	for _, f := range opts {
		f(appErr)
	}

	return appErr
}

func wrap(err error) *Error {
	pkgErr := extractPkgError(err)

	if appErr, ok := pkgErr.Err.(*Error); ok {
		return appErr.Copy()
	}

	stackTrace := pkgErr.StackTrace
	if stackTrace == nil {
		stackTrace = newStackTrace(1)
	}

	return &Error{
		Err:        pkgErr.Err,
		StackTrace: stackTrace,
		Message:    pkgErr.Message,
	}
}

// Unwrap extracts an underlying *fail.Error from an error.
// If the given error isn't eligible for retriving context from,
// it returns nil
func Unwrap(err error) *Error {
	if appErr, ok := err.(*Error); ok {
		return appErr
	}

	return nil
}
