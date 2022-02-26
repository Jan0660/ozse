package config

import (
	"encoding/json"
	"net/http"
)

func Url(endpoint string) string {
	return Config.MasterUrl + endpoint
}

func GetJson(endpoint string, target interface{}) error {
	// todo: use a http client with a timeout
	r, err := http.Get(Url(endpoint))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
