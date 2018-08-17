package fail

// Option annotates an errors.
type Option func(*Error)

// WithMessage annotates with the message.
func WithMessage(msg string) Option {
	return func(err *Error) {
		if msg == "" {
			return
		}
		err.Messages = append([]string{msg}, err.Messages...)
	}
}

// WithStatusCode annotates with the status code.
func WithStatusCode(code interface{}) Option {
	return func(err *Error) {
		err.StatusCode = code
	}
}

// WithIgnorable annotates with the reportability.
func WithIgnorable() Option {
	return func(err *Error) {
		err.Ignorable = true
	}
}

// WithTags annotates with tags.
func WithTags(tags ...string) Option {
	return func(err *Error) {
		err.Tags = append(err.Tags, tags...)
	}
}

// WithParam annotates with a key-value pair.
func WithParam(key string, value interface{}) Option {
	return WithParams(H{key: value})
}

// WithParams annotates with key-value pairs.
func WithParams(h H) Option {
	return func(err *Error) {
		err.Params = err.Params.Merge(h)
	}
}
