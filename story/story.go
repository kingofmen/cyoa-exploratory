// Package story implements validates actions and changes story state.
package story

import (
	"fmt"
	"log"

	"github.com/kingofmen/cyoa-exploratory/logic"
	"google.golang.org/protobuf/proto"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

// gameState implements logic.Lookup with a Playthrough as its scope.
type gameState struct {
	game *storypb.Playthrough
}

func (g *gameState) GetInt(key string) (int64, error) {
	if g == nil || g.game == nil {
		return 0, fmt.Errorf("game state not initialized")
	}
	return g.game.Values[key], nil
}

func (g *gameState) GetStr(key string) (string, error) {
	return "", fmt.Errorf("strings are not implemented")
}

func (g *gameState) GetStrArr(key string) ([]string, error) {
	return nil, fmt.Errorf("strings are not implemented")
}

func (g *gameState) GetScope(key string) logic.Lookup {
	return g
}

func (g *gameState) SetScope(key string, scope logic.Lookup) {}
func (g *gameState) ListScopes() []string {
	return []string{}
}

// allowed returns true if the action is in the location's available list.
func allowed(act *storypb.Action, loc *storypb.Location) bool {
	for _, av := range loc.GetAvailableActions() {
		if av == act.GetId() {
			return true
		}
	}
	return false
}

// apply sets the new state of the playthrough according to the effect.
func apply(eff *storypb.Effect, game *storypb.Playthrough) {
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

func HandleAction(act *storypb.Action, loc *storypb.Location, game *storypb.Playthrough, story *storypb.Story) error {
	if loc.GetId() != game.GetLocation() {
		return fmt.Errorf("cannot apply action %d (%s) to location %d (%s) when current location is %d", act.GetId(), act.GetTitle(), loc.GetId(), loc.GetTitle(), game.GetLocation())
	}
	if !allowed(act, loc) {
		return fmt.Errorf("action %d (%s) not allowed in location %d (%s)", act.GetId(), act.GetTitle(), loc.GetId(), loc.GetTitle())
	}

	// TODO: This should be conditional.
	for _, eff := range act.GetEffects() {
		apply(eff, game)
	}

	state := &gameState{game: game}
	for idx, tap := range story.GetEvents() {
		trigger, err := logic.Eval(tap.GetCondition(), state)
		if err != nil {
			// TODO: Escalate this in some manner.
			log.Printf("Could not evaluate predicate for TAP %d in story %d (%q): %v", idx, story.GetId(), story.GetTitle(), err)
			continue
		}
		if !trigger {
			continue
		}
		for _, effect := range tap.GetEffects() {
			apply(effect, game)
		}
	}

	return nil
}
