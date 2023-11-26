package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)


var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
}


func randomUserAgent() string {
	rand.New(rand.NewSource(time.Now().Unix()))
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}

func discoverLinks(response *http.Response, baseURL string) []string {
	if response != nil {
		doc, _ := goquery.NewDocumentFromReader(response.Body)
		foundUrls := []string{}
		if doc != nil {
			doc.Find("a").Each(func(i int, s *goquery.Selection){
				res, _ := s.Attr("href")
				foundUrls = append(foundUrls, res)
			})
		}
		return foundUrls
	} else {
		return []string{}
	}
}

func getRequest(targetURL string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", randomUserAgent())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}

}

func checkRelative(href string, baseUrl string) string {
	if strings.HasPrefix(href, "/") {
		return fmt.Sprintf("%s%s", baseUrl, href)
	} else {
		return href
	}
}

func resolveRelativeLinks(href string, baseUrl string) (bool, string) {
	resultHref := checkRelative(href, baseUrl)
	baseParse, _ := url.Parse(baseUrl)
	resultParse, _ := url.Parse(resultHref)

	if baseParse != nil && resultParse != nil {
		if baseParse.Host == resultParse.Host {
			return true, resultHref
		} else {
			return false, ""
		}
	}
	return false, ""
}

var tokens = make(chan struct{}, 5)

func Crawl(targetURL string, baseURL string) []string {
	fmt.Println(targetURL)

	tokens <-struct{}{}
	resp, _ := getRequest(targetURL)
	<-tokens

	links := discoverLinks(resp, baseURL)
	foundUrls := []string{}

	for _, link := range links{
		ok, correctLink := resolveRelativeLinks(link, baseURL)
		if ok {
			if correctLink != "" {
				foundUrls = append(foundUrls, correctLink)
			}
		}
	}
	// ParseHTML(resp)
	return foundUrls
}

// func ParseHTML(response *http.Response) {
	// 
// }

func main() {
	worklist := make(chan []string)
	var n int
	n++
	baseDomain := "https://www.baidu.com"
	go func(){worklist <-[]string{"https://www.baidu.com"}}()
	seen := make(map[string]bool)

	for ;n>0 ; n-- {
		list := <-worklist
		for _, link := range list {
			if !seen[link] {
				seen[link] = true
				n++
				go func(link string, baseURL string) {
					foundLinks := Crawl(link, baseDomain)
					if foundLinks != nil {
						worklist <-foundLinks
					}
				}(link, baseDomain)
			}
		}

	}
}
