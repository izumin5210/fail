package fail

import (
	"errors"
	"fmt"
	"strings"
)

const (
	messageDelimiter = ": "
)

// Error is an error that has contextual metadata
type Error struct {
	// Err is the original error (you might call it the root cause)
	Err error
	// Messages is an annotated description of the error
	Messages []string
	// Code is a status code that is desired to be contained in responses, such as HTTP Status code.
	Code interface{}
	// Ignorable represents whether the error should be reported to administrators
	Ignorable bool
	// Tags represents tags of the error which is classified errors.
	Tags []string
	// Params is an annotated parameters of the error.
	Params H
	// StackTrace is a stack trace of the original error
	// from the point where it was created
	StackTrace StackTrace
}

// New returns an error that formats as the given text.
// It also records the stack trace at the point it was called.
func New(text string) error {
	err := &Error{Err: errors.New(text)}
	withStackTrace(0)(err)
	return err
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// It also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	err := &Error{Err: fmt.Errorf(format, args...)}
	withStackTrace(0)(err)
	return err
}

// Error implements error interface.
// It returns a string of messages and the root error concatenated with ": ".
func (e *Error) Error() string {
	messages := append(e.Messages, e.Err.Error())
	return strings.Join(messages, messageDelimiter)
}

// Copy creates a copy of the current object
func (e *Error) Copy() *Error {
	return &Error{
		Err:        e.Err,
		Messages:   e.Messages,
		Code:       e.Code,
		Ignorable:  e.Ignorable,
		Tags:       e.Tags,
		Params:     e.Params,
		StackTrace: e.StackTrace,
	}
}

// LastMessage returns the last message
func (e *Error) LastMessage() string {
	if len(e.Messages) == 0 {
		return ""
	}
	return e.Messages[0]
}

// Wrap returns an error annotated with a stack trace from the point it was called,
// and with the specified annotators.
// It returns nil if err is nil.
func Wrap(err error, annotators ...Annotator) error {
	if err == nil {
		return nil
	}

	failErr := Unwrap(err)
	if failErr == nil {
		failErr = &Error{
			Err: err,
		}
	}

	withStackTrace(0)(failErr)

	for _, f := range annotators {
		f(failErr)
	}

	return failErr
}

// Unwrap extracts an underlying *fail.Error from an error.
// If the given error isn't eligible for retriving context from,
// it returns nil
func Unwrap(err error) *Error {
	if err == nil {
		return nil
	}

	if failErr, ok := err.(*Error); ok {
		return failErr.Copy()
	}

	if failErr := convertPkgError(err); failErr != nil {
		return failErr
	}

	return nil
}
