package fail

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
			stackTraces = append([]StackTrace{stackTrace}, stackTraces...)
		}

		if cause, ok := rootErr.(causer); ok {
			rootErr = cause.Cause()
			continue
		}

		break
	}

	var stackTrace StackTrace
	if len(stackTraces) > 0 {
		stackTrace = stackTraces[0] // TODO
	}

	var msg string
	if err.Error() != rootErr.Error() {
		msg = err.Error()
	}

	return pkgError{
		Err:        rootErr,
		Message:    msg,
		StackTrace: stackTrace,
	}
}

// convertStackTrace converts pkg/errors.StackTrace into fail.StackTrace
func convertStackTrace(stackTrace pkgerrors.StackTrace) (frames StackTrace) {
	if stackTrace == nil {
		return
	}

	for _, t := range stackTrace {
		file := fmt.Sprintf("%s", t)
		line, _ := strconv.ParseInt(fmt.Sprintf("%d", t), 10, 64)
		funcName := fmt.Sprintf("%n", t)

		frames = append(frames, Frame{
			Func: funcName,
			Line: line,
			File: file,
		})
	}

	return
}
