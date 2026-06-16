package translation

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Provider interface {
	Translate(text, sourceLang, targetLang string) (string, error)
}

type Service struct {
	db       *sql.DB
	provider Provider
}

type Config struct {
	Provider   string `json:"provider"`
	APIKey     string `json:"api_key"`
	Endpoint   string `json:"endpoint"`
	Model      string `json:"model"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

func NewService(db *sql.DB, config Config) (*Service, error) {
	var provider Provider
	switch config.Provider {
	case "google":
		provider = NewGoogleProvider()
	case "deepl":
		provider = NewDeepLProvider(config.APIKey)
	case "ai":
		provider = NewAIProvider(config.Endpoint, config.APIKey, config.Model)
	default:
		provider = NewGoogleProvider()
	}

	return &Service{
		db:       db,
		provider: provider,
	}, nil
}

func (s *Service) Translate(text, sourceLang, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Check cache
	hash := fmt.Sprintf("%x", md5.Sum([]byte(text)))
	var cached string
	err := s.db.QueryRow(
		"SELECT translated_text FROM translation_cache WHERE source_text_hash = ? AND target_lang = ? AND provider = ?",
		hash, targetLang, "",
	).Scan(&cached)
	if err == nil {
		return cached, nil
	}

	// Translate
	translated, err := s.provider.Translate(text, sourceLang, targetLang)
	if err != nil {
		return "", err
	}

	// Cache result
	_, err = s.db.Exec(
		"INSERT OR REPLACE INTO translation_cache (source_text_hash, source_text, target_lang, translated_text, provider) VALUES (?, ?, ?, ?, ?)",
		hash, text, targetLang, translated, "",
	)
	if err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to cache translation: %v\n", err)
	}

	return translated, nil
}

// GoogleProvider uses Google Translate API (free tier)
type GoogleProvider struct {
	client *http.Client
}

func NewGoogleProvider() *GoogleProvider {
	return &GoogleProvider{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *GoogleProvider) Translate(text, sourceLang, targetLang string) (string, error) {
	url := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&q=%s",
		sourceLang, targetLang, text,
	)

	resp, err := p.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to call Google Translate: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("empty translation result")
	}

	// Extract translated text
	translations, ok := result[0].([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	var translated string
	for _, t := range translations {
		if parts, ok := t.([]interface{}); ok && len(parts) > 0 {
			if str, ok := parts[0].(string); ok {
				translated += str
			}
		}
	}

	return translated, nil
}

// DeepLProvider uses DeepL API
type DeepLProvider struct {
	apiKey string
	client *http.Client
}

func NewDeepLProvider(apiKey string) *DeepLProvider {
	return &DeepLProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *DeepLProvider) Translate(text, sourceLang, targetLang string) (string, error) {
	url := "https://api-free.deepl.com/v2/translate"

	payload := map[string]interface{}{
		"text":       []string{text},
		"source_lang": sourceLang,
		"target_lang": targetLang,
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
	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call DeepL API: %w", err)
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

	translations, ok := result["translations"].([]interface{})
	if !ok || len(translations) == 0 {
		return "", fmt.Errorf("no translations returned")
	}

	translation, ok := translations[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid translation format")
	}

	textResult, ok := translation["text"].(string)
	if !ok {
		return "", fmt.Errorf("invalid text format")
	}

	return textResult, nil
}

// AIProvider uses OpenAI-compatible API
type AIProvider struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

func NewAIProvider(endpoint, apiKey, model string) *AIProvider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	return &AIProvider{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *AIProvider) Translate(text, sourceLang, targetLang string) (string, error) {
	url := p.endpoint + "/chat/completions"

	prompt := fmt.Sprintf("Translate the following text from %s to %s. Only return the translated text, nothing else.\n\n%s", sourceLang, targetLang, text)

	payload := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
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
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
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
