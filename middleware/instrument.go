package middleware

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type Instrument struct {
	Duration *prometheus.HistogramVec
}

func (i Instrument) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// send metrics from here
		fmt.Println("instrument", i.Duration)

		next.ServeHTTP(w, r)
	})
}
