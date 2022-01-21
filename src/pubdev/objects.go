package pubdev

import "time"

type Package struct {
	Name     string           `json:"name"`
	Latest   PackageVersion   `json:"latest"`
	Versions []PackageVersion `json:"versions"`
}

type PackageVersion struct {
	Version    string    `json:"version"`
	PubSpec    PubSpec   `json:"pubspec"`
	ArchiveUrl string    `json:"archive_url"`
	Published  time.Time `json:"published"`
}

type PubSpec struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Homepage        string            `json:"homepage"`
	Version         string            `json:"version"`
	Environment     map[string]string `json:"environment"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"dev_dependencies"`
}
