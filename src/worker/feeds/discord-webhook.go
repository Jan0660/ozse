package feeds

import (
	"bytes"
	"encoding/json"
	"ozse/shared"
	. "ozse/worker/config"
)

type DiscordWebhookFeed struct{}

func (dwf *DiscordWebhookFeed) Init() error {
	return nil
}

func (dwf *DiscordWebhookFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)
	jsonBytes, _ := json.Marshal(struct {
		Content string `json:"content"`
	}{
		Content: job.Data["content"].(string),
	})
	_, err := HttpClient.Post(job.Data["url"].(string), "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	done(task.Id)
	return nil
}
