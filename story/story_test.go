package story

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	lpb "github.com/kingofmen/cyoa-exploratory/logic/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

func TestHandleAction(t *testing.T) {
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	//uuid3 := uuid.New().String()
	// There is no 4.
	//uuid5 := uuid.New().String()

	loc1 := &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}}
	loc2 := &storypb.Location{Id: proto.String(uuid2)}
	cases := []struct {
		desc string
		act  *storypb.Action
		loc  *storypb.Location
		str  *storypb.Story
		game *storypb.Playthrough
		want *storypb.GameEvent
	}{
		{
			desc: "No-op",
			act:  &storypb.Action{Id: proto.String(uuid1)},
			loc:  loc1,
			game: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid1)},
			want: &storypb.GameEvent{
				Location: loc1,
				State:    storypb.RunState_RS_UNKNOWN.Enum(),
			},
		},
		{
			desc: "New location",
			act: &storypb.Action{
				Id: proto.String(uuid1),
				Triggers: []*storypb.TriggerAction{
					&storypb.TriggerAction{
						Effects: []*storypb.Effect{
							&storypb.Effect{NewLocationId: proto.String(uuid2)},
						},
					},
				},
			},
			loc:  loc1,
			game: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid1)},
			want: &storypb.GameEvent{
				Location: loc2,
				State:    storypb.RunState_RS_UNKNOWN.Enum(),
			},
		},
		{
			desc: "Conditional effect (yes)",
			act: &storypb.Action{
				Id: proto.String(uuid1),
				Triggers: []*storypb.TriggerAction{
					&storypb.TriggerAction{
						Condition: &lpb.Predicate{
							Test: &lpb.Predicate_Comp{
								Comp: &lpb.Compare{
									KeyOne:    proto.String("strength"),
									KeyTwo:    proto.String("1"),
									Operation: lpb.Compare_CMP_GT.Enum(),
								},
							},
						},
						Effects: []*storypb.Effect{
							&storypb.Effect{NewLocationId: proto.String(uuid2)},
						},
					},
				},
			},
			loc: loc1,
			game: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
				Values:     map[string]int64{"strength": 10},
			},
			want: &storypb.GameEvent{
				Location: loc2,
				Values:   map[string]int64{"strength": 10},
				State:    storypb.RunState_RS_UNKNOWN.Enum(),
			},
		},
	}
	ignore := protocmp.IgnoreFields(&storypb.GameEvent{}, "player_action")
	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			evt := &storypb.GameEvent{
				PlayerAction: cc.act,
				Location:     cc.loc,
				Values:       cc.game.GetValues(),
				State:        cc.game.GetState().Enum(),
				Story:        cc.str,
			}
			got, err := HandleEvent(evt)
			if err != nil {
				t.Errorf("%s: HandleAction() => %v, want nil", cc.desc, err)
			}
			if diff := cmp.Diff(got, cc.want, protocmp.Transform(), ignore); diff != "" {
				t.Errorf("%s: HandleAction() => %s, want %s, diff %s", cc.desc, prototext.Format(got), prototext.Format(cc.want), diff)
			}
		})
	}
}

func TestHandleActionSad(t *testing.T) {
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	cases := []struct {
		desc string
		act  *storypb.Action
		loc  *storypb.Location
		str  *storypb.Story
		game *storypb.Playthrough
		want string
	}{
		{
			desc: "Disallowed action",
			act:  &storypb.Action{Id: proto.String(uuid1)},
			loc:  &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid2)}}},
			want: fmt.Sprintf("%s not in possible-actions", uuid1),
		},
	}

	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			evt := &storypb.GameEvent{
				PlayerAction: cc.act,
				Location:     cc.loc,
				Values:       cc.game.GetValues(),
				State:        cc.game.GetState().Enum(),
				Story:        cc.str,
			}
			_, err := HandleEvent(evt)
			if got := fmt.Sprintf("%v", err); !strings.Contains(got, cc.want) {
				t.Errorf("%s: HandleAction() => %v, want %q", cc.desc, err, cc.want)
			}
		})
	}
}
