package handlers

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
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

		// 处理 rsshub:// 协议
		feedURL := req.URL
		isRsshub := false
		rsshubPath := ""
		if strings.HasPrefix(feedURL, "rsshub://") {
			isRsshub = true
			rsshubPath = strings.TrimPrefix(feedURL, "rsshub://")
			rsshubURL := os.Getenv("RSSHUB_URL")
			if rsshubURL == "" {
				rsshubURL = "http://rsshub:1200"
			}
			feedURL = rsshubURL + "/" + rsshubPath
		}

		// 获取 RSS 源信息
		title := feedURL
		description := ""
		link := ""
		imageURL := ""

		// 使用 gofeed 解析 RSS 源
		client := &http.Client{Timeout: 15 * time.Second}
		httpReq, err := http.NewRequest("GET", feedURL, nil)
		if err == nil {
			httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
			httpReq.Header.Set("Accept", "application/rss+xml,application/atom+xml,application/xml;q=0.9,*/*;q=0.8")
			resp, err := client.Do(httpReq)
			if err == nil {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					parser := gofeed.NewParser()
					feed, err := parser.ParseString(string(body))
					if err == nil {
						// 获取真实标题
						if feed.Title != "" {
							title = feed.Title
						}
						// 获取描述
						if feed.Description != "" {
							description = feed.Description
						}
						// 获取网站链接
						if feed.Link != "" {
							link = feed.Link
						}
						// 获取图标
						if feed.Image != nil && feed.Image.URL != "" {
							imageURL = feed.Image.URL
						}
					}
				}
			}
		}

		// 保存 URL：对于 rsshub:// 协议，保存原始格式以便后续解析
		saveURL := feedURL
		if isRsshub {
			saveURL = "rsshub://" + rsshubPath
		}

		result, err := db.Exec(`
			INSERT INTO feeds (url, title, link, description, category, image_url, last_updated)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, saveURL, title, link, description, req.Category, imageURL, time.Now())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to add feed")
			return
		}

		id, _ := result.LastInsertId()
		jsonResponse(w, map[string]interface{}{
			"id":          id,
			"url":         saveURL,
			"title":       title,
			"link":        link,
			"description": description,
			"image_url":   imageURL,
			"message":     "Feed added successfully",
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
		if err := fetcher.FetchAndSave(&feed); err != nil {
			errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to refresh feed: %v", err))
			return
		}

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

// OPML structures
type OPML struct {
	XMLName xml.Name   `xml:"opml"`
	Version string     `xml:"version,attr"`
	Head    OPMLHead   `xml:"head"`
	Body    OPMLBody   `xml:"body"`
}

type OPMLHead struct {
	Title string `xml:"title"`
}

type OPMLBody struct {
	Outlines []OPMLOutline `xml:"outline"`
}

type OPMLOutline struct {
	Text        string         `xml:"text,attr"`
	Title       string         `xml:"title,attr,omitempty"`
	XMLURL      string         `xml:"xmlUrl,attr"`
	HTMLURL     string         `xml:"htmlUrl,attr,omitempty"`
	Description string         `xml:"description,attr,omitempty"`
	Category    string         `xml:"category,attr,omitempty"`
	Outlines    []OPMLOutline  `xml:"outline,omitempty"`
}

func handleExportFeeds(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		rows, err := db.Query("SELECT title, url, link, description, category FROM feeds ORDER BY position, title")
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to fetch feeds")
			return
		}
		defer rows.Close()

		// Group by category
		categoryMap := map[string][]OPMLOutline{}
		var uncategorized []OPMLOutline

		for rows.Next() {
			var title, url string
			var link, description, category sql.NullString
			if err := rows.Scan(&title, &url, &link, &description, &category); err != nil {
				continue
			}

			outline := OPMLOutline{
				Text:        title,
				Title:       title,
				XMLURL:      url,
				HTMLURL:     link.String,
				Description: description.String,
			}

			if category.Valid && category.String != "" {
				categoryMap[category.String] = append(categoryMap[category.String], outline)
			} else {
				uncategorized = append(uncategorized, outline)
			}
		}

		// Build outlines: categories as folders + uncategorized
		var outlines []OPMLOutline
		for cat, items := range categoryMap {
			outlines = append(outlines, OPMLOutline{
				Text:     cat,
				Title:    cat,
				Outlines: items,
			})
		}
		outlines = append(outlines, uncategorized...)

		opml := OPML{
			Version: "2.0",
			Head:    OPMLHead{Title: "OneRSS Subscriptions"},
			Body:    OPMLBody{Outlines: outlines},
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=onerss-subscriptions.xml")
		fmt.Fprint(w, xml.Header)
		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		enc.Encode(opml)
	}
}

func handleImportFeeds(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		var opml OPML
		if err := xml.Unmarshal(body, &opml); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid OPML format")
			return
		}

		imported := 0
		skipped := 0

		// Flatten nested outlines
		var flatOutlines []struct {
			outline  OPMLOutline
			category string
		}
		var flatten func(outlines []OPMLOutline, category string)
		flatten = func(outlines []OPMLOutline, category string) {
			for _, o := range outlines {
				if o.XMLURL != "" {
					flatOutlines = append(flatOutlines, struct {
						outline  OPMLOutline
						category string
					}{o, category})
				}
				if len(o.Outlines) > 0 {
					cat := category
					if o.XMLURL == "" {
						if cat != "" {
							cat = cat + "/" + o.Text
						} else {
							cat = o.Text
						}
					}
					flatten(o.Outlines, cat)
				}
			}
		}
		flatten(opml.Body.Outlines, "")

		for _, item := range flatOutlines {
			url := item.outline.XMLURL
			if url == "" {
				continue
			}

			// Check if already exists
			var exists int
			db.QueryRow("SELECT COUNT(*) FROM feeds WHERE url = ?", url).Scan(&exists)
			if exists > 0 {
				skipped++
				continue
			}

			title := item.outline.Title
			if title == "" {
				title = item.outline.Text
			}
			if title == "" {
				title = url
			}

			category := item.category
			if item.outline.Category != "" {
				category = item.outline.Category
			}

			_, err := db.Exec(`
				INSERT INTO feeds (url, title, link, description, category, last_updated)
				VALUES (?, ?, ?, ?, ?, ?)
			`, url, title, item.outline.HTMLURL, item.outline.Description, category, time.Now())
			if err != nil {
				continue
			}
			imported++
		}

		jsonResponse(w, map[string]interface{}{
			"message":  "Import completed",
			"imported": imported,
			"skipped":  skipped,
		})
	}
}
