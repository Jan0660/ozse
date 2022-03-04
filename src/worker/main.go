package main

import (
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"ozse/shared"
	. "ozse/worker/config"
	feeds2 "ozse/worker/feeds"
	"strings"
	"time"
)

var feeds map[string]interface{}

func main() {
	feeds = make(map[string]interface{})
	feeds["discord-webhook"] = &feeds2.DiscordWebhookFeed{}
	feeds["gelbooru"] = &feeds2.GelbooruFeed{}
	feeds["github"] = &feeds2.GitHubFeed{}
	feeds["npm"] = &feeds2.NpmFeed{}
	feeds["pubdev"] = &feeds2.PubDevFeed{}
	feeds["reddit"] = &feeds2.RedditFeed{}
	feeds["twitter"] = &feeds2.TwitterFeed{}

	for i, val := range feeds {
		v, ok := val.(feeds2.Feed)
		if !ok {
			log.Fatal("Could not cast to Feed")
		}
		err := v.Init()
		if err != nil {
			log.Fatal(err)
		}
		feeds[i] = v
	}

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

	conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(Url("/worker/ws"), "http", "ws", 1), nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var packet shared.Packet
		for {
			conn.ReadJSON(&packet)
			log.Println(packet)
			m := packet.Data.(map[string]interface{})
			if packet.Type == "worker-event" {
				switch m["type"] {
				case shared.NewTask:
					var task shared.Task
					mapstructure.Decode(m["data"], &task)
					log.Println(task)
					handleTask(task)
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
