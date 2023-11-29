package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SeoData struct {
	URL 			string
	Title			string
	H1				string
	MetaDescription string
	StatusCode 		int
}

type Parser interface{
	getSEOData(resp *http.Response) (SeoData, error)
}

type DefaultParser struct {

}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/119.0",
}

func randomUserAgent() string {
	rand.New(rand.NewSource(time.Now().Unix()))
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}

func isSitemap(urls []string)([]string, []string) {
	sitemapFiles := []string{}
	pages := []string{}
	for _, page := range urls {
		foundSitemap := strings.Contains(page, "xml")
		if foundSitemap == true {
			fmt.Println("Found Sitemap", page)
			sitemapFiles = append(sitemapFiles, page)
		} else {
			pages = append(pages, page)
		}
	}
	return sitemapFiles, pages
}

func extractSiteMapURLs(startURL string) []string {
	Worklist := make(chan []string)
	toCrawl := []string{}
	var n int
	n++
	go func(){Worklist <- []string{startURL}}()

	for ; n>0 ; n-- {
		list := <- Worklist
		for _, link := range list {
			n++
			go func(link string){
				response, err := makeRequest(link)
				if err != nil {
					log.Printf("error retrieving URL: %s with error: %v", link, err)
				}
				urls, err := extractURLs(response)
				if err != nil {
					log.Printf("error extracting document from response, URL: %s with error: %v", link, err)
				}
				sitemapFiles, pages := isSitemap(urls)
				if sitemapFiles != nil {
					Worklist <- sitemapFiles
				}
				for _, page := range pages {
					toCrawl = append(toCrawl, page)
				}
			}(link)
		}
	}

	return toCrawl
}

func makeRequest(url string) (*http.Response , error) {
	client := http.Client{
		Timeout: 10*time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", randomUserAgent())
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}


func scrapeURLs(urls []string, parser Parser, concurrency int) []SeoData {
	tokens := make(chan struct{}, concurrency)
	var n int
	n++
	worklist := make(chan []string)
	results := []SeoData{}

	go func(){worklist <- urls}()
	for ; n>0; n-- {
		list := <- worklist
		for _, url := range list {
			if url != "" {
				n++
				go func(url string, token chan struct{}){
					log.Printf("requesting URL: %s", url)
					res, err := scrapePage(url, token, parser)
					if err != nil {
						log.Printf("encountered error, URL: %s with error: %v", url, err)
					} else {
						results = append(results, res)
					}
				}(url, tokens)
			}
		}
	}
	return results
}

func extractURLs(response *http.Response) ([]string, error){
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}
	results := []string{}
	sel := doc.Find("loc")
	for i := range sel.Nodes{
		loc := sel.Eq(i)
		result := loc.Text()
		results = append(results, result)
	}
	return results, nil
}


func scrapePage(url string,  token chan struct{}, parser Parser) (SeoData, error) {
	res, err := crawlPage(url, token)
	if err != nil {
		return SeoData{}, err
	}
	data, err := parser.getSEOData(res)
	if err != nil {
		return SeoData{}, err
	}
	return data, nil
}

func crawlPage(url string, tokens chan struct{})(*http.Response, error) {
	tokens <- struct{}{}
	resp, err := makeRequest(url)
	<- tokens
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func(d DefaultParser) getSEOData(resp *http.Response) (SeoData, error){
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return SeoData{}, err
	}
	result := SeoData{}
	result.URL = resp.Request.URL.String()
	result.StatusCode = resp.StatusCode
	result.Title = doc.Find("title").First().Text()
	result.H1 = doc.Find("h1").First().Text()
	result.MetaDescription,_ = doc.Find("meta[name^=description]").Attr("content")
	return result, nil
}

func ScrapeSiteMap(url string, parser Parser, concurrency int) []SeoData {
	results := extractSiteMapURLs(url)
	res := scrapeURLs(results, parser, concurrency)
	return res
}

func main() {
	p := DefaultParser{}
	results := ScrapeSiteMap("https://www.quicksprout.com/sitemap.xml", p, 10)	
	for _, res := range results {
		fmt.Println(res)
	}
}
