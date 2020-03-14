package main

import (
	"webCrawler/crawler"
)

func main() {
	url := "https://www.reddit.com/r/Overwatch/comments/fhuvvb/i_noticed_something_interesting_in_junkrats/"
	cr := crawler.NewCrawler()
	cr.Crawl(url, 5)
}

