package fail

// Annotator is a function that annotates an error with information
type Annotator func(*Error)

// WithMessage annotates an error with the message
func WithMessage(msg string) Annotator {
	return func(err *Error) {
		if msg == "" {
			return
		}
		err.Messages = append([]string{msg}, err.Messages...)
	}
}

// WithCode annotates an error with the code
func WithCode(code interface{}) Annotator {
	return func(err *Error) {
		err.Code = code
	}
}

// WithIgnorable annotates an error with the reportability
func WithIgnorable() Annotator {
	return func(err *Error) {
		err.Ignorable = true
	}
}

// WithTags annotates an error with tags
func WithTags(tags ...string) Annotator {
	return func(err *Error) {
		err.Tags = append(err.Tags, tags...)
	}
}

// WithParam annotates an error with a key-value pair
func WithParam(key string, value interface{}) Annotator {
	return WithParams(H{key: value})
}

// WithParams annotates an error with key-value pairs
func WithParams(h H) Annotator {
	return func(err *Error) {
		err.Params = err.Params.Merge(h)
	}
}

// withStackTrace annotates an error with the stack trace from the point it was called
func withStackTrace(offset int) Annotator {
	stackTrace := newStackTrace(offset + 1)

	return func(err *Error) {
		err.StackTrace = mergeStackTraces(err.StackTrace, stackTrace)
	}
}
