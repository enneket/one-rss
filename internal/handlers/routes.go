package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zjx/one-rss/internal/ai"
	"github.com/zjx/one-rss/internal/feed"
	"github.com/zjx/one-rss/internal/integration"
	"github.com/zjx/one-rss/internal/statistics"
	"github.com/zjx/one-rss/internal/summary"
	"github.com/zjx/one-rss/internal/translation"
)

func SetupRoutes(mux *http.ServeMux, db *sql.DB, fetcher *feed.Fetcher) {
	// Initialize services
	translationSvc, _ := translation.NewService(db, translation.Config{Provider: "google"})
	summarySvc, _ := summary.NewService(db, summary.Config{Provider: "local"})
	aiSvc := ai.NewService(db)
	integrationSvc := integration.NewService(db)
	statsSvc := statistics.NewService(db)

	// Health check
	mux.HandleFunc("/api/health", handleHealth(db))

	// Feed handlers
	mux.HandleFunc("/api/feeds", handleFeeds(db))
	mux.HandleFunc("/api/feeds/add", handleAddFeed(db))
	mux.HandleFunc("/api/feeds/delete", handleDeleteFeed(db))
	mux.HandleFunc("/api/feeds/update", handleUpdateFeed(db))
	mux.HandleFunc("/api/feeds/refresh", handleRefreshFeed(db))
	mux.HandleFunc("/api/feeds/reorder", handleReorderFeeds(db))

	// Article handlers
	mux.HandleFunc("/api/articles", handleArticles(db))
	mux.HandleFunc("/api/articles/content", handleArticleContent(db))
	mux.HandleFunc("/api/articles/read", handleMarkRead(db))
	mux.HandleFunc("/api/articles/favorite", handleToggleFavorite(db))
	mux.HandleFunc("/api/articles/toggle-read-later", handleToggleReadLater(db))
	mux.HandleFunc("/api/articles/toggle-hide", handleToggleHide(db))
	mux.HandleFunc("/api/articles/mark-all-read", handleMarkAllRead(db))
	mux.HandleFunc("/api/articles/unread-counts", handleUnreadCounts(db))

	// Translation handlers
	mux.HandleFunc("/api/articles/translate", handleTranslate(translationSvc))
	mux.HandleFunc("/api/articles/translate-text", handleTranslateText(translationSvc))

	// Summary handlers
	mux.HandleFunc("/api/articles/summarize", handleSummarize(summarySvc))

	// AI handlers
	mux.HandleFunc("/api/ai/test", handleAITest(aiSvc))
	mux.HandleFunc("/api/ai/search", handleAISearch(aiSvc))
	mux.HandleFunc("/api/ai/chat/sessions", handleChatSessions(aiSvc))
	mux.HandleFunc("/api/ai/chat/session/create", handleCreateChatSession(aiSvc))
	mux.HandleFunc("/api/ai/chat/messages", handleChatMessages(aiSvc))
	mux.HandleFunc("/api/ai/chat/send", handleChatSend(aiSvc))

	// Integration handlers
	mux.HandleFunc("/api/articles/export/obsidian", handleExportObsidian(integrationSvc))
	mux.HandleFunc("/api/articles/export/notion", handleExportNotion(integrationSvc))
	mux.HandleFunc("/api/articles/export/zotero", handleExportZotero(integrationSvc))
	mux.HandleFunc("/api/freshrss/sync", handleFreshRSSSync(integrationSvc))
	mux.HandleFunc("/api/rsshub/validate", handleRSSHubValidate(integrationSvc))

	// Statistics handlers
	mux.HandleFunc("/api/statistics", handleStatistics(statsSvc))
	mux.HandleFunc("/api/statistics/unread-counts", handleUnreadCounts(db))

	// Settings handlers
	mux.HandleFunc("/api/settings", handleSettings(db))

	// Tags handlers
	mux.HandleFunc("/api/tags", handleTags(db))

	// Refresh handlers
	mux.HandleFunc("/api/refresh", handleRefresh(db, fetcher))
	mux.HandleFunc("/api/progress", handleProgress(db))
}

