package main

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"os"
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Topic struct {
	Value   string `json:"value"`
	Creator string `json:"Creator"`
	LastSet uint `json:"last_set"`
}

type Purpose struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet uint `json:"last_set"`
}

type Channel struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	IsChannel      bool `json:"is_channel"`
	Created        uint `json:"created"`
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

type ChannelsList struct {
	Ok       bool `json:"ok"`
	Channels []Channel `json:"channels"`
}

type OAuth struct {
	OK     bool `json:"ok"`
	Url    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamId string `json:"team_id"`
	UserId string `json:"user_id"`
}

type Message struct {
	Type string `json:"type"`
	User string `json:"user"`
	Text string `json:"text"`
	Ts   string `json:"ts"`
}

type ChannelHistory struct {
	OK       bool `json:"ok"`
	Messages []Message `json:"messages"`
	HasMore  bool `json:"has_more"`
}

var Token string

const BaseSlackURL string = "https://slack.com/api/"

func main() {
	Token = os.Args[1]
	channels := getChannels()
	db, err := openDB("./slack.db")

	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	createDBTable(db)
	insertChannels(db, channels)
	createChannelTables(db, channels)

	//for _, message := range getChannelMessages(data.Channels[0]) {
	//	fmt.Println(message.Text)
	//}
}

func insertChannels(db *sql.DB, channels []Channel) {
	tx, err := db.Begin()
	stmt, err := tx.Prepare("insert into channels(id, name) values(?, ?)")

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, channel := range channels {
		if _, err := stmt.Exec(channel.Id, channel.Name); err != nil {
			// TODO: データ衝突時の処理を実装
		}
	}
	tx.Commit()
}

func createChannelTables(db *sql.DB, channels []Channel) {
	for _, channel := range channels {
		createChannelTable(db, channel)
	}
}

func createDBFile(fileName string) {
	if _, err := os.Stat(fileName); err != nil {
		os.Create(fileName)
	}
}

func openDB(fileName string) (db *sql.DB, err error) {
	createDBFile(fileName)
	return sql.Open("sqlite3", fileName)
}

func createDBTable(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS channels (id TEXT NOT NULL PRIMARY KEY, name TEXT NOT NULL)")

	if err != nil {
		fmt.Println(err)
		return
	}
}

func createChannelTable(db *sql.DB, channel Channel) {
	stmt := "create table if not exists " + channel.Id + " (type text, user text, text text, ts text)"
	_, err := db.Exec(stmt)

	if err != nil {
		fmt.Println(err)
		return
	}
}

func getSlackAPI(values url.Values, endpoint string) (b []byte, err error) {
	res, err := http.Get(BaseSlackURL + endpoint + "?" + values.Encode())

	if err != nil {
		return nil, err
	} else {
		defer res.Body.Close()
		return ioutil.ReadAll(res.Body)
	}
}

func getChannels() []Channel {
	values := url.Values{
		"token":  {Token},
		"pretty": {"1"},
	}
	b, _ := getSlackAPI(values, "channels.list")
	data := new(ChannelsList)

	if err := json.Unmarshal(b, data); err != nil {
		fmt.Println("JSON Unmarshal error", err)
		return nil
	}

	return data.Channels
}

func getChannelMessages(channel Channel) []Message {
	values := url.Values{
		"token":   {Token},
		"channel": {channel.Id},
		"pretty":  {"1"},
	}
	b, _ := getSlackAPI(values, "channels.history")
	data := new(ChannelHistory)

	if err := json.Unmarshal(b, data); err != nil {
		fmt.Println("JSON Unmarshal error", err)
		return nil
	}

	return data.Messages
}
