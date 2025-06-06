// Package server implements a template-based HTTP server.
package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
)

const (
	CreateLocationURL = "/locations/create"
	UpdateLocationURL = "/location/update"

	createCtx  = "create"
	updateCtx  = "update"
	titleKey   = "title_key"
	contentKey = "content_key"
	locIdKey   = "location_id_key"
	deleteKey  = "delete_key"
)

// indexData holds data for the front page.
type indexData struct {
	Timestamp        string
	Locations        []*spb.Location
	CreateLoc        string
	CreateLocTitle   string
	CreateLocContent string

	UpdateLoc        string
	UpdateLocId      string
	UpdateLocTitle   string
	UpdateLocContent string
	DeleteLoc        string
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

func makeKey(ctx, key string) string {
	return fmt.Sprintf("%s_%s", ctx, key)
}

func makeIndexData() indexData {
	return indexData{
		Timestamp:        fmt.Sprintf("%s", time.Now()),
		CreateLoc:        CreateLocationURL,
		CreateLocTitle:   makeKey(createCtx, titleKey),
		CreateLocContent: makeKey(createCtx, contentKey),
		UpdateLoc:        UpdateLocationURL,
		UpdateLocId:      makeKey(updateCtx, locIdKey),
		UpdateLocTitle:   makeKey(updateCtx, titleKey),
		UpdateLocContent: makeKey(updateCtx, contentKey),
		DeleteLoc:        makeKey(updateCtx, deleteKey),
	}
}

// ServeHTTP writes a response to the request into the writer.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	locResp, err := h.client.ListLocations(req.Context(), &spb.ListLocationsRequest{})
	if err != nil {
		http.Error(w, fmt.Errorf("could not load locations: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	data := makeIndexData()
	data.Locations = locResp.GetLocations()
	h.index.Execute(w, data)
}

// CreateLocation passes the request to the gRPC backend and returns
// the created location.
func (h *Handler) CreateLocation(w http.ResponseWriter, req *http.Request) {
	data := makeIndexData()
	title := req.FormValue(data.CreateLocTitle)
	content := req.FormValue(data.CreateLocContent)
	locData := &spb.Location{
		Title:   &title,
		Content: &content,
	}
	_, err := h.client.CreateLocation(req.Context(), &spb.CreateLocationRequest{
		Location: locData,
	})
	if err != nil {
		http.Error(w, fmt.Errorf("error creating location: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Location with title %q updated by frontend handler.", title)
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func ptr(x int64) *int64 {
	val := x
	return &val
}

// deleteLocation deletes the location with the given ID.
func (h *Handler) deleteLocation(ctx context.Context, locID int64) error {
	_, err := h.client.DeleteLocation(ctx, &spb.DeleteLocationRequest{LocationId: ptr(locID)})
	return err
}

func str(s string) *string {
	copy := s
	return &copy
}

// updateLocation updates the provided location.
func (h *Handler) updateLocation(ctx context.Context, locID int64, title, content string) error {
	listResp, err := h.client.ListLocations(ctx, &spb.ListLocationsRequest{})
	if err != nil {
		return fmt.Errorf("error fetching locations to prepare update for ID %d: %v", locID, err)
	}

	found := false
	for _, loc := range listResp.GetLocations() {
		if loc.GetId() == locID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Location with ID %d not found, cannot update.", locID)
	}

	if _, err = h.client.UpdateLocation(ctx, &spb.UpdateLocationRequest{
		LocationId: ptr(locID),
		Location: &spb.Location{
			Id:      ptr(locID),
			Title:   str(title),
			Content: str(content),
		},
	}); err != nil {
		return fmt.Errorf("Error updating location with ID %d: %v", locID, err)
	}
	return nil
}

// UpdateLocationHandler handles updates or deletions of locations.
func (h *Handler) UpdateLocationHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusBadRequest)
		return
	}

	data := makeIndexData()
	locIDStr := req.FormValue(data.UpdateLocId)
	newTitle := req.FormValue(data.UpdateLocTitle)
	newContent := req.FormValue(data.UpdateLocContent)
	deleteFlag := req.FormValue(data.DeleteLoc) == data.DeleteLoc

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
		if err := h.deleteLocation(ctx, locID); err != nil {
			http.Error(w, fmt.Sprintf("Error deleting location with ID %d: %v", locID, err), http.StatusInternalServerError)
			return
		}
		log.Printf("Location with ID %d deleted by frontend handler.", locID)
	} else {
		if err := h.updateLocation(ctx, locID, newTitle, newContent); err != nil {
			http.Error(w, fmt.Sprintf("Error updating location with ID %d: %v", locID, err), http.StatusInternalServerError)
			return
		}
		log.Printf("Location with ID %d updated by frontend handler.", locID)
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
}
