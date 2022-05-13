package feeds

import (
	"github.com/mmcdole/gofeed"
	"ozse/shared"
)

type GelbooruFeed struct {
}

func (gf *GelbooruFeed) Init() error {
	return nil
}

func (gf *GelbooruFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://gelbooru.com/index.php?page=cooliris")
	if err != nil {
		return err
	}

	results := make([]interface{}, 50)

	lastPost, ok := job.Data["lastPost"].(string)
	if !ok {
		lastPost = ""
	}

	for i, item := range feed.Items {
		if item.Link == lastPost {
			results = results[:i]
			break
		}
		result := make(map[string]interface{})
		result["thumbnail"] = (item.Extensions["media"])["thumbnail"][0].Attrs["url"]
		result["image"] = (item.Extensions["media"])["content"][0].Attrs["url"]
		result["link"] = item.Link
		result["tags"] = item.Title
		results[i] = result
	}
	lastPost = feed.Items[0].Link
	jobDataPropertyUpdate(task.JobId, "lastPost", lastPost)
	doneResults(task.Id, results)
	return nil
}
