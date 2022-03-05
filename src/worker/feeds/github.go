package feeds

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"ozse/shared"
	. "ozse/worker/config"
)

type GitHubFeed struct {
	client *github.Client
}

func (gf *GitHubFeed) Init() error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: Config.GitHubAccessToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	gf.client = github.NewClient(tc)
	// todo: check if token is valid
	return nil
}

func (gf *GitHubFeed) Run(task *shared.Task) error {
	job := getJob(task.JobId)

	lastId, ok := job.Data["lastId"].(int64)
	if !ok {
		lastId = int64(job.Data["lastId"].(float64))
	}
	broke := false

	releases, _, err := gf.client.Repositories.ListReleases(context.Background(), job.Data["owner"].(string), job.Data["repo"].(string), nil)
	if err != nil {
		return err
	}

	results := make([]interface{}, len(releases))

	for i, item := range releases {
		if *item.ID == lastId {
			results = results[:i]
			if i != 0 {
				lastId = *releases[0].ID
			}
			broke = true
			break
		}
		result := make(map[string]interface{})

		result["id"] = *item.ID
		result["tagName"] = *item.TagName
		result["name"] = *item.Name
		result["body"] = *item.Body
		result["authorName"] = *item.Author.Login
		result["authorAvatar"] = *item.Author.AvatarURL
		result["url"] = *item.URL
		result["htmlUrl"] = *item.HTMLURL
		result["tarball"] = *item.TarballURL
		result["zipball"] = *item.ZipballURL
		result["createdAt"] = item.CreatedAt.UnixMilli()
		result["publishedAt"] = item.PublishedAt.UnixMilli()
		result["assets"] = item.Assets
		result["prerelease"] = *item.Prerelease
		result["targetCommitish"] = *item.TargetCommitish
		result["draft"] = *item.Draft

		results[i] = result
	}
	if !broke {
		lastId = *releases[0].ID
	}
	jobDataPropertyUpdate(task.JobId, "lastId", lastId)

	doneResults(task.Id, results)
	return nil
}

func (gf *GitHubFeed) Validate(job *shared.Job) error {
	_, _, err := gf.client.Repositories.Get(context.Background(), job.Data["owner"].(string), job.Data["repo"].(string))
	return err
}
