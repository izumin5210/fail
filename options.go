package apperrors

// Option annotates an errors.
type Option func(*Error)

// WithMessage annotates with the message.
func WithMessage(msg string) Option {
	return func(err *Error) {
		err.Message = msg
	}
}

// WithStatusCode annotates with the status code.
func WithStatusCode(code int) Option {
	return func(err *Error) {
		err.StatusCode = code
	}
}

// WithReport annotates with the reportability.
func WithReport() Option {
	return func(err *Error) {
		err.Report = true
	}
}

// WithTags annotates with tags.
func WithTags(tags ...string) Option {
	return func(err *Error) {
		err.Tags = append(err.Tags, tags...)
	}
}
