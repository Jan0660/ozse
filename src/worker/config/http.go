package config

import (
	"encoding/json"
	"net/http"
	"time"
)

var HttpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func Url(endpoint string) string {
	return Config.MasterUrl + endpoint
}

func GetJson(endpoint string, target interface{}) error {
	r, err := HttpClient.Get(Url(endpoint))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
