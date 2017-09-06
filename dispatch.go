package main

import (
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

type Message struct {
	Session *Session
	Message *slack.MessageEvent
	Text    string
}

var gAtHambot *regexp.Regexp

func dispatchMessage(
	session *Session,
	slackMessage *slack.MessageEvent,
) {
	if gAtHambot == nil {
		var err error
		gAtHambot, err = regexp.Compile("^\\s*<@" + session.Info.User.ID + ">(\\s*:)?\\s+(.*?)\\s*$")
		if err != nil {
			panic(err)
		}
	}

	message := Message{Session: session, Message: slackMessage}
	matches := gAtHambot.FindStringSubmatch(slackMessage.Text)
	if matches == nil {
		// accept messages without @hambot tag if sent directly to hambot
		if session.GetIM(slackMessage.Channel) != nil {
			message.Text = slackMessage.Text
		} else {
			return
		}
	} else {
		message.Text = matches[2]
	}

	// send to other modules for processing
	if strings.ToLower(message.Text) == "ham" {
		reply := "ham"
		if session.GetIM(slackMessage.Channel) == nil {
			reply = "<@" + message.Message.User + "> " + reply
		}
		session.RTM.SendMessage(session.RTM.NewOutgoingMessage(reply, message.Message.Channel))
	}
}
