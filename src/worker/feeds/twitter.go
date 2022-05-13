package feeds

import (
	twitter "github.com/n0madic/twitter-scraper"
	"ozse/shared"
)

type TwitterFeed struct {
	scraper *twitter.Scraper
}

func (tf *TwitterFeed) Init() error {
	tf.scraper = twitter.New()
	tf.scraper.WithDelay(1)
	tf.scraper.SetSearchMode(twitter.SearchLatest)
	return nil
}

func (tf *TwitterFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)

	lastId := job.Data["lastId"].(string)

	results := make([]interface{}, 40)
	items, _, err := tf.scraper.FetchTweets(job.Data["name"].(string), 40, "")
	if err != nil {
		return err
	}
	for i, item := range items {
		if item.ID == lastId {
			results = results[:i]
			break
		}
		result := make(map[string]interface{})

		result["tweet"] = item

		results[i] = result
	}
	lastId = items[0].ID
	jobDataPropertyUpdate(task.JobId, "lastId", lastId)

	doneResults(task.Id, results)
	return nil
}

func (tf *TwitterFeed) Validate(job *shared.Job) error {
	_, err := tf.scraper.GetUserIDByScreenName(job.Data["name"].(string))
	return err
}
