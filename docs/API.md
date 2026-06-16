# OneRSS API Documentation

## Base URL

```
http://localhost:6011/api
```

## Authentication

Currently, the API does not require authentication. In production, consider adding authentication.

## Endpoints

### Health Check

#### GET /api/health

Check if the service is running.

**Response:**
```json
{
  "status": "ok"
}
```

### Feeds

#### GET /api/feeds

Get all feeds.

**Response:**
```json
[
  {
    "id": 1,
    "title": "Example Feed",
    "url": "https://example.com/feed.xml",
    "link": "https://example.com",
    "description": "An example feed",
    "category": "tech",
    "image_url": "",
    "position": 0,
    "last_updated": "2024-01-01T00:00:00Z",
    "last_error": "",
    "hide_from_timeline": false,
    "proxy_enabled": false,
    "proxy_url": "",
    "refresh_interval": 0,
    "is_image_mode": false
  }
]
```

#### POST /api/feeds/add

Add a new feed.

**Request:**
```json
{
  "url": "https://example.com/feed.xml",
  "category": "tech"
}
```

**Response:**
```json
{
  "id": 1,
  "url": "https://example.com/feed.xml",
  "title": "Example Feed",
  "message": "Feed added successfully"
}
```

#### POST /api/feeds/delete

Delete a feed.

**Request:**
```json
{
  "id": 1
}
```

**Response:**
```json
{
  "message": "Feed deleted successfully"
}
```

#### POST /api/feeds/update

Update a feed.

**Request:**
```json
{
  "id": 1,
  "title": "Updated Title",
  "category": "news"
}
```

#### POST /api/feeds/refresh

Refresh a feed.

**Request:**
```json
{
  "id": 1
}
```

### Articles

#### GET /api/articles

Get articles with filtering.

**Query Parameters:**
- `feed_id` (optional): Filter by feed ID
- `filter` (optional): Filter type (unread, favorites, readLater, hidden)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 50)

**Response:**
```json
[
  {
    "id": 1,
    "feed_id": 1,
    "title": "Article Title",
    "url": "https://example.com/article1",
    "image_url": "",
    "audio_url": "",
    "video_url": "",
    "published_at": "2024-01-01T00:00:00Z",
    "is_read": false,
    "is_favorite": false,
    "is_hidden": false,
    "is_read_later": false,
    "feed_title": "Example Feed",
    "author": "John Doe",
    "translated_title": "",
    "summary": ""
  }
]
```

#### GET /api/articles/content

Get article content.

**Query Parameters:**
- `id` (required): Article ID

**Response:**
```json
{
  "content": "<p>Article content here...</p>"
}
```

#### POST /api/articles/read

Mark article as read/unread.

**Request:**
```json
{
  "id": 1,
  "is_read": true
}
```

#### POST /api/articles/favorite

Toggle favorite status.

**Request:**
```json
{
  "id": 1
}
```

#### POST /api/articles/toggle-read-later

Toggle read later status.

**Request:**
```json
{
  "id": 1
}
```

#### POST /api/articles/toggle-hide

Toggle hide status.

**Request:**
```json
{
  "id": 1
}
```

#### POST /api/articles/mark-all-read

Mark all articles as read.

**Request:**
```json
{
  "feed_id": 1  // optional, omit for all feeds
}
```

#### GET /api/articles/unread-counts

Get unread counts.

**Response:**
```json
{
  "total": 42,
  "feedCounts": {
    "1": 10,
    "2": 32
  }
}
```

### Translation

#### POST /api/articles/translate

Translate text.

**Request:**
```json
{
  "text": "Hello world",
  "source_lang": "en",
  "target_lang": "zh"
}
```

**Response:**
```json
{
  "translated": "你好世界"
}
```

### Summary

#### POST /api/articles/summarize

Generate summary.

**Request:**
```json
{
  "text": "Long article text...",
  "max_len": 200
}
```

**Response:**
```json
{
  "summary": "Short summary..."
}
```

### AI

#### POST /api/ai/test

Test AI configuration.

**Request:**
```json
{
  "endpoint": "https://api.openai.com/v1",
  "api_key": "sk-...",
  "model": "gpt-3.5-turbo"
}
```

#### POST /api/ai/search

AI-powered search.

**Request:**
```json
{
  "config": {
    "endpoint": "https://api.openai.com/v1",
    "api_key": "sk-...",
    "model": "gpt-3.5-turbo"
  },
  "query": "Find articles about machine learning",
  "articles": [
    {
      "title": "Article 1",
      "content": "Content..."
    }
  ]
}
```

### Statistics

#### GET /api/statistics

Get reading statistics.

**Response:**
```json
{
  "total_articles": 100,
  "read_articles": 50,
  "favorite_articles": 10,
  "total_feeds": 5,
  "daily_stats": [
    {
      "date": "2024-01-01",
      "count": 10
    }
  ],
  "feed_stats": [
    {
      "feed_id": 1,
      "feed_title": "Example Feed",
      "count": 50
    }
  ],
  "category_stats": [
    {
      "category": "tech",
      "count": 30
    }
  ]
}
```

### Settings

#### GET /api/settings

Get all settings.

**Response:**
```json
{
  "language": "en-US",
  "theme": "auto",
  "update_interval": "30",
  "translation_enabled": "false"
}
```

#### POST /api/settings

Update settings.

**Request:**
```json
{
  "language": "zh-CN",
  "theme": "dark"
}
```

### Tags

#### GET /api/tags

Get all tags.

**Response:**
```json
[
  {
    "id": 1,
    "name": "tech",
    "color": "#3b82f6",
    "position": 0
  }
]
```

## Error Handling

All errors return a JSON response with an `error` field:

```json
{
  "error": "Error message"
}
```

Common HTTP status codes:
- `200`: Success
- `400`: Bad request
- `404`: Not found
- `500`: Internal server error
