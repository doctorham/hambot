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

func (m *Message) Reply(text string) {
	rtm := m.Session.RTM
	rtm.SendMessage(rtm.NewOutgoingMessage(m.ReplyPrefix+text, m.Channel))
}

func (m *Message) IsDirect() bool {
	return m.Session.GetIM(m.Channel) != nil
}

type MessageHandler interface {
	HandleMessage(Message) bool // returns true if handled
}

type Dispatcher struct {
	session    *Session
	reAtHambot *regexp.Regexp
	handlers   []MessageHandler
}

func NewDispatcher(session *Session) (*Dispatcher, error) {
	dispatcher := Dispatcher{session: session}

	var err error
	dispatcher.reAtHambot, err = regexp.Compile(
		"^\\s*<@" + session.Info.User.ID + ">(\\s*:)?\\s+(.*?)\\s*$")
	if err != nil {
		return nil, err
	}

	return &dispatcher, nil
}

func (d *Dispatcher) AddHandler(handler MessageHandler) {
	d.handlers = append(d.handlers, handler)
}

func (d *Dispatcher) Dispatch(slackMessage *slack.MessageEvent) {
	var message Message
	message.MessageEvent = slackMessage
	message.Session = d.session

	// ignore messages sent by another hambot
	if message.User == d.session.Info.User.ID {
		return
	}

	matches := d.reAtHambot.FindStringSubmatch(message.Text)
	if matches == nil {
		// accept messages without @hambot tag if sent directly to hambot
		if message.IsDirect() {
			message.DirectText = message.Text
		}
	} else {
		message.DirectText = matches[2]
		if !message.IsDirect() {
			message.ReplyPrefix = "<@" + message.User + "> "
		}
	}

	for _, handler := range d.handlers {
		if handler.HandleMessage(message) {
			break
		}
	}
}
