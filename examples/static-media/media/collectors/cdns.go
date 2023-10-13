package collectors

import (
	"encoding/json"
	"net/http"

	"golang.org/x/exp/slices"
)

type FlowerCDN struct {
	*http.Client
	URL string
}

func (c *FlowerCDN) Prepare(r *http.Request) (string, error) {
	return c.URL, nil
}

func (c *FlowerCDN) Collect(r *http.Request) (json.RawMessage, error) {
	resp, err := c.Client.Get(c.URL)
	if err != nil {
		return nil, err
	}

	var v []string
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	slices.Sort(v)

	return json.Marshal(map[string][]string{
		"flowers": v,
	})
}
