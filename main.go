package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const baseUrl = "https://www.muet.edu.pk"

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.3",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36 Edg/103.0.1264.6",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/117.",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.",
}

func main() {
	urls := make(chan []string)
	go func() {
		urls <- []string{baseUrl}
	}()
	seen := make(map[string]bool)
	time := 0

	for {
		for _, url := range <-urls {
			go func(url string) {
				if !seen[url] {
					fmt.Println("Visiting", url)
					urls <- crawl(url)
					fmt.Println("Done.")
					seen[url] = true
				}

			}(url)
		}
		time++
		if time > 5 {
			break
		}
	}

}

func resolveRelativePath(link string) (string, bool) {
	if strings.HasPrefix(link, "/") {
		link = baseUrl + link
	}
	bUrl, _ := url.Parse(baseUrl)
	linkUrl, _ := url.Parse(link)
	if bUrl.Host != linkUrl.Host {
		return "", false
	}
	return link, true
}

var semaphore = make(chan struct{}, 5)

func crawl(url string) []string {
	semaphore <- struct{}{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", randomUserAgent())
	resp, err := http.DefaultClient.Do(req) // T
	if err != nil {
		log.Fatal(err)
		return []string{}
	}
	<-semaphore
	urls := parseResponse(resp)
	foundUrls := []string{}
	for _, url := range urls {
		link, ok := resolveRelativePath(url)
		if ok {
			foundUrls = append(foundUrls, link)
		}

	}
	return foundUrls
}

func parseResponse(resp *http.Response) []string {
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	urls := []string{}
	doc.Find("a").Each(func(n int, s *goquery.Selection) {
		v, _ := s.Attr("href")
		urls = append(urls, v)
	})
	defer resp.Body.Close()
	return urls
}

func randomUserAgent() string {
	n := rand.Int() % len(userAgents)
	userAgent := userAgents[n]
	return userAgent
}
