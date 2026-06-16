package models

import "time"

type Feed struct {
	ID                 int64      `json:"id"`
	Title              string     `json:"title"`
	URL                string     `json:"url"`
	Link               string     `json:"link"`
	Description        string     `json:"description"`
	Category           string     `json:"category"`
	ImageURL           string     `json:"image_url"`
	Position           int        `json:"position"`
	LastUpdated        time.Time  `json:"last_updated"`
	LastError          string     `json:"last_error,omitempty"`
	ScriptPath         string     `json:"script_path,omitempty"`
	HideFromTimeline   bool       `json:"hide_from_timeline"`
	ProxyURL           string     `json:"proxy_url,omitempty"`
	ProxyEnabled       bool       `json:"proxy_enabled"`
	RefreshInterval    int        `json:"refresh_interval"`
	IsImageMode        bool       `json:"is_image_mode"`
	Type               string     `json:"type"`
	XPathItem          string     `json:"xpath_item"`
	XPathItemTitle     string     `json:"xpath_item_title"`
	XPathItemContent   string     `json:"xpath_item_content"`
	XPathItemURI       string     `json:"xpath_item_uri"`
	XPathItemAuthor    string     `json:"xpath_item_author"`
	XPathItemTimestamp string     `json:"xpath_item_timestamp"`
	XPathItemTimeFormat string   `json:"xpath_item_time_format"`
	XPathItemThumbnail string   `json:"xpath_item_thumbnail"`
	XPathItemCategories string  `json:"xpath_item_categories"`
	XPathItemUID       string    `json:"xpath_item_uid"`
	ArticleViewMode    string    `json:"article_view_mode"`
	AutoExpandContent  string    `json:"auto_expand_content"`
	EmailAddress       string    `json:"email_address,omitempty"`
	EmailIMAPServer    string    `json:"email_imap_server,omitempty"`
	EmailIMAPPort      int       `json:"email_imap_port"`
	EmailUsername      string    `json:"email_username,omitempty"`
	EmailPassword      string    `json:"email_password,omitempty"`
	EmailFolder        string    `json:"email_folder"`
	EmailLastUID       int       `json:"email_last_uid"`
	IsFreshRSSSource   bool      `json:"is_freshrss_source"`
	FreshRSSStreamID   string    `json:"freshrss_stream_id"`
	LatestArticleTime  *time.Time `json:"latest_article_time,omitempty"`
	ArticlesPerMonth   float64   `json:"articles_per_month,omitempty"`
	LastUpdateStatus   string    `json:"last_update_status,omitempty"`
	Tags               []Tag     `json:"tags,omitempty"`
}

type Article struct {
	ID              int64     `json:"id"`
	FeedID          int64     `json:"feed_id"`
	Title           string    `json:"title"`
	URL             string    `json:"url"`
	ImageURL        string    `json:"image_url"`
	AudioURL        string    `json:"audio_url"`
	VideoURL        string    `json:"video_url"`
	PublishedAt     time.Time `json:"published_at"`
	IsRead          bool      `json:"is_read"`
	IsFavorite      bool      `json:"is_favorite"`
	IsHidden        bool      `json:"is_hidden"`
	IsReadLater     bool      `json:"is_read_later"`
	FeedTitle       string    `json:"feed_title,omitempty"`
	Author          string    `json:"author,omitempty"`
	TranslatedTitle string    `json:"translated_title"`
	Summary         string    `json:"summary"`
	UniqueID        string    `json:"unique_id"`
	FreshRSSItemID  string    `json:"freshrss_item_id"`
}

type ArticleContent struct {
	ID        int64     `json:"id"`
	ArticleID int64     `json:"article_id"`
	Content   string    `json:"content"`
	FetchedAt time.Time `json:"fetched_at"`
}

type Tag struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Position int    `json:"position"`
}

type ChatSession struct {
	ID        int64     `json:"id"`
	ArticleID int64     `json:"article_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Thinking  string    `json:"thinking,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type AIProfile struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	APIKey        string    `json:"api_key"`
	Endpoint      string    `json:"endpoint"`
	Model         string    `json:"model"`
	CustomHeaders string    `json:"custom_headers"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SavedFilter struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Conditions string    `json:"conditions"`
	Position   int       `json:"position"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Rule struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Enabled   bool          `json:"enabled"`
	Condition RuleCondition `json:"condition"`
	Actions   []RuleAction  `json:"actions"`
}

type RuleCondition struct {
	Type   string           `json:"type"`
	Filter []FilterCondition `json:"filter,omitempty"`
}

type FilterCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	Logic    string `json:"logic"`
}

type RuleAction struct {
	Type string `json:"type"`
}

type Settings struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Encrypted bool   `json:"encrypted"`
}

type UnreadCounts struct {
	Total      int            `json:"total"`
	FeedCounts map[int64]int  `json:"feedCounts"`
}

type RefreshProgress struct {
	IsRunning bool    `json:"isRunning"`
	Current   int     `json:"current"`
	Total     int     `json:"total"`
	FeedTitle string  `json:"feedTitle"`
	Percent   float64 `json:"percent"`
}
