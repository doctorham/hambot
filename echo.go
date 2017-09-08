package main

import (
	"fmt"
	"strings"
)

type HamEcho struct{}

func (this *HamEcho) HandleMessage(message Message) bool {
	if strings.ToLower(message.DirectText) == "ham" {
		fmt.Println("Echoing ham from @" + message.Session.GetUser(message.User).Name)

		rtm := message.Session.RTM
		rtm.SendMessage(rtm.NewOutgoingMessage(
			message.ReplyPrefix+"ham", message.Channel))
		return true
	}
	return false
}
