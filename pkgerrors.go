package fail

import (
	"strings"

	pkgerrors "github.com/pkg/errors"
)

type pkgError struct {
	Err        error
	Messages   []string
	StackTrace StackTrace
}

const (
	pkgErrorsMessageDelimiter = ": "
)

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

	var stackTraces []StackTrace
	var messages []string
	var lastMessage string

	// Retrive stacks and trace back the root cause
	rootErr := err
	for {
		if t, ok := rootErr.(traceable); ok {
			stackTrace := convertStackTrace(t.StackTrace())
			stackTraces = append(stackTraces, stackTrace)
		}

		if cause, ok := rootErr.(causer); ok {
			msg := rootErr.Error()
			if lastMessage != msg {
				lastMessage = msg
				messages = append(messages, msg)
			}

			rootErr = cause.Cause()
			continue
		}

		break
	}

	// Extract annotated messages by removing the trailing message.
	//
	// w2 := errors.Wrap(e0, "message 2") // w2.Error() == "mesasge 2: message 1: e0"
	// w1 := errors.Wrap(e0, "message 1") // w1.Error() ==            "message 1: e0"
	// e0 := errors.New("e0")             // e0.Error() ==                       "e0"
	//
	//                       "e0"
	//                          \
	//                           '-.
	//                              \
	//            "message 1: e0" : "e0" --> ": e0" --> "messages 1"
	//                          \
	//                           '-.
	//                              \
	// "mesasge 2: message 1: e0" : "message 1: e0" --> ": message 1: e0" --> "messages 2"
	trailingMessage := rootErr.Error()
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.HasSuffix(messages[i], pkgErrorsMessageDelimiter+trailingMessage) {
			trimed := strings.TrimSuffix(messages[i], pkgErrorsMessageDelimiter+trailingMessage)
			trailingMessage = messages[i]
			messages[i] = trimed
		}
	}

	return pkgError{
		Err:        rootErr,
		Messages:   messages,
		StackTrace: reduceStackTraces(stackTraces),
	}
}

// convertStackTrace converts pkg/errors.StackTrace into fail.StackTrace
func convertStackTrace(stackTrace pkgerrors.StackTrace) StackTrace {
	pcs := make([]uintptr, len(stackTrace))
	for i, t := range stackTrace {
		pcs[i] = uintptr(t)
	}
	return newStackTraceFromPCs(pcs)
}
