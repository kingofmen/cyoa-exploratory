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

	"github.com/google/uuid"
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
	DeleteStoryURL         = "/api/story/delete"

	createCtx  = "create"
	updateCtx  = "update"
	titleKey   = "title_key"
	contentKey = "content_key"
	locIdKey   = "location_id_key"
	deleteKey  = "delete_key"
	storyIdKey = "story_id"
)

// indexData holds data for the front page.
type indexData struct {
	Timestamp        string
	Stories          []*storypb.Story
	CurrentStoryJSON string
	EditStoryURI     string
	DeleteStoryURI   string
	StoryIdKey       string

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
		Timestamp:      fmt.Sprintf("%s", time.Now()),
		EditStoryURI:   VueEditStoryURL,
		DeleteStoryURI: DeleteStoryURL,
		StoryIdKey:     storyIdKey,
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
func (h *Handler) deleteLocation(ctx context.Context, locID string) error {
	_, err := h.client.DeleteLocation(ctx, &spb.DeleteLocationRequest{LocationId: proto.String(locID)})
	return err
}

// updateLocation updates the provided location.
func (h *Handler) updateLocation(ctx context.Context, locID string, title, content string) error {
	_, err := h.client.GetLocation(ctx, &spb.GetLocationRequest{LocationId: proto.String(locID)})
	if err != nil {
		return fmt.Errorf("error fetching location to prepare update for ID %s: %v", locID, err)
	}

	if _, err = h.client.UpdateLocation(ctx, &spb.UpdateLocationRequest{
		LocationId: proto.String(locID),
		Location: &storypb.Location{
			Id:      proto.String(locID),
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
	lid := req.FormValue(data.UpdateLocId)
	newTitle := req.FormValue(data.UpdateLocTitle)
	newContent := req.FormValue(data.UpdateLocContent)
	deleteFlag := req.FormValue(data.DeleteLoc) == data.DeleteLoc

	if lid == "" {
		http.Error(w, "Location ID is required for update/delete.", http.StatusBadRequest)
		return
	}

	if err := uuid.Validate(lid); err != nil {
		http.Error(w, fmt.Sprintf("Invalid Location ID %q: %v", lid, err), http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	if deleteFlag {
		if err := h.deleteLocation(ctx, lid); err != nil {
			http.Error(w, fmt.Sprintf("Error deleting location with ID %d: %v", lid, err), http.StatusInternalServerError)
			return
		}
		log.Printf("Location with ID %d deleted by frontend handler.", lid)
	} else {
		if err := h.updateLocation(ctx, lid, newTitle, newContent); err != nil {
			http.Error(w, fmt.Sprintf("Error updating location with ID %s: %v", lid, err), http.StatusInternalServerError)
			return
		}
		log.Printf("Location with ID %d updated by frontend handler.", lid)
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

// VueExperimentalHandler handles the experimental Vue story editor.
func (h *Handler) VueExperimentalHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	params := req.URL.Query()
	data := makeIndexData()
	if strid := params.Get(storyIdKey); len(strid) > 0 {
		sid, err := strconv.ParseInt(strid, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot edit story with bad ID %q: %v", strid, err), http.StatusBadRequest)
			return
		}
		resp, err := h.client.GetStory(ctx, &spb.GetStoryRequest{
			Id:   proto.Int64(sid),
			View: spb.StoryView_VIEW_CONTENT.Enum(),
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
	updReq := &spb.UpdateStoryRequest{
		Story: str,
	}

	updResp, err := h.client.UpdateStory(ctx, updReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("backend error: %v", err), http.StatusInternalServerError)
		return
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

// DeleteStory deletes the story with the given ID.
func (h *Handler) DeleteStoryHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	params := req.URL.Query()
	if strid := params.Get(storyIdKey); len(strid) > 0 {
		sid, err := strconv.ParseInt(strid, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot edit story with bad ID %q: %v", strid, err), http.StatusBadRequest)
			return
		}
		if sid > 0 {
			if _, err := h.client.DeleteStory(ctx, &spb.DeleteStoryRequest{
				Id: proto.Int64(sid),
			}); err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete story with ID %q: %v", strid, err), http.StatusBadRequest)
				return
			}
		}
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
