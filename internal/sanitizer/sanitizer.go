package sanitizer

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

// SanitizeHTML 清理 HTML 内容，移除危险元素
func SanitizeHTML(content string) string {
	if content == "" {
		return ""
	}

	// 使用 UGC 策略，允许用户生成内容中的常见 HTML
	p := bluemonday.UGCPolicy()

	// 允许常用 HTML 元素
	p.AllowElements("p", "br", "hr", "h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowElements("div", "span", "section", "article", "aside", "header", "footer", "main", "nav")
	p.AllowElements("ul", "ol", "li", "dl", "dt", "dd")
	p.AllowElements("table", "thead", "tbody", "tfoot", "tr", "th", "td", "caption", "colgroup", "col")
	p.AllowElements("a", "strong", "em", "b", "i", "u", "s", "strike", "del", "ins", "sub", "sup", "mark", "small", "big")
	p.AllowElements("pre", "code", "kbd", "samp", "var", "blockquote", "cite", "q")
	p.AllowElements("figure", "figcaption", "picture", "source")
	p.AllowElements("details", "summary", "dialog")
	p.AllowElements("abbr", "acronym", "address", "dfn", "time")
	p.AllowElements("ruby", "rt", "rp", "rb", "rtc")

	// 允许多媒体元素
	p.AllowElements("img", "video", "audio", "iframe")

	// 允许图片属性
	p.AllowAttrs("src", "alt", "width", "height", "loading", "decoding", "srcset", "sizes").OnElements("img")
	p.AllowAttrs("src", "poster", "controls", "preload", "loop", "muted", "playsinline", "width", "height").OnElements("video")
	p.AllowAttrs("src", "controls", "preload", "loop", "muted").OnElements("audio")
	p.AllowAttrs("src", "width", "height", "frameborder", "allowfullscreen", "allow", "sandbox", "loading").OnElements("iframe")

	// 允许链接属性
	p.AllowAttrs("href", "target", "rel", "title", "type").OnElements("a")

	// 允许表格属性
	p.AllowAttrs("colspan", "rowspan", "headers", "scope", "abbr").OnElements("td", "th")
	p.AllowAttrs("span", "align", "valign", "width").OnElements("col", "colgroup")
	p.AllowAttrs("align", "valign").OnElements("tr", "tbody", "thead", "tfoot")

	// 允许全局属性
	p.AllowAttrs("dir", "lang", "title").OnElements("div", "span", "p", "table", "td", "th", "a", "img")
	p.AllowAttrs("align", "valign", "width", "height").OnElements("div", "table", "img")
	p.AllowAttrs("type", "start", "value", "reversed").OnElements("ol", "ul", "li")
	p.AllowAttrs("open").OnElements("details")
	p.AllowAttrs("datetime", "pubdate").OnElements("time")
	p.AllowAttrs("cite").OnElements("blockquote", "q", "del", "ins")
	p.AllowAttrs("cols", "rows", "disabled", "placeholder", "readonly").OnElements("textarea")

	// 允许 data-* 属性
	p.AllowDataAttributes()

	// 允许常用 aria 属性
	p.AllowAttrs("aria-label", "aria-hidden", "aria-expanded", "aria-controls", "aria-labelledby", "aria-describedby").OnElements("div", "span", "button", "input", "a", "img", "section", "article")

	return p.Sanitize(content)
}

// FixRelativeURLs 将相对 URL 转换为绝对 URL
func FixRelativeURLs(content, baseURL string) string {
	if content == "" || baseURL == "" {
		return content
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return content
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return content
	}

	// 处理 img src
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absURL := resolveURL(base, src); absURL != "" {
				s.SetAttr("src", absURL)
			}
		}
		if srcset, exists := s.Attr("srcset"); exists {
			s.SetAttr("srcset", fixSrcset(base, srcset))
		}
	})

	// 处理 a href
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if absURL := resolveURL(base, href); absURL != "" {
				s.SetAttr("href", absURL)
			}
		}
	})

	// 处理 video/audio source
	doc.Find("video, audio, source").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absURL := resolveURL(base, src); absURL != "" {
				s.SetAttr("src", absURL)
			}
		}
		if poster, exists := s.Attr("poster"); exists {
			if absURL := resolveURL(base, poster); absURL != "" {
				s.SetAttr("poster", absURL)
			}
		}
	})

	// 处理 iframe
	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absURL := resolveURL(base, src); absURL != "" {
				s.SetAttr("src", absURL)
			}
		}
	})

	// 处理 picture/source
	doc.Find("source").Each(func(i int, s *goquery.Selection) {
		if srcset, exists := s.Attr("srcset"); exists {
			s.SetAttr("srcset", fixSrcset(base, srcset))
		}
	})

	html, err := doc.Html()
	if err != nil {
		return content
	}

	// 移除 goquery 添加的 html/head/body 标签
	html = cleanGoqueryOutput(html)

	return html
}

