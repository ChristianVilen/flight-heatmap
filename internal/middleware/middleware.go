// Package middleware logs all requests
package middleware

import (
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(middlwareStack ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		// apply each middleware in stack
		for i := len(middlwareStack) - 1; i >= 0; i-- {
			x := middlwareStack[i]
			next = x(next)
		}

		return next
	}
}

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: res,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, req)

		log.Println(wrapped.statusCode, req.Method, req.URL.Path, time.Since(start))
	})
}
