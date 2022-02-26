package config

type Configuration struct {
	MasterUrl         string `yaml:"masterUrl"`
	GitHubAccessToken string `yaml:"gitHubAccessToken"`
}

var Config Configuration
