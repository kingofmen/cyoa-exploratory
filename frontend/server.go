// Package server implements a template-based HTTP server.
package server

import (
	//"context"
	"fmt"
	"html/template"
	"log" // Added
	"net/http"
	"strconv" // Added
	"time"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
)

const (
	CreateLocationURL   = "/locations/create"
	createLocTitleKey   = "create_location_title"
	createLocContentKey = "create_location_content"

	UpdateLocationURL   = "/location/update"
	updateLocIdKey      = "location_id"
	updateLocContentKey = "content"
	deleteLocKey        = "delete"
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

// UpdateLocationHandler handles updates or deletions of locations.
func (h *Handler) UpdateLocationHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusBadRequest)
		return
	}

	locIDStr := req.FormValue(updateLocIdKey)
	newContentFromForm := req.FormValue(updateLocContentKey)
	deleteFlag := req.FormValue(deleteLocKey) == "true"

	if locIDStr == "" {
		http.Error(w, "Location ID is required for update/delete.", http.StatusBadRequest)
		return
	}

	locID, err := strconv.ParseInt(locIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid Location ID format: %v", err), http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	if deleteFlag {
		_, err := h.client.DeleteLocation(ctx, &spb.DeleteLocationRequest{LocationId: locID})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error deleting location with ID %d: %v", locID, err), http.StatusInternalServerError)
			return
		}
		log.Printf("Location with ID %d marked for deletion by frontend handler.", locID)
	} else {
		// Update operation
		listResp, err := h.client.ListLocations(ctx, &spb.ListLocationsRequest{})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching locations to prepare update for ID %d: %v", locID, err), http.StatusInternalServerError)
			return
		}

		var originalTitlePtr *string
		found := false
		for _, loc := range listResp.GetLocations() {
			if loc.GetId() == locID {
				originalTitlePtr = loc.Title
				found = true
				break
			}
		}

		if !found {
			http.Error(w, fmt.Sprintf("Location with ID %d not found, cannot update.", locID), http.StatusNotFound)
			return
		}

		titleForUpdate := originalTitlePtr
		if originalTitlePtr == nil {
			// The backend's UpdateLocationImpl requires a non-nil title.
			emptyStr := ""
			titleForUpdate = &emptyStr
		}

		locationToUpdate := &spb.Location{
			Id:      locID,
			Title:   titleForUpdate,
			Content: &newContentFromForm,
		}

		_, err = h.client.UpdateLocation(ctx, &spb.UpdateLocationRequest{
			LocationId: locID,
			Location:   locationToUpdate,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating location with ID %d: %v", locID, err), http.StatusInternalServerError)
			return
		}
		log.Printf("Location with ID %d updated by frontend handler.", locID)
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
