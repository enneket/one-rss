package summary

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"
)

type Provider interface {
	Summarize(text string, maxLen int) (string, error)
}

type Service struct {
	db       *sql.DB
	provider Provider
}

type Config struct {
	Provider string `json:"provider"` // "local" or "ai"
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
	MaxLen   int    `json:"max_len"`
}

func NewService(db *sql.DB, config Config) (*Service, error) {
	var provider Provider
	switch config.Provider {
	case "local":
		provider = NewLocalProvider()
	case "ai":
		provider = NewAIProvider(config.Endpoint, config.APIKey, config.Model)
	default:
		provider = NewLocalProvider()
	}

	maxLen := config.MaxLen
	if maxLen == 0 {
		maxLen = 200
	}

	return &Service{
		db:       db,
		provider: provider,
	}, nil
}

func (s *Service) Summarize(text string, maxLen int) (string, error) {
	if text == "" {
		return "", nil
	}

	if maxLen == 0 {
		maxLen = 200
	}

	return s.provider.Summarize(text, maxLen)
}

// LocalProvider uses TF-IDF + TextRank algorithm
type LocalProvider struct{}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{}
}

func (p *LocalProvider) Summarize(text string, maxLen int) (string, error) {
	// Split text into sentences
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return "", nil
	}

	// If text is short enough, return it
	if len(text) <= maxLen {
		return text, nil
	}

	// Calculate word frequency (TF)
	wordFreq := calculateWordFrequency(text)

	// Score sentences
	sentenceScores := scoreSentences(sentences, wordFreq)

	// Sort by score
	type sentenceScore struct {
		index int
		score float64
		text  string
	}

	scored := make([]sentenceScore, len(sentences))
	for i, score := range sentenceScores {
		scored[i] = sentenceScore{
			index: i,
			score: score,
			text:  sentences[i],
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Select top sentences
	var summary strings.Builder
	selectedIndices := []int{}

	for _, ss := range scored {
		if summary.Len()+len(ss.text) > maxLen {
			break
		}
		selectedIndices = append(selectedIndices, ss.index)
		summary.WriteString(ss.text)
	}

	// Sort by original order
	sort.Ints(selectedIndices)

	var result strings.Builder
	for _, idx := range selectedIndices {
		result.WriteString(sentences[idx])
	}

	return result.String(), nil
}

func splitSentences(text string) []string {
	// Split by common sentence delimiters
	var sentences []string
	var current strings.Builder

	for _, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '!' || r == '?' || r == '。' || r == '！' || r == '？' {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}

	// Add remaining text
	if current.Len() > 0 {
		s := strings.TrimSpace(current.String())
		if s != "" {
			sentences = append(sentences, s)
		}
	}

	return sentences
}

func calculateWordFrequency(text string) map[string]float64 {
	words := tokenize(text)
	freq := make(map[string]float64)

	for _, word := range words {
		freq[word]++
	}

	// Normalize
	maxFreq := 0.0
	for _, f := range freq {
		if f > maxFreq {
			maxFreq = f
		}
	}

	if maxFreq > 0 {
		for word := range freq {
			freq[word] /= maxFreq
		}
	}

	return freq
}

func tokenize(text string) []string {
	var words []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			word := strings.ToLower(current.String())
			if len(word) > 1 { // Skip single characters
				words = append(words, word)
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		word := strings.ToLower(current.String())
		if len(word) > 1 {
			words = append(words, word)
		}
	}

	return words
}

func scoreSentences(sentences []string, wordFreq map[string]float64) []float64 {
	scores := make([]float64, len(sentences))

	for i, sentence := range sentences {
		words := tokenize(sentence)
		if len(words) == 0 {
			continue
		}

		score := 0.0
		for _, word := range words {
			score += wordFreq[word]
		}
		scores[i] = score / float64(len(words))
	}

	return scores
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

func (p *AIProvider) Summarize(text string, maxLen int) (string, error) {
	url := p.endpoint + "/chat/completions"

	prompt := fmt.Sprintf("Summarize the following text in about %d characters. Only return the summary, nothing else.\n\n%s", maxLen, text)

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
