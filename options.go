package fail

// Option annotates an errors.
type Option func(*Error)

// WithMessage annotates with the message.
func WithMessage(msg string) Option {
	return func(err *Error) {
		err.Message = msg
	}
}

// WithCode annotates with the code.
func WithCode(code interface{}) Option {
	return func(err *Error) {
		err.Code = code
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
