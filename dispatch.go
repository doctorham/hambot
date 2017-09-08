package main

import (
	"regexp"

	"github.com/nlopes/slack"
)

type Message struct {
	Session     *Session
	Message     *slack.MessageEvent
	DirectText  string // non-empty if direct message or begins with @hambot
	ReplyPrefix string // non-empty if begins with @hambot
}

type MessageHandler interface {
	HandleMessage(Message) bool // returns true if handled
}

type Dispatcher struct {
	session  *Session
	atHambot *regexp.Regexp
	handlers []MessageHandler
}

func NewDispatcher(session *Session) *Dispatcher {
	dispatcher := Dispatcher{session: session}

	var err error
	dispatcher.atHambot, err = regexp.Compile(
		"^\\s*<@" + session.Info.User.ID + ">(\\s*:)?\\s+(.*?)\\s*$")
	if err != nil {
		panic(err)
	}

	return &dispatcher
}

func (this *Dispatcher) AddHandler(handler MessageHandler) {
	this.handlers = append(this.handlers, handler)
}

func (this *Dispatcher) Dispatch(slackMessage *slack.MessageEvent) {
	// ignore messages sent by another hambot
	if slackMessage.User == this.session.Info.User.ID {
		return
	}

	message := Message{Session: this.session, Message: slackMessage}
	matches := this.atHambot.FindStringSubmatch(slackMessage.Text)
	if matches == nil {
		// accept messages without @hambot tag if sent directly to hambot
		if this.session.GetIM(slackMessage.Channel) != nil {
			message.DirectText = slackMessage.Text
		}
	} else {
		message.DirectText = matches[2]
		if this.session.GetIM(slackMessage.Channel) == nil {
			message.ReplyPrefix = "<@" + message.Message.User + "> "
		}
	}

	for _, handler := range this.handlers {
		if handler.HandleMessage(message) {
			break
		}
	}
}
