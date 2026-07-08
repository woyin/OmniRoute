package main

import (
	"embed"
	"io"
	"net/http"
)

//go:embed web/*
var webFS embed.FS

func webFileServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "" {
			path = "/index.html"
		}

		// Try to read the file from embedded FS
		data, err := webFS.ReadFile("web" + path)
		if err != nil {
			// File not found - serve index.html for SPA routing
			data, err = webFS.ReadFile("web/index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			path = "/index.html"
		}

		// Set content type
		contentType := "application/octet-stream"
		switch {
		case len(path) >= 5 && path[len(path)-5:] == ".html":
			contentType = "text/html; charset=utf-8"
		case len(path) >= 4 && path[len(path)-4:] == ".css":
			contentType = "text/css; charset=utf-8"
		case len(path) >= 3 && path[len(path)-3:] == ".js":
			contentType = "application/javascript; charset=utf-8"
		case len(path) >= 5 && path[len(path)-5:] == ".json":
			contentType = "application/json; charset=utf-8"
		case len(path) >= 4 && path[len(path)-4:] == ".svg":
			contentType = "image/svg+xml"
		case len(path) >= 4 && path[len(path)-4:] == ".png":
			contentType = "image/png"
		case len(path) >= 4 && path[len(path)-4:] == ".ico":
			contentType = "image/x-icon"
		}

		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	}
}

// Suppress unused import
var _ = io.EOF
