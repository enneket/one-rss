package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = getDBPath()
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	DB = db
	return db, nil
}

func getDBPath() string {
	// Check for portable mode
	if _, err := os.Stat("portable.txt"); err == nil {
		return "data/one-rss.db"
	}

	// Use current directory for development
	return "one-rss.db"
}

func createTables(db *sql.DB) error {
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
			discovery_completed BOOLEAN DEFAULT 0,
			script_path TEXT DEFAULT '',
			hide_from_timeline BOOLEAN DEFAULT 0,
			proxy_url TEXT DEFAULT '',
			proxy_enabled BOOLEAN DEFAULT 0,
			refresh_interval INTEGER DEFAULT 0,
			is_image_mode BOOLEAN DEFAULT 0,
			type TEXT DEFAULT '',
			xpath_item TEXT DEFAULT '',
			xpath_item_title TEXT DEFAULT '',
			xpath_item_content TEXT DEFAULT '',
			xpath_item_uri TEXT DEFAULT '',
			xpath_item_author TEXT DEFAULT '',
			xpath_item_timestamp TEXT DEFAULT '',
			xpath_item_time_format TEXT DEFAULT '',
			xpath_item_thumbnail TEXT DEFAULT '',
			xpath_item_categories TEXT DEFAULT '',
			xpath_item_uid TEXT DEFAULT '',
			article_view_mode TEXT DEFAULT 'global',
			auto_expand_content TEXT DEFAULT 'global',
			email_address TEXT DEFAULT '',
			email_imap_server TEXT DEFAULT '',
			email_imap_port INTEGER DEFAULT 993,
			email_username TEXT DEFAULT '',
			email_password TEXT DEFAULT '',
			email_folder TEXT DEFAULT 'INBOX',
			email_last_uid INTEGER DEFAULT 0,
			is_freshrss_source BOOLEAN DEFAULT 0,
			freshrss_stream_id TEXT DEFAULT '',
			latest_article_time DATETIME,
			articles_per_month REAL DEFAULT 0,
			last_update_status TEXT DEFAULT ''
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
			freshrss_item_id TEXT DEFAULT '',
			FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS article_contents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_id INTEGER NOT NULL UNIQUE,
			content TEXT NOT NULL,
			fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS translation_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_text_hash TEXT NOT NULL,
			source_text TEXT NOT NULL,
			target_lang TEXT NOT NULL,
			translated_text TEXT NOT NULL,
			provider TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source_text_hash, target_lang, provider)
		)`,
		`CREATE TABLE IF NOT EXISTS chat_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
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
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			encrypted BOOLEAN DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS saved_filters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			conditions TEXT NOT NULL,
			position INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			color TEXT DEFAULT '#3b82f6',
			position INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS feed_tags (
			feed_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (feed_id, tag_id),
			FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
			FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS ai_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			api_key TEXT DEFAULT '',
			endpoint TEXT DEFAULT '',
			model TEXT DEFAULT '',
			custom_headers TEXT DEFAULT '{}',
			is_default BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_articles_feed_id ON articles(feed_id)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_is_read ON articles(is_read)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_is_favorite ON articles(is_favorite)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_is_hidden ON articles(is_hidden)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_is_read_later ON articles(is_read_later)`,
		`CREATE INDEX IF NOT EXISTS idx_feeds_category ON feeds(category)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_feed_published ON articles(feed_id, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_read_published ON articles(is_read, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_fav_published ON articles(is_favorite, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_readlater_published ON articles(is_read_later, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_hidden_published ON articles(is_hidden, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_translation_cache_lookup ON translation_cache(source_text_hash, target_lang, provider)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_sessions_article_id ON chat_sessions(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated_at ON chat_sessions(updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_session_id ON chat_messages(session_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w\nQuery: %s", err, query)
		}
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
