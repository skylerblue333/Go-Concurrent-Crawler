package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Crawler manages concurrent web crawling
type Crawler struct {
	visited sync.Map
	wg      sync.WaitGroup
	results chan string
}

func NewCrawler() *Crawler {
	return &Crawler{
		results: make(chan string, 100),
	}
}

func (c *Crawler) Crawl(url string, depth int) {
	if depth == 0 {
		return
	}
	if _, loaded := c.visited.LoadOrStore(url, true); loaded {
		return
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("Error fetching %s: %v", url, err)
			return
		}
		defer resp.Body.Close()

		c.results <- fmt.Sprintf("Crawled: %s [Status: %d]", url, resp.StatusCode)
	}()
}

func main() {
	crawler := NewCrawler()

	seedURLs := []string{
		"https://example.com",
		"https://httpbin.org/get",
	}

	for _, u := range seedURLs {
		crawler.Crawl(u, 1)
	}

	go func() {
		crawler.wg.Wait()
		close(crawler.results)
	}()

	for result := range crawler.results {
		fmt.Println(result)
	}
	fmt.Println("Crawl complete.")
}
