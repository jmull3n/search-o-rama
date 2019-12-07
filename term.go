package main

import (
	b64 "encoding/base64"
	"sort"
	"sync"
)

// Terms is a map of a term (eg "cat") and a []*Page object
type Terms struct {
	// use a RW mutex, since the search bits will be super read heavy
	sync.RWMutex
	// collection to store which terms match to given page ids
	internal map[string]map[string]*Page
}


// IndexPageTerms updates the inverse-index for allowing term-based search
func (t *Terms) IndexPageTerms(terms []string, page *Page) {
	// acquire an exclusive lock
	t.Lock()
	if t.internal == nil { // NOTE: must recheck for nil
		t.internal = make(map[string]map[string]*Page)
	}
	for _, term := range terms {
		if t.internal[term] == nil {
			t.internal[term] = make(map[string]*Page)
		}
		t.internal[term][page.EncodedURL] = page
	}

	t.Unlock()
}

// GetTermPages returns a slice of Pages for a given term
func (t *Terms) GetTermPages(term string) []SearchResult {
	// acquire an exclusive lock
	t.RLock()
	if t.internal == nil { // NOTE: must recheck for nil
		t.internal = make(map[string]map[string]*Page)
	}
	pages, exists := t.internal[term]
	t.RUnlock()

	if exists {
		results := []SearchResult{}
		for _, p := range pages {
			count := p.IndexedTerms[term]
			decoded, _ := b64.URLEncoding.DecodeString(p.EncodedURL)
			result := SearchResult{
				URL:   string(decoded),
				Count: count,
				Term:  term,
				Title: p.Title,
			}
			results = append(results, result)
		}
		sort.Slice(results[:], func(i, j int) bool {
			return results[i].Count > results[j].Count
		})
		return results
	}
	return nil
}

// Reset clears the collection of Pages
func (t *Terms) Reset() {
	// acquire an exclusive lock
	t.Lock()
	t.internal = make(map[string]map[string]*Page)
	t.Unlock()
}
