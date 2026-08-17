package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return LoggingWithLogger(log.Default())(next)
}

func LoggingWithLogger(logger *log.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = log.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			next.ServeHTTP(w, r)
			logger.Printf(
				"%s %s %s",
				r.Method,
				r.URL.RequestURI(),
				time.Since(startedAt),
			)
		})
	}
}
