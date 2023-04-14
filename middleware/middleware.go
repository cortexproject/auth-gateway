package middleware

import (
	"net/http"
)

type Interface interface {
	Wrap(http.Handler) http.Handler
}

type Adapter func(http.Handler) http.Handler

func (m Adapter) Wrap(next http.Handler) http.Handler {
	return m(next)
}

func Merge(middlesware ...Interface) Interface {
	return Adapter(func(next http.Handler) http.Handler {
		for i := len(middlesware) - 1; i >= 0; i-- {
			next = middlesware[i].Wrap(next)
		}
		return next
	})
}
