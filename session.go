package main

import (
	"errors"

	"github.com/nlopes/slack"
)

// Session contains information about a Slack session.
type Session struct {
	Client    *slack.Client
	Info      *slack.Info
	RTM       *slack.RTM
	Callbacks chan func()

	usersByID    map[string]*slack.User
	channelsByID map[string]*slack.Channel
	groupsByID   map[string]*slack.Group
	imsByID      map[string]*slack.IM

	usersByName    map[string]*slack.User
	channelsByName map[string]*slack.Channel
	groupsByName   map[string]*slack.Group
	imsByName      map[string]*slack.IM

	hamBase string
}

// Start associates a Session with a running Slack session.
func (s *Session) Start(
	client *slack.Client,
	info *slack.Info,
	rtm *slack.RTM,
) {
	const channelBufferSize int = 32

	s.Client = client
	s.Info = info
	s.RTM = rtm
	s.Callbacks = make(chan func(), channelBufferSize)

	s.usersByID = make(map[string]*slack.User)
	s.usersByName = make(map[string]*slack.User)
	for index := range info.Users {
		s.usersByID[info.Users[index].ID] = &info.Users[index]
		s.usersByName[info.Users[index].Name] = &info.Users[index]
	}

	s.channelsByID = make(map[string]*slack.Channel)
	s.channelsByName = make(map[string]*slack.Channel)
	for index := range info.Channels {
		s.channelsByID[info.Channels[index].ID] = &info.Channels[index]
		s.channelsByName[info.Channels[index].Name] = &info.Channels[index]
	}

	s.groupsByID = make(map[string]*slack.Group)
	s.groupsByName = make(map[string]*slack.Group)
	for index := range info.Groups {
		s.groupsByID[info.Groups[index].ID] = &info.Groups[index]
		s.groupsByName[info.Groups[index].Name] = &info.Groups[index]
	}

	s.imsByID = make(map[string]*slack.IM)
	s.imsByName = make(map[string]*slack.IM)
	for index := range info.IMs {
		s.imsByID[info.IMs[index].ID] = &info.IMs[index]
		if imUser, ok := s.usersByName[info.IMs[index].User]; ok {
			s.imsByName[imUser.Name] = &info.IMs[index]
		}
	}
}

// HamBase returns the channel ID used for announcements.
func (s *Session) HamBase() (string, error) {
	if s.hamBase != "" {
		return s.hamBase, nil
	}
	if Settings.HamBase == "" {
		return "", errors.New("No ham base configured")
	}

	var channelID string
	if channel := s.ChannelByName(Settings.HamBase); channel != nil {
		channelID = channel.ID
	} else if group := s.GroupByName(Settings.HamBase); group != nil {
		channelID = group.ID
	} else {
		return "", errors.New("Channel '" + Settings.HamBase + "' not found")
	}

	s.hamBase = channelID
	return channelID, nil
}

// Announce sends a message to the channel returned by HamBase().
func (s *Session) Announce(text string) error {
	var channelID string
	var err error
	if channelID, err = s.HamBase(); err != nil {
		return err
	}
	s.RTM.SendMessage(s.RTM.NewOutgoingMessage(text, channelID))
	return nil
}

// User returns the User identified by the given ID.
func (s *Session) User(id string) *slack.User {
	return s.usersByID[id]
}

// Channel returns the Channel identified by the given ID.
func (s *Session) Channel(id string) *slack.Channel {
	return s.channelsByID[id]
}

// Group returns the Group identified by the given ID.
func (s *Session) Group(id string) *slack.Group {
	return s.groupsByID[id]
}

// IM returns the IM identified by the given ID.
func (s *Session) IM(id string) *slack.IM {
	return s.imsByID[id]
}

// UserByName returns the User nameentified by the given name.
func (s *Session) UserByName(name string) *slack.User {
	return s.usersByName[name]
}

// ChannelByName returns the Channel nameentified by the given name.
func (s *Session) ChannelByName(name string) *slack.Channel {
	return s.channelsByName[name]
}

// GroupByName returns the Group nameentified by the given name.
func (s *Session) GroupByName(name string) *slack.Group {
	return s.groupsByName[name]
}

// IMByName returns the IM nameentified by the given name.
func (s *Session) IMByName(name string) *slack.IM {
	return s.imsByName[name]
}
