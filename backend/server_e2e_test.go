package handlers

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/kingofmen/cyoa-exploratory/narrate"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
	lpb "github.com/kingofmen/cyoa-exploratory/logic/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

var db *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0"),
		mysql.WithDatabase("test_db"),
		mysql.WithUsername("test_user"),
		mysql.WithPassword("test_password"),
	)
	if err != nil {
		log.Fatalf("could not start mysql container: %v", err)
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("could not stop mysql container: %v", err)
		}
	}()

	connStr, err := container.ConnectionString(ctx, "multiStatements=true")
	if err != nil {
		log.Fatalf("could not get connection string: %v", err)
	}

	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("could not open database connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping database: %v", err)
	}

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}
	if err := goose.Up(db, "../db/migrations"); err != nil {
		log.Fatalf("failed to run goose migrations: %v", err)
	}
	log.Println("Goose migrations applied successfully.")

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestStoryE2E(t *testing.T) {
	ctx := context.Background()
	srv := New(db).WithNarrator(narrate.NewDebug())
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	uuid3 := uuid.New().String()
	uuid4 := uuid.New().String()
	charFighter := &storypb.Action{
		Id:          proto.String(uuid1),
		Title:       proto.String("Fighter"),
		Description: proto.String("A mighty warrior!"),
		Triggers: []*storypb.TriggerAction{
			&storypb.TriggerAction{
				Effects: []*storypb.Effect{
					&storypb.Effect{
						NewLocationId: proto.String(uuid2),
						TweakValue:    proto.String("Strength"),
						TweakAmount:   proto.Int64(5),
					},
				},
			},
		},
	}
	charThief := &storypb.Action{
		Id:          proto.String(uuid2),
		Title:       proto.String("Rogue"),
		Description: proto.String("A cunning thief!"),
		Triggers: []*storypb.TriggerAction{
			&storypb.TriggerAction{
				Effects: []*storypb.Effect{
					&storypb.Effect{
						NewLocationId: proto.String(uuid2),
						TweakValue:    proto.String("Dexterity"),
						TweakAmount:   proto.Int64(5),
					},
				},
			},
		},
	}
	fightOgre := &storypb.Action{
		Id:          proto.String(uuid3),
		Title:       proto.String("Attack!"),
		Description: proto.String("Fight the ogre with your sword."),
		Triggers: []*storypb.TriggerAction{
			&storypb.TriggerAction{
				Condition: &lpb.Predicate{
					Test: &lpb.Predicate_Comp{
						Comp: &lpb.Compare{
							KeyOne:    proto.String("Strength"),
							KeyTwo:    proto.String("3"),
							Operation: lpb.Compare_CMP_GT.Enum(),
						},
					},
				},
				Effects: []*storypb.Effect{
					&storypb.Effect{
						TweakValue:  proto.String("ogre_defeated"),
						TweakAmount: proto.Int64(1),
					},
				},
				IsFinal: proto.Bool(true),
			},
			&storypb.TriggerAction{
				Effects: []*storypb.Effect{
					&storypb.Effect{
						TweakValue:  proto.String("player_killed"),
						TweakAmount: proto.Int64(1),
					},
				},
			},
		},
	}
	sneakOgre := &storypb.Action{
		Id:          proto.String(uuid4),
		Title:       proto.String("Slow and sneaky wins the race..."),
		Description: proto.String("Sneak past the ogre."),
		Triggers: []*storypb.TriggerAction{
			&storypb.TriggerAction{
				Condition: &lpb.Predicate{
					Test: &lpb.Predicate_Comp{
						Comp: &lpb.Compare{
							KeyOne:    proto.String("Dexterity"),
							KeyTwo:    proto.String("3"),
							Operation: lpb.Compare_CMP_GT.Enum(),
						},
					},
				},
				Effects: []*storypb.Effect{
					&storypb.Effect{
						TweakValue:  proto.String("ogre_defeated"),
						TweakAmount: proto.Int64(1),
					},
				},
				IsFinal: proto.Bool(true),
			},
			&storypb.TriggerAction{
				Effects: []*storypb.Effect{
					&storypb.Effect{
						TweakValue:  proto.String("player_killed"),
						TweakAmount: proto.Int64(1),
					},
				},
			},
		},
	}

	csresp, err := srv.UpdateStory(ctx, &spb.UpdateStoryRequest{
		Story: &storypb.Story{
			Title:           proto.String("E2E test story"),
			Description:     proto.String("Story for end-to-end testing"),
			StartLocationId: proto.String(uuid1),
			Events: []*storypb.TriggerAction{
				&storypb.TriggerAction{
					Condition: &lpb.Predicate{
						Test: &lpb.Predicate_Comp{
							Comp: &lpb.Compare{
								KeyOne:    proto.String("ogre_defeated"),
								KeyTwo:    proto.String("0"),
								Operation: lpb.Compare_CMP_GT.Enum(),
							},
						},
					},
					Effects: []*storypb.Effect{
						&storypb.Effect{
							NewState: storypb.RunState_RS_COMPLETE.Enum(),
						},
					},
				},
				&storypb.TriggerAction{
					Condition: &lpb.Predicate{
						Test: &lpb.Predicate_Comp{
							Comp: &lpb.Compare{
								KeyOne:    proto.String("player_killed"),
								KeyTwo:    proto.String("0"),
								Operation: lpb.Compare_CMP_GT.Enum(),
							},
						},
					},
					Effects: []*storypb.Effect{
						&storypb.Effect{
							NewState: storypb.RunState_RS_COMPLETE.Enum(),
						},
					},
				},
			},
		},
		Content: &spb.StoryContent{
			Locations: []*storypb.Location{
				&storypb.Location{
					Id:      proto.String(uuid1),
					Title:   proto.String("Choose Character"),
					Content: proto.String("Choose which character to play as."),
					PossibleActions: []*storypb.ActionCondition{
						&storypb.ActionCondition{ActionId: proto.String(uuid1)},
						&storypb.ActionCondition{ActionId: proto.String(uuid2)},
					},
				},
				&storypb.Location{
					Id:      proto.String(uuid2),
					Title:   proto.String("Ogre Encounter"),
					Content: proto.String("Either fight the ogre or attempt to sneak past it."),
					PossibleActions: []*storypb.ActionCondition{
						&storypb.ActionCondition{ActionId: proto.String(uuid3)},
						&storypb.ActionCondition{ActionId: proto.String(uuid4)},
					},
				},
			},
			Actions: []*storypb.Action{
				charFighter, charThief, fightOgre, sneakOgre,
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdateStory() => %v, want nil", err)
	}
	stid := csresp.GetStory().GetId()
	if stid != 1 {
		t.Fatalf("UpdateStory() returned story ID %d, want 1", stid)
	}

	if llresp, err := srv.ListLocations(ctx, &spb.ListLocationsRequest{}); err != nil || len(llresp.GetLocations()) != 2 {
		t.Fatalf("Created 2 locations but List finds %d: %s error: %v", len(llresp.GetLocations()), prototext.Format(llresp), err)
	}

	cases := []struct {
		desc      string
		actions   []string
		expect    []*storypb.Playthrough
		narrative []string
	}{
		{
			desc:    "Fighter, attack",
			actions: []string{charFighter.GetId(), fightOgre.GetId()},
			expect: []*storypb.Playthrough{
				&storypb.Playthrough{
					LocationId: proto.String(uuid2),
					Values:     map[string]int64{"Strength": 5},
					State:      storypb.RunState_RS_ACTIVE.Enum(),
				},
				&storypb.Playthrough{
					LocationId: proto.String(uuid2),
					Values:     map[string]int64{"Strength": 5, "ogre_defeated": 1},
					State:      storypb.RunState_RS_COMPLETE.Enum(),
				},
			},
			narrative: []string{
				"Fighter",
				"Fighter\nAttack!",
			},
		},
		{
			desc:    "Rogue, attack",
			actions: []string{charThief.GetId(), fightOgre.GetId()},
			expect: []*storypb.Playthrough{
				&storypb.Playthrough{
					LocationId: proto.String(uuid2),
					Values:     map[string]int64{"Dexterity": 5},
					State:      storypb.RunState_RS_ACTIVE.Enum(),
				},
				&storypb.Playthrough{
					LocationId: proto.String(uuid2),
					Values:     map[string]int64{"Dexterity": 5, "player_killed": 1},
					State:      storypb.RunState_RS_COMPLETE.Enum(),
				},
			},
			narrative: []string{
				"Rogue",
				"Rogue\nAttack!",
			},
		},
		{
			desc:    "Fighter, sneak",
			actions: []string{charFighter.GetId(), sneakOgre.GetId()},
			expect: []*storypb.Playthrough{
				&storypb.Playthrough{
					LocationId: proto.String(uuid2),
					Values:     map[string]int64{"Strength": 5},
					State:      storypb.RunState_RS_ACTIVE.Enum(),
				},
				&storypb.Playthrough{
					LocationId: proto.String(uuid2),
					Values:     map[string]int64{"Strength": 5, "player_killed": 1},
					State:      storypb.RunState_RS_COMPLETE.Enum(),
				},
			},
			narrative: []string{
				"Fighter",
				"Fighter\nSlow and sneaky wins the race...",
			},
		},
		{
			desc:    "Rogue, sneak",
			actions: []string{charThief.GetId(), sneakOgre.GetId()},
			expect: []*storypb.Playthrough{
				&storypb.Playthrough{
					LocationId: proto.String(uuid2),
					Values:     map[string]int64{"Dexterity": 5},
					State:      storypb.RunState_RS_ACTIVE.Enum(),
				},
				&storypb.Playthrough{
					LocationId: proto.String(uuid2),
					Values:     map[string]int64{"Dexterity": 5, "ogre_defeated": 1},
					State:      storypb.RunState_RS_COMPLETE.Enum(),
				},
			},
			narrative: []string{
				"Rogue",
				"Rogue\nSlow and sneaky wins the race...",
			},
		},
	}

	ignore := protocmp.IgnoreFields(&storypb.Playthrough{}, "id", "story_id")
	for cid, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			gresp, err := srv.CreateGame(ctx, &spb.CreateGameRequest{
				StoryId: proto.Int64(stid),
			})
			if err != nil {
				t.Fatalf("%s: Could not create playthrough: %v", cc.desc, err)
			}
			gid := gresp.GetGameId()
			if expid := int64(cid + 1); gid != expid {
				t.Errorf("%s: CreateGame() => unexpected game ID %d, want %d", cc.desc, gid, expid)
			}
			for idx, actid := range cc.actions {
				resp, err := srv.PlayerAction(ctx, &spb.PlayerActionRequest{
					GameId:   proto.Int64(gid),
					ActionId: proto.String(actid),
				})
				if err != nil {
					t.Errorf("%s: Action %d had unexpected error %v", cc.desc, idx, err)
					continue
				}
				got, want := resp.GetGameState(), cc.expect[idx]
				if diff := cmp.Diff(got, want, protocmp.Transform(), ignore); diff != "" {
					t.Errorf("%s: PlayerAction(%d) => %s, want %s, diff %s", cc.desc, idx, prototext.Format(got), prototext.Format(want), diff)
				}
				gn, wn := resp.GetNarrative(), cc.narrative[idx]
				if diff := cmp.Diff(gn, wn); diff != "" {
					t.Errorf("%s: PlayerAction(%d) => narrative %q, want %q, diff %s", cc.desc, idx, gn, wn, diff)
				}
			}
		})
	}
}
