package main

import (
	"log"
	"strings"
)

// HamEcho echoes "ham."
type HamEcho struct{}

// NewHamEcho creates a new HamEcho.
func NewHamEcho() *HamEcho {
	return &HamEcho{}
}

// HandleMessage handles a Message.
func (*HamEcho) HandleMessage(msg Message) bool {
	if strings.ToLower(msg.DirectText) != "ham" {
		return false
	}

	log.Println("Echoing ham from @" + msg.Session.User(msg.User).Name)

	msg.Reply("ham :ham:")
	return true
}
