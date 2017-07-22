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

func main() {
	token := os.Args[1]
	values := url.Values{
		"token":  {token},
		"pretty": {"1"},
	}

	res, err := http.Get("https://slack.com/api/channels.list?" + values.Encode())

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	} else {
		defer res.Body.Close()
	}

	b, err := ioutil.ReadAll(res.Body)

	data := new(ChannelsList)

	if err := json.Unmarshal(b, data); err != nil {
		fmt.Println("JSON Unmarshal error", err)
		return
	}

	for _, c := range data.Channels {
		fmt.Println(c.Name, c.Id)
	}

	db, err := openDB("./slack.db")

	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	createDBTable(db)

	tx, err := db.Begin()
	stmt, err := tx.Prepare("insert into channels(id, name) values(?, ?)")

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, channel := range data.Channels {
		if _, err := stmt.Exec(channel.Id, channel.Name); err != nil {
			fmt.Println(err)
			// TODO: データ衝突時の処理を実装
		}
	}
	tx.Commit()
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
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS channels (id TEXT NOT NULL PRIMARY KEY, name TEXT NOT NULL)`)

	if err != nil {
		log.Fatal(err)
	}
}
