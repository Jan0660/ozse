package config

type Configuration struct {
	MasterUrl            string `yaml:"masterUrl"`
	GitHubAccessToken    string `yaml:"gitHubAccessToken"`
	GoogleApiKey         string `yaml:"googleApiKey"`
	TwitchClientId       string `yaml:"twitchClientId"`
	TwitchClientSecret   string `yaml:"twitchClientSecret"`
	TwitchAppAccessToken string `yaml:"twitchAppAccessToken"`
}

var Config Configuration
