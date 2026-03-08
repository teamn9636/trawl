package crawl

import (
	"sync"

	"github.com/akdavidsson/trawl/internal/extract"
	"github.com/akdavidsson/trawl/internal/fetch"
	"github.com/akdavidsson/trawl/internal/strategy"
)

// WorkerPool manages concurrent page fetching and extraction.
type WorkerPool struct {
	Concurrency int
	FetchOpts   fetch.Options
	Strategy    *strategy.ExtractionStrategy

	mu      sync.Mutex
	results []*extract.Result
	errors  []error
}

// ProcessURLs fetches and extracts data from multiple URLs concurrently.
func (wp *WorkerPool) ProcessURLs(urls []string) ([]*extract.Result, []error) {
	sem := make(chan struct{}, wp.Concurrency)
	var wg sync.WaitGroup

	for _, u := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }()

			html, err := fetch.Fetch(url, wp.FetchOpts)
			if err != nil {
				wp.addError(err)
				return
			}

			result, err := extract.Apply(wp.Strategy, html)
			if err != nil {
				wp.addError(err)
				return
			}

			wp.addResult(result)
		}(u)
	}

	wg.Wait()
	return wp.results, wp.errors
}

func (wp *WorkerPool) addResult(r *extract.Result) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.results = append(wp.results, r)
}

func (wp *WorkerPool) addError(err error) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.errors = append(wp.errors, err)
}
