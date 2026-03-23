package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"server/internal/config"
	"server/internal/models"

	"github.com/google/uuid"
)

type D2LVersions struct {
	LE *string
	LP *string
}

type D2LClient struct {
	orgID    string
	versions D2LVersions
	token    string
	baseURL  string
	http     *http.Client
}

func NewD2LClient(userID uuid.UUID) (*D2LClient, error) {
	var session models.D2LLocalStorageSession
	if result := config.DBClient.Where("user_id = ?", userID).Last(&session); result.Error != nil {
		return nil, fmt.Errorf("d2l: no session found for user: %w", result.Error)
	}

	if session.FetchAccessToken == "" {
		return nil, fmt.Errorf("d2l: access token is empty in stored session")
	}

	var user models.User
	if result := config.DBClient.Preload("Org").First(&user, "id = ?", userID); result.Error != nil {
		return nil, fmt.Errorf("d2l: no user found: %w", result.Error)
	}

	if user.Org == nil {
		return nil, fmt.Errorf("d2l: user has no associated org")
	}

	return &D2LClient{
		orgID:   user.Org.ID.String(),
		versions: D2LVersions{LE: user.Org.LEVersion, LP: user.Org.LPVersion},
		token:   session.FetchAccessToken,
		baseURL: user.Org.D2LBaseURL,
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
