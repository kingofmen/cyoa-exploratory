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

const (
	CreateLocationURL   = "/locations/create"
	createLocTitleKey   = "create_location_title"
	createLocContentKey = "create_location_content"
)

// indexData holds data for the front page.
type indexData struct {
	Timestamp        string
	Locations        []*spb.Location
	CreateLoc        string
	CreateLocTitle   string
	CreateLocContent string
}

// locationData holds data to display a Location.
type locationData struct {
	Proto *spb.Location
}

// Handler handles incoming requests. It implements http.Handler.
type Handler struct {
	index    *template.Template
	location *template.Template
	client   spb.CyoaClient
}

// NewHandler returns an initialized Handler object.
func NewHandler(cl spb.CyoaClient) *Handler {
	return &Handler{
		index:    template.Must(template.ParseFiles("frontend/content/index.html")),
		location: template.Must(template.ParseFiles("frontend/content/location.html")),
		client:   cl,
	}
}

// ServeHTTP writes a response to the request into the writer.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	locResp, err := h.client.ListLocations(req.Context(), &spb.ListLocationsRequest{})
	if err != nil {
		http.Error(w, fmt.Errorf("could not load locations: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	data := indexData{
		Timestamp:        fmt.Sprintf("%s", time.Now()),
		Locations:        locResp.GetLocations(),
		CreateLoc:        CreateLocationURL,
		CreateLocTitle:   createLocTitleKey,
		CreateLocContent: createLocContentKey,
	}
	h.index.Execute(w, data)
}

// CreateLocation passes the request to the gRPC backend and returns
// the created location.
func (h *Handler) CreateLocation(w http.ResponseWriter, req *http.Request) {
	title := req.FormValue(createLocTitleKey)
	content := req.FormValue(createLocContentKey)
	locData := &locationData{
		Proto: &spb.Location{
			Title:   &title,
			Content: &content,
		},
	}
	_, err := h.client.CreateLocation(req.Context(), &spb.CreateLocationRequest{
		Location: locData.Proto,
	})
	if err != nil {
		http.Error(w, fmt.Errorf("error creating location: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	h.location.Execute(w, &locData)
}
