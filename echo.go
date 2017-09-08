package main

import (
	"fmt"
	"strings"
)

type HamEcho struct{}

func (this *HamEcho) HandleMessage(message Message) bool {
	if strings.ToLower(message.DirectText) == "ham" {
		fmt.Println("Echoing ham from @" + message.Session.GetUser(message.Message.User).Name)

		rtm := message.Session.RTM
		rtm.SendMessage(rtm.NewOutgoingMessage(
			message.ReplyPrefix+"ham", message.Message.Channel))
		return true
	}
	return false
}
