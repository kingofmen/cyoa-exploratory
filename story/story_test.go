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
	uuid3 := uuid.New().String()
	// There is no 4.
	uuid5 := uuid.New().String()

	cases := []struct {
		desc string
		act  *storypb.Action
		loc  *storypb.Location
		str  *storypb.Story
		game *storypb.Playthrough
		want *storypb.Playthrough
	}{
		{
			desc: "No-op",
			act:  &storypb.Action{Id: proto.String(uuid1)},
			loc:  &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid1)},
			want: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid1)},
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
			loc:  &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid1)},
			want: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid2)},
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
			loc: &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
				Values:     map[string]int64{"strength": 10},
			},
			want: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid2),
				Values:     map[string]int64{"strength": 10},
			},
		},
		{
			desc: "Conditional effect (no)",
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
			loc: &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
			},
			want: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
			},
		},
		{
			desc: "Final effect (yes)",
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
						IsFinal: proto.Bool(true),
					},
					&storypb.TriggerAction{
						Effects: []*storypb.Effect{
							&storypb.Effect{NewLocationId: proto.String(uuid3)},
						},
						IsFinal: proto.Bool(true),
					},
				},
			},
			loc: &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
				Values:     map[string]int64{"strength": 10},
			},
			want: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid2),
				Values:     map[string]int64{"strength": 10},
			},
		},
		{
			desc: "Final effect (no)",
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
						IsFinal: proto.Bool(true),
					},
					&storypb.TriggerAction{
						Effects: []*storypb.Effect{
							&storypb.Effect{NewLocationId: proto.String(uuid3)},
						},
						IsFinal: proto.Bool(true),
					},
				},
			},
			loc: &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
				Values:     map[string]int64{"strength": 1},
			},
			want: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid3),
				Values:     map[string]int64{"strength": 1},
			},
		},
		{
			desc: "Tweak values",
			act: &storypb.Action{
				Id: proto.String(uuid1),
				Triggers: []*storypb.TriggerAction{
					&storypb.TriggerAction{
						Effects: []*storypb.Effect{
							&storypb.Effect{
								TweakValue:  proto.String("a"),
								TweakAmount: proto.Int64(1),
							},
							&storypb.Effect{
								TweakValue:  proto.String("b"),
								TweakAmount: proto.Int64(10),
							},
						},
					},
				},
			},
			loc:  &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid1)},
			want: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
				Values:     map[string]int64{"a": 1, "b": 10},
			},
		},
		{
			desc: "Location and value",
			act: &storypb.Action{
				Id: proto.String(uuid1),
				Triggers: []*storypb.TriggerAction{
					&storypb.TriggerAction{
						Effects: []*storypb.Effect{
							&storypb.Effect{
								NewLocationId: proto.String(uuid2),
								TweakValue:    proto.String("a"),
								TweakAmount:   proto.Int64(1),
							},
						},
					},
				},
			},
			loc: &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
				Values:     map[string]int64{"a": 1, "b": 10},
			},
			want: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid2),
				Values:     map[string]int64{"a": 2, "b": 10},
			},
		},
		{
			desc: "Story trigger",
			act: &storypb.Action{
				Id: proto.String(uuid1),
				Triggers: []*storypb.TriggerAction{
					&storypb.TriggerAction{
						Effects: []*storypb.Effect{
							&storypb.Effect{
								TweakValue:  proto.String("hit_points"),
								TweakAmount: proto.Int64(-1),
							},
						},
					},
				},
			},
			str: &storypb.Story{
				Id: proto.Int64(1),
				Events: []*storypb.TriggerAction{
					&storypb.TriggerAction{
						Condition: &lpb.Predicate{
							Test: &lpb.Predicate_Comp{
								Comp: &lpb.Compare{
									KeyOne:    proto.String("hit_points"),
									KeyTwo:    proto.String("1"),
									Operation: lpb.Compare_CMP_LT.Enum(),
								},
							},
						},
						Effects: []*storypb.Effect{
							&storypb.Effect{
								NewLocationId: proto.String(uuid5),
								TweakValue:    proto.String("deadness"),
								TweakAmount:   proto.Int64(100),
								NewState:      storypb.RunState_RS_COMPLETE.Enum(),
							},
						},
					},
				},
			},
			loc: &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid1),
				Values:     map[string]int64{"hit_points": 1},
			},
			want: &storypb.Playthrough{
				Id:         proto.Int64(1),
				LocationId: proto.String(uuid5),
				Values: map[string]int64{
					"hit_points": 0,
					"deadness":   100,
				},
				State: storypb.RunState_RS_COMPLETE.Enum(),
			},
		},
	}

	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			evt := &storypb.GameEvent{
				Action:       cc.act,
				Location:     cc.loc,
				GameSnapshot: cc.game,
				Story:        cc.str,
			}
			got, err := HandleEvent(evt)
			if err != nil {
				t.Errorf("%s: HandleAction() => %v, want nil", cc.desc, err)
			}
			if diff := cmp.Diff(got, cc.want, protocmp.Transform()); diff != "" {
				t.Errorf("%s: HandleAction() => %s, want %s, diff %s", cc.desc, prototext.Format(got), prototext.Format(cc.want), diff)
			}
		})
	}
}

func TestHandleActionSad(t *testing.T) {
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	uuid10 := uuid.New().String()
	cases := []struct {
		desc string
		act  *storypb.Action
		loc  *storypb.Location
		str  *storypb.Story
		game *storypb.Playthrough
		want string
	}{
		{
			desc: "Bad ID",
			act:  &storypb.Action{Id: proto.String(uuid1)},
			loc:  &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid1)}}},
			game: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid2)},
			want: fmt.Sprintf("when current location is %s", uuid2),
		},
		{
			desc: "Disallowed action",
			act:  &storypb.Action{Id: proto.String(uuid1)},
			loc:  &storypb.Location{Id: proto.String(uuid1), PossibleActions: []*storypb.ActionCondition{&storypb.ActionCondition{ActionId: proto.String(uuid10)}}},
			game: &storypb.Playthrough{Id: proto.Int64(1), LocationId: proto.String(uuid1)},
			want: fmt.Sprintf("%s not in possible-actions", uuid1),
		},
	}

	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			evt := &storypb.GameEvent{
				Action:       cc.act,
				Location:     cc.loc,
				GameSnapshot: cc.game,
				Story:        cc.str,
			}
			_, err := HandleEvent(evt)
			if got := fmt.Sprintf("%v", err); !strings.Contains(got, cc.want) {
				t.Errorf("%s: HandleAction() => %v, want %q", cc.desc, err, cc.want)
			}
		})
	}
}
