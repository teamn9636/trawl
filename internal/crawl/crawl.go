package crawl

import (
	"context"
	"fmt"
	"os"

	"github.com/akdavidsson/trawl/internal/analyze"
	"github.com/akdavidsson/trawl/internal/config"
	"github.com/akdavidsson/trawl/internal/extract"
	"github.com/akdavidsson/trawl/internal/fetch"
	"github.com/akdavidsson/trawl/internal/strategy"
)

// Options configures a crawl run.
type Options struct {
	URL         string
	FetchOpts   fetch.Options
	Strategy    *strategy.ExtractionStrategy // pre-loaded strategy (nil = derive)
	FieldNames  []string
	FieldDescs  map[string]string
	Model       string
	APIKey      string
	MaxPages    int
	NoCache     bool
	NoHeal      bool
	Verbose     bool
}

// Result holds the output of a full crawl run.
type Result struct {
	Strategy *strategy.ExtractionStrategy
	Extract  *extract.Result
	Pages    int
}

// Run executes the full crawl pipeline: fetch, analyze, derive/load strategy, extract.
func Run(ctx context.Context, opts Options) (*Result, error) {
	// 1. Fetch the page
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[fetch] %s\n", opts.URL)
	}
	html, err := fetch.Fetch(opts.URL, opts.FetchOpts)
	if err != nil {
		return nil, fmt.Errorf("fetching page: %w", err)
	}

	// 2. Get or derive strategy
	strat, err := resolveStrategy(ctx, opts, html)
	if err != nil {
		return nil, err
	}

	// 3. Apply strategy to extract data
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[extract] applying strategy (item_selector: %s)\n", strat.ItemSelector)
	}
	result, err := extract.Apply(strat, html)
	if err != nil {
		// Self-healing: if extraction fails completely, try re-deriving
		if !opts.NoHeal {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "[heal] extraction failed, re-deriving strategy...\n")
			}
			strat, err = deriveNewStrategy(ctx, opts, html)
			if err != nil {
				return nil, fmt.Errorf("re-derivation failed: %w", err)
			}
			result, err = extract.Apply(strat, html)
			if err != nil {
				return nil, fmt.Errorf("extraction after re-derivation: %w", err)
			}
		} else {
			return nil, fmt.Errorf("extraction: %w", err)
		}
	}

	// 4. Check health and self-heal if needed
	if !opts.NoHeal && result != nil {
		health := extract.ComputeHealth(result)
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "[health] %d records, %.0f%% fields populated\n",
				health.TotalRecords, health.SuccessRate())
		}
		if health.NeedsReInference(70) && health.TotalRecords > 0 {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "[heal] low success rate (%.0f%%), re-deriving strategy...\n", health.SuccessRate())
			}
			newStrat, err := deriveNewStrategy(ctx, opts, html)
			if err == nil {
				newResult, err := extract.Apply(newStrat, html)
				if err == nil {
					newHealth := extract.ComputeHealth(newResult)
					if newHealth.SuccessRate() > health.SuccessRate() {
						strat = newStrat
						result = newResult
						if opts.Verbose {
							fmt.Fprintf(os.Stderr, "[heal] improved to %.0f%% success rate\n", newHealth.SuccessRate())
						}
					}
				}
			}
		}
	}

	return &Result{
		Strategy: strat,
		Extract:  result,
		Pages:    1,
	}, nil
}

func resolveStrategy(ctx context.Context, opts Options, html []byte) (*strategy.ExtractionStrategy, error) {
	// If a strategy was pre-loaded (--strategy flag), use it
	if opts.Strategy != nil {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "[strategy] using provided strategy\n")
		}
		return opts.Strategy, nil
	}

	// Compute fingerprint for cache lookup
	fingerprint, err := analyze.Fingerprint(html)
	if err != nil {
		return nil, fmt.Errorf("fingerprinting page: %w", err)
	}

	// Try cache
	if !opts.NoCache {
		cached, err := strategy.LoadCached(opts.URL, fingerprint)
		if err != nil {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "[strategy] cache error: %v\n", err)
			}
		} else if cached != nil {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "[strategy] using cached strategy (fingerprint match)\n")
			}
			return cached, nil
		}
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "[strategy] no cached strategy found\n")
		}
	}

	// Derive new strategy
	strat, err := deriveNewStrategy(ctx, opts, html)
	if err != nil {
		return nil, err
	}
	strat.Fingerprint = fingerprint

	// Cache it
	if !opts.NoCache {
		if err := strategy.SaveCache(strat); err != nil {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "[strategy] cache write error: %v\n", err)
			}
		} else if opts.Verbose {
			fmt.Fprintf(os.Stderr, "[strategy] cached strategy\n")
		}
	}

	return strat, nil
}

func deriveNewStrategy(ctx context.Context, opts Options, html []byte) (*strategy.ExtractionStrategy, error) {
	// Validate API config
	cfg := &config.Config{APIKey: opts.APIKey, Model: opts.Model}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Simplify HTML for the LLM
	simplified, err := analyze.SimplifyHTML(html)
	if err != nil {
		return nil, fmt.Errorf("simplifying HTML: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[strategy] deriving via LLM (%s)...\n", opts.Model)
	}

	strat, err := strategy.Derive(ctx, strategy.DeriveRequest{
		SimplifiedHTML: simplified,
		URL:            opts.URL,
		FieldNames:     opts.FieldNames,
		FieldDescs:     opts.FieldDescs,
		Model:          opts.Model,
		APIKey:         opts.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("strategy derivation: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[strategy] derived (confidence: %.2f)\n", strat.Confidence)
	}

	return strat, nil
}
