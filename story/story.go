// Package story implements validates actions and changes story state.
package story

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

// allowed returns true if the action is in the location's available list.
func allowed(act *storypb.Action, loc *storypb.Location) bool {
	for _, av := range loc.GetAvailableActions() {
		if av == act.GetId() {
			return true
		}
	}
	return false
}

func HandleAction(act *storypb.Action, loc *storypb.Location, game *storypb.Playthrough) error {
	if loc.GetId() != game.GetLocation() {
		return fmt.Errorf("cannot apply action %d (%s) to location %d (%s) when current location is %d", act.GetId(), act.GetTitle(), loc.GetId(), loc.GetTitle(), game.GetLocation())
	}
	if !allowed(act, loc) {
		return fmt.Errorf("action %d (%s) not allowed in location %d (%s)", act.GetId(), act.GetTitle(), loc.GetId(), loc.GetTitle())
	}

	for _, eff := range act.GetEffects() {
		if nl := eff.GetNewLocation(); nl != 0 {
			game.Location = proto.Int64(nl)
		}
		if k, v := eff.GetTweakValue(), eff.GetTweakAmount(); len(k) > 0 && v != 0 {
			if len(game.Values) == 0 {
				game.Values = make(map[string]int64)
			}
			game.Values[k] += v
		}
	}
	return nil
}
