package rules

import (
	"testing"

	"github.com/zjx/one-rss/internal/models"
)

func TestEvaluateCondition(t *testing.T) {
	engine := &Engine{}

	article := &models.Article{
		ID:        1,
		Title:     "Test Article",
		Author:    "John Doe",
		FeedTitle: "Test Feed",
		URL:       "https://example.com/article",
		Summary:   "This is a test summary",
	}

	tests := []struct {
		name      string
		condition RuleCondition
		expected  bool
	}{
		{
			name: "Always condition",
			condition: RuleCondition{
				Type: "always",
			},
			expected: true,
		},
		{
			name: "Filter - contains title",
			condition: RuleCondition{
				Type: "filter",
				Filter: []FilterCondition{
					{Field: "article_title", Operator: "contains", Value: "Test"},
				},
			},
			expected: true,
		},
		{
			name: "Filter - equals author",
			condition: RuleCondition{
				Type: "filter",
				Filter: []FilterCondition{
					{Field: "author", Operator: "equals", Value: "John Doe"},
				},
			},
			expected: true,
		},
		{
			name: "Filter - not equals",
			condition: RuleCondition{
				Type: "filter",
				Filter: []FilterCondition{
					{Field: "author", Operator: "not_equals", Value: "Jane Doe"},
				},
			},
			expected: true,
		},
		{
			name: "Filter - starts_with",
			condition: RuleCondition{
				Type: "filter",
				Filter: []FilterCondition{
					{Field: "article_title", Operator: "starts_with", Value: "Test"},
				},
			},
			expected: true,
		},
		{
			name: "Filter - ends_with",
			condition: RuleCondition{
				Type: "filter",
				Filter: []FilterCondition{
					{Field: "article_title", Operator: "ends_with", Value: "Article"},
				},
			},
			expected: true,
		},
		{
			name: "Filter - no match",
			condition: RuleCondition{
				Type: "filter",
				Filter: []FilterCondition{
					{Field: "article_title", Operator: "contains", Value: "Nonexistent"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.evaluateCondition(tt.condition, article)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateFilter(t *testing.T) {
	engine := &Engine{}

	article := &models.Article{
		Title:     "Test Article",
		Author:    "John Doe",
		FeedTitle: "Test Feed",
	}

	tests := []struct {
		name     string
		filters  []FilterCondition
		expected bool
	}{
		{
			name: "AND logic",
			filters: []FilterCondition{
				{Field: "article_title", Operator: "contains", Value: "Test", Logic: "and"},
				{Field: "author", Operator: "equals", Value: "John Doe"},
			},
			expected: true,
		},
		{
			name: "OR logic",
			filters: []FilterCondition{
				{Field: "article_title", Operator: "contains", Value: "Nonexistent", Logic: "or"},
				{Field: "author", Operator: "equals", Value: "John Doe"},
			},
			expected: true,
		},
		{
			name:     "Empty filters",
			filters:  []FilterCondition{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.evaluateFilter(tt.filters, article)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
