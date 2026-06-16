package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
			script_path TEXT DEFAULT '',
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
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			encrypted BOOLEAN DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			color TEXT DEFAULT '#3b82f6',
			position INTEGER DEFAULT 0
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	return db
}

func TestHandleHealth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleHealth(db)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestHandleFeeds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleFeeds(db)

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Test POST (should fail)
	req = httptest.NewRequest(http.MethodPost, "/api/feeds", nil)
	w = httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleAddFeed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleAddFeed(db)

	// Test adding a feed
	body := map[string]string{
		"url":      "https://example.com/feed.xml",
		"category": "test",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/feeds/add", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Feed added successfully" {
		t.Errorf("Expected message 'Feed added successfully', got '%s'", response["message"])
	}
}

func TestHandleDeleteFeed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a feed
	result, err := db.Exec("INSERT INTO feeds (title, url) VALUES (?, ?)", "Test Feed", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}
	feedID, _ := result.LastInsertId()

	handler := handleDeleteFeed(db)

	// Test deleting the feed
	body := map[string]int64{
		"id": feedID,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/feeds/delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleArticles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleArticles(db)

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleMarkRead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a feed and article
	_, err := db.Exec("INSERT INTO feeds (id, title, url) VALUES (1, 'Test Feed', 'https://example.com/feed.xml')")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	_, err = db.Exec("INSERT INTO articles (id, feed_id, title, url, unique_id) VALUES (1, 1, 'Test Article', 'https://example.com/article', 'unique1')")
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	handler := handleMarkRead(db)

	// Test marking as read
	body := map[string]interface{}{
		"id":      1,
		"is_read": true,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/articles/read", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleToggleFavorite(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a feed and article
	_, err := db.Exec("INSERT INTO feeds (id, title, url) VALUES (1, 'Test Feed', 'https://example.com/feed.xml')")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	_, err = db.Exec("INSERT INTO articles (id, feed_id, title, url, unique_id) VALUES (1, 1, 'Test Article', 'https://example.com/article', 'unique1')")
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	handler := handleToggleFavorite(db)

	// Test toggling favorite
	body := map[string]int64{
		"id": 1,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/articles/favorite", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleUnreadCounts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleUnreadCounts(db)

	req := httptest.NewRequest(http.MethodGet, "/api/articles/unread-counts", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleSettings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleSettings(db)

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Test POST
	body := map[string]string{
		"theme": "dark",
	}
	jsonBody, _ := json.Marshal(body)

	req = httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleTags(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleTags(db)

	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleRefresh(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleRefresh(db)

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleProgress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handleProgress(db)

	req := httptest.NewRequest(http.MethodGet, "/api/progress", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
