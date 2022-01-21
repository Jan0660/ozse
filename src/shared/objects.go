package shared

type BaseObject struct {
	Id string `json:"_id" bson:"_id"`
}

type Job struct {
	Id   string `json:"_id" bson:"_id"`
	Name string `json:"name"`
	// Timer second interval for running the job
	Timer               uint32 `json:"timer"`
	AllowTaskDuplicates bool   `json:"allowTaskDuplicates"`
	// Data information for running the job
	Data map[string]interface{} `json:"data"`

	Duplicates []string `json:"duplicates"`
}

type Task struct {
	Id    string `json:"_id" bson:"_id"`
	Name  string `json:"name"`
	JobId string `json:"jobId"`
	// FirstRun if true, results are discarded
	FirstRun bool `json:"firstRun"`
}

type Result struct {
	Id      string                 `json:"_id" bson:"_id"`
	TaskId  string                 `json:"taskId"`
	JobId   string                 `json:"jobId"`
	JobName string                 `json:"jobName"`
	Data    map[string]interface{} `json:"data"`
}
