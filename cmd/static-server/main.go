package main

import (
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	// Serve static files from the "web" directory.
	webDir := filepath.Join(".", "web")
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	log.Println("Static file server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
