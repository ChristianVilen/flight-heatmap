// Package middleware logs all requests
package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()

		next.ServeHTTP(res, req)

		log.Println(req.Method, req.URL.Path, time.Since(start))
	})
}
