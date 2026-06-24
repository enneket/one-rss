package rsshub

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Proxy struct {
	client    *http.Client
	rsshubURL string
}

func NewProxy() *Proxy {
	rsshubURL := os.Getenv("RSSHUB_URL")
	if rsshubURL == "" {
		rsshubURL = "http://rsshub:1200"
	}
	return &Proxy{
		client:    &http.Client{Timeout: 60 * time.Second},
		rsshubURL: rsshubURL,
	}
}

func (p *Proxy) HandleProxy(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/rsshub")
	if path == "" {
		path = "/"
	}

	targetURL := p.rsshubURL + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	// 转发原始请求的 Header
	for key, values := range r.Header {
		// 跳过 Host 和 Connection 相关 Header
		keyLower := strings.ToLower(key)
		if keyLower == "host" || keyLower == "connection" || keyLower == "keep-alive" {
			continue
		}
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	// 确保设置 User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "OneRSS/1.0")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("RSSHub error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 转发响应 Header
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
