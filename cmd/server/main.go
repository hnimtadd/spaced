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
	baseDir string
}

func (fs httpFs) Open(name string) (fs.File, error) {
	name = filepath.Join(fs.baseDir, name)
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
	var staticFilesDir string
	if len(os.Args) == 1 {
		staticFilesDir = "./" // Current directory
	} else {
		// go run ./cmd/server folder
		staticFilesDir = os.Args[1]
	}

	if _, err := os.Stat(staticFilesDir); os.IsNotExist(err) {
		log.Fatalf("Error: Static files directory '%s' does not exist. Please create it and place your files there.", staticFilesDir)
	}
	svc := http.NewServeMux()
	fs := http.FileServerFS(httpFs{
		baseDir: staticFilesDir,
	})

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
