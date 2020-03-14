package crawler

import (
	"fmt"
	"golang.org/x/net/html"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type UrlStats struct {
	Url             string
	NoRetrievalErrs bool
	VisitCount      uint64
}

type Crawler struct {
	stats map[string]*UrlStats
	isCrawling bool
}

func NewCrawler() Crawler {
	return Crawler{make(map[string]*UrlStats), false}
}

func (cr *Crawler) Crawl(url string, depth int, verbose bool) {
	ch := make(chan UrlStats)
	if cr.isCrawling {
		fmt.Println("This Crawler is already on the prowl...")
		return
	}
	cr.isCrawling = true
	go cr.crawlRoutine(url, depth, ch)
	for stat := range ch {
		if _, found := cr.stats[stat.Url]; !found {
			// newly visited site
			cr.stats[stat.Url] = &UrlStats{stat.Url, stat.NoRetrievalErrs, stat.VisitCount}
		} else {
			// site we've been to before
			tmp := cr.stats[stat.Url]
			tmp.VisitCount++
			cr.stats[stat.Url] = tmp
		}
	}
	cr.isCrawling = false
	fmt.Println(cr.report(verbose))
}

func (cr *Crawler) report(verbose bool) string {
	if cr.isCrawling {
		return "Please wait for the Crawler to finish before generating a report"
	}
	
	deadLinkCount := 0
	urlsVisitedMoreThanOnce := 0
	verboseReport := ""
	for _, stat := range cr.stats {
		if !stat.NoRetrievalErrs { deadLinkCount++ }
		if stat.VisitCount > 1 { urlsVisitedMoreThanOnce++ }
		if verbose {
			verboseReport += fmt.Sprintln("URL:", stat.Url, "Visited:", stat.VisitCount, "Successfully Retrieved:", stat.NoRetrievalErrs)
		}
	}
	
	rv := fmt.Sprintln("Unique URLs visited:", len(cr.stats))
	rv += fmt.Sprintln("Unable to retrieve (#URLs):", deadLinkCount)
	rv += fmt.Sprintln("URLs visited more than once:", urlsVisitedMoreThanOnce)
	rv += verboseReport
	
	return rv
}

func (cr *Crawler) crawlRoutine(url string, depth int, ch chan UrlStats) {
	fmt.Println("Beginning web crawl at", url)
	fmt.Println("Crawling...")
	rand.Seed(time.Now().UTC().UnixNano()) // random seed for sanity output
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
		
		// Some output for sanity
		if rand.Int() % 5 == 0 {
			fmt.Printf(".")
		}
		
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
							if strings.Contains(a.Val, url) { continue } // let's only follow links outside of the current domain
							go internalCrawl(a.Val, depth-1, ch)
						}
					}
				}
			}
		}
	}
	
	internalCrawl(url, depth, ch)
	
	wg.Wait()
	fmt.Println()
	close(ch)
}
