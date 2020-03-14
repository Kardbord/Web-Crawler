package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
)

func main() {
	url := "https://google.com/"
	fmt.Println("Retrieving HTML from", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed to retrieve", url)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Failed to close response Body")
		}
	}()
	
	tokenizer := html.NewTokenizer(resp.Body)
	
	for token := tokenizer.Next(); token != html.ErrorToken; token = tokenizer.Next() {
		if token == html.StartTagToken {
			t := tokenizer.Token()
			if t.Data == "a" {
				fmt.Println("Found a link!")
			}
		}
	}
	
}
