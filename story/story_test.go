package story

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

func TestHandleAction(t *testing.T) {
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
			act:  &storypb.Action{Id: proto.Int64(1)},
			loc:  &storypb.Location{Id: proto.Int64(1), AvailableActions: []int64{1}},
			game: &storypb.Playthrough{Id: proto.Int64(1), Location: proto.Int64(1)},
			want: &storypb.Playthrough{Id: proto.Int64(1), Location: proto.Int64(1)},
		},
		{
			desc: "New location",
			act: &storypb.Action{
				Id: proto.Int64(1),
				Effects: []*storypb.Effect{
					&storypb.Effect{NewLocation: proto.Int64(2)},
				},
			},
			loc:  &storypb.Location{Id: proto.Int64(1), AvailableActions: []int64{1}},
			game: &storypb.Playthrough{Id: proto.Int64(1), Location: proto.Int64(1)},
			want: &storypb.Playthrough{Id: proto.Int64(1), Location: proto.Int64(2)},
		},
		{
			desc: "Tweak values",
			act: &storypb.Action{
				Id: proto.Int64(1),
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
			loc:  &storypb.Location{Id: proto.Int64(1), AvailableActions: []int64{1}},
			game: &storypb.Playthrough{Id: proto.Int64(1), Location: proto.Int64(1)},
			want: &storypb.Playthrough{
				Id:       proto.Int64(1),
				Location: proto.Int64(1),
				Values:   map[string]int64{"a": 1, "b": 10},
			},
		},
		{
			desc: "Location and value",
			act: &storypb.Action{
				Id: proto.Int64(1),
				Effects: []*storypb.Effect{
					&storypb.Effect{
						NewLocation: proto.Int64(2),
						TweakValue:  proto.String("a"),
						TweakAmount: proto.Int64(1),
					},
				},
			},
			loc: &storypb.Location{Id: proto.Int64(1), AvailableActions: []int64{1}},
			game: &storypb.Playthrough{
				Id:       proto.Int64(1),
				Location: proto.Int64(1),
				Values:   map[string]int64{"a": 1, "b": 10},
			},
			want: &storypb.Playthrough{
				Id:       proto.Int64(1),
				Location: proto.Int64(2),
				Values:   map[string]int64{"a": 2, "b": 10},
			},
		},
	}

	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			if err := HandleAction(cc.act, cc.loc, cc.game, cc.str); err != nil {
				t.Errorf("%s: HandleAction() => %v, want nil", cc.desc, err)
			}
			if diff := cmp.Diff(cc.game, cc.want, protocmp.Transform()); diff != "" {
				t.Errorf("%s: HandleAction() => %s, want %s, diff %s", cc.desc, prototext.Format(cc.game), prototext.Format(cc.want), diff)
			}
		})
	}
}

func TestHandleActionSad(t *testing.T) {
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
			act:  &storypb.Action{Id: proto.Int64(1)},
			loc:  &storypb.Location{Id: proto.Int64(1), AvailableActions: []int64{1}},
			game: &storypb.Playthrough{Id: proto.Int64(1), Location: proto.Int64(2)},
			want: "when current location is 2",
		},
		{
			desc: "Disallowed action",
			act:  &storypb.Action{Id: proto.Int64(1)},
			loc:  &storypb.Location{Id: proto.Int64(1), AvailableActions: []int64{10}},
			game: &storypb.Playthrough{Id: proto.Int64(1), Location: proto.Int64(1)},
			want: "not allowed in location 1",
		},
	}

	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			err := HandleAction(cc.act, cc.loc, cc.game, cc.str)
			if got := fmt.Sprintf("%v", err); !strings.Contains(got, cc.want) {
				t.Errorf("%s: HandleAction() => %v, want %q", cc.desc, err, cc.want)
			}
		})
	}
}
