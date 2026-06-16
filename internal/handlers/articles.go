package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zjx/one-rss/internal/models"
)

func handleArticles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		feedID := r.URL.Query().Get("feed_id")
		filter := r.URL.Query().Get("filter")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 50
		}
		offset := (page - 1) * limit

		query := `
			SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url,
				   a.video_url, a.published_at, a.is_read, a.is_favorite,
				   a.is_hidden, a.is_read_later, a.translated_title, a.summary,
				   a.author, f.title as feed_title
			FROM articles a
			LEFT JOIN feeds f ON a.feed_id = f.id
			WHERE 1=1
		`
		args := []interface{}{}

		if feedID != "" {
			query += " AND a.feed_id = ?"
			args = append(args, feedID)
		}

		switch filter {
		case "unread":
			query += " AND a.is_read = 0 AND a.is_hidden = 0"
		case "favorites":
			query += " AND a.is_favorite = 1"
		case "readLater":
			query += " AND a.is_read_later = 1"
		case "hidden":
			query += " AND a.is_hidden = 1"
		default:
			query += " AND a.is_hidden = 0"
		}

		query += " ORDER BY a.published_at DESC LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to fetch articles")
			return
		}
		defer rows.Close()

		var articles []models.Article
		for rows.Next() {
			var article models.Article
			err := rows.Scan(
				&article.ID, &article.FeedID, &article.Title, &article.URL,
				&article.ImageURL, &article.AudioURL, &article.VideoURL,
				&article.PublishedAt, &article.IsRead, &article.IsFavorite,
				&article.IsHidden, &article.IsReadLater, &article.TranslatedTitle,
				&article.Summary, &article.Author, &article.FeedTitle,
			)
			if err != nil {
				continue
			}
			articles = append(articles, article)
		}

		if articles == nil {
			articles = []models.Article{}
		}
		jsonResponse(w, articles)
	}
}

func handleArticleContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		articleID := r.URL.Query().Get("id")
		if articleID == "" {
			errorResponse(w, http.StatusBadRequest, "Article ID is required")
			return
		}

		// 首先尝试从 article_contents 表获取内容
		var content string
		err := db.QueryRow("SELECT content FROM article_contents WHERE article_id = ?", articleID).Scan(&content)
		if err == sql.ErrNoRows {
			// 如果 article_contents 没有数据，尝试从 articles 表获取
			err = db.QueryRow("SELECT content FROM articles WHERE id = ?", articleID).Scan(&content)
			if err == sql.ErrNoRows {
				jsonResponse(w, map[string]string{"content": ""})
				return
			}
			if err != nil {
				errorResponse(w, http.StatusInternalServerError, "Failed to fetch content")
				return
			}
		} else if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to fetch content")
			return
		}

		jsonResponse(w, map[string]string{"content": content})
	}
}

func handleMarkRead(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			ID     int64 `json:"id"`
			IsRead bool  `json:"is_read"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		_, err := db.Exec("UPDATE articles SET is_read = ? WHERE id = ?", req.IsRead, req.ID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to update article")
			return
		}

		jsonResponse(w, map[string]string{"message": "Article updated"})
	}
}

func handleToggleFavorite(db *sql.DB) http.HandlerFunc {
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

		_, err := db.Exec("UPDATE articles SET is_favorite = NOT is_favorite WHERE id = ?", req.ID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to toggle favorite")
			return
		}

		jsonResponse(w, map[string]string{"message": "Favorite toggled"})
	}
}

func handleToggleReadLater(db *sql.DB) http.HandlerFunc {
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

		_, err := db.Exec("UPDATE articles SET is_read_later = NOT is_read_later WHERE id = ?", req.ID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to toggle read later")
			return
		}

		jsonResponse(w, map[string]string{"message": "Read later toggled"})
	}
}

func handleToggleHide(db *sql.DB) http.HandlerFunc {
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

		_, err := db.Exec("UPDATE articles SET is_hidden = NOT is_hidden WHERE id = ?", req.ID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to toggle hide")
			return
		}

		jsonResponse(w, map[string]string{"message": "Hide toggled"})
	}
}

func handleMarkAllRead(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			FeedID *int64 `json:"feed_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		query := "UPDATE articles SET is_read = 1"
		args := []interface{}{}

		if req.FeedID != nil {
			query += " WHERE feed_id = ?"
			args = append(args, *req.FeedID)
		}

		_, err := db.Exec(query, args...)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to mark all read")
			return
		}

		jsonResponse(w, map[string]string{"message": "All articles marked as read"})
	}
}

func handleUnreadCounts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		counts := models.UnreadCounts{
			FeedCounts: make(map[int64]int),
		}

		// Get total unread
		err := db.QueryRow("SELECT COUNT(*) FROM articles WHERE is_read = 0 AND is_hidden = 0").Scan(&counts.Total)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to get unread counts")
			return
		}

		// Get per-feed unread
		rows, err := db.Query("SELECT feed_id, COUNT(*) FROM articles WHERE is_read = 0 AND is_hidden = 0 GROUP BY feed_id")
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to get feed counts")
			return
		}
		defer rows.Close()

		for rows.Next() {
			var feedID int64
			var count int
			if err := rows.Scan(&feedID, &count); err != nil {
				continue
			}
			counts.FeedCounts[feedID] = count
		}

		jsonResponse(w, counts)
	}
}


