package internalhttp

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		ip, err := getIP(r)
		if err != nil {
			log.Errorf("failed to get client IP: %v", err)
		}
		log.WithField("ip", ip).WithField("method", r.Method).WithField("path", r.URL).
			WithField("HTTP version", r.Proto).WithField("user-agent", r.Header.Get("user-agent")).
			WithField("latency", time.Since(start)).
			Info("http request processed")
	})
}
