package narrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

const (
	grokURL = "https://api.x.ai/v1/chat/completions"
)

// Grokker is a Narrator which talks to Grok.
type Grokker struct {
	apiKey string
}

func (d *Grokker) Event(ctx context.Context, ostate, nstate *storypb.GameEvent) (string, error) {
	goData := map[string]any{
		"messages": []map[string]string{
			map[string]string{
				"role":    "system",
				"content": "TODO",
			},
			map[string]string{
				"role":    "user",
				"content": "TODO",
			},
		},
		"model":  "grok-3",
		"stream": false,
	}
	jsonData, err := json.Marshal(goData)
	if err != nil {
		return "", fmt.Errorf("could not marshal narration data: %w", err)
	}
	req, err := http.NewRequest(http.MethodGet, grokURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer debug placeholder")

	// TODO: HTTP client goes here.

	return fmt.Sprintf("%v", req), nil
}

func NewGrokker(ak string) *Grokker {
	return &Grokker{
		apiKey: ak,
	}
}
