package main

import (
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// CrawlResult is a thread-safe counter for storing the results of a crawl/indexing operation
type CrawlResult struct {
	sync.RWMutex

	PagesCrawled int
	WordsIndexed int
	crawlErrors  []error
	internal     map[string]bool
}

// SetCrawlSummary sets the summary of a crawl
func (c *CrawlResult) SetCrawlSummary(pageTerms []string) {
	// acquire an exclusive lock
	c.Lock()
	if c.internal == nil { // NOTE: must recheck for nil
		c.internal = make(map[string]bool)
	}
	// add the terms to the crawl/indexing operation map
	for _, pageTerm := range pageTerms {
		if _, ok := c.internal[pageTerm]; !ok {
			c.internal[pageTerm] = true
		}
	}
	c.PagesCrawled++
	c.WordsIndexed = len(c.internal)

	c.Unlock()
}

// AddCrawlError sets the summary of a crawl
func (c *CrawlResult) AddCrawlError(err error) {
	// acquire an exclusive lock
	c.Lock()

	c.crawlErrors = append(c.crawlErrors, err)

	c.Unlock()
}

// GetErrors gets the errors of a crawl
func (c *CrawlResult) GetErrors() []error {
	// acquire an exclusive lock
	c.RLock()
	defer c.RUnlock()

	return c.crawlErrors

}

func getRawPage(urlString string) (*goquery.Document, error) {
	response, err := http.Get(urlString)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer response.Body.Close()
	// Create goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	return document, err
}
func cleanPage(doc *goquery.Document) {
	//strip scripts and other bits that muddle the finding of text
	doc.Find("script").Each(func(i int, el *goquery.Selection) {
		el.Remove()
	})
	doc.Find("style").Each(func(i int, el *goquery.Selection) {
		el.Remove()
	})
	// Lets not index the css!
	doc.Find("link").Each(func(i int, el *goquery.Selection) {
		el.Remove()
	})
}
func extractTitle(doc *goquery.Document) string {
	if doc != nil {
		title := doc.Find("head title").Text()
		return title
	}
	return ""
}

func extractText(doc *goquery.Document) []string {
	words := []string{}
	// clean lines to just words and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Error(err)
	}
	if doc != nil {
		doc.Find("*").Each(func(index int, element *goquery.Selection) {
			elementText := strings.TrimSpace(reg.ReplaceAllString(element.Contents().Text(), " "))
			elementWords := strings.Fields(elementText)
			words = append(words, elementWords...)
		})
		return words
	}
	return words
}

func extractLinks(doc *goquery.Document, depth int, input CrawlerInput) []string {
	links := []string{}
	if doc != nil {
		doc.Find("a").Each(func(index int, element *goquery.Selection) {
			href, exists := element.Attr("href")

			if exists {
				// attempt to scrub bad links. Plenty more to do here
				if href != "/" &&
					href != input.URLString &&
					href != input.PreviousURLString &&
					!strings.HasPrefix(href, "mailto") &&
					!strings.HasSuffix(href, ".zip") {
					links = append(links, href)
				}
			}
		})
		return links
	}
	return links
}

func resolveRelativeLinks(baseURL string, foundLinks []string) []string {
	internalUrls := []string{}

	for _, href := range foundLinks {
		if strings.HasPrefix(href, baseURL) {
			internalUrls = append(internalUrls, href)
		}

		if strings.HasPrefix(href, "/") {
			resolvedURL := fmt.Sprintf("%s%s", baseURL, href)
			internalUrls = append(internalUrls, resolvedURL)
		}
	}

	return internalUrls
}

func crawl(crawlerInput CrawlerInput, queue chan struct{}, depth int) ([]CrawlerInput, *Page, error) {
	queue <- struct{}{}
	log.Debug("Requesting: ", crawlerInput.URLString)
	doc, err := getRawPage(crawlerInput.URLString)

	baseURL := getBaseURL(crawlerInput.URLString)
	childLinks := []CrawlerInput{}

	links := extractLinks(doc, depth, crawlerInput)
	text := extractText(doc)
	title := extractTitle(doc)
	encodedURL := b64.URLEncoding.EncodeToString([]byte(crawlerInput.URLString))

	foundURLs := resolveRelativeLinks(baseURL, links)

	for _, foundURL := range foundURLs {
		newCrawlerInput := CrawlerInput{
			URLString:         foundURL,
			PreviousURLString: crawlerInput.URLString,
			CrawlDepth:        crawlerInput.CrawlDepth + 1,
		}
		childLinks = append(childLinks, newCrawlerInput)
	}
	page := &Page{
		EncodedURL: encodedURL,
		Title:      title,
		TermCount:  len(text),
		CreatedAt:  time.Now(),
	}
	page.IndexPageTerms(text)
	<-queue

	return childLinks, page, err
}

func getBaseURL(urlString string) string {
	parsed, _ := url.Parse(urlString)

	// This may have some issues creating valid URLs, but it'll at least work for a prototype
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
}

// CrawlAndCatalog does the main work of crawling a page and getting each page text
func CrawlAndCatalog(crawlerInput CrawlerInput, concurrency int, depth int, pages *Pages, terms *Terms) (int, int, []error) {
	var result CrawlResult
	// var crawlErrors []error
	// errorChan := make(chan error)
	worklist := make(chan []CrawlerInput)

	processed := make(map[string]bool)
	n := 1
	var queue = make(chan struct{}, concurrency)
	go func() {
		worklist <- []CrawlerInput{crawlerInput}
	}()

	for ; n > 0; n-- {
		list := <-worklist
		for _, link := range list {
			if link.CrawlDepth < depth {
				if _, exists := processed[link.URLString]; !exists {
					processed[link.URLString] = true
					n++
					go func(link CrawlerInput, queue chan struct{}, depth int, pages *Pages, terms *Terms, crawlResult *CrawlResult) {

						foundLinks, page, err := crawl(link, queue, depth)
						if err != nil {
							crawlResult.AddCrawlError(err)
						}
						if page != nil {
							pageTerms := make([]string, 0, len(page.IndexedTerms))

							// just get a list of the terms on the page
							for k := range page.IndexedTerms {
								pageTerms = append(pageTerms, k)
							}
							// index the terms / page combo
							pages.AddPage(page)
							terms.IndexPageTerms(pageTerms, page)

							// summarize crawl statistics
							crawlResult.SetCrawlSummary(pageTerms)
						}
						// keep running on any links that were found in the page
						if foundLinks != nil {
							worklist <- foundLinks
						}
					}(link, queue, depth, pages, terms, &result)
				}
			}
		}
	}
	return result.PagesCrawled, result.WordsIndexed, result.GetErrors()
}
