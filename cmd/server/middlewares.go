package main

import (
	"log"
	"net/http"
	"time"
)

// loggingMiddleware wraps an http.Handler to add logging functionality.
func disableCacheMiddelware(next http.Handler) http.Handler {
	return disableCacheMiddlewareFunc(next.ServeHTTP)
}

func disableCacheMiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Serve the request using the next handler in the chain
		log.Println("disabling cache")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache") // For older browsers
		w.Header().Set("Expires", "0")       // For older browsers
		next(w, r)
	}
}

// loggingResponseWriter is a custom http.ResponseWriter to capture the status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before calling the underlying WriteHeader.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware wraps an http.Handler to add logging functionality.
func loggingMiddleware(next http.Handler) http.Handler {
	return loggingMiddlewareFunc(next.ServeHTTP)
}

func loggingMiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // Record start time

		// Create a custom ResponseWriter to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK} // Default to 200 OK

		// Serve the request using the next handler in the chain
		next(lrw, r)

		// Log the access details
		log.Printf("%s %s \"%s\" %d %s",
			r.RemoteAddr,       // Client IP address and port
			r.Method,           // HTTP method (GET, POST, etc.)
			r.URL.RequestURI(), // Requested URI (including query string)
			lrw.statusCode,     // HTTP status code
			time.Since(start),  // Request duration
		)
	}
}
