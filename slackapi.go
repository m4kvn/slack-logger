package main

import (
	"net/url"
	"net/http"
	"io/ioutil"
	"log"
	"encoding/json"
)

type OAuth struct {
	OK     bool `json:"ok"`
	Url    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamId string `json:"team_id"`
	UserId string `json:"user_id"`
}

type API struct {
	Token        string
	Notification Notification
	Channels     []Channel
}

type Notification struct {
	ChannelName string
	Time        string
}

const baseSlackURL string = "https://slack.com/api/"

func (api *API) getSlackAPI(values url.Values, endpoint string) (b []byte, err error) {
	res, err := http.Get(baseSlackURL + endpoint + "?" + values.Encode())

	if err != nil {
		return nil, err
	} else {
		defer res.Body.Close()
		return ioutil.ReadAll(res.Body)
	}
}

func (api *API) UpdateChannels() {
	values := url.Values{
		"token":  {api.Token},
		"pretty": {"1"},
	}
	b, _ := api.getSlackAPI(values, "channels.list")
	data := new(ChannelsList)

	if err := json.Unmarshal(b, data); err != nil {
		log.Println("JSON Unmarshal error", err)
		return
	}

	api.Channels = data.Channels
}

func (api *API) GetChannelMessages(channel Channel, ts string) []Message {
	values := url.Values{}
	values.Add("token", api.Token)
	values.Add("channel", channel.Id)
	values.Add("count", "1000")
	if ts != "" {
		values.Add("oldest", ts)
	}
	values.Add("pretty", "1")
	b, _ := api.getSlackAPI(values, "channels.history")
	data := new(ChannelHistory)

	if err := json.Unmarshal(b, data); err != nil {
		log.Println("JSON Unmarshal error", err)
		return nil
	}

	return data.Messages
}

func (api *API) PostSlackAPI(values url.Values, endpoint string) {
	http.PostForm(baseSlackURL+endpoint+"?", values)
}
