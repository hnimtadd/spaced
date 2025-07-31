package main

import (
	"fmt"
	"net/http"

	handler "github.com/hnimtadd/spaced/api/sound"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("StatusOK"))
}

func main() {
	svc := http.NewServeMux()
	svc.HandleFunc("/healthcheck", Handler)
	svc.HandleFunc("/api/sound/index", handler.Handler)
	fmt.Println("Listening on :8088")
	if err := http.ListenAndServe(":8088", svc); err != nil {
		fmt.Println(err)
	}
}
