package database

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Use a temporary database for testing
	os.Setenv("DB_PATH", ":memory:")
	defer os.Unsetenv("DB_PATH")

	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test that tables were created
	tables := []string{
		"feeds",
		"articles",
		"translation_cache",
		"chat_sessions",
		"chat_messages",
		"settings",
		"saved_filters",
		"tags",
		"feed_tags",
		"ai_profiles",
	}

	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("Table %s not created: %v", table, err)
		}
	}
}

func TestCreateTables(t *testing.T) {
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test inserting a feed
	result, err := db.Exec(`
		INSERT INTO feeds (title, url, category) VALUES (?, ?, ?)
	`, "Test Feed", "https://example.com/feed.xml", "test")
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	id, _ := result.LastInsertId()
	if id == 0 {
		t.Error("Expected non-zero feed ID")
	}

	// Test inserting an article
	result, err = db.Exec(`
		INSERT INTO articles (feed_id, title, url, unique_id) VALUES (?, ?, ?, ?)
	`, id, "Test Article", "https://example.com/article1", "unique1")
	if err != nil {
		t.Fatalf("Failed to insert article: %v", err)
	}

	articleID, _ := result.LastInsertId()
	if articleID == 0 {
		t.Error("Expected non-zero article ID")
	}

	// Test cascade delete
	_, err = db.Exec("DELETE FROM feeds WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to delete feed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles WHERE feed_id = ?", id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count articles: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 articles after cascade delete, got %d", count)
	}
}
