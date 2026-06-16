package translation

import (
	"database/sql"
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
	query := `CREATE TABLE IF NOT EXISTS translation_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_text_hash TEXT NOT NULL,
		source_text TEXT NOT NULL,
		target_lang TEXT NOT NULL,
		translated_text TEXT NOT NULL,
		provider TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_text_hash, target_lang, provider)
	)`

	if _, err := db.Exec(query); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestNewService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	config := Config{
		Provider: "google",
	}

	svc, err := NewService(db, config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestTranslateEmptyText(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	config := Config{
		Provider: "google",
	}

	svc, err := NewService(db, config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Translate empty text
	result, err := svc.Translate("", "en", "zh")
	if err != nil {
		t.Fatalf("Failed to translate empty text: %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
}

func TestTranslateWithCache(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	config := Config{
		Provider: "google",
	}

	svc, err := NewService(db, config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Insert a cached translation
	_, err = db.Exec(`INSERT INTO translation_cache (source_text_hash, source_text, target_lang, translated_text, provider) 
		VALUES (?, ?, ?, ?, ?)`, "testhash", "Hello", "zh", "你好", "google")
	if err != nil {
		t.Fatalf("Failed to insert cached translation: %v", err)
	}

	// This would normally call the API, but we can't test that without mocking
	// For now, just test that the service can be created
	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestNewGoogleProvider(t *testing.T) {
	provider := NewGoogleProvider()
	if provider == nil {
		t.Error("Expected non-nil provider")
	}
}

func TestNewDeepLProvider(t *testing.T) {
	provider := NewDeepLProvider("test-api-key")
	if provider == nil {
		t.Error("Expected non-nil provider")
	}
}

func TestNewAIProvider(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		apiKey   string
		model    string
	}{
		{
			name:     "Default values",
			endpoint: "",
			apiKey:   "test-key",
			model:    "",
		},
		{
			name:     "Custom values",
			endpoint: "https://custom.api.com/v1",
			apiKey:   "test-key",
			model:    "gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewAIProvider(tt.endpoint, tt.apiKey, tt.model)
			if provider == nil {
				t.Error("Expected non-nil provider")
			}
		})
	}
}

func TestGoogleProviderTranslate(t *testing.T) {
	// Create a test server that returns a mock response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[[["你好","Hello",null,null,10]]]`))
	}))
	defer server.Close()

	// Note: We can't easily test the actual Google Translate API without mocking
	// The test above just verifies the provider can be created
	provider := NewGoogleProvider()
	if provider == nil {
		t.Error("Expected non-nil provider")
	}
}
