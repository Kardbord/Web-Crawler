package crawler

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"sync"
	"time"
)

type UrlStats struct {
	Url string
	IsValid bool
	VisitCount uint64
}

type Crawler struct {
	stats map[string]UrlStats
	mu sync.RWMutex
}

func NewCrawler() Crawler {
	return Crawler{make(map[string]UrlStats), sync.RWMutex{}}
}

func (cr *Crawler) Crawl(url string, depth int) {
	ch := make(chan UrlStats)
	go cr.crawlRoutine(url, depth, ch)
	for stat := range ch {
		if _, found := cr.stats[stat.Url]; found {
			// newly visited site
			cr.mu.Lock()
			cr.stats[stat.Url] = UrlStats{stat.Url, stat.IsValid, stat.VisitCount}
			cr.mu.Unlock()
		} else {
			// site we've been to before
			cr.mu.RLock()
			tmp := cr.stats[stat.Url]
			tmp.VisitCount++
			cr.mu.RUnlock()
			cr.mu.Lock()
			cr.stats[stat.Url] = tmp
			cr.mu.Unlock()
		}
		time.Sleep(50 * time.Millisecond) // Give other writer threads a chance at the mutex
	}
}

func (cr *Crawler) hasVisited(url string) bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	if _, ok := cr.stats[url]; ok {
		return true
	}
	return false
}

func (cr *Crawler) incrVisitCount(url string) {
	if !cr.hasVisited(url) { return }
	
	cr.mu.Lock()
	defer cr.mu.Unlock()
	tmp := cr.stats[url]
	tmp.VisitCount++
	cr.stats[url] = tmp
}

func (cr *Crawler) crawlRoutine(url string, depth int, ch chan UrlStats) {
	
	wg := sync.WaitGroup{}
	
	// Declare internalCrawl function
	var internalCrawl func(url string, depth int, ch chan UrlStats)
	
	// Define internalCrawl function
	// Can't declare and define internalCrawl at the same time
	// because it is a recursive function.
	internalCrawl = func(url string, depth int, ch chan UrlStats){
		wg.Add(1)
		defer wg.Done()
		if depth <= 0 {
			return
		}
		fmt.Println("Crawling to", url, "at depth", depth)
		
		// Send our results back home
		resp, err := http.Get(url)
		if err != nil {
			ch <- UrlStats{url, false, 1}
			return
		}
		ch <- UrlStats{url, true, 1}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				fmt.Println("Failed to close response Body for ", url)
			}
		}()
		
		// Continue crawling
		tokenizer := html.NewTokenizer(resp.Body)
		for token := tokenizer.Next(); token != html.ErrorToken; token = tokenizer.Next() {
			if token == html.StartTagToken {
				t := tokenizer.Token()
				if t.Data == "a" {
					for _, a := range t.Attr {
						if a.Key == "href" {
							if strings.Index(a.Val, "https") != 0 { continue } // let's only follow https links
							if cr.hasVisited(url) {
								cr.incrVisitCount(url)
								return
							}
							go internalCrawl(a.Val, depth-1, ch)
						}
					}
				}
			}
		}
	}
	
	internalCrawl(url, depth, ch)
	
	wg.Wait()
	close(ch)
}
