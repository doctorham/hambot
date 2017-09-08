package main

import "github.com/nlopes/slack"

const channelBufferSize int = 32

type Session struct {
	Client    *slack.Client
	Info      *slack.Info
	RTM       *slack.RTM
	Callbacks chan func()

	users    map[string]*slack.User
	channels map[string]*slack.Channel
	groups   map[string]*slack.Group
	ims      map[string]*slack.IM
}

func (this *Session) Start(client *slack.Client, info *slack.Info, rtm *slack.RTM) {
	this.Client = client
	this.Info = info
	this.RTM = rtm
	this.Callbacks = make(chan func(), channelBufferSize)

	this.users = make(map[string]*slack.User)
	for index := range info.Users {
		this.users[info.Users[index].ID] = &info.Users[index]
	}

	this.channels = make(map[string]*slack.Channel)
	for index := range info.Channels {
		this.channels[info.Channels[index].ID] = &info.Channels[index]
	}

	this.groups = make(map[string]*slack.Group)
	for index := range info.Groups {
		this.groups[info.Groups[index].ID] = &info.Groups[index]
	}

	this.ims = make(map[string]*slack.IM)
	for index := range info.IMs {
		this.ims[info.IMs[index].ID] = &info.IMs[index]
	}
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