// ExtractThumbnail 从内容中智能提取缩略图
func ExtractThumbnail(content string, existingThumbnail string) string {
	if existingThumbnail != "" {
		return existingThumbnail
	}

	if content == "" {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return ""
	}

	// 1. 查找 og:image 或 twitter:image
	if ogImage, exists := doc.Find("meta[property='og:image']").Attr("content"); exists && ogImage != "" {
		return ogImage
	}
	if twImage, exists := doc.Find("meta[name='twitter:image']").Attr("content"); exists && twImage != "" {
		return twImage
	}

	// 2. 查找文章中的第一张图片
	var thumbnail string
	doc.Find("article img, .content img, .post img, .entry img, main img").Each(func(i int, s *goquery.Selection) {
		if thumbnail != "" {
			return
		}
		src, exists := s.Attr("src")
		if exists && src != "" && !isTinyImage(s) {
			thumbnail = src
		}
	})

	// 3. 如果没找到，查找任意图片
	if thumbnail == "" {
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			if thumbnail != "" {
				return
			}
			src, exists := s.Attr("src")
			if exists && src != "" && !isTinyImage(s) {
				thumbnail = src
			}
		})
	}

	return thumbnail
}

// ExtractContentBySelector 使用 CSS 选择器提取内容
func ExtractContentBySelector(html, cssSelector string) (string, error) {
	if html == "" || cssSelector == "" {
		return html, nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	// 支持多个选择器（逗号分隔）
	selectors := strings.Split(cssSelector, ",")
	var result strings.Builder

	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		if selector == "" {
			continue
		}

		selection := doc.Find(selector)
		selection.Each(func(i int, s *goquery.Selection) {
			html, err := s.Html()
			if err == nil {
				result.WriteString(html)
				result.WriteString("\n")
			}
		})
	}

	if result.Len() == 0 {
		return "", nil
	}

	return result.String(), nil
}

// FilterContent 使用 CSS 选择器移除不需要的元素
func FilterContent(content, filterSelector string) string {
	if content == "" || filterSelector == "" {
		return content
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return content
	}

	// 支持多个选择器（逗号分隔）
	selectors := strings.Split(filterSelector, ",")
	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		if selector != "" {
			doc.Find(selector).Remove()
		}
	}

	html, err := doc.Html()
	if err != nil {
		return content
	}

	return cleanGoqueryOutput(html)
}

