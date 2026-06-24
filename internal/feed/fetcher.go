package feed

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/zjx/one-rss/internal/models"
	"github.com/zjx/one-rss/internal/sanitizer"
)

type Fetcher struct {
	db        *sql.DB
	client    *http.Client
	parser    *gofeed.Parser
	rsshubURL string
}

func NewFetcher(db *sql.DB) *Fetcher {
	rsshubURL := os.Getenv("RSSHUB_URL")
	if rsshubURL == "" {
		rsshubURL = "http://rsshub:1200"
	}
	return &Fetcher{
		db: db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		parser:    gofeed.NewParser(),
		rsshubURL: rsshubURL,
	}
}

type FeedItem struct {
	Title       string
	URL         string
	Content     string
	ImageURL    string
	AudioURL    string
	VideoURL    string
	Author      string
	PublishedAt time.Time
	UniqueID    string
}

func (f *Fetcher) FetchFeed(feed *models.Feed) ([]FeedItem, error) {
	// 处理 rsshub:// 协议
	feedURL := feed.URL
	if strings.HasPrefix(feedURL, "rsshub://") {
		path := strings.TrimPrefix(feedURL, "rsshub://")
		feedURL = f.rsshubURL + "/" + path
	}

	// Fetch feed content
	resp, err := f.client.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	// Parse feed
	parsedFeed, err := f.parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	var items []FeedItem

	for _, item := range parsedFeed.Items {
		author := ""
		if item.Author != nil {
			author = item.Author.Name
		}

		feedItem := FeedItem{
			Title:   item.Title,
			URL:     item.Link,
			Content: item.Content,
			Author:  author,
		}

		// Get unique ID
		if item.GUID != "" {
			feedItem.UniqueID = item.GUID
		} else {
			feedItem.UniqueID = fmt.Sprintf("%x", md5.Sum([]byte(item.Link+item.Title)))
		}

		// Get published time
		if item.PublishedParsed != nil {
			feedItem.PublishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			feedItem.PublishedAt = *item.UpdatedParsed
		} else {
			feedItem.PublishedAt = time.Now()
		}

		// Get image
		if item.Image != nil {
			feedItem.ImageURL = item.Image.URL
		}

		// Check for enclosures (audio/video)
		for _, enc := range item.Enclosures {
			switch {
			case strings.HasPrefix(enc.Type, "audio"):
				feedItem.AudioURL = enc.URL
			case strings.HasPrefix(enc.Type, "video"):
				feedItem.VideoURL = enc.URL
			}
		}

		items = append(items, feedItem)
	}

	return items, nil
}

func (f *Fetcher) SaveArticles(feedID int64, items []FeedItem) error {
	tx, err := f.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO articles (feed_id, title, url, content, image_url, audio_url, video_url, author, published_at, unique_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		// 处理内容：清理 HTML、修复相对 URL、处理附件
		content := sanitizer.SanitizeHTML(item.Content)
		if item.URL != "" {
			content = sanitizer.FixRelativeURLs(content, item.URL)
		}

		// 处理附件
		var enclosures []sanitizer.Enclosure
		if item.AudioURL != "" {
			enclosures = append(enclosures, sanitizer.Enclosure{
				URL:    item.AudioURL,
				Type:   "audio/mpeg",
				Medium: "audio",
			})
		}
		if item.VideoURL != "" {
			enclosures = append(enclosures, sanitizer.Enclosure{
				URL:    item.VideoURL,
				Type:   "video/mp4",
				Medium: "video",
			})
		}
		content = sanitizer.ProcessEnclosures(content, enclosures)

		// 智能提取缩略图
		imageURL := sanitizer.ExtractThumbnail(content, item.ImageURL)

		_, err := stmt.Exec(
			feedID,
			item.Title,
			item.URL,
			content,
			imageURL,
			item.AudioURL,
			item.VideoURL,
			item.Author,
			item.PublishedAt,
			item.UniqueID,
		)
		if err != nil {
			// Log but continue
			fmt.Printf("Failed to insert article: %v\n", err)
		}
	}

	return tx.Commit()
}

func (f *Fetcher) FetchAndSave(feed *models.Feed) error {
	items, err := f.FetchFeed(feed)
	if err != nil {
		// Update feed with error
		f.db.Exec("UPDATE feeds SET last_error = ? WHERE id = ?", err.Error(), feed.ID)
		return err
	}

	if len(items) > 0 {
		// 保存文章
		if err := f.SaveArticles(feed.ID, items); err != nil {
			return err
		}

		// 异步抓取文章内容
		go f.fetchArticleContents(items)

		// Update feed last_updated
		_, err = f.db.Exec("UPDATE feeds SET last_updated = ?, last_error = '' WHERE id = ?", time.Now(), feed.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Fetcher) fetchArticleContents(items []FeedItem) {
	for _, item := range items {
		if item.URL == "" {
			continue
		}

		// 检查是否已经有内容
		var count int
		err := f.db.QueryRow("SELECT COUNT(*) FROM article_contents WHERE article_id = (SELECT id FROM articles WHERE unique_id = ?)", item.UniqueID).Scan(&count)
		if err == nil && count > 0 {
			continue
		}

		// 抓取文章内容
		content, err := f.FetchContent(item.URL)
		if err != nil {
			fmt.Printf("Failed to fetch content for %s: %v\n", item.URL, err)
			continue
		}

		// 获取文章 ID
		var articleID int64
		err = f.db.QueryRow("SELECT id FROM articles WHERE unique_id = ?", item.UniqueID).Scan(&articleID)
		if err != nil {
			continue
		}

		// 保存内容
		_, err = f.db.Exec(`
			INSERT OR REPLACE INTO article_contents (article_id, content, fetched_at)
			VALUES (?, ?, ?)
		`, articleID, content, time.Now())
		if err != nil {
			fmt.Printf("Failed to save content for article %d: %v\n", articleID, err)
		}
	}
}

func (f *Fetcher) FetchContent(url string) (string, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch content: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	content := string(body)

	// 清理 HTML
	content = sanitizer.SanitizeHTML(content)

	// 修复相对 URL
	content = sanitizer.FixRelativeURLs(content, url)

	return content, nil
}

func (f *Fetcher) DiscoverFeeds(url string) ([]string, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	content := string(body)
	var feeds []string

	// Look for RSS/Atom links in HTML
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "type=\"application/rss+xml\"") ||
			strings.Contains(lower, "type=\"application/atom+xml\"") {
			// Extract href
			start := strings.Index(lower, "href=\"")
			if start != -1 {
				start += 6
				end := strings.Index(line[start:], "\"")
				if end != -1 {
					feedURL := line[start : start+end]
					if !strings.HasPrefix(feedURL, "http") {
						// Relative URL
						feedURL = url + "/" + strings.TrimPrefix(feedURL, "/")
					}
					feeds = append(feeds, feedURL)
				}
			}
		}
	}

	// If no feeds found, try the URL directly
	if len(feeds) == 0 {
		_, err := f.parser.ParseString(content)
		if err == nil {
			feeds = append(feeds, url)
		}
	}

	return feeds, nil
}
