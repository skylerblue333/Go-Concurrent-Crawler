package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func allowTestURL(raw string) (*url.URL, error) { return url.Parse(raw) }

func TestCrawlerFetchAndTitle(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><title> Example   Page </title></html>"))
	}))
	defer upstream.Close()

	crawler := NewCrawler(upstream.Client())
	crawler.validate = allowTestURL
	result := crawler.Fetch(context.Background(), upstream.URL)
	if result.StatusCode != http.StatusOK || result.Title != "Example Page" || result.Error != "" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestCrawlerConcurrentBatch(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<title>x</title>"))
	}))
	defer upstream.Close()
	crawler := NewCrawler(upstream.Client())
	crawler.validate = allowTestURL
	urls := []string{upstream.URL + "/1", upstream.URL + "/2", upstream.URL + "/3"}
	results := crawler.Crawl(context.Background(), urls, 3)
	if len(results) != len(urls) {
		t.Fatalf("expected %d results, got %d", len(urls), len(results))
	}
}

func TestPrivateAndMalformedTargetsRejected(t *testing.T) {
	for _, raw := range []string{"file:///etc/passwd", "http://127.0.0.1", "http://localhost"} {
		if _, err := validatePublicURL(raw); err == nil {
			t.Fatalf("expected rejection for %s", raw)
		}
	}
}

func TestHandlerValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := &Server{crawler: NewCrawler(http.DefaultClient), logger: logger}
	cases := []string{
		`{"urls":[],"workers":1}`,
		`{"urls":["https://example.com"],"workers":0}`,
		`{"urls":["https://example.com"],"workers":17}`,
		`{"urls":["https://example.com"],"workers":1,"extra":true}`,
	}
	for _, raw := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawl", bytes.NewBufferString(raw))
		rr := httptest.NewRecorder()
		server.routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d", raw, rr.Code)
		}
	}
}
