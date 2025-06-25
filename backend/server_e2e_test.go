package handlers

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
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
	srv := New(db)
	csresp, err := srv.CreateStory(ctx, &spb.CreateStoryRequest{
		Story: &storypb.Story{
			Title:       proto.String("E2E test story"),
			Description: proto.String("Story for end-to-end testing"),
		},
	})
	if err != nil {
		t.Fatalf("CreateStory() => %v, want nil", err)
	}
	stid := csresp.GetStory().GetId()
	if stid != 1 {
		t.Fatalf("CreateStory() returned story ID %d, want 1", stid)
	}

	clresp1, err := srv.CreateLocation(ctx, &spb.CreateLocationRequest{
		Location: &storypb.Location{
			Title:   proto.String("Choose Character"),
			Content: proto.String("Choose which character to play as."),
		},
	})
	if err != nil {
		t.Fatalf("CreateLocation(1) => %v, want nil", err)
	}
	exploc := &storypb.Location{
		Id:      proto.Int64(1),
		Title:   proto.String("Choose Character"),
		Content: proto.String("Choose which character to play as."),
	}
	if diff := cmp.Diff(clresp1.GetLocation(), exploc, protocmp.Transform()); diff != "" {
		t.Errorf("After CreateLocation(1): %s, want %s, diff %s", prototext.Format(clresp1.GetLocation()), prototext.Format(exploc), diff)
	}

	loc1id := clresp1.GetLocation().GetId()
	clresp2, err := srv.CreateLocation(ctx, &spb.CreateLocationRequest{
		Location: &storypb.Location{
			Title:   proto.String("Ogre Encounter"),
			Content: proto.String("Either fight the ogre or attempt to sneak past it."),
		},
	})
	if err != nil {
		t.Fatalf("CreateLocation(2) => %v, want nil", err)
	}
	loc2id := clresp2.GetLocation().GetId()

	exploc = &storypb.Location{
		Id:      proto.Int64(2),
		Title:   proto.String("Ogre Encounter"),
		Content: proto.String("Either fight the ogre or attempt to sneak past it."),
	}
	if diff := cmp.Diff(clresp2.GetLocation(), exploc, protocmp.Transform()); diff != "" {
		t.Errorf("After CreateLocation(2): %s, want %s, diff %s", prototext.Format(clresp2.GetLocation()), prototext.Format(exploc), diff)
	}
	if llresp, err := srv.ListLocations(ctx, &spb.ListLocationsRequest{}); err != nil || len(llresp.GetLocations()) != 2 {
		t.Fatalf("Created 2 locations but List finds %d: %s error: %v", len(llresp.GetLocations()), prototext.Format(llresp), err)
	}

	usresp, err := srv.UpdateStory(ctx, &spb.UpdateStoryRequest{
		Story: &storypb.Story{
			Id:              proto.Int64(stid),
			StartLocationId: proto.Int64(loc1id),
		},
	})
	if err != nil {
		t.Fatalf("UpdateStory() => %v, want nil", err)
	}

	got := usresp.GetStory()
	want := &storypb.Story{
		Id:              proto.Int64(1),
		Title:           proto.String("E2E test story"),
		Description:     proto.String("Story for end-to-end testing"),
		StartLocationId: proto.Int64(1),
	}
	if diff := cmp.Diff(got, want, protocmp.Transform()); diff != "" {
		t.Errorf("After UpdateStory: %s, want %s, diff %s (%s)", prototext.Format(got), prototext.Format(want), diff, prototext.Format(clresp1))
	}

	charFighter := &storypb.Action{
		Title:       proto.String("Fighter"),
		Description: proto.String("A mighty warrior!"),
		Triggers: []*storypb.TriggerAction{
			&storypb.TriggerAction{
				Effects: []*storypb.Effect{
					&storypb.Effect{
						NewLocation: proto.Int64(loc2id),
						TweakValue:  proto.String("Strength"),
						TweakAmount: proto.Int64(5),
					},
				},
			},
		},
	}
	charThief := &storypb.Action{
		Title:       proto.String("Rogue"),
		Description: proto.String("A cunning thief!"),
		Triggers: []*storypb.TriggerAction{
			&storypb.TriggerAction{
				Effects: []*storypb.Effect{
					&storypb.Effect{
						NewLocation: proto.Int64(loc2id),
						TweakValue:  proto.String("Dexterity"),
						TweakAmount: proto.Int64(5),
					},
				},
			},
		},
	}
	fightOgre := &storypb.Action{
		Title:       proto.String("Attack!"),
		Description: proto.String("Fight the ogre with your sword."),
	}
	sneakOgre := &storypb.Action{
		Title:       proto.String("Slow and sneaky wins the race..."),
		Description: proto.String("Sneak past the ogre."),
	}

	actions := []*storypb.Action{charFighter, charThief, fightOgre, sneakOgre}
	for idx, act := range actions {
		resp, err := srv.CreateAction(ctx, &spb.CreateActionRequest{
			Action: act,
		})
		if err != nil {
			t.Fatalf("Could not create action %d: %v", idx, err)
		}
		actions[idx] = resp.GetAction()
	}

	cases := []struct {
		desc    string
		actions []int64
	}{
		{
			desc:    "Fighter, attack",
			actions: []int64{charFighter.GetId(), fightOgre.GetId()},
		},
		{
			desc:    "Rogue, attack",
			actions: []int64{charThief.GetId(), fightOgre.GetId()},
		},
		{
			desc:    "Fighter, sneak",
			actions: []int64{charFighter.GetId(), sneakOgre.GetId()},
		},
		{
			desc:    "Rogue, sneak",
			actions: []int64{charThief.GetId(), sneakOgre.GetId()},
		},
	}

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
				_, err := srv.PlayerAction(ctx, &spb.PlayerActionRequest{
					GameId:   proto.Int64(gid),
					ActionId: proto.Int64(actid),
				})
				if err != nil {
					t.Errorf("%s: Action %d had unexpected error %v", cc.desc, idx, err)
					continue
				}
			}
		})
	}
}
