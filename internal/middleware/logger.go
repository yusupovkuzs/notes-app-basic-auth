package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// [INFO] POST /users - 201 Created (15ms)
// - метод
// - путь
// - статус
// - время выполнения

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			sw := &statusResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(sw, r)

			duration := time.Since(start)

			log.Info(
				r.Method+" "+r.URL.Path,
				slog.Int("status", sw.status),
				slog.String("status_text", http.StatusText(sw.status)),
				slog.Duration("duration", duration),
			)
		})
	}
}

// time=2026-01-23T16:39:51.598+03:00 level=INFO msg="POST /users" status=400 status_text="Bad Request" duration=25.7912ms
