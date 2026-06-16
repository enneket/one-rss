package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	feedpkg "github.com/zjx/one-rss/internal/feed"
	"github.com/zjx/one-rss/internal/models"
)

func handleFeeds(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		rows, err := db.Query(`
			SELECT id, title, url, link, description, category, image_url, position,
				   last_updated, last_error, script_path, hide_from_timeline,
				   proxy_url, proxy_enabled, refresh_interval, is_image_mode, type,
				   article_view_mode, auto_expand_content
			FROM feeds ORDER BY position, title
		`)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to fetch feeds")
			return
		}
		defer rows.Close()

		var feeds []models.Feed
		for rows.Next() {
			var feed models.Feed
			err := rows.Scan(
				&feed.ID, &feed.Title, &feed.URL, &feed.Link, &feed.Description,
				&feed.Category, &feed.ImageURL, &feed.Position, &feed.LastUpdated,
				&feed.LastError, &feed.ScriptPath, &feed.HideFromTimeline,
				&feed.ProxyURL, &feed.ProxyEnabled, &feed.RefreshInterval,
				&feed.IsImageMode, &feed.Type, &feed.ArticleViewMode,
				&feed.AutoExpandContent,
			)
			if err != nil {
				continue
			}
			feeds = append(feeds, feed)
		}

		if feeds == nil {
			feeds = []models.Feed{}
		}
		jsonResponse(w, feeds)
	}
}

func handleAddFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			URL      string `json:"url"`
			Category string `json:"category"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.URL == "" {
			errorResponse(w, http.StatusBadRequest, "URL is required")
			return
		}

		// TODO: Fetch feed info from URL
		title := req.URL
		description := ""

		result, err := db.Exec(`
			INSERT INTO feeds (url, title, description, category, last_updated)
			VALUES (?, ?, ?, ?, ?)
		`, req.URL, title, description, req.Category, time.Now())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to add feed")
			return
		}

		id, _ := result.LastInsertId()
		jsonResponse(w, map[string]interface{}{
			"id":      id,
			"url":     req.URL,
			"title":   title,
			"message": "Feed added successfully",
		})
	}
}

func handleDeleteFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		_, err := db.Exec("DELETE FROM feeds WHERE id = ?", req.ID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to delete feed")
			return
		}

		jsonResponse(w, map[string]string{"message": "Feed deleted successfully"})
	}
}

func handleUpdateFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var feed models.Feed
		if err := json.NewDecoder(r.Body).Decode(&feed); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		_, err := db.Exec(`
			UPDATE feeds SET title=?, url=?, link=?, description=?, category=?,
				   image_url=?, hide_from_timeline=?, proxy_url=?, proxy_enabled=?,
				   refresh_interval=?, is_image_mode=?, type=?, article_view_mode=?,
				   auto_expand_content=?
			WHERE id=?
		`, feed.Title, feed.URL, feed.Link, feed.Description, feed.Category,
			feed.ImageURL, feed.HideFromTimeline, feed.ProxyURL, feed.ProxyEnabled,
			feed.RefreshInterval, feed.IsImageMode, feed.Type, feed.ArticleViewMode,
			feed.AutoExpandContent, feed.ID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to update feed")
			return
		}

		jsonResponse(w, map[string]string{"message": "Feed updated successfully"})
	}
}

func handleRefreshFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Get feed from database
		var feed models.Feed
		err := db.QueryRow(`
			SELECT id, title, url, refresh_interval, last_updated
			FROM feeds WHERE id = ?
		`, req.ID).Scan(&feed.ID, &feed.Title, &feed.URL, &feed.RefreshInterval, &feed.LastUpdated)
		if err != nil {
			errorResponse(w, http.StatusNotFound, "Feed not found")
			return
		}

		// Create fetcher and refresh feed
		fetcher := feedpkg.NewFetcher(db)
		go fetcher.FetchAndSave(&feed)

		jsonResponse(w, map[string]string{"message": "Feed refresh started"})
	}
}

func handleReorderFeeds(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			FeedIDs []int64 `json:"feedIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		tx, err := db.Begin()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to start transaction")
			return
		}

		for i, id := range req.FeedIDs {
			_, err := tx.Exec("UPDATE feeds SET position = ? WHERE id = ?", i, id)
			if err != nil {
				tx.Rollback()
				errorResponse(w, http.StatusInternalServerError, "Failed to reorder feeds")
				return
			}
		}

		if err := tx.Commit(); err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to commit transaction")
			return
		}

		jsonResponse(w, map[string]string{"message": "Feeds reordered successfully"})
	}
}
