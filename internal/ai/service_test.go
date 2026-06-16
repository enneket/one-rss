package ai

import (
	"database/sql"
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
		`CREATE TABLE IF NOT EXISTS chat_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			thinking TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			content TEXT
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	// Insert test article
	_, err = db.Exec("INSERT INTO articles (id, title, content) VALUES (1, 'Test Article', 'Test content')")
	if err != nil {
		t.Fatalf("Failed to insert test article: %v", err)
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

func TestCreateSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	session, err := svc.CreateSession(1, "Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ArticleID != 1 {
		t.Errorf("Expected article ID 1, got %d", session.ArticleID)
	}

	if session.Title != "Test Session" {
		t.Errorf("Expected title 'Test Session', got '%s'", session.Title)
	}
}

func TestGetSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	// Create sessions
	_, err := svc.CreateSession(1, "Session 1")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	_, err = svc.CreateSession(1, "Session 2")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get sessions
	sessions, err := svc.GetSessions(1)
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
}

func TestGetMessages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db)

	// Create session
	session, err := svc.CreateSession(1, "Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Insert messages
	_, err = db.Exec("INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)",
		session.ID, "user", "Hello")
	if err != nil {
		t.Fatalf("Failed to insert message: %v", err)
	}

	_, err = db.Exec("INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)",
		session.ID, "assistant", "Hi there!")
	if err != nil {
		t.Fatalf("Failed to insert message: %v", err)
	}

	// Get messages
	messages, err := svc.GetMessages(session.ID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "user" {
		t.Errorf("Expected first message role 'user', got '%s'", messages[0].Role)
	}

	if messages[1].Role != "assistant" {
		t.Errorf("Expected second message role 'assistant', got '%s'", messages[1].Role)
	}
}
