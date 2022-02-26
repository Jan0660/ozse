package feeds

import (
	"ozse/npm"
	"ozse/shared"
	"sort"
)

type NpmFeed struct {
	client *npm.Client
}

func (nf *NpmFeed) Init() error {
	nf.client = npm.DefaultClient()
	return nil
}

func (nf *NpmFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)

	lastVersion := job.Data["lastVersion"].(string)
	broke := false

	pkg, err := nf.client.GetPackageMetadata(job.Data["name"].(string))
	if err != nil {
		return err
	}

	results := make([]interface{}, len(pkg.Versions))

	reversed := make([]npm.PackageVersion, len(pkg.Versions))

	keys := make([]string, len(pkg.Versions))
	i := 0
	for v := range pkg.Versions {
		keys[i] = v
		i++
	}
	sort.Strings(keys)
	i = 0
	for _, k := range keys {
		reversed[len(pkg.Versions)-1-i] = pkg.Versions[k]
		i++
	}
	for i, item := range reversed {
		if item.Version == lastVersion {
			results = results[:i]
			if i != 0 {
				lastVersion = reversed[0].Version
			}
			broke = true
			break
		}
		result := make(map[string]interface{})

		result["previousVersion"] = lastVersion
		result["previous"] = pkg.Versions[lastVersion]
		result["package"] = item

		results[i] = result
	}
	if !broke {
		lastVersion = reversed[0].Version
	}
	jobDataPropertyUpdate(task.JobId, "lastVersion", lastVersion)

	doneResults(task.Id, results)
	return nil
}
