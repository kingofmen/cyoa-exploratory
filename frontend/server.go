// Package server implements a template-based HTTP server.
package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// IndexData holds data for the front page.
type IndexData struct {
	Timestamp string
}

// Handler handles incoming requests. It implements http.Handler.
type Handler struct {
	index *template.Template
}

// NewHandler returns an initialized Handler object.
func NewHandler() *Handler {
	return &Handler{
		index: template.Must(template.ParseFiles("content/index.html")),
	}
}

// ServeHTTP writes a response to the request into the writer.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	data := IndexData{
		Timestamp: fmt.Sprintf("%s", time.Now()),
	}
	h.index.Execute(w, data)
}
