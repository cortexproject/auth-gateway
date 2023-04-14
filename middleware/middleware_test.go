package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdapter_Wrap(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test handler"))
	})

	adapter := Adapter(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("adapter "))
			next.ServeHTTP(w, r)
		})
	})

	wrappedHandler := adapter.Wrap(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, "adapter test handler", rec.Body.String())
}

func TestMerge(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("handler"))
	})

	middleware1 := Adapter(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("1"))
			next.ServeHTTP(w, r)
		})
	})

	middleware2 := Adapter(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("2"))
			next.ServeHTTP(w, r)
		})
	})

	mergedMiddleware := Merge(middleware1, middleware2)
	wrappedHandler := mergedMiddleware.Wrap(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, "12handler", rec.Body.String())
}
