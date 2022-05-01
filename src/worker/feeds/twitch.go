package feeds

import (
	"container/list"
	"errors"
	"github.com/ahmetb/go-linq/v3"
	"github.com/nicklaw5/helix"
	"ozse/shared"
	. "ozse/worker/config"
)

type TwitchFeed struct {
	Client *helix.Client
}

func (tf *TwitchFeed) Init() error {
	client, err := helix.NewClient(&helix.Options{
		ClientID: Config.TwitchClientId,
		//ClientSecret:   Config.TwitchClientSecret,
		AppAccessToken: Config.TwitchAppAccessToken,
	})
	tf.Client = client
	return err
}

func (tf *TwitchFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)

	// todo(cleanup): look if there's a smarted way to do this or just make a function for this
	oldOnlineIdsInterfaces := job.Data["onlineIds"].([]interface{})
	oldOnlineIds := make([]string, len(oldOnlineIdsInterfaces))
	for i, item := range oldOnlineIdsInterfaces {
		oldOnlineIds[i] = item.(string)
	}
	usersInterfaces := job.Data["users"].([]interface{})
	users := make([]string, len(usersInterfaces))
	for i, item := range usersInterfaces {
		users[i] = item.(string)
	}
	results := list.New()
	res, err := tf.Client.GetStreams(&helix.StreamsParams{
		UserLogins: users,
	})
	if err != nil {
		return err
	}
	onlineIds := make([]string, len(res.Data.Streams))
	for i, item := range res.Data.Streams {
		result := make(map[string]interface{})
		if item.Type == "live" {
			onlineIds[i] = item.UserID
		}
		if item.Type == "live" && linq.From(oldOnlineIds).Contains(item.UserID) {
			// not changed
			continue
		} else if item.Type != "live" && linq.From(oldOnlineIds).Contains(item.UserID) {
			result["type"] = "goOffline"
			results.PushFront(map[string]interface{}{
				"type": "goOffline",
				"item": item,
			})
		} else {
			result["type"] = "goOnline"
		}
		result["item"] = item

		results.PushFront(result)
	}
	jobDataPropertyUpdate(task.JobId, "onlineIds", onlineIds)

	resultsArray := make([]interface{}, results.Len())
	i := 0
	for item := results.Back(); item != nil; item = item.Next() {
		resultsArray[i] = item.Value
		i++
	}
	doneResultsPtrTest(task.Id, &resultsArray)
	return nil
}

func (tf *TwitchFeed) Validate(job *shared.Job) error {
	usersInterfaces := job.Data["users"].([]interface{})
	users := make([]string, len(usersInterfaces))
	for i, item := range usersInterfaces {
		users[i] = item.(string)
	}
	resp, err := tf.Client.GetUsers(&helix.UsersParams{
		Logins: users,
	})
	if err != nil {
		return err
	}
	if len(resp.Data.Users) == 0 {
		return errors.New("no users found")
	}
	return nil
}
