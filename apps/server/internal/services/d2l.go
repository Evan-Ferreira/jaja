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
	orgID   string
	vesions map[string]string
	token   string
	baseURL string
	http    *http.Client
}

func NewD2LClient(userID uuid.UUID) (*D2LClient, error) {
	var session models.D2LLocalStorageSession
	if result := config.DBClient.Where("user_id = ?", userID).Last(&session); result.Error != nil {
		return nil, fmt.Errorf("d2l: no session found for user: %w", result.Error)
	}

	if session.FetchAccessToken == "" {
		return nil, fmt.Errorf("d2l: access token is empty in stored session")
	}

	return &D2LClient{
		orgID: "111111", // TODO: store real org ID in DB
		vesions: map[string]string{ // TODO: store real API versions in DB or fetch from D2L
			"le": "1.67",
			"lp": "1.30",
		},
		token:   session.FetchAccessToken,
		baseURL: config.D2LBaseURL,
		http:    &http.Client{},
	}, nil
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
