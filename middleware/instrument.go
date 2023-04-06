package middleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type Instrument struct {
	Duration *prometheus.HistogramVec
}

func (i Instrument) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// send metrics from here

		// startTime := time.Now()
		// duration := time.Since(startTime)
		// requestDuration.WithLabelValues(r.Method, r.URL.Path, "status_code", "false").Observe(duration.Seconds())

		next.ServeHTTP(w, r)
	})
}
