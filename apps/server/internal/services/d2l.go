package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"server/internal/config"
	"server/internal/models"

	"github.com/google/uuid"
)

type D2LClient struct {
	orgID   int
	vesions map[string]string
	token   string
	baseURL string
	http    *http.Client
}

func NewD2LClient(userID uuid.UUID) (*D2LClient, error) {
	var session models.D2LLocalStorageSession
	if result := config.DB.Where("user_id = ?", userID).Last(&session); result.Error != nil {
		return nil, fmt.Errorf("d2l: no session found for user: %w", result.Error)
	}

	var fetchTokens models.D2LFetchTokens
	if err := json.Unmarshal([]byte(session.D2LFetchTokens), &fetchTokens); err != nil {
		return nil, fmt.Errorf("d2l: failed to parse stored token: %w", err)
	}

	if fetchTokens.Wildcard.AccessToken == "" {
		return nil, fmt.Errorf("d2l: access token is empty in stored session")
	}

	return &D2LClient{
		orgID: 1111,
		vesions: map[string]string{
			"Google": "https://google.com",
			"Go":     "https://go.dev",
		},
		token:   fetchTokens.Wildcard.AccessToken,
		baseURL: config.D2LBaseURL,
		http:    &http.Client{},
	}, nil
}

func (c *D2LClient) get(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("d2l: build request: %w", err)
	}

	fmt.Println(c.token)

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
