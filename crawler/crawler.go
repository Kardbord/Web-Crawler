package crawler

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"sync"
)

type UrlStats struct {
	Url string
	IsValid bool
	VisitCount uint64
}

type Crawler struct {
	stats map[string]UrlStats
}

func NewCrawler() Crawler {
	return Crawler{make(map[string]UrlStats)}
}

func (cr *Crawler) Crawl(url string, depth int) {
	ch := make(chan UrlStats)
	go cr.crawlRoutine(url, depth, ch)
	for stat := range ch {
		if _, found := cr.stats[stat.Url]; found {
			// newly visited site
			cr.stats[stat.Url] = UrlStats{stat.Url, stat.IsValid, stat.VisitCount}
		} else {
			// site we've been to before
			tmp := cr.stats[stat.Url]
			tmp.VisitCount++
			cr.stats[stat.Url] = tmp
		}
	}
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
