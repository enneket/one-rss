package feed

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/zjx/one-rss/internal/models"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	os.Setenv("DB_PATH", ":memory:")
	defer os.Unsetenv("DB_PATH")

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS feeds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			url TEXT UNIQUE,
			link TEXT DEFAULT '',
			description TEXT,
			category TEXT DEFAULT '',
			image_url TEXT DEFAULT '',
			position INTEGER DEFAULT 0,
			last_updated DATETIME,
			last_error TEXT DEFAULT '',
			hide_from_timeline BOOLEAN DEFAULT 0,
			proxy_url TEXT DEFAULT '',
			proxy_enabled BOOLEAN DEFAULT 0,
			refresh_interval INTEGER DEFAULT 0,
			is_image_mode BOOLEAN DEFAULT 0,
			type TEXT DEFAULT '',
			article_view_mode TEXT DEFAULT 'global',
			auto_expand_content TEXT DEFAULT 'global'
		)`,
		`CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			feed_id INTEGER,
			title TEXT,
			url TEXT,
			content TEXT DEFAULT '',
			image_url TEXT DEFAULT '',
			audio_url TEXT DEFAULT '',
			video_url TEXT DEFAULT '',
			translated_title TEXT DEFAULT '',
			published_at DATETIME,
			is_read BOOLEAN DEFAULT 0,
			is_favorite BOOLEAN DEFAULT 0,
			is_hidden BOOLEAN DEFAULT 0,
			is_read_later BOOLEAN DEFAULT 0,
			summary TEXT DEFAULT '',
			unique_id TEXT UNIQUE,
			author TEXT DEFAULT '',
			FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	return db
}

func TestNewFetcher(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fetcher := NewFetcher(db)
	if fetcher == nil {
		t.Error("Expected non-nil fetcher")
	}
}

func TestSaveArticles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fetcher := NewFetcher(db)

	// Insert a feed
	result, err := db.Exec("INSERT INTO feeds (title, url) VALUES (?, ?)", "Test Feed", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}
	feedID, _ := result.LastInsertId()

	// Create test articles
	items := []FeedItem{
		{
			Title:    "Article 1",
			URL:      "https://example.com/article1",
			Content:  "Content 1",
			Author:   "Author 1",
			UniqueID: "unique1",
		},
		{
			Title:    "Article 2",
			URL:      "https://example.com/article2",
			Content:  "Content 2",
			Author:   "Author 2",
			UniqueID: "unique2",
		},
	}

	// Save articles
	err = fetcher.SaveArticles(feedID, items)
	if err != nil {
		t.Fatalf("Failed to save articles: %v", err)
	}

	// Verify articles were saved
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles WHERE feed_id = ?", feedID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count articles: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 articles, got %d", count)
	}
}

func TestFetchFeed(t *testing.T) {
	// Create a test RSS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <description>A test feed</description>
    <item>
      <title>Article 1</title>
      <link>https://example.com/article1</link>
      <description>Content 1</description>
      <pubDate>Mon, 15 Jun 2026 00:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>`))
	}))
	defer server.Close()

	db := setupTestDB(t)
	defer db.Close()

	fetcher := NewFetcher(db)

	// Create a feed with the test server URL
	feed := &models.Feed{
		ID:  1,
		URL: server.URL,
	}

	items, err := fetcher.FetchFeed(feed)
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}

	if items[0].Title != "Article 1" {
		t.Errorf("Expected title 'Article 1', got '%s'", items[0].Title)
	}
}

func TestFetchAndSave(t *testing.T) {
	// Create a test RSS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <description>A test feed</description>
    <item>
      <title>Article 1</title>
      <link>https://example.com/article1</link>
      <description>Content 1</description>
    </item>
  </channel>
</rss>`))
	}))
	defer server.Close()

	db := setupTestDB(t)
	defer db.Close()

	fetcher := NewFetcher(db)

	// Insert a feed
	result, err := db.Exec("INSERT INTO feeds (title, url) VALUES (?, ?)", "Test Feed", server.URL)
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}
	feedID, _ := result.LastInsertId()

	feed := &models.Feed{
		ID:  feedID,
		URL: server.URL,
	}

	err = fetcher.FetchAndSave(feed)
	if err != nil {
		t.Fatalf("Failed to fetch and save: %v", err)
	}

	// Verify article was saved
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles WHERE feed_id = ?", feedID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count articles: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 article, got %d", count)
	}
}

func TestDiscoverFeeds(t *testing.T) {
	// Create a test server with RSS link in HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
<head>
  <link rel="alternate" type="application/rss+xml" href="/feed.xml" />
</head>
<body>Test</body>
</html>`))
	}))
	defer server.Close()

	db := setupTestDB(t)
	defer db.Close()

	fetcher := NewFetcher(db)

	feeds, err := fetcher.DiscoverFeeds(server.URL)
	if err != nil {
		t.Fatalf("Failed to discover feeds: %v", err)
	}

	if len(feeds) != 1 {
		t.Errorf("Expected 1 feed, got %d", len(feeds))
	}
}
