package main

import "github.com/nlopes/slack"

type Session struct {
	Client *slack.Client
	Info   *slack.Info
	RTM    *slack.RTM

	users    map[string]*slack.User
	channels map[string]*slack.Channel
	groups   map[string]*slack.Group
	ims      map[string]*slack.IM
}

func NewSession(client *slack.Client, info *slack.Info, rtm *slack.RTM) *Session {
	s := Session{Client: client, Info: info, RTM: rtm}

	s.users = make(map[string]*slack.User)
	for index := range info.Users {
		s.users[info.Users[index].ID] = &info.Users[index]
	}

	s.channels = make(map[string]*slack.Channel)
	for index := range info.Channels {
		s.channels[info.Channels[index].ID] = &info.Channels[index]
	}

	s.groups = make(map[string]*slack.Group)
	for index := range info.Groups {
		s.groups[info.Groups[index].ID] = &info.Groups[index]
	}

	s.ims = make(map[string]*slack.IM)
	for index := range info.IMs {
		s.ims[info.IMs[index].ID] = &info.IMs[index]
	}

	return &s
}

func (s *Session) GetUser(id string) *slack.User {
	return s.users[id]
}

func (s *Session) GetChannel(id string) *slack.Channel {
	return s.channels[id]
}

func (s *Session) GetGroup(id string) *slack.Group {
	return s.groups[id]
}

func (s *Session) GetIM(id string) *slack.IM {
	return s.ims[id]
}
