package errorx

import "errors"

type Option func(*HttpError)

// WithCause Method
func WithCause(err error) Option {
	return func(e *HttpError) {
		e.cause = err
	}
}

// WithMeta Method
func WithMeta(key string, value any) Option {
	return func(e *HttpError) {
		if e.Meta == nil {
			e.Meta = make(map[string]any)
		}
		e.Meta[key] = value
	}
}

// IsHttpError Method
func IsHttpError(err error) (*HttpError, bool) {
	httpErr, ok := errors.AsType[*HttpError](err)
	return httpErr, ok
}
