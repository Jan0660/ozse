package main

import (
	"context"
	. "github.com/ahmetb/go-linq/v3"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	. "ozse/shared"
)

func Dedup() {
	runAgain := true
	for runAgain {
		runAgain = dedupRun()
	}
}

func dedupRun() bool {
	f, _ := jobsCol.Find(context.Background(), bson.M{})
	var jobs []Job
	f.All(context.Background(), &jobs)
	for _, job := range jobs {
		q := From(jobs).Where(func(obj interface{}) bool {
			potentialDupe := obj.(Job)
			if potentialDupe.Id == job.Id {
				return false
			}
			switch potentialDupe.Name {
			case "gelbooru":
				return true
			case "reddit":
				return potentialDupe.Data["url"] == job.Data["url"]
			}
			return false
		})
		if q.Any() {
			println(q.Count())
			q.ForEach(func(obj interface{}) {
				dupe := obj.(Job)
				_, err := jobsCol.DeleteOne(context.Background(), bson.M{"_id": dupe.Id})
				if err != nil {
					log.Fatalln(err)
				}
				job.Duplicates = append(job.Duplicates, dupe.Id)
			})
			jobsCol.FindOneAndReplace(context.Background(), bson.M{"_id": job.Id}, job)
			return true
		}
	}
	return false
}
