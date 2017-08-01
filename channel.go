package main

type Channel struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	IsChannel      bool `json:"is_channel"`
	Created        int64 `json:"created"`
	Creator        string `json:"creator"`
	IsArchived     bool `json:"is_archived"`
	IsGeneral      bool `json:"is_general"`
	NameNormalized string `json:"name_normalized"`
	IsShared       bool `json:"is_shared"`
	IsOrgShared    bool `json:"is_org_shared"`
	IsMember       bool `json:"is_member"`
	IsPrivate      bool `json:"is_private"`
	IsMpim         bool `json:"is_mpim"`
	Members        []string `json:"members"`
	Topic          Topic `json:"topic"`
	Purpose        Purpose `json:"purpose"`
	PreviousNames  []string `json:"previous_names"`
	NumMembers     int `json:"num_members"`
}

type Topic struct {
	Value   string `json:"value"`
	Creator string `json:"Creator"`
	LastSet int64 `json:"last_set"`
}

type Purpose struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64 `json:"last_set"`
}

type ChannelsList struct {
	Ok       bool `json:"ok"`
	Channels []Channel `json:"channels"`
}

type ChannelHistory struct {
	OK       bool `json:"ok"`
	Messages []Message `json:"messages"`
	HasMore  bool `json:"has_more"`
}
