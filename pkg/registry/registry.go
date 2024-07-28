// Package registry provides helpers for interacting with Go module registries.
package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is an HTTP client for Go registry operations.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new registry client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GitHubRepo represents a GitHub repository search result.
type GitHubRepo struct {
	FullName    string   `json:"full_name"`
	HTMLURL     string   `json:"html_url"`
	Description string   `json:"description"`
	Stars       int      `json:"stargazers_count"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	UpdatedAt   string   `json:"updated_at"`
}

// SearchGitHub searches GitHub for Go repositories.
func (c *Client) SearchGitHub(query string, limit int, sort string) ([]GitHubRepo, error) {
	u, _ := url.Parse("https://api.github.com/search/repositories")
	q := u.Query()
	q.Set("q", query+" language:go")
	if sort != "" {
		q.Set("sort", sort)
	} else {
		q.Set("sort", "stars")
	}
	q.Set("per_page", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("github request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "mygo")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	var result struct {
		Items []GitHubRepo `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode github response: %w", err)
	}
	return result.Items, nil
}
