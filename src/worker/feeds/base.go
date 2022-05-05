package feeds

import (
	"bytes"
	"encoding/json"
	"log"
	"ozse/shared"
	. "ozse/worker/config"
)

type Feed interface {
	Init() error
	Run(task *shared.Task) error
}

type ValidatableFeed interface {
	Feed
	Validate(job *shared.Job) error
}

func jobDataPropertyUpdate(jobId string, property string, value interface{}) {
	h := make(map[string]interface{})
	h[property] = value
	body, _ := json.Marshal(&h)
	HttpClient.Post(Url("/jobs/"+jobId+"/data/update/"), "application/json", bytes.NewBuffer(body))
}

func getJob(id string) *shared.Job {
	var job shared.Job
	err := GetJson("/jobs/get/"+id, &job)
	if err != nil {
		log.Println(err)
	}
	return &job
}
func done(id string) {
	HttpClient.Post(Url("/tasks/done/"+id), "application/json", nil)
}

func doneResults(id string, results []interface{}) {
	obj := struct {
		Results []interface{} `json:"results"`
	}{
		Results: results,
	}
	body, _ := json.Marshal(&obj)
	HttpClient.Post(Url("/tasks/done/"+id), "application/json", bytes.NewBuffer(body))
}

// todo(cleanup): twitch feed wasn't working for some reason and this seems to have fixed it?
func doneResultsPtrTest(id string, results *[]interface{}) {
	obj := struct {
		Results *[]interface{} `json:"results"`
	}{
		Results: results,
	}
	body, _ := json.Marshal(&obj)
	HttpClient.Post(Url("/tasks/done/"+id), "application/json", bytes.NewBuffer(body))
}
