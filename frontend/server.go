// Package server implements a template-based HTTP server.
package server

import (
	//"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
)

// IndexData holds data for the front page.
type IndexData struct {
	Timestamp string
	Locations []*spb.Location
}

// Handler handles incoming requests. It implements http.Handler.
type Handler struct {
	index  *template.Template
	client spb.CyoaClient
}

// NewHandler returns an initialized Handler object.
func NewHandler(cl spb.CyoaClient) *Handler {
	return &Handler{
		index:  template.Must(template.ParseFiles("frontend/content/index.html")),
		client: cl,
	}
}

// ServeHTTP writes a response to the request into the writer.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	locResp, err := h.client.ListLocations(req.Context(), &spb.ListLocationsRequest{})
	if err != nil {
		http.Error(w, fmt.Errorf("could not load locations: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	data := IndexData{
		Timestamp: fmt.Sprintf("%s", time.Now()),
		Locations: locResp.GetLocations(),
	}
	h.index.Execute(w, data)
}
