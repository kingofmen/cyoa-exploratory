package handlers

import (
	"log"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func printProto(comment string, msg proto.Message) {
	log.Printf("%s: %s", comment, prototext.Format(msg))
}
