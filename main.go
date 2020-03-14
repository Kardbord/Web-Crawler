package main

import (
	"webCrawler/crawler"
)

func main() {
	url := "https://www.pinkbike.com"
	cr := crawler.NewCrawler()
	cr.Crawl(url, 3)
}

