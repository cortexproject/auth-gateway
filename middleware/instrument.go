package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Instrument struct {
	Duration *prometheus.HistogramVec
}

func (i Instrument) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &StatusRecorder{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}
		next.ServeHTTP(recorder, r)
		var (
			statusCode = strconv.Itoa(recorder.Status)
			took       = time.Since(start).Seconds()
		)
		i.Duration.WithLabelValues(r.Method, r.URL.Path, statusCode, "false").Observe(took)
	})
}

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (sr *StatusRecorder) WriteHeader(status int) {
	sr.Status = status
	sr.ResponseWriter.WriteHeader(status)
}
