package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zjx/one-rss/internal/feed"
	"github.com/zjx/one-rss/internal/models"
)

type Scheduler struct {
	db       *sql.DB
	fetcher  *feed.Fetcher
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
	running  bool
	mu       sync.Mutex
}

func NewScheduler(db *sql.DB, fetcher *feed.Fetcher, interval time.Duration) *Scheduler {
	return &Scheduler{
		db:       db,
		fetcher:  fetcher,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run()
	}()

	log.Println("Scheduler started")
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) run() {
	// Run immediately on start
	s.refreshAllFeeds()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.refreshAllFeeds()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Scheduler) refreshAllFeeds() {
	log.Println("Refreshing all feeds...")

	rows, err := s.db.Query(`
		SELECT id, title, url, refresh_interval, last_updated
		FROM feeds
		WHERE hide_from_timeline = 0
	`)
	if err != nil {
		log.Printf("Failed to query feeds: %v", err)
		return
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var feed models.Feed
		if err := rows.Scan(&feed.ID, &feed.Title, &feed.URL, &feed.RefreshInterval, &feed.LastUpdated); err != nil {
			continue
		}
		feeds = append(feeds, feed)
	}

	for _, feed := range feeds {
		// Check if feed needs refresh
		if feed.RefreshInterval > 0 {
			nextRefresh := feed.LastUpdated.Add(time.Duration(feed.RefreshInterval) * time.Minute)
			if time.Now().Before(nextRefresh) {
				continue
			}
		}

		if err := s.fetcher.FetchAndSave(&feed); err != nil {
			log.Printf("Failed to refresh feed %s: %v", feed.Title, err)
		} else {
			log.Printf("Refreshed feed: %s", feed.Title)
		}
	}

	log.Printf("Refreshed %d feeds", len(feeds))
}

func (s *Scheduler) RefreshFeed(feedID int64) error {
	var feed models.Feed
	err := s.db.QueryRow(`
		SELECT id, title, url, refresh_interval, last_updated
		FROM feeds
		WHERE id = ?
	`, feedID).Scan(&feed.ID, &feed.Title, &feed.URL, &feed.RefreshInterval, &feed.LastUpdated)
	if err != nil {
		return fmt.Errorf("feed not found: %w", err)
	}

	return s.fetcher.FetchAndSave(&feed)
}
