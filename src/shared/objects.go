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
	Data       map[string]interface{} `json:"data"`
	Duplicates []string               `json:"duplicates"`
	// LastAdded is the last time(Unix seconds) the job was last added or last supposed to be added.
	// Used to keep a regular interval even between master restarts.
	LastAdded int64 `json:"lastAdded"`
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

type Packet struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type WorkerEvent struct {
	Type WorkerEventType `json:"type"`
	Data interface{}     `json:"data"`
}

type WorkerEventType uint8

const (
	NewTask WorkerEventType = iota
	ValidateJob
)

type ValidateJobResult struct {
	Valid bool        `json:"valid"`
	Error interface{} `json:"error"`
	Job   Job         `json:"job"`
}
