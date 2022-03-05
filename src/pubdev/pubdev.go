package pubdev

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Client struct {
	http *http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{
		http: client,
	}
}

var defaultClient = NewClient(http.DefaultClient)

func DefaultClient() *Client {
	return defaultClient
}

func (c *Client) GetPackage(name string) (*Package, error) {
	resp, err := c.http.Get("https://pub.dev/api/packages/" + name)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New(resp.Status)
	}
	var pkg Package
	err = json.NewDecoder(resp.Body).Decode(&pkg)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}
