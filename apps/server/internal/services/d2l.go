package services

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type D2LClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewD2LClient(token, baseURL string) *D2LClient {
	return &D2LClient{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

func (c *D2LClient) get(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("d2l: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("d2l: request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("d2l: unexpected status %d for %s", res.StatusCode, path)
	}

	return json.NewDecoder(res.Body).Decode(out)
}

type WhoAmI struct {
	Identifier        string `json:"Identifier"`
	FirstName         string `json:"FirstName"`
	LastName          string `json:"LastName"`
	UniqueName        string `json:"UniqueName"`
	ExternalEmail     string `json:"ExternalEmail"`
	OrgDefinedId      string `json:"OrgDefinedId"`
	ProfileIdentifier string `json:"ProfileIdentifier"`
}

func (c *D2LClient) GetWhoAmI() (*WhoAmI, error) {
	var out WhoAmI
	if err := c.get("/d2l/api/lp/1.30/users/whoami", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
