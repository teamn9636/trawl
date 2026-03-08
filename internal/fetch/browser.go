package fetch

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// FetchWithBrowser uses a headless Chromium browser to fetch a URL,
// waiting for JavaScript to render the page before returning the HTML.
// Chromium is auto-downloaded to ~/.cache/rod/ on first use.
//
// If the page contains iframes (e.g. HuggingFace Spaces, embedded apps),
// it captures iframe content and returns the richest HTML found.
func FetchWithBrowser(url string, opts Options) ([]byte, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	u, err := launcher.New().Headless(true).Launch()
	if err != nil {
		return nil, fmt.Errorf("launching browser: %w", err)
	}

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return nil, fmt.Errorf("opening page: %w", err)
	}
	defer page.Close()

	// Wait for network idle (JS finished loading)
	if err := page.WaitIdle(timeout); err != nil {
		fmt.Fprintln(os.Stderr, "warning: page idle timeout, using partial render")
	}

	// Wait for DOM to stabilize — React/Next.js SPAs often render data after idle
	waitForDOMStable(page)

	// Auto-expand: click "Show more" / "Load more" / "View all" buttons to reveal hidden data
	expandContent(page)

	// Extra wait for SPAs that load data after idle
	if opts.WaitExtra > 0 {
		time.Sleep(opts.WaitExtra)
	}

	html, err := page.HTML()
	if err != nil {
		return nil, fmt.Errorf("getting rendered HTML: %w", err)
	}

	// Check if the main page has meaningful content or if it's mostly an iframe wrapper.
	// Sites like HuggingFace Spaces embed their actual app in an iframe — the outer
	// page HTML has almost no extractable data.
	iframeHTML := captureIframeContent(page)
	if iframeHTML != "" && len(iframeHTML) > len(html)/2 {
		// The iframe has substantial content; check if it has more data elements
		mainTags := countDataTags(html)
		iframeTags := countDataTags(iframeHTML)
		if iframeTags > mainTags {
			fmt.Fprintln(os.Stderr, "[fetch] using iframe content (richer than outer page)")
			return []byte(iframeHTML), nil
		}
	}

	return []byte(html), nil
}

// captureIframeContent finds iframes on the page and returns the HTML of the
// largest one. Returns empty string if no iframes or on error.
func captureIframeContent(page *rod.Page) string {
	iframes, err := page.Elements("iframe")
	if err != nil || len(iframes) == 0 {
		return ""
	}

	fmt.Fprintf(os.Stderr, "[fetch] found %d iframe(s), checking for content...\n", len(iframes))

	var best string
	for _, iframe := range iframes {
		frame, err := iframe.Frame()
		if err != nil {
			continue
		}

		// Always wait for the iframe to settle
		_ = frame.WaitIdle(10 * time.Second)

		html, err := frame.HTML()
		if err != nil {
			continue
		}
		if len(html) > len(best) {
			best = html
		}
	}
	return best
}

// waitForDOMStable polls the page until the DOM element count stops changing,
// indicating that client-side rendering (React, Vue, etc.) has finished.
func waitForDOMStable(page *rod.Page) {
	var lastCount int
	stableRounds := 0
	for i := 0; i < 20; i++ { // max 10 seconds (20 * 500ms)
		res, err := page.Eval(`() => document.querySelectorAll('*').length`)
		if err != nil {
			return
		}
		count := res.Value.Int()
		if count == lastCount && count > 0 {
			stableRounds++
			if stableRounds >= 2 {
				return // DOM stable for 1+ second
			}
		} else {
			stableRounds = 0
		}
		lastCount = count
		time.Sleep(500 * time.Millisecond)
	}
}

// expandContent finds and clicks "Show more" / "Load more" / "View all" buttons
// to reveal hidden data. Repeats up to 3 rounds in case clicking one button reveals
// another. This is a generalizable pattern — many sites paginate data inline.
func expandContent(page *rod.Page) {
	// JavaScript that finds and clicks expand-type buttons/links.
	// Returns the number of buttons clicked.
	const expandJS = `() => {
		const patterns = [
			/show\s*(more|all)/i,
			/load\s*(more|all)/i,
			/view\s*(more|all)/i,
			/see\s*(more|all)/i,
			/expand/i,
			/more\s*results/i,
		];
		let clicked = 0;
		const clickable = document.querySelectorAll('button, a, [role="button"], [onclick]');
		for (const el of clickable) {
			const text = (el.textContent || '').trim();
			if (text.length > 50) continue;
			for (const pat of patterns) {
				if (pat.test(text)) {
					try {
						el.click();
						clicked++;
					} catch(e) {}
					break;
				}
			}
		}
		return clicked;
	}`

	for round := 0; round < 3; round++ {
		res, err := page.Eval(expandJS)
		if err != nil {
			return
		}
		clicked := res.Value.Int()
		if clicked == 0 {
			return
		}
		fmt.Fprintf(os.Stderr, "[fetch] clicked %d expand button(s)\n", clicked)
		// Wait for new content to load
		time.Sleep(1 * time.Second)
		waitForDOMStable(page)
	}
}

// countDataTags counts HTML elements that typically contain extractable data
// (tables, lists, repeated divs with content). Used as a heuristic to decide
// whether iframe content is richer than the outer page.
func countDataTags(html string) int {
	lower := strings.ToLower(html)
	count := 0
	for _, tag := range []string{"<table", "<tbody", "<tr", "<td", "<th", "<li", "<article"} {
		count += strings.Count(lower, tag)
	}
	return count
}
