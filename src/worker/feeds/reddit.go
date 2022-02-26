package feeds

import (
	"context"
	"github.com/mmcdole/gofeed"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"ozse/shared"
)

type RedditFeed struct {
	client *reddit.Client
}

func (rf *RedditFeed) Init() error {
	var err error
	rf.client, err = reddit.NewReadonlyClient()
	return err
}

func (rf *RedditFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(job.Data["url"].(string))

	results := make([]interface{}, 25)

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
		result["title"] = item.Title
		result["content"] = item.Content
		result["link"] = item.Link
		result["updated"] = item.Updated
		result["published"] = item.Published
		result["id"] = item.GUID

		result["author"] = item.Author.Name
		post, res, _ := rf.client.Post.Get(context.Background(), item.GUID[3:])
		result["nsfw"] = post.Post.NSFW
		result["spoiler"] = post.Post.Spoiler
		result["contentUrl"] = post.Post.URL
		result["contentText"] = post.Post.Body
		result["subredditName"] = post.Post.SubredditName
		println(post, res)

		results[i] = result
	}
	if !broke {
		lastPost = feed.Items[0].Link
	}
	jobDataPropertyUpdate(task.JobId, "lastPost", lastPost)

	doneResults(task.Id, results)
	return nil
}
