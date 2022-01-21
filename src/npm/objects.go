package npm

import "time"

type FullSearchResult struct {
	Objects []SearchResult `json:"objects"`
	Total   int            `json:"total"`
	Time    time.Time      `json:"time"`
}

type SearchResult struct {
	Package Package `json:"package"`
}

type Package struct {
	Name        string      `json:"name"`
	DistTags    DistTags    `json:"dist-tags"`
	Description string      `json:"description"`
	Keywords    []string    `json:"keywords"`
	Scope       string      `json:"scope"`
	Date        string      `json:"date"`
	Version     string      `json:"version"`
	Links       Links       `json:"links"`
	Author      Author      `json:"author"`
	Publisher   Publisher   `json:"publisher"`
	Maintainers []Publisher `json:"maintainers"`
}

type PackageMetadata struct {
	Id             string                    `json:"_id"`
	Rev            string                    `json:"_rev"`
	Name           string                    `json:"name"`
	DistTags       DistTags                  `json:"dist-tags"`
	Versions       map[string]PackageVersion `json:"versions"`
	Time           map[string]time.Time      `json:"time"`
	Maintainers    []Human                   `json:"maintainers"`
	Description    string                    `json:"description"`
	Homepage       string                    `json:"homepage"`
	Keywords       []string                  `json:"keywords"`
	Repository     Repository                `json:"repository"`
	Author         Author                    `json:"author"`
	Bugs           Bugs                      `json:"bugs"`
	License        string                    `json:"license"`
	Readme         string                    `json:"readme"`
	ReadmeFilename string                    `json:"readmeFilename"`
}

type Publisher struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Links struct {
	Npm        string `json:"npm"`
	Homepage   string `json:"homepage"`
	Repository string `json:"repository"`
	Bugs       string `json:"bugs"`
}

type DistTags struct {
	Latest string `json:"latest"`
}

type PackageVersion struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Main            string            `json:"main"`
	Author          Author            `json:"author"`
	Keywords        []string          `json:"keywords"`
	License         string            `json:"license"`
	Repository      Repository        `json:"repository"`
	Private         bool              `json:"private"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	GitHead         string            `json:"gitHead"`
	Bugs            Bugs              `json:"bugs"`
	Homepage        string            `json:"homepage"`
	Id              string            `json:"_id"`
	NodeVersion     string            `json:"_nodeVersion"`
	NpmVersion      string            `json:"_npmVersion"`
	Dist            Dist              `json:"dist"`
	Maintainers     []Human           `json:"maintainers"`
	NpmUser         Human             `json:"_npmUser"`
}

type Author struct {
	Name string `json:"name"`
}

type Repository struct {
	Url  string `json:"url"`
	Type string `json:"type"`
}
type Bugs struct {
	Url string `json:"url"`
}

type Dist struct {
	Integrity    string `json:"integrity"`
	ShaSum       string `json:"shasum"`
	Tarball      string `json:"tarball"`
	FileCount    int    `json:"fileCount"`
	UnpackedSize int    `json:"unpackedSize"`
	NpmSignature string `json:"npm-signature"`
}

type Human struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
