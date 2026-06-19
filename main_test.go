package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCrawlAll(t *testing.T) {
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer s1.Close()

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer s2.Close()

	results := CrawlAll([]string{s1.URL, s2.URL}, 2)

	if results[0].Status != 200 {
		t.Errorf("expected 200, got %d", results[0].Status)
	}
	if results[1].Status != 404 {
		t.Errorf("expected 404, got %d", results[1].Status)
	}
}
