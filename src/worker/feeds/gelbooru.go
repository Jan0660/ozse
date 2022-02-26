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
	feed, _ := fp.ParseURL("https://gelbooru.com/index.php?page=cooliris")

	results := make([]interface{}, 50)

	lastPost, _ := job.Data["lastPost"].(string)
	broke := false

	for i, item := range feed.Items {
		if item.Link == lastPost {
			results = results[:i]
			if i != 0 {
				lastPost = feed.Items[0].Link
			}
			broke = true
			break
		}
		result := make(map[string]interface{})
		result["thumbnail"] = (item.Extensions["media"])["thumbnail"][0].Attrs["url"]
		result["image"] = (item.Extensions["media"])["content"][0].Attrs["url"]
		result["link"] = item.Link
		result["tags"] = item.Title
		results[i] = result
	}
	if !broke {
		lastPost = feed.Items[0].Link
	}
	jobDataPropertyUpdate(task.JobId, "lastPost", lastPost)
	doneResults(task.Id, results)
	return nil
}
