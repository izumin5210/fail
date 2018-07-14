package apperrors

// H represents a JSON-like key-value object.
type H map[string]interface{}

// Merge returns a new H object contains self and other H contents.
func (h H) Merge(other map[string]interface{}) H {
	out := H{}

	for k, v := range h {
		out[k] = v
	}
	for k, v := range other {
		out[k] = v
	}

	return out
}
