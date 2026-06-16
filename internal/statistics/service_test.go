package statistics

import (
	"database/sql"
	"os"
	"testing"
	"time"

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
			category TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			feed_id INTEGER,
			title TEXT,
			url TEXT,
			published_at DATETIME,
			is_read BOOLEAN DEFAULT 0,
			is_favorite BOOLEAN DEFAULT 0,
			is_hidden BOOLEAN DEFAULT 0,
			is_read_later BOOLEAN DEFAULT 0,
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

func TestNewService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestGetStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	// Insert test data
	_, err := db.Exec("INSERT INTO feeds (title, url, category) VALUES (?, ?, ?)", "Feed 1", "https://example.com/feed1", "tech")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	_, err = db.Exec("INSERT INTO feeds (title, url, category) VALUES (?, ?, ?)", "Feed 2", "https://example.com/feed2", "news")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite) 
		VALUES (1, 'Article 1', 'https://example.com/article1', ?, 1, 1)`, time.Now())
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite) 
		VALUES (1, 'Article 2', 'https://example.com/article2', ?, 0, 0)`, time.Now())
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite) 
		VALUES (2, 'Article 3', 'https://example.com/article3', ?, 1, 0)`, time.Now())
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	// Get stats
	stats, err := svc.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalArticles != 3 {
		t.Errorf("Expected 3 total articles, got %d", stats.TotalArticles)
	}

	if stats.ReadArticles != 2 {
		t.Errorf("Expected 2 read articles, got %d", stats.ReadArticles)
	}

	if stats.FavoriteArticles != 1 {
		t.Errorf("Expected 1 favorite article, got %d", stats.FavoriteArticles)
	}

	if stats.TotalFeeds != 2 {
		t.Errorf("Expected 2 total feeds, got %d", stats.TotalFeeds)
	}
}

func TestGetUnreadCounts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	// Insert test data
	_, err := db.Exec("INSERT INTO feeds (title, url) VALUES (?, ?)", "Feed 1", "https://example.com/feed1")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	_, err = db.Exec("INSERT INTO feeds (title, url) VALUES (?, ?)", "Feed 2", "https://example.com/feed2")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, is_read) VALUES (1, 'Article 1', 'https://example.com/article1', 0)`)
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, is_read) VALUES (1, 'Article 2', 'https://example.com/article2', 0)`)
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, is_read) VALUES (2, 'Article 3', 'https://example.com/article3', 0)`)
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	// Get unread counts
	counts, err := svc.GetUnreadCounts()
	if err != nil {
		t.Fatalf("Failed to get unread counts: %v", err)
	}

	if counts[1] != 2 {
		t.Errorf("Expected 2 unread for feed 1, got %d", counts[1])
	}

	if counts[2] != 1 {
		t.Errorf("Expected 1 unread for feed 2, got %d", counts[2])
	}
}

func TestGetTotalUnread(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	// Insert test data
	_, err := db.Exec("INSERT INTO feeds (title, url) VALUES (?, ?)", "Feed 1", "https://example.com/feed1")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, is_read) VALUES (1, 'Article 1', 'https://example.com/article1', 0)`)
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, is_read) VALUES (1, 'Article 2', 'https://example.com/article2', 1)`)
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	// Get total unread
	total, err := svc.GetTotalUnread()
	if err != nil {
		t.Fatalf("Failed to get total unread: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 total unread, got %d", total)
	}
}
