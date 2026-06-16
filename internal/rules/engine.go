package rules

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zjx/one-rss/internal/models"
)

type Engine struct {
	db *sql.DB
}

func NewEngine(db *sql.DB) *Engine {
	return &Engine{db: db}
}

type Rule struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Enabled   bool          `json:"enabled"`
	Condition RuleCondition `json:"condition"`
	Actions   []RuleAction  `json:"actions"`
}

type RuleCondition struct {
	Type   string           `json:"type"`   // "always" or "filter"
	Filter []FilterCondition `json:"filter,omitempty"`
}

type FilterCondition struct {
	Field    string `json:"field"`    // feed_name, article_title, is_read, etc.
	Operator string `json:"operator"` // contains, equals, not_equals, starts_with, ends_with
	Value    string `json:"value"`
	Logic    string `json:"logic"`    // and, or
}

type RuleAction struct {
	Type string `json:"type"` // favorite, hide, mark_read, read_later
}

func (e *Engine) GetRules() ([]Rule, error) {
	// For now, return empty rules since we don't have a rules table yet
	// In production, this would query a rules table
	return []Rule{}, nil
}

func (e *Engine) ApplyRules(article *models.Article) error {
	rules, err := e.GetRules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		matches, err := e.evaluateCondition(rule.Condition, article)
		if err != nil {
			continue
		}

		if matches {
			for _, action := range rule.Actions {
				e.executeAction(action, article)
			}
		}
	}

	return nil
}

func (e *Engine) evaluateCondition(condition RuleCondition, article *models.Article) (bool, error) {
	switch condition.Type {
	case "always":
		return true, nil
	case "filter":
		return e.evaluateFilter(condition.Filter, article)
	default:
		return false, fmt.Errorf("unknown condition type: %s", condition.Type)
	}
}

func (e *Engine) evaluateFilter(filters []FilterCondition, article *models.Article) (bool, error) {
	if len(filters) == 0 {
		return true, nil
	}

	result := true
	logic := "and"

	for _, filter := range filters {
		match, err := e.evaluateSingleFilter(filter, article)
		if err != nil {
			return false, err
		}

		switch logic {
		case "and":
			result = result && match
		case "or":
			result = result || match
		}

		logic = filter.Logic
	}

	return result, nil
}

func (e *Engine) evaluateSingleFilter(filter FilterCondition, article *models.Article) (bool, error) {
	var fieldValue string

	switch filter.Field {
	case "feed_name":
		fieldValue = article.FeedTitle
	case "article_title":
		fieldValue = article.Title
	case "author":
		fieldValue = article.Author
	case "content":
		fieldValue = article.Summary
	case "url":
		fieldValue = article.URL
	default:
		return false, fmt.Errorf("unknown field: %s", filter.Field)
	}

	switch filter.Operator {
	case "contains":
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(filter.Value)), nil
	case "equals":
		return strings.EqualFold(fieldValue, filter.Value), nil
	case "not_equals":
		return !strings.EqualFold(fieldValue, filter.Value), nil
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(fieldValue), strings.ToLower(filter.Value)), nil
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(fieldValue), strings.ToLower(filter.Value)), nil
	default:
		return false, fmt.Errorf("unknown operator: %s", filter.Operator)
	}
}

func (e *Engine) executeAction(action RuleAction, article *models.Article) error {
	switch action.Type {
	case "favorite":
		_, err := e.db.Exec("UPDATE articles SET is_favorite = 1 WHERE id = ?", article.ID)
		article.IsFavorite = true
		return err
	case "hide":
		_, err := e.db.Exec("UPDATE articles SET is_hidden = 1 WHERE id = ?", article.ID)
		article.IsHidden = true
		return err
	case "mark_read":
		_, err := e.db.Exec("UPDATE articles SET is_read = 1 WHERE id = ?", article.ID)
		article.IsRead = true
		return err
	case "read_later":
		_, err := e.db.Exec("UPDATE articles SET is_read_later = 1 WHERE id = ?", article.ID)
		article.IsReadLater = true
		return err
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (e *Engine) SaveRule(rule Rule) error {
	// In production, this would save to a rules table
	rulesJSON, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	// For now, just log it
	fmt.Printf("Saving rule: %s\n", string(rulesJSON))
	return nil
}

func (e *Engine) DeleteRule(id int64) error {
	// In production, this would delete from a rules table
	return nil
}
