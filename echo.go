package main

import (
	"fmt"
	"strings"
)

// HamEcho echoes "ham."
type HamEcho struct{}

// NewHamEcho creates a new HamEcho.
func NewHamEcho() *HamEcho {
	return &HamEcho{}
}

// HandleMessage handles a Message.
func (*HamEcho) HandleMessage(message Message) bool {
	if strings.ToLower(message.DirectText) != "ham" {
		return false
	}

	fmt.Println("Echoing ham from @" + message.Session.User(message.User).Name)

	message.Reply("ham :ham:")
	return true
}
