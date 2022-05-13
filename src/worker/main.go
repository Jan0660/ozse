package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"ozse/shared"
	. "ozse/worker/config"
	feeds2 "ozse/worker/feeds"
	"strings"
	"time"
)

var feeds map[string]interface{}

func main() {
	{
		filesBytes, err := ioutil.ReadFile("./config.yaml")
		if err != nil {
			log.Fatal(err)
		}
		err = yaml.Unmarshal(filesBytes, &Config)
		if err != nil {
			log.Fatal(err)
		}
	}

	feeds = make(map[string]interface{})
	feeds["discord-webhook"] = &feeds2.DiscordWebhookFeed{}
	feeds["gelbooru"] = &feeds2.GelbooruFeed{}
	feeds["github"] = &feeds2.GitHubFeed{}
	feeds["npm"] = &feeds2.NpmFeed{}
	feeds["pubdev"] = &feeds2.PubDevFeed{}
	feeds["reddit"] = &feeds2.RedditFeed{}
	feeds["twitter"] = &feeds2.TwitterFeed{}
	feeds["youtube"] = &feeds2.YouTubeFeed{}
	feeds["twitch"] = &feeds2.TwitchFeed{}
	feeds["rss"] = &feeds2.RssFeed{}

	for i, val := range feeds {
		v, ok := val.(feeds2.Feed)
		if !ok {
			log.Fatal("Could not cast to Feed")
		}
		err := v.Init()
		if err != nil {
			log.Println("error initializing feed", i)
			log.Fatal(err)
		}
		_, ok = val.(feeds2.ValidatableFeed)
		log.Println("Initialized feed", i, "validatable:", ok)
		feeds[i] = v
	}

	var conn *websocket.Conn
	var err error
	for conn, _, err = websocket.DefaultDialer.Dial(strings.Replace(Url("/worker/ws"), "http", "ws", 1), nil); err != nil; {
		log.Println("Could not connect to ws:", err)
		time.Sleep(time.Second * 2)
	}
	log.Println("Connected to ws")

	go func() {
		var packet shared.Packet
		for {
			conn.ReadJSON(&packet)
			log.Println(packet)
			m := packet.Data.(map[string]interface{})
			if packet.Type == "worker-event" {
				switch shared.WorkerEventType(uint8(m["type"].(float64))) {
				case shared.NewTask:
					var task shared.Task
					mapstructure.Decode(m["data"], &task)
					log.Println(task)
					handleTask(task)
					break
				case shared.ValidateJob:
					var job shared.Job
					mapstructure.Decode(m["data"], &job)
					job.Id = m["data"].(map[string]interface{})["_id"].(string)
					log.Println(job)
					handleValidateJob(job)
					break
				}
			}
		}
	}()

	log.Println("sus")
	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		var tasks []shared.Task
		err := GetJson("/tasks", &tasks)
		if err != nil {
			log.Println("Error getting /tasks", err)
		}
		for _, task := range tasks {
			handleTask(task)
		}
	}
}

func handleTask(task shared.Task) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("handleTask panic:", r)
		}
	}()
	log.Println("Handling task", task)

	f, ok := feeds[task.Name]
	if !ok {
		log.Println("Could not find feed", task.Name)
		return
	}
	feed := f.(feeds2.Feed)
	err := feed.Run(&task)
	if err != nil {
		log.Println("Error running feed", err)
	}
}

func handleValidateJob(job shared.Job) {
	log.Println("Handling validate job", job)

	f, ok := feeds[job.Name]
	if !ok {
		log.Println("Could not find feed", job.Name)
		return
	}
	feed, ok := f.(feeds2.ValidatableFeed)
	if ok == false {
		log.Println("Feed is not validatable")
		return
	}
	err := feed.Validate(&job)
	var errorStr interface{} = nil
	if err != nil {
		errorStr = err.Error()
	}
	obj := shared.ValidateJobResult{
		Valid: err == nil,
		Error: errorStr,
		Job:   job,
	}
	body, _ := json.Marshal(&obj)
	http.Post(Url("/worker/jobValidated"), "application/json", bytes.NewBuffer(body))
}
