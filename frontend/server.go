// Package server implements a template-based HTTP server.
package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

const (
	CreateLocationURL      = "/locations/create"
	UpdateLocationURL      = "/location/update"
	VueEditStoryURL        = "/edit_story"
	CreateOrUpdateStoryURL = "/api/story/update"

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
	Stories          []*storypb.Story
	CurrentStoryJSON string
	EditStoryURI     string

	CreateLocTitle   string
	CreateLocContent string
	UpdateLocId      string
	UpdateLocTitle   string
	UpdateLocContent string
	DeleteLoc        string
}

// Handler handles incoming requests. It implements http.Handler.
type Handler struct {
	index    *template.Template
	location *template.Template
	vuetmpl  *template.Template
	client   spb.CyoaClient
}

// NewHandler returns an initialized Handler object.
func NewHandler(cl spb.CyoaClient) *Handler {
	return &Handler{
		index:    template.Must(template.ParseFiles("frontend/content/index.html")),
		location: template.Must(template.ParseFiles("frontend/content/location.html")),
		vuetmpl:  template.Must(template.ParseFiles("frontend/story_editor_app/dist/story_editor.html")),
		client:   cl,
	}
}

func makeKey(ctx, key string) string {
	return fmt.Sprintf("%s_%s", ctx, key)
}

func makeIndexData() indexData {
	return indexData{
		Timestamp:    fmt.Sprintf("%s", time.Now()),
		EditStoryURI: VueEditStoryURL,
	}
}

// ServeHTTP writes a response to the request into the writer.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	strResp, err := h.client.ListStories(req.Context(), &spb.ListStoriesRequest{})
	if err != nil {
		http.Error(w, fmt.Errorf("could not load stories: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	data := makeIndexData()
	data.Stories = strResp.GetStories()
	if err := h.index.Execute(w, data); err != nil {
		log.Printf("Template error: %v", err)
	}
}

// CreateLocation passes the request to the gRPC backend and returns
// the created location.
func (h *Handler) CreateLocation(w http.ResponseWriter, req *http.Request) {
	data := makeIndexData()
	title := req.FormValue(data.CreateLocTitle)
	content := req.FormValue(data.CreateLocContent)
	locData := &storypb.Location{
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

// deleteLocation deletes the location with the given ID.
func (h *Handler) deleteLocation(ctx context.Context, locID int64) error {
	_, err := h.client.DeleteLocation(ctx, &spb.DeleteLocationRequest{LocationId: proto.Int64(locID)})
	return err
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
		LocationId: proto.Int64(locID),
		Location: &storypb.Location{
			Id:      proto.Int64(locID),
			Title:   proto.String(title),
			Content: proto.String(content),
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

// VueExperimentalHandler handles the experimental Vue story editor.
func (h *Handler) VueExperimentalHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	params := req.URL.Query()
	data := makeIndexData()
	if strid := params.Get("story_id"); len(strid) > 0 {
		sid, err := strconv.ParseInt(strid, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot edit story with bad ID %q: %v", strid, err), http.StatusBadRequest)
			return
		}
		resp, err := h.client.GetStory(ctx, &spb.GetStoryRequest{
			Id: proto.Int64(sid),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot find story with ID %q: %v", strid, err), http.StatusBadRequest)
			return
		}
		bts, err := protojson.Marshal(resp.GetStory())
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling proto: %v", err), http.StatusInternalServerError)
			return
		}
		data.CurrentStoryJSON = string(bts)
	}
	if err := h.vuetmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// CreateOrUpdateStoryHandler saves story data to the database.
func (h *Handler) CreateOrUpdateStoryHandler(w http.ResponseWriter, req *http.Request) {
	bts, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not read request body: %v", err), http.StatusBadRequest)
		return
	}
	str := &storypb.Story{}
	if err := protojson.Unmarshal(bts, str); err != nil {
		http.Error(w, fmt.Sprintf("could not parse Story object: %v", err), http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	updResp := &spb.UpdateStoryResponse{}
	if str.GetId() > 0 {
		updResp, err = h.client.UpdateStory(ctx, &spb.UpdateStoryRequest{
			Story: str,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("update error: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		cr, err := h.client.CreateStory(ctx, &spb.CreateStoryRequest{
			Story: str,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("create error: %v", err), http.StatusInternalServerError)
			return
		}
		updResp.Story = cr.GetStory()
	}

	bts, err = protojson.Marshal(updResp)
	if err != nil {
		http.Error(w, fmt.Sprintf("error marshaling proto: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(bts); err != nil {
		http.Error(w, fmt.Sprintf("error writing JSON: %v", err), http.StatusInternalServerError)
		return
	}
}
