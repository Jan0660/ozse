package main

import (
	"container/list"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"ozse/shared"
	"time"
)

type Config struct {
	MongoUrl            string `yaml:"mongoUrl"`
	DatabaseName        string `yaml:"databaseName"`
	Address             string `yaml:"address"`
	EnableLastAddedWait bool   `yaml:"enableLastAddedWait"`
}

var config Config

var jobsCol *mongo.Collection
var tasksCol *mongo.Collection
var resultsCol *mongo.Collection

var upgrader = websocket.Upgrader{}

var userWriteChanList = list.New()
var workerEventChanList = list.New()

func main() {
	bytes, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(config.MongoUrl))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	database := client.Database(config.DatabaseName)
	database.CreateCollection(ctx, "jobs")
	database.CreateCollection(ctx, "tasks")
	database.CreateCollection(ctx, "results")
	jobsCol = database.Collection("jobs")
	tasksCol = database.Collection("tasks")
	resultsCol = database.Collection("results")

	r := gin.Default()
	r.GET("/ws", wsHandler)
	r.GET("/worker/ws", workerWsHandler)
	r.POST("/worker/jobValidated", func(c *gin.Context) {
		var result shared.ValidateJobResult
		err := c.BindJSON(&result)
		if err != nil {
			return
		}
		p := &shared.Packet{
			Type: "job-validated",
			Data: result,
		}
		for e := userWriteChanList.Front(); e != nil; e = e.Next() {
			ch := e.Value.(chan *shared.Packet)
			ch <- p
		}
	})
	r.POST("/jobs/new", func(c *gin.Context) {
		var job shared.Job
		err := c.BindJSON(&job)
		if err != nil {
			return
		}
		job.Id = shared.NewUlid().String()

		_, err = jobsCol.InsertOne(context.Background(), &job)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err,
			})
		}
		go JobTicker(&job, true)
		c.JSON(200, &job)
	})
	r.GET("/jobs/get/:id", func(c *gin.Context) {
		type Params struct {
			Id string `uri:"id" binding:"required"`
		}
		var params Params
		c.BindUri(&params)
		c.JSON(200, getJob(params.Id))
	})
	r.POST("/jobs/:jobId/data/update", func(c *gin.Context) {
		type Params struct {
			JobId string `uri:"jobId" binding:"required"`
		}
		var params Params
		c.BindUri(&params)

		var body map[string]interface{}
		c.BindJSON(&body)

		res := jobsCol.FindOne(context.Background(), bson.M{"_id": params.JobId})
		var job shared.Job
		res.Decode(&job)
		for i, e := range body {
			job.Data[i] = e
		}
		_, err := jobsCol.ReplaceOne(context.Background(), bson.M{"_id": job.Id}, job)

		if err != nil {
			log.Println(err)
		}
		c.Status(200)
	})
	r.POST("/jobs/:jobId/taskAdd", func(c *gin.Context) {
		job := getJob(c.Param("jobId"))
		if job.Id == "" {
			c.String(404, "Job not found")
			return
		}
		AddTask(job, false)
		c.Status(200)
	})
	r.POST("/jobs/validate", func(c *gin.Context) {
		var job shared.Job
		err := c.BindJSON(&job)
		if err != nil {
			return
		}
		event := &shared.WorkerEvent{
			Type: shared.ValidateJob,
			Data: job,
		}
		for e := workerEventChanList.Front(); e != nil; e = e.Next() {
			e.Value.(chan *shared.WorkerEvent) <- event
		}
		c.Status(200)
	})
	r.GET("/tasks", func(c *gin.Context) {
		cursor, err := tasksCol.Find(context.Background(), bson.D{})
		if err != nil {
			c.Status(500)
			return
		}
		var results []shared.Task
		cursor.All(context.Background(), &results)
		c.JSON(200, &results)
	})
	r.POST("/tasks/done/:id", func(c *gin.Context) {
		type Params struct {
			Id string `uri:"id" binding:"required"`
		}
		var params Params
		c.BindUri(&params)

		type BodyParams struct {
			Results []map[string]interface{} `json:"results"`
		}
		var body BodyParams
		if err := c.ShouldBindJSON(&body); err == nil {
			res := tasksCol.FindOne(context.Background(), bson.M{"_id": params.Id})
			var task shared.Task
			res.Decode(&task)

			if task.FirstRun {
				goto afterAdd
			}

			results := make([]shared.Result, len(body.Results))

			job := getJob(task.JobId)

			if job.Duplicates == nil {
				job.Duplicates = []string{}
			}

			for _, j := range append(job.Duplicates, job.Id) {
				for i, result := range body.Results {
					results[i] = shared.Result{
						Id:      shared.NewUlid().String(),
						TaskId:  task.Id,
						JobId:   j,
						JobName: task.Name,
						Data:    result,
					}

					packet := &shared.Packet{
						Type: "new-result",
						Data: results[i],
					}
					for e := userWriteChanList.Front(); e != nil; e = e.Next() {
						e.Value.(chan *shared.Packet) <- packet
					}
				}
				documents := make([]interface{}, len(results))
				for i := range results {
					documents[i] = results[i]
				}
				resultsCol.InsertMany(context.Background(), documents)
			}
		}

	afterAdd:
		del, err := tasksCol.DeleteOne(context.Background(), bson.M{"_id": params.Id})
		if err != nil || del.DeletedCount == 0 {
			c.Status(404)
			return
		}
		c.Status(200)
	})
	r.GET("/results", func(c *gin.Context) {
		type Params struct {
			Count int64 `form:"count"`
		}
		var params Params
		c.BindQuery(&params)
		if params.Count == 0 {
			params.Count = 10
		}
		opts := options.Find()
		opts.Limit = &params.Count
		cursor, _ := resultsCol.Find(context.Background(), bson.M{}, opts)
		var results []shared.Result
		cursor.All(context.Background(), &results)

		if results == nil {
			results = make([]shared.Result, 0)
		}

		c.JSON(200, &results)
	})
	r.DELETE("/results/:id", func(c *gin.Context) {
		type Params struct {
			Id string `uri:"id" binding:"required"`
		}
		var params Params
		c.BindUri(&params)
		resultsCol.DeleteOne(context.Background(), bson.M{"_id": params.Id})
		c.Status(200)
	})

	r.GET("/dedup", func(c *gin.Context) {
		Dedup()
		c.Status(200)
	})

	go func() {
		cursor, err := jobsCol.Find(context.Background(), bson.D{})
		if err != nil {
			log.Fatalln(err)
		}
		var jobs []shared.Job
		cursor.All(context.Background(), &jobs)
		for i := range jobs {
			go JobTicker(&jobs[i], false)
		}
		log.Println("job tickers started")
	}()

	r.Run(config.Address)
}

