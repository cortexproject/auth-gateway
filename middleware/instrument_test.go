package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusRecorder_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	statusRecorder := StatusRecorder{
		ResponseWriter: rec,
		Status:         http.StatusOK,
	}

	statusRecorder.WriteHeader(http.StatusForbidden)

	assert.Equal(t, http.StatusForbidden, statusRecorder.Status)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
