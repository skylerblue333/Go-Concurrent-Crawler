package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type CrawlResult struct {
	URL    string
	Status int
	Err    error
}

func crawl(url string) CrawlResult {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return CrawlResult{URL: url, Err: err}
	}
	defer resp.Body.Close()
	return CrawlResult{URL: url, Status: resp.StatusCode}
}

func CrawlAll(urls []string, concurrency int) []CrawlResult {
	sem := make(chan struct{}, concurrency)
	results := make([]CrawlResult, len(urls))
	var wg sync.WaitGroup

	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = crawl(u)
		}(i, url)
	}
	wg.Wait()
	return results
}

func main() {
	urls := []string{
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
	}
	results := CrawlAll(urls, 5)
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("ERROR %s: %v\n", r.URL, r.Err)
		} else {
			fmt.Printf("OK    %s -> %d\n", r.URL, r.Status)
		}
	}
}
