package main

import (
	"sync"
	"time"
)

// Page is the base object of the search engine
type Page struct {
	Title        string
	EncodedURL   string // base64 url encoded url of page
	TermCount    int
	CreatedAt    time.Time
	IndexedTerms map[string]int
}

// IndexPageTerms will add terms to a page's term index
func (p *Page) IndexPageTerms(terms []string) {
	if p.IndexedTerms == nil { // NOTE: must recheck for nil
		p.IndexedTerms = make(map[string]int)
	}
	for _, term := range terms {
		p.IndexedTerms[term]++
		p.TermCount++
	}
}

// Pages is a slice of page objects that have been index.
type Pages struct {
	// use a RW mutex, since the search bits will be super read heavy
	sync.RWMutex
	// collection to store which pages have been visited and indexed
	internal map[string]*Page
}

// AddPage accepts a page object and addes it to the thread-safe collection
func (p *Pages) AddPage(newPage *Page) {
	p.Lock()
	if p.internal == nil {
		p.internal = make(map[string]*Page)
	}
	p.internal[newPage.EncodedURL] = newPage
	p.Unlock()
}

// GetPage looks for a page based on its encoded url
func (p *Pages) GetPage(encodedURL string) *Page {
	p.RLock()
	if p.internal != nil {
		page, exists := p.internal[encodedURL]
		p.RUnlock()
		if exists {
			return page
		}
		return nil
	}
	p.RUnlock()

	// acquire an exclusive lock
	p.Lock()
	if p.internal == nil {
		p.internal = make(map[string]*Page)
	}
	page, exists := p.internal[encodedURL]
	p.Unlock()
	if exists {
		return page
	}
	return nil
}

// Reset clears the collection of Pages
func (p *Pages) Reset() {
	// acquire an exclusive lock
	p.Lock()
	p.internal = make(map[string]*Page)
	p.Unlock()
}
