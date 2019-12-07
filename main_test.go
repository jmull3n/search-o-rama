package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The tests here are primarily integration level
// since the goal of the crawler is a functional prototype.

// I'd do some rearranging for a production-grade test suite, once
// the api was dialed and we've iterated with Product on tuning
// the behavior to what they were thinking.

func crawlTestHelper(urlString string, terms *Terms, pages *Pages) bool {

	url := CrawlerInput{
		URLString:         urlString,
		PreviousURLString: "",
	}
	var termCount int
	depth := 3
	concurrency := 15
	crawledPages, termCount, _ := CrawlAndCatalog(url, concurrency, depth, pages, terms)

	s := fmt.Sprintf("pages crawled: %d\ntotal words found: %d\n", crawledPages, termCount)
	fmt.Println(s)

	results := terms.GetTermPages("http")
	for _, p := range results {
		fmt.Printf("%+v\n", p)
	}
	return true
}

func TestCrawl(t *testing.T) {
	var pages Pages
	var terms Terms
	urlString := "https://golang.org/"

	crawlTestHelper(urlString, &terms, &pages)

}

func TestConcurrentCrawls(t *testing.T) {
	var pages Pages
	var terms Terms
	urlStrings := []string{"https://golang.org/", "http://wikipedia.org", "https://pinpoint.com/"}

	resultChannel := make(chan bool)

	for _, url := range urlStrings {
		go func(urlString string) {
			resultChannel <- crawlTestHelper(urlString, &terms, &pages)
		}(url)
	}

	for i := 0; i < len(urlStrings); i++ {
		<-resultChannel
	}
}

func TestCrawlingError(t *testing.T) {

	var pages Pages
	var terms Terms

	urlString := "golang.org/"

	url := CrawlerInput{
		URLString:         urlString,
		PreviousURLString: "",
	}
	depth := 3
	concurrency := 5
	_, _, err := CrawlAndCatalog(url, concurrency, depth, &pages, &terms)

	if err == nil {
		t.Error("Expected error was not encountered")
	}
}

func TestCrawlEndpoint(t *testing.T) {

	testPayload := CrawlerInput{
		URLString:         "http://golang.org",
		PreviousURLString: "",
	}

	requestBody, err := json.Marshal(&testPayload)
	if err != nil {
		t.Error("error marshalling test request body")
	}
	req, err := http.NewRequest("POST", "/api/crawl", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(crawlPageHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestCrawlEndpointFailure(t *testing.T) {

	testPayload := CrawlerInput{
		URLString:         "golang.org",
		PreviousURLString: "",
	}

	requestBody, err := json.Marshal(&testPayload)
	if err != nil {
		t.Error("error marshalling test request body")
	}
	req, err := http.NewRequest("POST", "/api/crawl", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(crawlPageHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
