package integration

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

func TestExportToObsidian(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	article := &models.Article{
		ID:          1,
		Title:       "Test Article",
		Author:      "Test Author",
		FeedTitle:   "Test Feed",
		URL:         "https://example.com/article",
		PublishedAt: time.Now(),
	}

	config := ObsidianConfig{
		VaultPath: "/tmp/test-vault",
	}

	// This should not error (even though path doesn't exist, it just logs)
	err := svc.ExportToObsidian(article, "Test content", config)
	if err != nil {
		t.Fatalf("Failed to export to Obsidian: %v", err)
	}
}

func TestExportToNotion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	article := &models.Article{
		ID:    1,
		Title: "Test Article",
	}

	config := NotionConfig{
		APIKey: "",
		PageID: "",
	}

	// Should fail with empty config
	err := svc.ExportToNotion(article, "Test content", config)
	if err == nil {
		t.Error("Expected error with empty config")
	}
}

func TestExportToZotero(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	article := &models.Article{
		ID:    1,
		Title: "Test Article",
	}

	config := ZoteroConfig{
		APIKey: "",
		UserID: "",
	}

	// Should fail with empty config
	err := svc.ExportToZotero(article, config)
	if err == nil {
		t.Error("Expected error with empty config")
	}
}

func TestSyncWithFreshRSS(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	config := FreshRSSConfig{
		Server:   "",
		Username: "",
		Password: "",
	}

	// Should fail with empty config
	err := svc.SyncWithFreshRSS(config)
	if err == nil {
		t.Error("Expected error with empty config")
	}
}

func TestValidateRSSHubRoute(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	config := RSSHubConfig{
		Endpoint: server.URL,
	}

	// Test valid route
	valid, err := svc.ValidateRSSHubRoute(config, "/test")
	if err != nil {
		t.Fatalf("Failed to validate route: %v", err)
	}

	if !valid {
		t.Error("Expected route to be valid")
	}
}
