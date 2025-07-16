package handlers

import (
	"log"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

func printProto(comment string, msg proto.Message) {
	log.Printf("%s: %s", comment, prototext.Format(msg))
}

type summarizable interface {
	GetTitle() string
	GetDescription() string
}

// summarize creates a storypb.Summary object.
func summarize(s summarizable) *storypb.Summary {
	return &storypb.Summary{
		Title:       proto.String(s.GetTitle()),
		Description: proto.String(s.GetDescription()),
	}
}
