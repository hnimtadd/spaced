package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

func main() {
	staticFilesDir := "./" // Current directory

	if _, err := os.Stat(staticFilesDir); os.IsNotExist(err) {
		log.Fatalf("Error: Static files directory '%s' does not exist. Please create it and place your files there.", staticFilesDir)
	}
	svc := http.NewServeMux()
	fs := http.FileServerFS(httpFs{})

	svc.HandleFunc("/api/sound/index", loggingMiddlewareFunc(disableCacheMiddlewareFunc(http.HandlerFunc(handler.Handler))))
	svc.Handle("/", loggingMiddleware(disableCacheMiddelware(fs)))

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
