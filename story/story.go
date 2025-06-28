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
	if nl := eff.GetNewLocationId(); nl != 0 {
		game.LocationId = proto.Int64(nl)
	}
	if k, v := eff.GetTweakValue(), eff.GetTweakAmount(); len(k) > 0 && v != 0 {
		if len(game.Values) == 0 {
			game.Values = make(map[string]int64)
		}
		game.Values[k] += v
	}
	if ns := eff.GetNewState(); ns != storypb.RunState_RS_UNKNOWN {
		game.State = ns.Enum()
	}
}

func HandleEvent(event *storypb.GameEvent) (*storypb.Playthrough, error) {
	game := proto.Clone(event.GetGameSnapshot()).(*storypb.Playthrough)
	act, loc, str := event.GetAction(), event.GetLocation(), event.GetStory()
	aid, lid, sid := act.GetId(), loc.GetId(), str.GetId()
	if clid := game.GetLocationId(); lid != clid {
		return nil, fmt.Errorf("cannot apply action %d (%s) to location %d (%s) when current location is %d", aid, act.GetTitle(), lid, loc.GetTitle(), clid)
	}
	if !allowed(act, loc) {
		return nil, fmt.Errorf("action %d (%s) not allowed in location %d (%s)", aid, act.GetTitle(), lid, loc.GetTitle())
	}

	state := &gameState{game: game}
	for idx, tap := range act.GetTriggers() {
		trigger, err := logic.Eval(tap.GetCondition(), state)
		if err != nil {
			log.Printf("Could not evaluate predicate for trigger %d in action %d (%q) of story %d (%q): %v", idx, aid, act.GetTitle(), sid, str.GetTitle(), err)
			continue
		}
		if !trigger {
			continue
		}
		for _, effect := range tap.GetEffects() {
			apply(effect, game)
		}
		if tap.GetIsFinal() {
			break
		}
	}

	for idx, tap := range str.GetEvents() {
		trigger, err := logic.Eval(tap.GetCondition(), state)
		if err != nil {
			// TODO: Escalate this in some manner.
			log.Printf("Could not evaluate predicate for TAP %d in story %d (%q): %v", idx, sid, str.GetTitle(), err)
			continue
		}
		if !trigger {
			continue
		}
		for _, effect := range tap.GetEffects() {
			apply(effect, game)
		}
		if tap.GetIsFinal() {
			break
		}
	}

	return game, nil
}
