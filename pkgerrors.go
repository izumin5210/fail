package rrr

import (
	"fmt"
	"strconv"

	pkgerrors "github.com/pkg/errors"
)

type pkgError struct {
	Err        error
	Message    string
	StackTrace StackTrace
}

// extractPkgError extracts the innermost error from the given error.
// It converts the stack trace that is annotated by pkg/errors into rrr.StackTrace.
// If the error doesn't have a stack trace or a causer of pkg/errors,
// it simply returns the original error
func extractPkgError(err error) pkgError {
	type traceable interface {
		StackTrace() pkgerrors.StackTrace
	}
	type causer interface {
		Cause() error
	}

	rootErr := err
	var st pkgerrors.StackTrace
	for {
		if stackTrace, ok := rootErr.(traceable); ok {
			st = stackTrace.StackTrace()
		}

		if cause, ok := rootErr.(causer); ok {
			rootErr = cause.Cause()
			continue
		}

		break
	}

	var frames []Frame
	if st != nil {
		for _, t := range st {
			file := fmt.Sprintf("%s", t)
			line, _ := strconv.ParseInt(fmt.Sprintf("%d", t), 10, 64)
			funcName := fmt.Sprintf("%n", t)

			frames = append(frames, Frame{
				Func: funcName,
				Line: line,
				File: file,
			})
		}
	}

	var msg string
	if err.Error() != rootErr.Error() {
		msg = err.Error()
	}

	return pkgError{
		Err:        rootErr,
		Message:    msg,
		StackTrace: frames,
	}
}
