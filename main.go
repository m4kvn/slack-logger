package main

import (
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"os"
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
	"strconv"
	"github.com/jasonlvhit/gocron"
	"flag"
	"io"
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

type Attachment struct {
	Text           string `json:"text"`
	Fallback       string `json:"fallback"`
	CallbackId     string `json:"callback_id"`
	Color          string `json:"color"`
	AttachmentType string `json:"attachment_type"`
	Pretext        string `json:"pretext"`
}

var token string
var channelName string
var notificationTime string

var updateChannels map[string]int

const BaseSlackURL string = "https://slack.com/api/"

func main() {
	logfile, err := os.OpenFile("./slack-logger.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("cannnot open slack-logger.log:", err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)

	log.Println("slack-logger start")

	flag.StringVar(&token, "token", os.Getenv("SLACK_API_TOKEN"), "Slack API Token")
	flag.StringVar(&channelName, "channel", os.Getenv("NOTIFICATION_CHANNEL"), "notification slack channel name")
	flag.StringVar(&notificationTime, "time", os.Getenv("NOTIFICATION_TIME"), "notification time")
	flag.Parse()

	log.Println("token:", token)
	log.Println("notificationChannelName:", channelName)
	log.Println("notificationTime:", notificationTime)

	gocron.Every(1).Day().At(notificationTime).Do(runLogger)
	<-gocron.Start()
}

func runLogger() {
	log.Println("slack-logger run")
	channels := getChannels()
	db, err := openDB("./slack.db")

	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	createDBTable(db)
	insertChannels(db, channels)
	insertHistory(db, channels)
	notification(db, channelName)

	log.Printf("slack-logger finish\n")
}

func notification(db *sql.DB, notificationChannelName string) {
	if notificationChannelName != "" {
		var channelId string
		stmt, _ := db.Prepare("SELECT id FROM channels WHERE name = ?")
		stmt.QueryRow(notificationChannelName).Scan(&channelId)
		stmt.Close()

		if channelId == "" {
			log.Println("Not found notification channel: ", notificationChannelName)
			return
		}

		text := ""
		if len(updateChannels) > 0 {
			text += "Update channels:\n"
			for channelName := range updateChannels {
				text += "\t" + channelName + ": +" + strconv.Itoa(updateChannels[channelName]) + "\n"
			}
		}

		attachments := []Attachment{
			{
				Fallback: "Required plain-text summary of the attachment.",
				Color:    "#36a64f",
				Pretext:  "Complete at " + time.Now().String() + "\n",
				Text:     text,
			},
		}

		attachmentsBytes, _ := json.Marshal(attachments)

		values := url.Values{}
		values.Add("token", token)
		values.Add("channel", channelId)
		values.Add("text", "")
		values.Add("icon_emoji", ":banana:")
		values.Add("username", "Slack Logger")
		values.Add("attachments", string(attachmentsBytes))
		http.PostForm(BaseSlackURL+"chat.postMessage?", values)
	}
}

func insertHistory(db *sql.DB, channels []Channel) {
	updateChannels = make(map[string]int)

	for _, channel := range channels {
		stmt, _ := db.Prepare("SELECT ts FROM channels WHERE id = ?")
		var ts string
		stmt.QueryRow(channel.Id).Scan(&ts)
		stmt.Close()

		tx, _ := db.Begin()
		stmt, _ = tx.Prepare("insert into history(id, type, user, text, ts) values(?, ?, ?, ?, ?)")
		messages := getChannelMessages(channel, ts)

		if len(messages) > 0 {
			updateChannels[channel.Name] = len(messages)
		}
		log.Println(channel.Id, channel.Name, len(messages), ts)

		for _, message := range messages {
			if _, err := stmt.Exec(channel.Id, message.Type, message.User, message.Text, message.Ts); err != nil {
				// TODO: データ衝突時の処理を実装
			}
		}

		tx.Commit()
		stmt.Close()

		stmt, _ = db.Prepare("SELECT max(ts) FROM history WHERE id = ?")
		var tsNew string
		stmt.QueryRow(channel.Id).Scan(&tsNew)
		stmt.Close()

		stmt, _ = db.Prepare("UPDATE channels SET ts = ? WHERE id = ?")
		stmt.Exec(tsNew, channel.Id)
		stmt.Close()
	}
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
	stmt := `
	create table if not exists channels (
	id text not null primary key,
	name text not null,
	ts text
	)
	`
	_, err := db.Exec(stmt)

	if err != nil {
		log.Println(err)
	}

	stmt = `
	create table if not exists history (
	id text not null,
	type text not null,
	user text not null,
	text text,
	ts text not null
	)
	`

	_, err = db.Exec(stmt)

	if err != nil {
		log.Println(err)
	}

	stmt = "create index if not exists id_ts on history(id, ts)"

	_, err = db.Exec(stmt)

	if err != nil {
		log.Println(err)
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
		"token":  {token},
		"pretty": {"1"},
	}
	b, _ := getSlackAPI(values, "channels.list")
	data := new(ChannelsList)

	if err := json.Unmarshal(b, data); err != nil {
		log.Println("JSON Unmarshal error", err)
		return nil
	}

	return data.Channels
}

func getChannelMessages(channel Channel, ts string) []Message {
	values := url.Values{}
	values.Add("token", token)
	values.Add("channel", channel.Id)
	values.Add("count", "1000")
	if ts != "" {
		values.Add("oldest", ts)
	}
	values.Add("pretty", "1")
	b, _ := getSlackAPI(values, "channels.history")
	data := new(ChannelHistory)

	if err := json.Unmarshal(b, data); err != nil {
		log.Println("JSON Unmarshal error", err)
		return nil
	}

	return data.Messages
}
