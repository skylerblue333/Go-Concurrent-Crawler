package main

import (
	"context"
	"sync"
	"testing"
)

func TestCrawlerConcurrency(t *testing.T) {
	crawler := NewCrawler(5)
	wg := &sync.WaitGroup{}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go crawler.Crawl(ctx, "http://test.com", 2, wg)

	go func() {
		wg.Wait()
		close(crawler.results)
	}()

	count := 0
	for range crawler.results {
		count++
	}

	if count == 0 {
		t.Errorf("Expected crawler to yield results")
	}
}
