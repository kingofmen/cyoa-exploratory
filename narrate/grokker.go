package narrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

const (
	grokURL = "https://api.x.ai/v1/chat/completions"

	// bkupTmpl holds the last-ditch backup template that will keep
	// the Grokker limping along if the database-stored one fails
	// to load or parse.
	bkupTmpl = `The current location is {{ .Ostate.Location.Title }} with the detailed description {{ .Ostate.Location.Description }}.
The player has chosen the action {{ .Ostate.PlayerAction.Title }} with the detailed description {{ .Ostate.PlayerAction.Description }}.
{{ if len .Nstate.Effects }}
  This triggered the following effects:
  {{ range $effect := .Nstate.Effects }}
   * {{ $effect }}
  {{ end }}
{{ end }}
{{ $oid := printf .Ostate.Location.Id }}
{{ $nid := printf .Nstate.Location.Id }}
{{ if ne $oid $nid }}
This results in the player arriving in {{ .Nstate.Location.Title }} with detailed description {{ .Nstate.Location.Description }}.
{{ end }}
Narrate the action and its results.
`
)

// Grokker is a Narrator which talks to Grok.
type Grokker struct {
	apiKey string
	client *http.Client
	tmpl   *template.Template
	debug  bool
}

// Event calls the Grok API to create the narration of an event.
func (d *Grokker) Event(ctx context.Context, ostate, nstate *storypb.GameEvent) (string, error) {
	var buf bytes.Buffer
	if err := d.tmpl.Execute(&buf, struct {
		Ostate *storypb.GameEvent
		Nstate *storypb.GameEvent
	}{
		Ostate: ostate,
		Nstate: nstate,
	}); err != nil {
		return "", fmt.Errorf("error with template: %w", err)
	}

	prompt := buf.String()
	if d.debug {
		return prompt, nil
	}

	// TODO: These strings should not be hardcoded.
	goData := map[string]any{
		"messages": []map[string]string{
			map[string]string{
				"role": "system",
				"content": `You are the narrator of a choose-your-own-adventure game.
                    Return a text that describes the player's chosen action,
                    its outcome, and any new location they enter, as outlined
                    in the prompt. Do not return any header or footer material
                    or comments on the request - just the narration text.
                   `,
			},
			map[string]string{
				"role":    "user",
				"content": prompt,
			},
		},
		"model":  "grok-3",
		"stream": false,
	}
	jsonData, err := json.Marshal(goData)
	if err != nil {
		return "", fmt.Errorf("could not marshal narration data: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, grokURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error contacting narrator: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK status from narrator: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading HTTP response from narrator: %w", err)
	}

	return string(body), nil
}

func NewGrokker(ak string) *Grokker {
	return &Grokker{
		apiKey: ak,
		client: &http.Client{},
		tmpl:   template.Must(template.New("content").Parse(bkupTmpl)),
		debug:  false,
	}
}

func DebugGrokker() *Grokker {
	return &Grokker{
		tmpl:  template.Must(template.New("content").Parse(bkupTmpl)),
		debug: true,
	}
}
