package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath" // Import for path manipulation
	"time"

	handler "github.com/hnimtadd/spaced/api/sound"
)

type httpFs struct {
	fs.FS
}

func (httpFs) Open(name string) (fs.File, error) {
	retried := false
retry:
	_, err := os.Stat(name)
	if err != nil {
		if retried {
			return nil, err
		}
		name = name + ".html"
		retried = true
		goto retry
	}

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return f, nil
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

func main() {
	staticFilesDir := "./" // Current directory

	if _, err := os.Stat(staticFilesDir); os.IsNotExist(err) {
		log.Fatalf("Error: Static files directory '%s' does not exist. Please create it and place your files there.", staticFilesDir)
	}
	svc := http.NewServeMux()
	fs := http.FileServerFS(httpFs{})

	svc.HandleFunc("/api/sound/index", loggingMiddlewareFunc(http.HandlerFunc(handler.Handler)))
	svc.Handle("/", loggingMiddleware(fs))

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	addr := fmt.Sprintf(":%s", port)

	fmt.Printf("Serving static files from '%s' on http://localhost%s\n", filepath.Join(staticFilesDir), addr)
	fmt.Println("Press Ctrl+C to stop the server.")

	// Start the HTTP server. log.Fatal will print an error and exit if the server fails to start.
	log.Fatal(http.ListenAndServe(addr, svc))
}
