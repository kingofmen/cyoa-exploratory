// Package narrate defines an interface for creating story text
// from game events.
package narrate

import (
	"context"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

type Narrator interface {
	Event(context.Context, *storypb.GameEvent) (string, error)
}

// Default is a very silly placeholder narrator which merely
// echoes back the title of the provided action.
type Default struct{}

func (d *Default) Event(_ context.Context, event *storypb.GameEvent) (string, error) {
	return event.GetAction().GetTitle(), nil
}

func DefaultNarrator() Narrator {
	return &Default{}
}
