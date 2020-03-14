package main

import (
	"flag"
	"fmt"
	"webCrawler/crawler"
)

func main() {
	
	urlArg := flag.String("url", "https://www.google.com", "The URL from which to start the crawl")
	depthArg := flag.Int("depth", 3, "How many layers deep each goroutine is allowed to crawl")
	verboseArg := flag.Bool("verbose", false, "Indicates whether or not the final report should be verbose")
	
	flag.Parse()
	
	if *depthArg < 1 {
		fmt.Println("Please provide a positive depth arg")
		flag.Usage()
		return
	}
	
	cr := crawler.NewCrawler()
	cr.Crawl(*urlArg, *depthArg, *verboseArg)
}

