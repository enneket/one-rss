package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zjx/one-rss/internal/models"
)

type Service struct {
	db     *sql.DB
	client *http.Client
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:     db,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Obsidian integration
type ObsidianConfig struct {
	VaultPath string `json:"vault_path"`
}

func (s *Service) ExportToObsidian(article *models.Article, content string, config ObsidianConfig) error {
	if config.VaultPath == "" {
		return fmt.Errorf("vault path is required")
	}

	// Create markdown content
	markdown := fmt.Sprintf(`---
title: %s
author: %s
source: %s
date: %s
url: %s
---

# %s

%s

---
*Exported from OneRSS*
`, article.Title, article.Author, article.FeedTitle, article.PublishedAt.Format(time.RFC3339), article.URL, article.Title, content)

	// In production, this would write to the vault path
	// For now, just log it
	fmt.Printf("Exporting to Obsidian: %s\n", markdown)

	return nil
}

// Notion integration
type NotionConfig struct {
	APIKey  string `json:"api_key"`
	PageID  string `json:"page_id"`
}

func (s *Service) ExportToNotion(article *models.Article, content string, config NotionConfig) error {
	if config.APIKey == "" || config.PageID == "" {
		return fmt.Errorf("api_key and page_id are required")
	}

	// Create Notion page
	url := "https://api.notion.com/v1/pages"

	payload := map[string]interface{}{
		"parent": map[string]string{
			"page_id": config.PageID,
		},
		"properties": map[string]interface{}{
			"title": []map[string]interface{}{
				{
					"text": map[string]string{
						"content": article.Title,
					},
				},
			},
		},
		"children": []map[string]interface{}{
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": content,
							},
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create Notion page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Notion API error: %s", string(body))
	}

	return nil
}

// Zotero integration
type ZoteroConfig struct {
	APIKey string `json:"api_key"`
	UserID string `json:"user_id"`
}

func (s *Service) ExportToZotero(article *models.Article, config ZoteroConfig) error {
	if config.APIKey == "" || config.UserID == "" {
		return fmt.Errorf("api_key and user_id are required")
	}

	url := fmt.Sprintf("https://api.zotero.org/users/%s/items", config.UserID)

	payload := map[string]interface{}{
		"itemType": "webpage",
		"title":    article.Title,
		"url":      article.URL,
		"creators": []map[string]string{
			{
				"creatorType": "author",
				"firstName":   article.Author,
			},
		},
		"date": article.PublishedAt.Format("2006-01-02"),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Zotero-API-Key", config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create Zotero item: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Zotero API error: %s", string(body))
	}

	return nil
}

// FreshRSS integration
type FreshRSSConfig struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Service) SyncWithFreshRSS(config FreshRSSConfig) error {
	if config.Server == "" || config.Username == "" || config.Password == "" {
		return fmt.Errorf("server, username, and password are required")
	}

	// Login to FreshRSS
	loginURL := config.Server + "/api/greader.php/accounts/ClientLogin"
	payload := fmt.Sprintf("Email=%s&Passwd=%s", config.Username, config.Password)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login to FreshRSS: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse auth token
	authToken := string(body)
	// In production, parse the actual token from the response

	fmt.Printf("FreshRSS auth token: %s\n", authToken)

	return nil
}

// RSSHub integration
type RSSHubConfig struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
}

func (s *Service) ValidateRSSHubRoute(config RSSHubConfig, route string) (bool, error) {
	if config.Endpoint == "" {
		return false, fmt.Errorf("endpoint is required")
	}

	url := config.Endpoint + route

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err
	}

	if config.APIKey != "" {
		req.Header.Set("X-API-Key", config.APIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
