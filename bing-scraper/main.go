package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var bingDomains = map[string]string{
	"cn": "&cc=CN",
	"hk": "&cc=HK",
}

type SearchResult struct {
	ResultRank	int
	ResultURL 	string
	ResultTitle string
	ResultDesc 	string
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

func buildBingUrls(searchTerm, country string, pages, count int) ([]string, error) {
	toScrape := []string{}
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	if countryCode, found := bingDomains[country]; found {
		for i := 0 ; i < pages; i++ {
			first := firstParameter(i, count);
			scrapeURL := fmt.Sprintf("https://bing.com/search?q=%s&first=%d&count=%d%s", searchTerm, first, count, countryCode)
			toScrape = append(toScrape, scrapeURL)
		}
	} else {
		err := fmt.Errorf("country(%s) is currently not supported", country)
		return nil, err
	}

	return toScrape, nil
}

func firstParameter(number, count int) int {
	if number == 0 {
		return number + 1
	}
	return number*count + 1
}

func getScrapeClient(proxyString interface{}) (*http.Client, error) {
	switch V := proxyString.(type) {
		case string:
			proxyUrl, _ := url.Parse(V)
			return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}, nil
		default:
			return &http.Client{}, nil
	}
}

func scrapeClientRequest(searchURL string, proxyString interface{}) (*http.Response, error) {
	baseClient, err := getScrapeClient(proxyString)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent",randomUserAgent())
	res, err := baseClient.Do(req)
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	if res.StatusCode != 200 {
		err := fmt.Errorf("scraper received a non-200 status code suggesting a ban")
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}


func bingResultParser(response *http.Response, rank int) ([]SearchResult, error) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	results := []SearchResult{}
	sel := doc.Find("li.b_algo")
	rank++

	for i := range sel.Nodes{
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		titleTag := item.Find("h2")
		descTag := item.Find("div.b_caption p")
		desc := descTag.Text()
		title := titleTag.Text()
		link = strings.Trim(link, " ")
		if link == "" && link != "#" && !strings.HasPrefix(link, "/") {
			result := SearchResult{
				rank,
				link,
				title,
				desc,
			}
			results = append(results, result)
			rank++
		}
	}
	return results, nil
}


func BingScrape(searchTerm, country string, proxyString interface{}, pages, count, backoff int)([]SearchResult, error) {
	results := []SearchResult{}
	bingPages, err := buildBingUrls(searchTerm, country, pages, count)
	if err != nil {
		return nil, err
	}
	for _, page := range bingPages {
		rank := len(results)
		res, err := scrapeClientRequest(page, proxyString)
		if err != nil {
			return nil, err
		}
		data, err := bingResultParser(res, rank)
		if err != nil {
			return nil, err
		}
		for _, result := range data {
			results = append(results, result)
		}
		time.Sleep(time.Duration(backoff)*time.Second)
	}

	return results, nil
}

func main() {
	res, err := BingScrape("abcd", "cn", nil, 2, 3, 30)
	if err != nil {
		for _, res := range res{
			fmt.Println(res)
		}
	} else {
		fmt.Println(err)
	}
}
