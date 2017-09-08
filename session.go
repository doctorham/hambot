package main

import (
	"errors"

	"github.com/nlopes/slack"
)

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
}

func (s *Session) Start(client *slack.Client, info *slack.Info, rtm *slack.RTM) {
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

func (s *Session) Announce(text string) error {
	if gConfig.AnnouncementChannel == "" {
		return errors.New("No announcement channel configured")
	}

	var channelID string
	if channel := s.GetChannelByName(gConfig.AnnouncementChannel); channel != nil {
		channelID = channel.ID
	} else if group := s.GetGroupByName(gConfig.AnnouncementChannel); group != nil {
		channelID = group.ID
	} else {
		return errors.New("Announcement channel '" + gConfig.AnnouncementChannel + "' not found")
	}

	s.RTM.SendMessage(s.RTM.NewOutgoingMessage(text, channelID))
	return nil
}

func (s *Session) GetUser(id string) *slack.User {
	return s.usersByID[id]
}

func (s *Session) GetChannel(id string) *slack.Channel {
	return s.channelsByID[id]
}

func (s *Session) GetGroup(id string) *slack.Group {
	return s.groupsByID[id]
}

func (s *Session) GetIM(id string) *slack.IM {
	return s.imsByID[id]
}

func (s *Session) GetUserByName(id string) *slack.User {
	return s.usersByName[id]
}

func (s *Session) GetChannelByName(id string) *slack.Channel {
	return s.channelsByName[id]
}

func (s *Session) GetGroupByName(id string) *slack.Group {
	return s.groupsByName[id]
}

func (s *Session) GetIMByName(id string) *slack.IM {
	return s.imsByName[id]
}
