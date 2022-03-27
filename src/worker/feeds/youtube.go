package feeds

import (
	"context"
	"errors"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"ozse/shared"
	. "ozse/worker/config"
)

type YouTubeFeed struct {
	Service *youtube.Service
}

func (yf *YouTubeFeed) Init() error {
	ys, err := youtube.NewService(context.TODO(), option.WithAPIKey(Config.GoogleApiKey))
	yf.Service = ys
	return err
}

func (yf *YouTubeFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)

	lastId := job.Data["lastId"].(string)

	results := make([]interface{}, 10)
	// todo(perf): save channelId to job data for next time
	channelId, err := getChannelId(yf.Service, job.Data["name"].(string))
	if err != nil {
		return err
	}
	// todo: option to add snippet part
	call := yf.Service.Search.List([]string{"id"}).Type("video").ChannelId(channelId).Order("date")
	call.MaxResults(10)
	res, err := call.Do()
	if err != nil {
		return err
	}
	for i, item := range res.Items {
		if item.Id.VideoId == lastId {
			results = results[:i]
			break
		}
		result := make(map[string]interface{})

		result["item"] = item

		results[i] = result
	}
	lastId = res.Items[0].Id.VideoId
	jobDataPropertyUpdate(task.JobId, "lastId", lastId)

	doneResults(task.Id, results)
	return nil
}

func (yf *YouTubeFeed) Validate(job *shared.Job) error {
	_, err := getChannelId(yf.Service, job.Data["name"].(string))
	return err
}

func getChannelId(service *youtube.Service, q string) (string, error) {
	call := service.Search.List([]string{"id"}).Q(q).Type("channel")
	res, err := call.Do()
	if err != nil {
		return "", err
	}
	if len(res.Items) == 0 {
		return "", errors.New("no results")
	}
	// assumes first result is correct
	return res.Items[0].Id.ChannelId, nil
}
