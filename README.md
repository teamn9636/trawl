![trawl](img/logo.png)

# trawl

Scrape structured data from any website using LLM-powered extraction.

trawl lets you define *what* you want semantically — not with CSS selectors — and it figures out *how* to extract it. When a site redesigns, trawl re-derives the extraction strategy automatically. The LLM is called **once per site structure**, not once per page. Steady-state scraping is pure Go at full speed with zero API cost.

---

## Install

```bash
go install github.com/akdavidsson/trawl@latest
```

Or build from source:

```bash
git clone https://github.com/akdavidsson/trawl
cd trawl
go build -o trawl .
```

---

## Quickstart

```bash
export ANTHROPIC_API_KEY=sk-ant-...

# Extract product data as JSON
trawl "https://books.toscrape.com" --fields "title, price, rating, in_stock"

# Output as CSV
trawl "https://books.toscrape.com" --fields "title, price" --format csv

# Preview the extraction plan without extracting
trawl "https://books.toscrape.com" --fields "title, price" --plan
```

---

## Usage

```
trawl [url] [flags]
```

### Examples

```bash
# Simple field extraction
trawl "https://example.com/products" --fields "name, price, rating, url" --format json

# Use a YAML schema for precise control
trawl "https://example.com/products" --schema products.yaml --format csv

# Natural language query
trawl "https://example.com/products" --query "extract all product listings with names, prices, and stock status"

# Save to a file
trawl "https://example.com/products" --fields "name, price" --output products.json

# Streaming JSONL output, pipe to jq
trawl "https://news.example.com" --fields "headline, date, author" --format jsonl | jq '.headline'

# Re-use a previously derived strategy (no LLM call)
trawl "https://example.com/products" --strategy cached-strategy.json --format csv

# Verbose output to see the full pipeline
trawl "https://example.com/products" --fields "name, price" -v

# Custom headers and cookies
trawl "https://example.com/dashboard" --fields "metric, value" \
    --headers "Authorization: Bearer token123" \
    --cookie "session=abc123"
```

---

## Flags

### Input

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--fields` | `-f` | | Comma-separated field names to extract |
| `--query` | `-q` | | Natural language description of what to extract |
| `--schema` | `-s` | | Path to YAML schema file |

### Output

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | | `json` | Output format: `json`, `jsonl`, `csv`, `parquet` |
| `--output` | `-o` | stdout | Write output to file instead of stdout |

### Crawling

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--max-pages` | `-n` | `1` | Maximum pages to crawl |
| `--paginate` | | | Auto-detect and follow pagination |
| `--concurrency` | `-c` | `10` | Number of concurrent workers |
| `--delay` | | `1s` | Delay between requests to same domain |
| `--no-robots` | | | Ignore robots.txt (use responsibly) |
| `--js` | | | Enable headless browser for JS-rendered pages |
| `--timeout` | | `30s` | Per-request timeout |
| `--headers` | | | Custom headers (`"Key: Value"` format) |
| `--cookie` | | | Cookie string to include |

### Strategy

