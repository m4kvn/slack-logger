package main

import (
	"os"
	"log"
	"flag"
	"io"
	"github.com/jasonlvhit/gocron"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"time"
	"errors"
	"net/url"
	"encoding/json"
	"strconv"
	"fmt"
)

func main() {
	defer getLogFile("./slack-logger.log").Close()

	log.Println("slack-logger start")
	token := flag.String("token", os.Getenv("SLACK_API_TOKEN"), "Set Slack API Token")
	channel := flag.String("channel", os.Getenv("NOTIFICATION_CHANNEL"), "Set notification slack channel name")
	notiTime := flag.String("time", os.Getenv("NOTIFICATION_TIME"), "Set notification time")
	dbhost := flag.String("dbhost", os.Getenv("DB_HOST"), "Set DB host")
	dbport := flag.String("dbport", os.Getenv("DB_PORT"), "Set DB port")
	dbuser := flag.String("dbuser", os.Getenv("DB_USER"), "Set DB user")
	dbpass := flag.String("dbpass", os.Getenv("DB_PASS"), "Set DB pass")
	dbname := flag.String("dbname", os.Getenv("DB_NAME"), "Set DB name")
	flag.Parse()

	db := connectDB(dbhost, dbport, dbuser, dbpass, dbname)
	defer db.Close()

	slackAPI := API{
		Token: *token,
		Notification: Notification{
			ChannelName: *channel,
			Time:        *notiTime,
		},
	}

	log.Println("token:", slackAPI.Token)
	log.Println("notificationChannelName:", slackAPI.Notification.ChannelName)
	log.Println("notificationTime:", slackAPI.Notification.Time)

	runLogger(db, slackAPI)

	gocron.Every(1).Day().At(slackAPI.Notification.Time).Do(runLogger, db, slackAPI)
	<-gocron.Start()
}

func connectDB(dbhost *string, dbport *string, dbuser *string, dbpass *string, dbname *string) mysql.Conn {
	db := mysql.New("tcp", "", *dbhost + ":" + *dbport, *dbuser, *dbpass, *dbname)

	connected := make(chan bool, 1)
	timeout := make(chan bool, 1)

	go func(s chan<- bool) {
		for {
			err := db.Connect()
			if err == nil {
				s <- true
				break
			} else {
				time.Sleep(2 * time.Second)
			}
		}
	}(connected)

	go func(s chan bool, t chan bool) {
		t <- true
		time.Sleep(20 * time.Second)
		if <-t {
			log.Println("Database connection was timeout.")
			s <- false
		}
	}(connected, timeout)

	log.Println("Waiting for database connection.")
	if <-connected {
		<-timeout
		timeout <- false
		log.Println("Connected the database.")
	} else {
		log.Println("Cannot connect the database.")
		os.Exit(1)
	}

	return db
}

func getLogFile(filePath string) *os.File {
	logfile, _ := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)
	return logfile
}

func runLogger(db mysql.Conn, api API) {
	log.Println("slack-logger run")
	api.UpdateChannels()

	newChannels := insertChannels(db, api)
	updates := insertHistory(db, api)
	go notification(api, newChannels, updates)

	if len(newChannels) > 0 {
		log.Println("New Channels:")
		for _, c := range newChannels {
			log.Println("Name:", c.Name, "Created:", time.Unix(c.Created, 0))
		}
	}

	if len(updates) > 0 {
		log.Println("Update Channels:")
		for key := range updates {
			log.Println("Name:", key, "Updates:", updates[key])
		}
	}

	log.Printf("slack-logger finish\n")
}

func notification(api API, newChannels []Channel, updates map[string]int) {
	if api.Notification.ChannelName == "" {
		return
	}

	notificationChannel, err := getNotificationChannel(api)

	if err != nil {
		log.Println(err)
		return
	}

	text := ""
	if len(newChannels) > 0 {
		text += "New channels:\n"
		for _, c := range newChannels {
			text += "\t<#" + c.Id + "> was created at " + getTimeStr(time.Unix(c.Created, 0)) + "\n"
		}
		text += "\n"
	}

	if len(updates) > 0 {
		text += "Update channels:\n"
		for key := range updates {
			text += "\t<#" + key + ">: " + strconv.Itoa(updates[key]) + "\n"
		}
		text += "\n"
	}

	attachments := []Attachment{
		{
			Fallback: "Required plain-text summary of the attachment.",
			Color:    "#36a64f",
			Pretext:  "Completed at " + getTimeStr(time.Now()) + "\n",
			Text:     text,
		},
	}

	attachmentsBytes, _ := json.Marshal(attachments)

	values := url.Values{}
	values.Add("token", api.Token)
	values.Add("channel", notificationChannel.Id)
	values.Add("text", "")
	values.Add("icon_emoji", ":banana:")
	values.Add("username", "Slack Logger")
	values.Add("attachments", string(attachmentsBytes))
	api.PostSlackAPI(values, "chat.postMessage")

	log.Println("Complete notification.")
}

func getTimeStr(t time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func getNotificationChannel(api API) (Channel, error) {
	for _, c := range api.Channels {
		if c.Name == api.Notification.ChannelName {
			return c, nil
		}
	}
	return Channel{}, errors.New("Not fount notification channel (" + api.Notification.ChannelName + ")")
}

func insertChannels(db mysql.Conn, api API) []Channel {
	log.Println("slack-logger insertChannels start")

	newChannels := []Channel{}
	for _, channel := range api.Channels {
		stmt, _ := db.Prepare("insert into channels values (?, ?, ?)")
		_, err := stmt.Run(channel.Id, channel.Name, channel.Created)
		if err == nil {
			newChannels = append(newChannels, channel)
		}
	}
	return newChannels
}

func insertHistory(db mysql.Conn, api API) map[string]int {
	log.Println("slack-logger insertHistory start")

	updates := map[string]int{}

	for _, channel := range api.Channels {
		stmt, _ := db.Prepare("select last_update from channels where channel_id = ?")
		row, _, _ := stmt.Exec(channel.Id)

		lastUpdate := ""
		for _, col := range row {
			if col != nil {
				lastUpdate = col.Str(0)
			}
			break
		}

		messages := api.GetChannelMessages(channel, lastUpdate)

		if len(messages) > 0 {
			for _, message := range messages {
				if message.Ts > lastUpdate {
					lastUpdate = message.Ts
				}
				stmt, _ := db.Prepare("insert into history values (?, ?, ?, ?, ?)")
				stmt.Run(channel.Id, message.Type, message.User, message.Text, message.Ts)
			}

			stmt, _ = db.Prepare("update channels set last_update = ? where channel_id = ?")
			stmt.Run(lastUpdate, channel.Id)
			updates[channel.Id] = len(messages)
		}
	}

	return updates
}
