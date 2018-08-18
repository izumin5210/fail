package fail

import (
	pkgerrors "github.com/pkg/errors"
)

type pkgError struct {
	Err        error
	Message    string
	StackTrace StackTrace
}

// extractPkgError extracts the innermost error from the given error.
// It converts the stack trace that is annotated by pkg/errors into fail.StackTrace.
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
	var stackTraces []StackTrace
	for {
		if t, ok := rootErr.(traceable); ok {
			stackTrace := convertStackTrace(t.StackTrace())
			stackTraces = append(stackTraces, stackTrace)
		}

		if cause, ok := rootErr.(causer); ok {
			rootErr = cause.Cause()
			continue
		}

		break
	}

	var msg string
	if err.Error() != rootErr.Error() {
		msg = err.Error()
	}

	return pkgError{
		Err:        rootErr,
		Message:    msg,
		StackTrace: reduceStackTraces(stackTraces),
	}
}

// convertStackTrace converts pkg/errors.StackTrace into fail.StackTrace
func convertStackTrace(stackTrace pkgerrors.StackTrace) (frames StackTrace) {
	for _, t := range stackTrace {
		if frame, ok := newFrameFrom(uintptr(t)); ok {
			frames = append(frames, frame)
		}
	}
	return
}
