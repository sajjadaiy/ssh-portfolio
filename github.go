package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type RepoData struct {
	Stars    int
	Language string
	PushedAt time.Time
}

type repoCache struct {
	mu   sync.RWMutex
	data map[string]RepoData
}

var cache = &repoCache{data: make(map[string]RepoData)}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func (c *repoCache) get(repo string) (RepoData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	d, ok := c.data[repo]
	return d, ok
}

func (c *repoCache) set(repo string, d RepoData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[repo] = d
}

type githubRepo struct {
	StargazersCount int    `json:"stargazers_count"`
	Language        string `json:"language"`
	PushedAt        string `json:"pushed_at"`
}

func fetchRepo(repo string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s", repo), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ssh-portfolio/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github API status %d", resp.StatusCode)
	}

	var gr githubRepo
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return err
	}

	pushed, _ := time.Parse(time.RFC3339, gr.PushedAt)
	cache.set(repo, RepoData{
		Stars:    gr.StargazersCount,
		Language: gr.Language,
		PushedAt: pushed,
	})
	return nil
}

func refreshGitHub() {
	for _, item := range items {
		if item.Repo != "" {
			_ = fetchRepo(item.Repo)
		}
	}
}

func startGitHubRefresh() {
	refreshGitHub()
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		for range ticker.C {
			refreshGitHub()
		}
	}()
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < 24*time.Hour:
		return "today"
	case d < 48*time.Hour:
		return "yesterday"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
}
