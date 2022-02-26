package feeds

import (
	"ozse/pubdev"
	"ozse/shared"
)

type PubDevFeed struct {
	client *pubdev.Client
}

func (pdf *PubDevFeed) Init() error {
	pdf.client = pubdev.DefaultClient()
	return nil
}

func (pdf *PubDevFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)

	lastVersion := job.Data["lastVersion"].(string)
	broke := false

	pkg, err := pdf.client.GetPackage(job.Data["name"].(string))
	if err != nil {
		return err
	}

	results := make([]interface{}, len(pkg.Versions))

	for i := range pkg.Versions {
		item := pkg.Versions[len(pkg.Versions)-1-i]
		if item.Version == lastVersion {
			results = results[:i]
			if i != 0 {
				lastVersion = pkg.Versions[len(pkg.Versions)-1].Version
			}
			broke = true
			break
		}
		result := make(map[string]interface{})

		// todo: rename to previousVersion
		result["lastVersion"] = lastVersion
		result["package"] = item

		results[i] = result
	}
	if !broke {
		lastVersion = pkg.Versions[len(pkg.Versions)-1].Version
	}
	jobDataPropertyUpdate(task.JobId, "lastVersion", lastVersion)

	doneResults(task.Id, results)
	return nil
}
