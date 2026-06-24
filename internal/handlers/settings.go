package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/zjx/one-rss/internal/feed"
	"github.com/zjx/one-rss/internal/models"
)

func handleSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query("SELECT key, value, encrypted FROM settings")
			if err != nil {
				errorResponse(w, http.StatusInternalServerError, "Failed to fetch settings")
				return
			}
			defer rows.Close()

			settings := make(map[string]interface{})
			for rows.Next() {
				var key, value string
				var encrypted bool
				if err := rows.Scan(&key, &value, &encrypted); err != nil {
					continue
				}
				// Try to parse JSON value
				var jsonValue interface{}
				if err := json.Unmarshal([]byte(value), &jsonValue); err != nil {
					// If not valid JSON, use raw string
					settings[key] = value
				} else {
					settings[key] = jsonValue
				}
			}
			jsonResponse(w, settings)

		case http.MethodPost:
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				errorResponse(w, http.StatusBadRequest, "Invalid request body")
				return
			}

			tx, err := db.Begin()
			if err != nil {
				errorResponse(w, http.StatusInternalServerError, "Failed to start transaction")
				return
			}

			for key, value := range req {
				// Store value as JSON string
				var valueStr string
				switch v := value.(type) {
				case string:
					// Store strings directly without JSON encoding
					valueStr = v
				default:
					// For other types, use JSON encoding
					jsonBytes, _ := json.Marshal(v)
					valueStr = string(jsonBytes)
				}
				
				_, err := tx.Exec(`
					INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
					ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
				`, key, valueStr, time.Now(), valueStr, time.Now())
				if err != nil {
					tx.Rollback()
					errorResponse(w, http.StatusInternalServerError, "Failed to update settings")
					return
				}
			}

			if err := tx.Commit(); err != nil {
				errorResponse(w, http.StatusInternalServerError, "Failed to commit settings")
				return
			}

			jsonResponse(w, map[string]string{"message": "Settings updated"})

		default:
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

func handleTags(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query("SELECT id, name, color, position FROM tags ORDER BY position")
			if err != nil {
				errorResponse(w, http.StatusInternalServerError, "Failed to fetch tags")
				return
			}
			defer rows.Close()

			var tags []struct {
				ID       int64  `json:"id"`
				Name     string `json:"name"`
				Color    string `json:"color"`
				Position int    `json:"position"`
			}
			for rows.Next() {
				var tag struct {
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Color    string `json:"color"`
					Position int    `json:"position"`
				}
				if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.Position); err != nil {
					continue
				}
				tags = append(tags, tag)
			}

			if tags == nil {
				tags = []struct {
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Color    string `json:"color"`
					Position int    `json:"position"`
				}{}
			}
			jsonResponse(w, tags)

		default:
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

func handleRefresh(db *sql.DB, fetcher *feed.Fetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		rows, err := db.Query("SELECT id, title, url, refresh_interval, last_updated FROM feeds WHERE hide_from_timeline = 0")
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to query feeds")
			return
		}
		defer rows.Close()

		var feedIDs []int64
		for rows.Next() {
			var id int64
			var title, url string
			var refreshInterval int
			var lastUpdated string
			if err := rows.Scan(&id, &title, &url, &refreshInterval, &lastUpdated); err != nil {
				continue
			}
			feedIDs = append(feedIDs, id)
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		var refreshed, failed int

		for _, id := range feedIDs {
			wg.Add(1)
			go func(feedID int64) {
				defer wg.Done()
				var feed models.Feed
				err := db.QueryRow("SELECT id, title, url, refresh_interval, last_updated FROM feeds WHERE id = ?", feedID).Scan(&feed.ID, &feed.Title, &feed.URL, &feed.RefreshInterval, &feed.LastUpdated)
				if err != nil {
					log.Printf("Refresh: failed to query feed %d: %v", feedID, err)
					mu.Lock()
					failed++
					mu.Unlock()
					return
				}
				if err := fetcher.FetchAndSave(&feed); err != nil {
					log.Printf("Refresh: failed to fetch feed %s: %v", feed.Title, err)
					mu.Lock()
					failed++
					mu.Unlock()
				} else {
					log.Printf("Refresh: feed %s updated", feed.Title)
					mu.Lock()
					refreshed++
					mu.Unlock()
				}
			}(id)
		}

		wg.Wait()

		jsonResponse(w, map[string]interface{}{
			"message":   "Refresh completed",
			"refreshed": refreshed,
			"failed":    failed,
			"total":     len(feedIDs),
		})
	}
}

func handleProgress(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// TODO: Implement progress tracking
		progress := struct {
			IsRunning bool    `json:"isRunning"`
			Current   int     `json:"current"`
			Total     int     `json:"total"`
			FeedTitle string  `json:"feedTitle"`
			Percent   float64 `json:"percent"`
		}{
			IsRunning: false,
			Current:   0,
			Total:     0,
		}
		jsonResponse(w, progress)
	}
}

func handleSavedFilters(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query("SELECT id, name, conditions, position FROM saved_filters ORDER BY position")
			if err != nil {
				errorResponse(w, http.StatusInternalServerError, "Failed to fetch saved filters")
				return
			}
			defer rows.Close()

			var filters []struct {
				ID         int64  `json:"id"`
				Name       string `json:"name"`
				Conditions string `json:"conditions"`
				Position   int    `json:"position"`
			}
			for rows.Next() {
				var filter struct {
					ID         int64  `json:"id"`
					Name       string `json:"name"`
					Conditions string `json:"conditions"`
					Position   int    `json:"position"`
				}
				if err := rows.Scan(&filter.ID, &filter.Name, &filter.Conditions, &filter.Position); err != nil {
					continue
				}
				filters = append(filters, filter)
			}

			if filters == nil {
				filters = []struct {
					ID         int64  `json:"id"`
					Name       string `json:"name"`
					Conditions string `json:"conditions"`
					Position   int    `json:"position"`
				}{}
			}
			jsonResponse(w, filters)

		default:
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}
