// Package story implements validates actions and changes story state.
package story

import (
	"fmt"
	"log"

	"github.com/kingofmen/cyoa-exploratory/logic"
	"google.golang.org/protobuf/proto"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

// gameState implements logic.Lookup with a GameEvent as its scope.
type gameState struct {
	game *storypb.GameEvent
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

// allowed returns an error if the action is not available in the location.
func allowed(act *storypb.Action, loc *storypb.Location, state logic.Lookup) error {
	for _, cand := range loc.GetPossibleActions() {
		if cand.GetActionId() != act.GetId() {
			continue
		}
		ok, err := logic.Eval(cand.GetCondition(), state)
		if err != nil {
			return fmt.Errorf("could not evaluate condition: %w", err)
		}
		if !ok {
			return fmt.Errorf("condition fails")
		}
		return nil
	}
	return fmt.Errorf("action ID %s not in possible-actions list", act.GetId())
}

// apply sets the new state of the playthrough according to the effect.
func apply(eff *storypb.Effect, game *storypb.GameEvent) {
	if nl := eff.GetNewLocationId(); len(nl) > 0 {
		game.Location = &storypb.Location{
			Id: proto.String(nl),
		}
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

func HandleEvent(event *storypb.GameEvent) (*storypb.GameEvent, error) {
	game := proto.Clone(event).(*storypb.GameEvent)
	act, loc, str := event.GetPlayerAction(), event.GetLocation(), event.GetStory()
	aid, lid, sid := act.GetId(), loc.GetId(), str.GetId()
	if clid := game.GetLocation().GetId(); lid != clid {
		return nil, fmt.Errorf("cannot apply action %s (%s) to location %s (%s) when current location is %s", aid, act.GetTitle(), lid, loc.GetTitle(), clid)
	}
	state := &gameState{game: game}
	if err := allowed(act, loc, state); err != nil {
		return nil, fmt.Errorf("action %s (%s) not available in location %s (%s): %w", aid, act.GetTitle(), lid, loc.GetTitle(), err)
	}

	for idx, tap := range act.GetTriggers() {
		trigger, err := logic.Eval(tap.GetCondition(), state)
		if err != nil {
			log.Printf("Could not evaluate predicate for trigger %d in action %s (%q) of story %d (%q): %v", idx, aid, act.GetTitle(), sid, str.GetTitle(), err)
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

// PossibleActions returns the actions of the current story location
// that are possible given the rest of the game state.
func PossibleActions(event *storypb.GameEvent) []string {
	state := &gameState{game: event}
	cands := event.GetLocation().GetPossibleActions()
	ret := make([]string, 0, len(cands))
	for idx, cand := range cands {
		good, err := logic.Eval(cand.GetCondition(), state)
		if err != nil {
			log.Printf("Could not evaluate predicate for possible action %d (%s) in story %q: %v", idx, cand.GetActionId(), event.GetStory().GetTitle(), err)
			continue
		}
		if !good {
			continue
		}
		ret = append(ret, cand.GetActionId())
	}
	return ret
}
