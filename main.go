package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const maxBodyBytes = 1 << 20

var titlePattern = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

type CrawlRequest struct {
	URLs    []string `json:"urls"`
	Workers int      `json:"workers"`
}

type CrawlResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Title      string `json:"title,omitempty"`
	Error      string `json:"error,omitempty"`
}

type URLValidator func(string) (*url.URL, error)

type Crawler struct {
	client   *http.Client
	validate URLValidator
}

func NewCrawler(client *http.Client) *Crawler {
	return &Crawler{client: client, validate: validatePublicURL}
}

func validatePublicURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("invalid http(s) URL")
	}
	if parsed.User != nil {
		return nil, errors.New("URL userinfo is not allowed")
	}
	ips, err := net.LookupIP(parsed.Hostname())
	if err != nil || len(ips) == 0 {
		return nil, errors.New("host resolution failed")
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
			return nil, errors.New("private or special-use destination is not allowed")
		}
	}
	return parsed, nil
}

func normalizeTitle(body []byte) string {
	match := titlePattern.FindSubmatch(body)
	if len(match) < 2 {
		return ""
	}
	title := strings.Join(strings.Fields(string(match[1])), " ")
	if len(title) > 200 {
		title = title[:200]
	}
	return title
}

func (c *Crawler) Fetch(ctx context.Context, raw string) CrawlResult {
	parsed, err := c.validate(raw)
	if err != nil {
		return CrawlResult{URL: raw, Error: err.Error()}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return CrawlResult{URL: raw, Error: "request creation failed"}
	}
	req.Header.Set("User-Agent", "SkyCrawler/0.1 (+engineering-beta)")
	resp, err := c.client.Do(req)
	if err != nil {
		return CrawlResult{URL: raw, Error: "fetch failed"}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return CrawlResult{URL: raw, StatusCode: resp.StatusCode, Error: "response read failed"}
	}
	if len(body) > maxBodyBytes {
		return CrawlResult{URL: raw, StatusCode: resp.StatusCode, Error: "response body exceeds 1 MiB limit"}
	}
	return CrawlResult{URL: raw, StatusCode: resp.StatusCode, Title: normalizeTitle(body)}
}

func (c *Crawler) Crawl(ctx context.Context, urls []string, workers int) []CrawlResult {
	if workers < 1 {
		workers = 1
	}
	if workers > 16 {
		workers = 16
	}
	jobs := make(chan string)
	results := make(chan CrawlResult, len(urls))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for raw := range jobs {
				results <- c.Fetch(ctx, raw)
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, raw := range urls {
			select {
			case jobs <- raw:
			case <-ctx.Done():
				return
			}
		}
	}()
	wg.Wait()
	close(results)
	collected := make([]CrawlResult, 0, len(urls))
	for result := range results {
		collected = append(collected, result)
	}
	return collected
}

type Server struct {
	crawler *Crawler
	logger  *slog.Logger
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "sky-crawler"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /api/v1/crawl", s.handleCrawl)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *Server) handleCrawl(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var input CrawlRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(input.URLs) == 0 || len(input.URLs) > 32 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "urls must contain 1 to 32 entries"})
		return
	}
	if input.Workers < 1 || input.Workers > 16 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "workers must be between 1 and 16"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	results := s.crawler.Crawl(ctx, input.URLs, input.Workers)
	s.logger.Info("crawl completed", "requested", len(input.URLs), "returned", len(results))
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			_, err := validatePublicURL(req.URL.String())
			return err
		},
	}
	server := &Server{crawler: NewCrawler(client), logger: logger}
	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           server.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	logger.Info("starting crawler service", "addr", httpServer.Addr)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", fmt.Sprint(err))
		os.Exit(1)
	}
}
