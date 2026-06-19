package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type CrawlResult struct {
	URL   string
	Title string
	Depth int
}

type Crawler struct {
	visited sync.Map
	results chan CrawlResult
	workers chan struct{} // Semaphore pattern
}

func NewCrawler(maxWorkers int) *Crawler {
	return &Crawler{
		results: make(chan CrawlResult, 1000),
		workers: make(chan struct{}, maxWorkers),
	}
}

func (c *Crawler) Crawl(ctx context.Context, url string, depth int, wg *sync.WaitGroup) {
	defer wg.Done()

	if depth <= 0 {
		return
	}

	if _, loaded := c.visited.LoadOrStore(url, true); loaded {
		return
	}

	// Acquire worker slot
	select {
	case c.workers <- struct{}{}:
	case <-ctx.Done():
		return
	}
	
	defer func() { <-c.workers }() // Release slot

	// Simulate network fetch and parse
	time.Sleep(50 * time.Millisecond)
	
	select {
	case c.results <- CrawlResult{URL: url, Title: fmt.Sprintf("Page %s", url), Depth: depth}:
	case <-ctx.Done():
		return
	}

	// Simulate discovering links
	links := []string{url + "/a", url + "/b"}
	for _, link := range links {
		wg.Add(1)
		go c.Crawl(ctx, link, depth-1, wg)
	}
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	crawler := NewCrawler(10)
	wg := &sync.WaitGroup{}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wg.Add(1)
	go crawler.Crawl(ctx, "http://example.com", 3, wg)

	// Close results channel when all workers finish
	go func() {
		wg.Wait()
		close(crawler.results)
	}()

	var collected []CrawlResult
	for res := range crawler.results {
		collected = append(collected, res)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "completed",
		"pages_crawled": len(collected),
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/crawl", handleStart)

	log.Println("Concurrent Crawler Engine running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
