package main

import (
	"regexp"

	"github.com/nlopes/slack"
)

// Message represents a Slack message.
type Message struct {
	*slack.MessageEvent

	Session     *Session
	DirectText  string // non-empty if direct message or begins with @hambot
	ReplyPrefix string // non-empty if begins with @hambot
}

// Reply sends a reply to a Message on its channel, prefixing it with the
// recepient's name if it is not a direct message.
func (m *Message) Reply(text string) {
	rtm := m.Session.RTM
	rtm.SendMessage(rtm.NewOutgoingMessage(m.ReplyPrefix+text, m.Channel))
}

// IsDirect returns whether the Message is a direct message (not on a channel).
func (m *Message) IsDirect() bool {
	return m.Session.IM(m.Channel) != nil
}

// MessageHandler handles messages.
type MessageHandler interface {
	HandleMessage(Message) bool // returns true if handled
}

// Dispatcher sends Messages to registered handlers.
type Dispatcher struct {
	session    *Session
	reAtHambot *regexp.Regexp
	handlers   []MessageHandler
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(session *Session) (*Dispatcher, error) {
	dispatcher := Dispatcher{session: session}

	var err error
	dispatcher.reAtHambot, err = regexp.Compile(
		`^\s*<@` + session.Info.User.ID + `>(\s*:)?\s+(.*?)\s*$`)
	if err != nil {
		return nil, err
	}

	return &dispatcher, nil
}

// AddHandler registers a MessageHandler with a Dispatcher.
func (d *Dispatcher) AddHandler(handler MessageHandler) {
	d.handlers = append(d.handlers, handler)
}

// Dispatch sends a Message to each registered MessageHandler in turn
// until it is handled.
func (d *Dispatcher) Dispatch(slackMsg *slack.MessageEvent) {
	var msg Message
	msg.MessageEvent = slackMsg
	msg.Session = d.session

	// ignore messages sent by another hambot
	if msg.User == d.session.Info.User.ID {
		return
	}

	matches := d.reAtHambot.FindStringSubmatch(msg.Text)
	if matches == nil {
		// accept messages without @hambot tag if sent directly to hambot
		if msg.IsDirect() {
			msg.DirectText = msg.Text
		}
	} else {
		msg.DirectText = matches[2]
		if !msg.IsDirect() {
			msg.ReplyPrefix = "<@" + msg.User + "> "
		}
	}

	for _, handler := range d.handlers {
		if handler.HandleMessage(msg) {
			break
		}
	}
}
