package feeds

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"ozse/shared"
	. "ozse/worker/config"
)

type Feed interface {
	Init() error
	Run(task *shared.Task) error
}

func jobDataPropertyUpdate(jobId string, property string, value interface{}) {
	h := make(map[string]interface{})
	h[property] = value
	body, _ := json.Marshal(&h)
	http.Post(Url("/jobs/"+jobId+"/data/update/"), "application/json", bytes.NewBuffer(body))
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
	http.Post(Url("/tasks/done/"+id), "application/json", nil)
}

func doneResults(id string, results []interface{}) {
	obj := struct {
		Results []interface{} `json:"results"`
	}{
		Results: results,
	}
	body, _ := json.Marshal(&obj)
	http.Post(Url("/tasks/done/"+id), "application/json", bytes.NewBuffer(body))
}