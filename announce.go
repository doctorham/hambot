package main

import (
	"log"
	"regexp"
)

// HamAnnounce forwards messages to hambot's ham base.
type HamAnnounce struct {
	reAnnounce *regexp.Regexp
}

// NewHamAnnounce creates a new HamAnnounce.
func NewHamAnnounce() (*HamAnnounce, error) {
	a := HamAnnounce{}
	var err error
	if a.reAnnounce, err = regexp.Compile(`^announce\s+(.*)$`); err != nil {
		return nil, err
	}
	return &a, nil
}

// HandleMessage handles a Message.
func (a *HamAnnounce) HandleMessage(msg Message) bool {
	matches := a.reAnnounce.FindStringSubmatch(msg.DirectText)
	if matches == nil {
		return false
	}

	allowedUser := msg.Session.UserByName(Settings.Announcer)
	if allowedUser == nil || msg.User != allowedUser.ID {
		log.Printf("announce failed: disallowed user (%v != %v)\n", msg.User, allowedUser.ID)
		return false
	}

	log.Printf("announce '%v'\n", matches[1])

	msg.Session.Announce(matches[1])
	return true
}