func getJob(id string) *shared.Job {
	res := jobsCol.FindOne(context.Background(), bson.M{"_id": id})
	var job shared.Job
	res.Decode(&job)
	return &job
}

func JobTicker(job *shared.Job, firstRun bool) {
	if config.EnableLastAddedWait {
		shouldWait := job.LastAdded + int64(job.Timer) - time.Now().Unix()
		if shouldWait > 0 {
			println(job.Name, job.Id, job.Timer, shouldWait)
			time.Sleep(time.Duration(shouldWait) * time.Second)
		}
	}
	ticker := time.NewTicker(time.Duration(job.Timer) * time.Second)
	AddTask(job, firstRun)
	for range ticker.C {
		log.Println(job.Name, "tick!")
		AddTask(job, false)
	}
}

func AddTask(job *shared.Job, firstRun bool) {
	if config.EnableLastAddedWait {
		go UpdateJobLastAdded(job.Id)
	}
	if !job.AllowTaskDuplicates {
		count, _ := tasksCol.CountDocuments(context.Background(), bson.M{"jobid": job.Id})
		if count != 0 {
			log.Println("nono duplicates")
			return
		}
	}
	task := &shared.Task{
		Name:     job.Name,
		JobId:    job.Id,
		Id:       shared.NewUlid().String(),
		FirstRun: firstRun,
	}
	tasksCol.InsertOne(context.Background(), task)
	go func() {
		event := &shared.WorkerEvent{
			Type: shared.NewTask,
			Data: task,
		}
		for e := workerEventChanList.Front(); e != nil; e = e.Next() {
			e.Value.(chan *shared.WorkerEvent) <- event
		}
	}()
	log.Println(job.Name, "task added!")
}

func UpdateJobLastAdded(jobId string) {
	jobsCol.UpdateOne(context.Background(), bson.M{"_id": jobId}, bson.M{"$set": bson.M{"lastadded": time.Now().Unix()}})
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed", err)
		return
	}

	packetWriteChan := make(chan *shared.Packet)
	userWriteChanList.PushFront(packetWriteChan)

	defer func() {
		conn.Close()
	}()

	var packet shared.Packet

	go func() {
		for {
			o := <-packetWriteChan
			conn.WriteJSON(o)
		}
	}()

	for {
		err = conn.ReadJSON(&packet)
		if err != nil {
			return
		}
		log.Println("ws", packet)
		switch packet.Type {
		case "ping":
			packet.Type = "pong"
			conn.WriteJSON(packet)
			break
		}
	}
}

func workerWsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Worker WebSocket upgrade failed", err)
		return
	}

	workerEventChan := make(chan *shared.WorkerEvent)
	workerEventChanList.PushFront(workerEventChan)

	defer func() {
		conn.Close()
	}()

	var packet shared.Packet

	go func() {
		for {
			o := <-workerEventChan
			conn.WriteJSON(shared.Packet{
				Type: "worker-event",
				Data: o,
			})
		}
	}()

	for {
		err = conn.ReadJSON(&packet)
		if err != nil {
			return
		}
		log.Println("worker ws", packet)
		switch packet.Type {
		case "ping":
			packet.Type = "pong"
			conn.WriteJSON(packet)
			break
		}
	}
}
