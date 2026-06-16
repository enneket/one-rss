package scheduler

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/zjx/one-rss/internal/feed"
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

func TestNewScheduler(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fetcher := feed.NewFetcher(db)
	interval := 30 * time.Minute

	scheduler := NewScheduler(db, fetcher, interval)
	if scheduler == nil {
		t.Error("Expected non-nil scheduler")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fetcher := feed.NewFetcher(db)
	interval := 1 * time.Second

	scheduler := NewScheduler(db, fetcher, interval)

	// Start scheduler
	scheduler.Start()

	// Verify it's running
	scheduler.mu.Lock()
	if !scheduler.running {
		t.Error("Expected scheduler to be running")
	}
	scheduler.mu.Unlock()

	// Stop scheduler
	scheduler.Stop()

	// Verify it's stopped
	scheduler.mu.Lock()
	if scheduler.running {
		t.Error("Expected scheduler to be stopped")
	}
	scheduler.mu.Unlock()
}

func TestSchedulerDoubleStart(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fetcher := feed.NewFetcher(db)
	interval := 1 * time.Second

	scheduler := NewScheduler(db, fetcher, interval)

	// Start scheduler twice
	scheduler.Start()
	scheduler.Start() // Should not panic or create duplicate goroutines

	// Stop scheduler
	scheduler.Stop()
}

func TestSchedulerDoubleStop(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fetcher := feed.NewFetcher(db)
	interval := 1 * time.Second

	scheduler := NewScheduler(db, fetcher, interval)

	// Start and stop scheduler twice
	scheduler.Start()
	scheduler.Stop()
	scheduler.Stop() // Should not panic
}

func TestRefreshFeed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fetcher := feed.NewFetcher(db)
	interval := 30 * time.Minute

	scheduler := NewScheduler(db, fetcher, interval)

	// Insert a feed
	_, err := db.Exec("INSERT INTO feeds (id, title, url) VALUES (1, 'Test Feed', 'https://example.com/feed.xml')")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	// Try to refresh (will fail because URL doesn't exist, but should not panic)
	err = scheduler.RefreshFeed(1)
	// We expect an error because the URL doesn't exist
	if err == nil {
		t.Log("RefreshFeed completed (URL might be valid or cached)")
	}
}

func TestRefreshFeedNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fetcher := feed.NewFetcher(db)
	interval := 30 * time.Minute

	scheduler := NewScheduler(db, fetcher, interval)

	// Try to refresh non-existent feed
	err := scheduler.RefreshFeed(999)
	if err == nil {
		t.Error("Expected error for non-existent feed")
	}
}
