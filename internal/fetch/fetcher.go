package fetch

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Options configures the HTTP fetcher.
type Options struct {
	Headers map[string]string
	Cookie  string
	Timeout time.Duration
}

// DefaultOptions returns sensible fetch defaults.
func DefaultOptions() Options {
	return Options{
		Timeout: 30 * time.Second,
	}
}

// Fetch downloads a URL and returns its body as bytes.
func Fetch(url string, opts Options) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; trawl/0.1; +https://github.com/akdavidsson/trawl)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}
	if opts.Cookie != "" {
		req.Header.Set("Cookie", opts.Cookie)
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
