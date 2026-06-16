package ai

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	db     *sql.DB
	client *http.Client
}

type Config struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

type ChatSession struct {
	ID        int64     `json:"id"`
	ArticleID int64     `json:"article_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Thinking  string    `json:"thinking,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:     db,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (s *Service) CreateSession(articleID int64, title string) (*ChatSession, error) {
	result, err := s.db.Exec(
		"INSERT INTO chat_sessions (article_id, title) VALUES (?, ?)",
		articleID, title,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	id, _ := result.LastInsertId()
	return &ChatSession{
		ID:        id,
		ArticleID: articleID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *Service) GetSessions(articleID int64) ([]ChatSession, error) {
	rows, err := s.db.Query(
		"SELECT id, article_id, title, created_at, updated_at FROM chat_sessions WHERE article_id = ? ORDER BY updated_at DESC",
		articleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ChatSession
	for rows.Next() {
		var session ChatSession
		if err := rows.Scan(&session.ID, &session.ArticleID, &session.Title, &session.CreatedAt, &session.UpdatedAt); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *Service) GetMessages(sessionID int64) ([]ChatMessage, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, role, content, COALESCE(thinking, ''), created_at FROM chat_messages WHERE session_id = ? ORDER BY created_at",
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Thinking, &msg.CreatedAt); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *Service) SendMessage(sessionID int64, config Config, articleContent, question string) (*ChatMessage, error) {
	// Save user message
	_, err := s.db.Exec(
		"INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)",
		sessionID, "user", question,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get chat history
	messages, err := s.GetMessages(sessionID)
	if err != nil {
		return nil, err
	}

	// Build API messages
	apiMessages := []map[string]string{
		{
			"role": "system",
			"content": fmt.Sprintf("You are a helpful assistant. The user is reading an article. Here is the article content:\n\n%s\n\nAnswer the user's questions based on this article.", articleContent),
		},
	}

	for _, msg := range messages {
		apiMessages = append(apiMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Call AI API
	response, err := s.callAPI(config, apiMessages)
	if err != nil {
		return nil, err
	}

	// Save assistant message
	result, err := s.db.Exec(
		"INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)",
		sessionID, "assistant", response,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// Update session
	s.db.Exec("UPDATE chat_sessions SET updated_at = ? WHERE id = ?", time.Now(), sessionID)

	id, _ := result.LastInsertId()
	return &ChatMessage{
		ID:        id,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   response,
		CreatedAt: time.Now(),
	}, nil
}

func (s *Service) callAPI(config Config, messages []map[string]string) (string, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	model := config.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	url := endpoint + "/chat/completions"

	payload := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call AI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	return content, nil
}

func (s *Service) TestConfig(config Config) (string, error) {
	messages := []map[string]string{
		{"role": "user", "content": "Hello, this is a test message. Please respond with 'Test successful'."},
	}

	return s.callAPI(config, messages)
}

func (s *Service) Search(config Config, query string, articles []map[string]string) (string, error) {
	// Build context from articles
	var context strings.Builder
	for i, article := range articles {
		context.WriteString(fmt.Sprintf("Article %d:\nTitle: %s\nContent: %s\n\n", i+1, article["title"], article["content"]))
	}

	messages := []map[string]string{
		{
			"role": "system",
			"content": "You are a helpful assistant. The user will provide a collection of articles and a search query. Find and summarize the most relevant articles based on the query.",
		},
		{
			"role": "user",
			"content": fmt.Sprintf("Here are some articles:\n\n%s\nSearch query: %s\n\nPlease find and summarize the most relevant articles.", context.String(), query),
		},
	}

	return s.callAPI(config, messages)
}
