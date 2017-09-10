package main

import (
	"fmt"
	"strings"
)

type HamEcho struct{}

func NewHamEcho() *HamEcho {
	return &HamEcho{}
}

func (*HamEcho) HandleMessage(message Message) bool {
	if strings.ToLower(message.DirectText) != "ham" {
		return false
	}

	fmt.Println("Echoing ham from @" + message.Session.GetUser(message.User).Name)

	message.Reply("ham :ham:")
	return true
}