| Flag | Default | Description |
|------|---------|-------------|
| `--strategy` | | Path to a cached extraction strategy JSON file |
| `--plan` | | Dry run: show the LLM-derived extraction plan, don't extract |
| `--no-cache` | | Don't cache or use cached strategies |
| `--no-heal` | | Disable self-healing (don't re-derive on failure) |

### LLM

| Flag | Default | Description |
|------|---------|-------------|
| `--model` | `claude-sonnet-4-6` | Anthropic model to use |
| `--no-llm` | | Disable LLM, use heuristic extraction only |

### General

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Verbose output (show strategy derivation, health stats) |
| `--help` | `-h` | Help |

---

## How it works

```
URL ──► Fetch (Go) ──► Simplify HTML ──► LLM Strategy Derivation ──► Extraction Strategy
                                                                            │
                                                                            ▼
         Output (JSON/CSV/JSONL) ◄───────────────────── Apply Strategy via CSS Selectors (Go)
                                                                │
                                                         [Strategy fails?]
                                                                │
                                                       Re-derive from new HTML
```

### The pipeline in detail

1. **Fetch** the target URL with configurable headers, cookies, and timeouts.
2. **Simplify** the HTML: strip scripts, styles, data attributes, and noise. Compute a structural fingerprint of the DOM.
3. **Check cache**: if a strategy exists for this URL pattern + fingerprint, skip the LLM entirely.
4. **Derive strategy** via Anthropic API: send the simplified HTML and field descriptions, get back CSS selectors, attribute mappings, transforms, and fallback selectors.
5. **Extract** data using pure Go + goquery: apply CSS selectors to every matching item on the page, coerce values to the correct types.
6. **Monitor health**: track what percentage of fields were populated. If it drops below 70%, trigger self-healing.
7. **Output** results as JSON, JSONL, or CSV.

The LLM is called **once** to figure out the selectors. Every subsequent page with the same structure uses the cached strategy — pure Go, no API calls, no token cost.

### Extraction strategy

The LLM returns a JSON strategy like this:

```json
{
  "site_pattern": "https://example.com/products/*",
  "item_selector": "div.product-card",
  "fields": [
    {
      "name": "name",
      "selector": "h2.product-title",
      "attribute": "text",
      "type": "string",
      "fallbacks": ["h3.title", ".product-name"]
    },
    {
      "name": "price",
      "selector": "span.price",
      "attribute": "text",
      "transform": "parse_price",
      "type": "float"
    }
  ],
  "pagination": {
    "type": "next_link",
    "selector": "a.next-page",
    "has_more": "a.next-page"
  },
  "confidence": 0.95,
  "fingerprint": "a8f3e2b1..."
}
```

Each field has a primary CSS selector, an attribute to read (`text`, `href`, `src`, or any HTML attribute), an optional transform (`parse_price`, `parse_date`, `trim`), and fallback selectors for resilience.

### Self-healing

```
Extract page
     │
     ├── All fields populated ──────────── continue
     │
     ├── Some fields empty (< 70%) ─────── re-derive strategy via LLM
     │                                       └── use new strategy if it improves success rate
     │
     └── Total failure (0 items matched) ── re-derive strategy via LLM
                                              └── resume with new strategy
```

When a site redesigns, the structural fingerprint changes, the cached strategy is bypassed, and trawl automatically derives a new one. No manual intervention needed.

### Preview the plan

Use `--plan` to see what trawl will do without extracting:

```
$ trawl "https://example.com/products" --fields "name, price" --plan

Strategy for https://example.com/products
  Item selector: div.product-card
  Fields:
    name:                h2.product-title -> text (string)
    price:               span.price -> text -> parse_price (float)
  Pagination: a.next-page -> href (next_link)
  Confidence: 0.95
  Fingerprint: a8f3e2b1
  Items found: 24
```

---

## Schema files

For complex or recurring extractions, define a YAML schema:

```yaml
name: product_listing
url: "https://example.com/products/*"
fields:
  - name: product_name
    type: string
    description: "The product's display name"
  - name: price
    type: float
    description: "Price in local currency"
  - name: currency
    type: string
    description: "Currency code (USD, EUR, etc.)"
  - name: in_stock
    type: bool
    description: "Whether the item is available"
  - name: rating
    type: float
    nullable: true
    description: "Star rating out of 5"
```

Field descriptions are passed to the LLM to improve selector accuracy. Supported types: `string`, `int`, `float`, `bool`, `date`, `datetime`.

See the `examples/` directory for more schema files.

---

## Configuration

trawl reads configuration from the environment and an optional config file.

**Environment variable:**

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

**Config file** (`~/.trawl/config.yaml`):

```yaml
api_key: sk-ant-...
model: claude-sonnet-4-6
```

Environment variables take precedence over the config file. The `--model` flag takes precedence over both.

**Strategy cache** is stored in `~/.trawl/strategies/`. Strategies are keyed by URL pattern + page structure fingerprint. Use `--no-cache` to bypass.

---

## Output

- **stdout** — structured data only (JSON, JSONL, or CSV)
- **stderr** — warnings, verbose logs, and strategy derivation status

This makes trawl pipeline-friendly:

```bash
trawl "https://example.com/products" --fields "name, price" --format csv | csvkit | ...
trawl "https://example.com/products" --fields "name, price" --format jsonl | jq 'select(.price > 50)'
```

Type coercion is soft: if a value cannot be parsed to the target type, trawl emits a warning on stderr and falls back to the raw string (or `null` for nullable fields), rather than aborting.

---

## How trawl compares

| Tool | Approach | LLM? | Self-heals? | Speed |
|------|----------|------|-------------|-------|
| Scrapy | Hardcoded selectors | No | No | Fast |
| Playwright/Puppeteer | Hardcoded selectors | No | No | Medium |
| ScrapeGraphAI | Full LLM extraction | Every page | Inherent | Slow, expensive |
| Firecrawl | LLM page-by-page | Every page | Inherent | Slow, expensive |
| **trawl** | LLM strategy + Go extraction | Once per structure | Yes | Fast |

trawl uses the LLM for **intelligence** (figuring out the right selectors) and Go for **throughput** (applying them at scale). Competitors either skip the LLM (brittle) or use it on every page (slow and expensive).

---

## Requirements

- Go 1.24+
- `ANTHROPIC_API_KEY` (not required for `--plan` with `--strategy`)
