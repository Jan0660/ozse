package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/go-github/github"
	"github.com/mmcdole/gofeed"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"ozse/npm"
	"ozse/pubdev"
	"ozse/shared"
	"sort"
	"time"
)

type Config struct {
	MasterUrl         string `yaml:"masterUrl"`
	GitHubAccessToken string `yaml:"gitHubAccessToken"`
}

var config Config
var redditClient *reddit.Client
var githubClient *github.Client
var pubdevClient *pubdev.Client
var npmClient *npm.Client

func main() {
	{
		filesBytes, err := ioutil.ReadFile("./config.yaml")
		if err != nil {
			log.Fatal(err)
		}
		err = yaml.Unmarshal(filesBytes, &config)
		if err != nil {
			log.Fatal(err)
		}
	}
	{
		// ready clients
		var err error
		redditClient, err = reddit.NewReadonlyClient()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GitHubAccessToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		githubClient = github.NewClient(tc)

		pubdevClient = pubdev.DefaultClient()
		npmClient = npm.DefaultClient()
	}

	log.Println("sus")
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		var tasks []shared.Task
		err := getJson("/tasks", &tasks)
		if err != nil {
			log.Fatal(err)
		}
		for _, task := range tasks {
			log.Println("Handling task", task)

			done := func() {
				http.Post(url("/tasks/done/"+task.Id), "application/json", nil)
			}
			doneResults := func(results []interface{}) {
				obj := struct {
					Results []interface{} `json:"results"`
				}{
					Results: results,
				}
				body, _ := json.Marshal(&obj)
				http.Post(url("/tasks/done/"+task.Id), "application/json", bytes.NewBuffer(body))
			}

			switch task.Name {
			case "discord-webhook":
				{
					log.Println("discord-webhook")
					job := getJob(task.JobId)
					jsonBytes, _ := json.Marshal(struct {
						Content string `json:"content"`
					}{
						Content: job.Data["content"].(string),
					})
					resp, err := http.Post(job.Data["url"].(string), "application/json", bytes.NewBuffer(jsonBytes))
					if err != nil {
						log.Println(err)
					}
					log.Println(resp.Status)

					done()
				}
			case "gelbooru":
				{
					log.Println("gelbooru")
					job := getJob(task.JobId)
					fp := gofeed.NewParser()
					feed, _ := fp.ParseURL("https://gelbooru.com/index.php?page=cooliris")

					results := make([]interface{}, 50)

					lastPost, _ := job.Data["lastPost"].(string)
					broke := false

					for i, item := range feed.Items {
						if item.Link == lastPost {
							results = results[:i]
							if i != 0 {
								lastPost = feed.Items[0].Link
							}
							broke = true
							break
						}
						result := make(map[string]interface{})
						result["thumbnail"] = (item.Extensions["media"])["thumbnail"][0].Attrs["url"]
						result["image"] = (item.Extensions["media"])["content"][0].Attrs["url"]
						result["link"] = item.Link
						result["tags"] = item.Title
						results[i] = result
					}
					if !broke {
						lastPost = feed.Items[0].Link
					}
					jobDataPropertyUpdate(task.JobId, "lastPost", lastPost)
					doneResults(results)
				}
			case "reddit":
				{
					log.Println("reddit")
					job := getJob(task.JobId)
					fp := gofeed.NewParser()
					feed, _ := fp.ParseURL(job.Data["url"].(string))

					results := make([]interface{}, 25)

					lastPost, _ := job.Data["lastPost"].(string)
					broke := false

					for i, item := range feed.Items {
						if item.Link == lastPost {
							results = results[:i]
							if i != 0 {
								lastPost = feed.Items[0].Link
							}
							broke = true
							break
						}
						result := make(map[string]interface{})
						result["title"] = item.Title
						result["content"] = item.Content
						result["link"] = item.Link
						result["updated"] = item.Updated
						result["published"] = item.Published
						result["id"] = item.GUID

						result["author"] = item.Author.Name
						post, res, _ := redditClient.Post.Get(context.Background(), item.GUID[3:])
						result["nsfw"] = post.Post.NSFW
						result["spoiler"] = post.Post.Spoiler
						result["contentUrl"] = post.Post.URL
						result["contentText"] = post.Post.Body
						result["subredditName"] = post.Post.SubredditName
						println(post, res)

						results[i] = result
					}
					if !broke {
						lastPost = feed.Items[0].Link
					}
					jobDataPropertyUpdate(task.JobId, "lastPost", lastPost)

					doneResults(results)
				}
			case "github":
				{
					log.Println("github")
					job := getJob(task.JobId)

					lastId, ok := job.Data["lastId"].(int64)
					if !ok {
						lastId = int64(job.Data["lastId"].(float64))
					}
					broke := false

					releases, _, err := githubClient.Repositories.ListReleases(context.Background(), job.Data["owner"].(string), job.Data["repo"].(string), nil)
					if err != nil {
						log.Fatal(err)
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

					doneResults(results)
				}
			case "pubdev":
				{
					log.Println("pubdev")
					job := getJob(task.JobId)

					lastVersion := job.Data["lastVersion"].(string)
					broke := false

					pkg, err := pubdevClient.GetPackage(job.Data["name"].(string))
					if err != nil {
						log.Fatal(err)
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

					doneResults(results)
				}
			case "npm":
				{
					log.Println("npm")
					job := getJob(task.JobId)

					lastVersion := job.Data["lastVersion"].(string)
					broke := false

					pkg, err := npmClient.GetPackageMetadata(job.Data["name"].(string))
					if err != nil {
						log.Fatal(err)
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

					doneResults(results)
				}
			}
		}
	}

}

func jobDataPropertyUpdate(jobId string, property string, value interface{}) {
	h := make(map[string]interface{})
	h[property] = value
	body, _ := json.Marshal(&h)
	http.Post(url("/jobs/"+jobId+"/data/update/"), "application/json", bytes.NewBuffer(body))
}

func url(endpoint string) string {
	return config.MasterUrl + endpoint
}

func getJson(endpoint string, target interface{}) error {
	// todo: use a http client with a timeout
	r, err := http.Get(url(endpoint))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getJob(id string) *shared.Job {
	var job shared.Job
	err := getJson("/jobs/get/"+id, &job)
	if err != nil {
		log.Println(err)
	}
	return &job
}
