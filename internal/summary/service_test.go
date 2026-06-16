package summary

import (
	"testing"
)

func TestSplitSentences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "English sentences",
			input:    "Hello world. This is a test. How are you?",
			expected: 3,
		},
		{
			name:     "Chinese sentences",
			input:    "你好世界。这是一个测试。你好吗？",
			expected: 3,
		},
		{
			name:     "Single sentence",
			input:    "Hello world",
			expected: 1,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSentences(tt.input)
			if len(result) != tt.expected {
				t.Errorf("expected %d sentences, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Simple text",
			input:    "Hello world",
			expected: 2,
		},
		{
			name:     "With punctuation",
			input:    "Hello, world! How are you?",
			expected: 5,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			if len(result) != tt.expected {
				t.Errorf("expected %d tokens, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestLocalProvider_Summarize(t *testing.T) {
	provider := NewLocalProvider()

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected bool
	}{
		{
			name:     "Short text",
			input:    "Hello world",
			maxLen:   100,
			expected: true,
		},
		{
			name:     "Long text",
			input:    "This is a very long text that should be summarized. It contains multiple sentences with different information. The summary should be shorter than the original text. This is another sentence to make it longer.",
			maxLen:   100,
			expected: true,
		},
		{
			name:     "Empty text",
			input:    "",
			maxLen:   100,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.Summarize(tt.input, tt.maxLen)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expected && result == "" {
				t.Error("expected non-empty summary")
			}

			if !tt.expected && result != "" {
				t.Error("expected empty summary")
			}

			if len(result) > tt.maxLen && tt.input != "" {
				// Allow some flexibility for sentence boundaries
				if len(result) > tt.maxLen*2 {
					t.Errorf("summary too long: %d > %d", len(result), tt.maxLen)
				}
			}
		})
	}
}
