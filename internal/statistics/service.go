package statistics

import (
	"database/sql"
	"time"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type Stats struct {
	TotalArticles   int            `json:"total_articles"`
	ReadArticles    int            `json:"read_articles"`
	FavoriteArticles int           `json:"favorite_articles"`
	TotalFeeds      int            `json:"total_feeds"`
	DailyStats      []DailyStat    `json:"daily_stats"`
	FeedStats       []FeedStat     `json:"feed_stats"`
	CategoryStats   []CategoryStat `json:"category_stats"`
}

type DailyStat struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type FeedStat struct {
	FeedID    int64  `json:"feed_id"`
	FeedTitle string `json:"feed_title"`
	Count     int    `json:"count"`
}

type CategoryStat struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

func (s *Service) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Get total articles
	err := s.db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&stats.TotalArticles)
	if err != nil {
		return nil, err
	}

	// Get read articles
	err = s.db.QueryRow("SELECT COUNT(*) FROM articles WHERE is_read = 1").Scan(&stats.ReadArticles)
	if err != nil {
		return nil, err
	}

	// Get favorite articles
	err = s.db.QueryRow("SELECT COUNT(*) FROM articles WHERE is_favorite = 1").Scan(&stats.FavoriteArticles)
	if err != nil {
		return nil, err
	}

	// Get total feeds
	err = s.db.QueryRow("SELECT COUNT(*) FROM feeds").Scan(&stats.TotalFeeds)
	if err != nil {
		return nil, err
	}

	// Get daily stats for last 30 days
	dailyStats, err := s.getDailyStats(30)
	if err != nil {
		return nil, err
	}
	stats.DailyStats = dailyStats

	// Get feed stats
	feedStats, err := s.getFeedStats()
	if err != nil {
		return nil, err
	}
	stats.FeedStats = feedStats

	// Get category stats
	categoryStats, err := s.getCategoryStats()
	if err != nil {
		return nil, err
	}
	stats.CategoryStats = categoryStats

	return stats, nil
}

func (s *Service) getDailyStats(days int) ([]DailyStat, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	rows, err := s.db.Query(`
		SELECT DATE(published_at) as date, COUNT(*) as count
		FROM articles
		WHERE published_at >= ?
		GROUP BY DATE(published_at)
		ORDER BY date
	`, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []DailyStat
	for rows.Next() {
		var stat DailyStat
		if err := rows.Scan(&stat.Date, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (s *Service) getFeedStats() ([]FeedStat, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.title, COUNT(a.id) as count
		FROM feeds f
		LEFT JOIN articles a ON f.id = a.feed_id
		GROUP BY f.id, f.title
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []FeedStat
	for rows.Next() {
		var stat FeedStat
		if err := rows.Scan(&stat.FeedID, &stat.FeedTitle, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (s *Service) getCategoryStats() ([]CategoryStat, error) {
	rows, err := s.db.Query(`
		SELECT category, COUNT(*) as count
		FROM feeds
		WHERE category != ''
		GROUP BY category
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []CategoryStat
	for rows.Next() {
		var stat CategoryStat
		if err := rows.Scan(&stat.Category, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (s *Service) GetUnreadCounts() (map[int64]int, error) {
	rows, err := s.db.Query(`
		SELECT feed_id, COUNT(*) as count
		FROM articles
		WHERE is_read = 0 AND is_hidden = 0
		GROUP BY feed_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[int64]int)
	for rows.Next() {
		var feedID int64
		var count int
		if err := rows.Scan(&feedID, &count); err != nil {
			continue
		}
		counts[feedID] = count
	}

	return counts, nil
}

func (s *Service) GetTotalUnread() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM articles WHERE is_read = 0 AND is_hidden = 0").Scan(&count)
	return count, err
}
