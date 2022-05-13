package feeds

import (
	"errors"
	"github.com/mmcdole/gofeed"
	"ozse/shared"
)

type RssFeed struct {
}

func (rf *RssFeed) Init() error {
	return nil
}

func (rf *RssFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(job.Data["url"].(string))

	results := make([]interface{}, len(feed.Items))

	lastLink, _ := job.Data["lastLink"].(string)

	for i, item := range feed.Items {
		if item.Link == lastLink {
			results = results[:i]
			if i == 0 {
				return nil
			}
			break
		}
		result := make(map[string]interface{})
		result["item"] = *item
		results[i] = result
	}
	lastLink = feed.Items[0].Link
	jobDataPropertyUpdate(task.JobId, "lastLink", lastLink)
	doneResults(task.Id, results)
	return nil
}

func (rf *RssFeed) Validate(job *shared.Job) error {
	if job.Data["url"] == nil {
		return errors.New("url is required")
	}
	fp := gofeed.NewParser()
	_, err := fp.ParseURL(job.Data["url"].(string))
	return err
}
