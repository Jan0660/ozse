package config

type Configuration struct {
	MasterUrl         string `yaml:"masterUrl"`
	GitHubAccessToken string `yaml:"gitHubAccessToken"`
	GoogleApiKey      string `yaml:"googleApiKey"`
}

var Config Configuration