// ProcessEnclosures 统一处理附件（图片/音频/视频）
func ProcessEnclosures(content string, enclosures []Enclosure) string {
	if len(enclosures) == 0 {
		return content
	}

	var extra strings.Builder

	for _, enc := range enclosures {
		if enc.URL == "" {
			continue
		}

		// 检查内容中是否已包含该 URL
		if strings.Contains(content, enc.URL) {
			continue
		}

		medium := strings.ToLower(enc.Medium)
		mimeType := strings.ToLower(enc.Type)

		switch {
		case medium == "image" || strings.HasPrefix(mimeType, "image"):
			extra.WriteString(`<figure class="enclosure">`)
			extra.WriteString(`<p class="enclosure-content"><img src="`)
			extra.WriteString(enc.URL)
			extra.WriteString(`" alt="`)
			extra.WriteString(enc.Title)
			extra.WriteString(`" loading="lazy" /></p>`)
			extra.WriteString("</figure>\n")

		case medium == "audio" || strings.HasPrefix(mimeType, "audio"):
			extra.WriteString(`<figure class="enclosure">`)
			extra.WriteString(`<p class="enclosure-content"><audio preload="none" src="`)
			extra.WriteString(enc.URL)
			extra.WriteString(`" controls`)
			if mimeType != "" {
				extra.WriteString(`" type="`)
				extra.WriteString(mimeType)
			}
			extra.WriteString(`"></audio></p>`)
			extra.WriteString("</figure>\n")

		case medium == "video" || strings.HasPrefix(mimeType, "video"):
			extra.WriteString(`<figure class="enclosure">`)
			extra.WriteString(`<p class="enclosure-content"><video preload="none" src="`)
			extra.WriteString(enc.URL)
			extra.WriteString(`" controls`)
			if mimeType != "" {
				extra.WriteString(`" type="`)
				extra.WriteString(mimeType)
			}
			extra.WriteString(`"></video></p>`)
			extra.WriteString("</figure>\n")
		}
	}

	if extra.Len() > 0 {
		content += "\n" + extra.String()
	}

	return content
}

// Enclosure 附件信息
type Enclosure struct {
	URL    string
	Type   string // MIME type
	Medium string // image, audio, video
	Title  string
}

// 辅助函数

func resolveURL(base *url.URL, relative string) string {
	if relative == "" {
		return ""
	}

	// 已经是绝对 URL
	if strings.HasPrefix(relative, "http://") || strings.HasPrefix(relative, "https://") {
		return relative
	}

	// data: URL
	if strings.HasPrefix(relative, "data:") {
		return ""
	}

	rel, err := url.Parse(relative)
	if err != nil {
		return ""
	}

	return base.ResolveReference(rel).String()
}

func fixSrcset(base *url.URL, srcset string) string {
	if srcset == "" {
		return ""
	}

	parts := strings.Split(srcset, ",")
	var result []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}

		// 第一个字段是 URL
		if absURL := resolveURL(base, fields[0]); absURL != "" {
			fields[0] = absURL
		}

		result = append(result, strings.Join(fields, " "))
	}

	return strings.Join(result, ", ")
}

func isTinyImage(s *goquery.Selection) bool {
	width, _ := s.Attr("width")
	height, _ := s.Attr("height")

	// 小于 50x50 的图片可能是图标
	if width != "" && height != "" {
		w := parseInt(width)
		h := parseInt(height)
		if w > 0 && w < 50 && h > 0 && h < 50 {
			return true
		}
	}

	// 检查 class 是否包含 icon 相关
	class, _ := s.Attr("class")
	class = strings.ToLower(class)
	if strings.Contains(class, "icon") || strings.Contains(class, "avatar") || strings.Contains(class, "emoji") {
		return true
	}

	return false
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// 简单的整数解析
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}

var bodyTagRegex = regexp.MustCompile(`(?i)</?body[^>]*>`)
var htmlTagRegex = regexp.MustCompile(`(?i)</?html[^>]*>`)
var headTagRegex = regexp.MustCompile(`(?i)<head[^>]*>.*?</head>`)

func cleanGoqueryOutput(html string) string {
	// 移除 goquery 添加的 html/head/body 标签
	html = headTagRegex.ReplaceAllString(html, "")
	html = htmlTagRegex.ReplaceAllString(html, "")
	html = bodyTagRegex.ReplaceAllString(html, "")
	return strings.TrimSpace(html)
}