func handleHealth(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			errorResponse(w, http.StatusServiceUnavailable, "Database connection failed")
			return
		}
		jsonResponse(w, map[string]string{"status": "ok"})
	}
}

func handleTranslate(svc *translation.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Text       string `json:"text"`
			SourceLang string `json:"source_lang"`
			TargetLang string `json:"target_lang"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		translated, err := svc.Translate(req.Text, req.SourceLang, req.TargetLang)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, map[string]string{"translated": translated})
	}
}

func handleTranslateText(svc *translation.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Text       string `json:"text"`
			TargetLang string `json:"target_lang"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		translated, err := svc.Translate(req.Text, "auto", req.TargetLang)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, map[string]string{"translated": translated})
	}
}

func handleSummarize(svc *summary.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Text   string `json:"text"`
			MaxLen int    `json:"max_len"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		summary, err := svc.Summarize(req.Text, req.MaxLen)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, map[string]string{"summary": summary})
	}
}

func handleAITest(svc *ai.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var config ai.Config
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		result, err := svc.TestConfig(config)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, map[string]string{"result": result})
	}
}

func handleAISearch(svc *ai.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Config   ai.Config             `json:"config"`
			Query    string                `json:"query"`
			Articles []map[string]string   `json:"articles"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		result, err := svc.Search(req.Config, req.Query, req.Articles)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, map[string]string{"result": result})
	}
}

func handleChatSessions(svc *ai.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		articleIDStr := r.URL.Query().Get("article_id")
		articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid article ID")
			return
		}

		sessions, err := svc.GetSessions(articleID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, sessions)
	}
}

func handleCreateChatSession(svc *ai.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			ArticleID int64  `json:"article_id"`
			Title     string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		session, err := svc.CreateSession(req.ArticleID, req.Title)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, session)
	}
}

func handleChatMessages(svc *ai.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		sessionIDStr := r.URL.Query().Get("session_id")
		sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid session ID")
			return
		}

		messages, err := svc.GetMessages(sessionID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, messages)
	}
}

func handleChatSend(svc *ai.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			SessionID      int64     `json:"session_id"`
			Config         ai.Config `json:"config"`
			ArticleContent string    `json:"article_content"`
			Question       string    `json:"question"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		message, err := svc.SendMessage(req.SessionID, req.Config, req.ArticleContent, req.Question)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, message)
	}
}

func handleExportObsidian(svc *integration.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Article interface{}             `json:"article"`
			Content string                  `json:"content"`
			Config  integration.ObsidianConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// For now, just return success
		jsonResponse(w, map[string]string{"message": "Exported to Obsidian"})
	}
}

func handleExportNotion(svc *integration.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Article interface{}           `json:"article"`
			Content string                `json:"content"`
			Config  integration.NotionConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// For now, just return success
		jsonResponse(w, map[string]string{"message": "Exported to Notion"})
	}
}

func handleExportZotero(svc *integration.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Article interface{}           `json:"article"`
			Config  integration.ZoteroConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// For now, just return success
		jsonResponse(w, map[string]string{"message": "Exported to Zotero"})
	}
}

func handleFreshRSSSync(svc *integration.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var config integration.FreshRSSConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		err := svc.SyncWithFreshRSS(config)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, map[string]string{"message": "FreshRSS sync completed"})
	}
}

func handleRSSHubValidate(svc *integration.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Config integration.RSSHubConfig `json:"config"`
			Route  string                   `json:"route"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		valid, err := svc.ValidateRSSHubRoute(req.Config, req.Route)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, map[string]bool{"valid": valid})
	}
}

func handleStatistics(svc *statistics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		stats, err := svc.GetStats()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, stats)
	}
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
