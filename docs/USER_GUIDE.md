# OneRSS User Guide

## Table of Contents

- [Getting Started](#getting-started)
- [Adding Feeds](#adding-feeds)
- [Reading Articles](#reading-articles)
- [Using AI Features](#using-ai-features)
- [Integrations](#integrations)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [FAQ](#faq)

## Getting Started

### Docker Installation (Recommended)

```bash
# Pull and run
docker run -d \
  --name one-rss \
  -p 6011:6011 \
  -v one-rss-data:/data \
  one-rss:latest
```

Then open http://localhost:6011 in your browser.

### Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  one-rss:
    image: one-rss:latest
    container_name: one-rss
    ports:
      - "6011:6011"
    volumes:
      - one-rss-data:/data
    restart: unless-stopped

volumes:
  one-rss-data:
```

Run with:
```bash
docker-compose up -d
```

## Adding Feeds

### Manual Addition

1. Click the "+" button in the sidebar
2. Enter the RSS feed URL
3. Optionally select a category
4. Click "Add"

### Feed Discovery

1. Click the search icon in the sidebar
2. Enter a website URL
3. OneRSS will automatically discover RSS feeds
4. Select the feeds you want to add

### Supported Feed Formats

- RSS 2.0
- Atom
- RDF

## Reading Articles

### Article List

- **Normal View**: Shows title, summary, and metadata
- **Compact View**: Minimal display for scanning
- **Card View**: Large images with title overlay
- **Gallery View**: Image grid for visual feeds

### Article Actions

- **Favorite**: Star important articles
- **Read Later**: Save for later reading
- **Hide**: Remove from timeline
- **Translate**: Translate to your language
- **Summarize**: Generate AI summary
- **Chat**: Ask AI about the article

### Filters

- **All**: Show all articles
- **Unread**: Show only unread articles
- **Favorites**: Show starred articles
- **Read Later**: Show saved articles

## Using AI Features

### Translation

1. Open an article
2. Click the translate icon
3. Select target language
4. Translation appears inline

Supported providers:
- Google Translate (free)
- DeepL (API key required)
- AI Translation (OpenAI-compatible API)

### Summarization

1. Open an article
2. Click the summarize icon
3. Summary appears above content

Providers:
- Local (TF-IDF algorithm, no API needed)
- AI (OpenAI-compatible API)

### AI Chat

1. Open an article
2. Click the chat icon
3. Ask questions about the article
4. AI responds with contextual answers

### AI Search

1. Click the AI search button in the toolbar
2. Enter your query in natural language
3. AI searches across all articles
4. Results are ranked by relevance

## Integrations

### Obsidian

1. Go to Settings > Integrations
2. Enter your Obsidian vault path
3. Click "Export to Obsidian" on any article

### Notion

1. Go to Settings > Integrations
2. Enter your Notion API key
3. Enter the target page ID
4. Click "Export to Notion" on any article

### Zotero

1. Go to Settings > Integrations
2. Enter your Zotero API key
3. Enter your user ID
4. Click "Export to Zotero" on any article

### FreshRSS

1. Go to Settings > Integrations
2. Enter your FreshRSS server URL
3. Enter your credentials
4. Enable sync

### RSSHub

1. Go to Settings > Integrations
2. Enter your RSSHub endpoint
3. Use RSSHub routes when adding feeds

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+N` | Add feed |
| `Ctrl+,` | Open settings |
| `Ctrl+R` | Refresh feeds |
| `Ctrl+Shift+A` | Mark all as read |
| `↑` / `↓` | Navigate articles |
| `Enter` | Open article |
| `Escape` | Close article |
| `F` | Toggle favorite |
| `L` | Toggle read later |
| `H` | Toggle hide |
| `T` | Translate |
| `S` | Summarize |

## FAQ

### Q: How often are feeds refreshed?

By default, feeds are refreshed every 30 minutes. You can change this in Settings > General.

### Q: Can I use my own AI API?

Yes! Go to Settings > AI and enter your OpenAI-compatible API endpoint and key.

### Q: Is my data private?

Yes. All data is stored locally in the SQLite database. No data is sent to external servers unless you explicitly enable AI features.

### Q: Can I export my data?

Yes. You can export your feeds as OPML from Settings > Storage.

### Q: How do I backup my data?

Simply copy the `data/` folder (or the database file) to a safe location.

### Q: Can I use a proxy?

Yes. Go to Settings > Network and enable proxy support. Supports HTTP, HTTPS, and SOCKS5 proxies.
