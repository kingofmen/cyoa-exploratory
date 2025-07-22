// Package narrate defines an interface for creating story text
// from game events.
package narrate

import (
	"context"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

type Narrator interface {
	Event(context.Context, *storypb.GameEvent, *storypb.GameEvent) (string, error)
}

// Debug is a narrator suitable for tests, which merely
// echoes back the title of the provided action.
type Debug struct{}

func (d *Debug) Event(_ context.Context, ostate, nstate *storypb.GameEvent) (string, error) {
	return ostate.GetPlayerAction().GetTitle(), nil
}

// Noop is a placeholder narrator which returns an empty string.
type Noop struct{}

func (d *Noop) Event(_ context.Context, o, n *storypb.GameEvent) (string, error) {
	return "", nil
}

func NewDebug() Narrator {
	return &Debug{}
}

func NewNoop() Narrator {
	return &Noop{}
}
