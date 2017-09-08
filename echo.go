package main

import (
	"fmt"
	"strings"
)

type HamEcho struct{}

func (*HamEcho) HandleMessage(message Message) bool {
	if strings.ToLower(message.DirectText) != "ham" {
		return false
	}

	fmt.Println("Echoing ham from @" + message.Session.GetUser(message.User).Name)

	message.Reply("ham")
	return true
}
