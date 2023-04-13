package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestInstrumentWrap(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello, test")
	})

	i := Instrument{
		Duration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "test_request_duration_seconds",
			Help:    "Time taken for requests.",
			Buckets: []float64{0.1, 0.2, 0.5, 1},
		}, []string{"method", "route", "status_code", "ws"}),
	}

	wrappedHandler := i.Wrap(testHandler)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Hello, test\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	expectedCount := 1
	count := testutil.CollectAndCount(i.Duration)
	if count != expectedCount {
		t.Errorf("expected histogram count to be %d, but got %d", expectedCount, count)
	}
}

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
