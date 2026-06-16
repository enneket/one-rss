package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zjx/one-rss/internal/database"
	"github.com/zjx/one-rss/internal/feed"
	"github.com/zjx/one-rss/internal/handlers"
	"github.com/zjx/one-rss/internal/scheduler"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize feed fetcher
	fetcher := feed.NewFetcher(db)

	// Initialize scheduler
	interval := 30 * time.Minute
	if envInterval := os.Getenv("REFRESH_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			interval = d
		}
	}
	sched := scheduler.NewScheduler(db, fetcher, interval)
	sched.Start()
	defer sched.Stop()

	// Setup HTTP handlers
	mux := http.NewServeMux()
	handlers.SetupRoutes(mux, db, fetcher)

	// Serve static files from embedded frontend/dist
	distFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to get sub filesystem: %v", err)
	}
	
	// Serve index.html for all non-API routes
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file
		f, err := distFS.Open(r.URL.Path[1:])
		if err != nil {
			// File not found, serve index.html for SPA routing
			r.URL.Path = "/"
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	})

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "6011"
	}

	// Start server
	log.Printf("OneRSS server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
