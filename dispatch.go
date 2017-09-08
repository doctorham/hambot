package main

import (
	"regexp"

	"github.com/nlopes/slack"
)

type Message struct {
	*slack.MessageEvent

	Session     *Session
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

func NewDispatcher(session *Session) (*Dispatcher, error) {
	dispatcher := Dispatcher{session: session}

	var err error
	dispatcher.atHambot, err = regexp.Compile(
		"^\\s*<@" + session.Info.User.ID + ">(\\s*:)?\\s+(.*?)\\s*$")
	if err != nil {
		return nil, err
	}

	return &dispatcher, nil
}

func (this *Dispatcher) AddHandler(handler MessageHandler) {
	this.handlers = append(this.handlers, handler)
}

func (this *Dispatcher) Dispatch(slackMessage *slack.MessageEvent) {
	var message Message
	message.MessageEvent = slackMessage
	message.Session = this.session

	// ignore messages sent by another hambot
	if message.User == this.session.Info.User.ID {
		return
	}

	matches := this.atHambot.FindStringSubmatch(message.Text)
	if matches == nil {
		// accept messages without @hambot tag if sent directly to hambot
		if this.session.GetIM(message.Channel) != nil {
			message.DirectText = message.Text
		}
	} else {
		message.DirectText = matches[2]
		if this.session.GetIM(message.Channel) == nil {
			message.ReplyPrefix = "<@" + message.User + "> "
		}
	}

	for _, handler := range this.handlers {
		if handler.HandleMessage(message) {
			break
		}
	}
}
