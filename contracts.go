// All the contracts for interacting with the REST api live here
package main

// CrawlerInput is the input to a crawl, contain the target and previous URLs
type CrawlerInput struct {
	URLString, PreviousURLString string
	CrawlDepth                   int
}

// CrawlReponse is the output datastructure of a crawl request
type CrawlReponse struct {
	DurationSeconds float64
	PagesCrawled    int
	WordsIndexed    int
	CrawlErrors     []string
}

// SearchRequest is the post body of a search request
type SearchRequest struct {
	Term string
}

// SearchResponse is the output of a search request
type SearchResponse struct {
	DurationSeconds float64
	Results         []SearchResult
}

// SearchResult is used to store the results of a search
type SearchResult struct {
	URL, Term, Title string
	Count            int
}
