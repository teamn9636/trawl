package strategy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	anthropicMessagesURL = "https://api.anthropic.com/v1/messages"
	anthropicVersion     = "2023-06-01"
	deriveMaxTokens      = 2048
	deriveTimeout        = 60 * time.Second
)

const systemPrompt = `You are a web scraping strategy engine. Your only output is a valid JSON object.

Given simplified HTML from a web page and a list of fields the user wants to extract, derive a CSS-selector-based extraction strategy.

Respond ONLY with a JSON object matching this exact structure:
{
  "site_pattern": "URL pattern this strategy applies to (e.g. https://example.com/products/*)",
  "item_selector": "CSS selector that matches each repeating item/row on the page",
  "fields": [
    {
      "name": "field_name",
      "selector": "CSS selector relative to each item",
      "attribute": "text or href or src or any HTML attribute name",
      "transform": "optional: trim, parse_price, parse_date, parse_int, parse_float",
      "type": "string or int or float or bool or date",
      "fallbacks": ["alternative CSS selector 1", "alternative CSS selector 2"]
    }
  ],
  "pagination": {
    "type": "next_link or url_increment or load_more or none",
    "selector": "CSS selector for next page element",
    "url_pattern": "URL with {page} placeholder if url_increment",
    "has_more": "CSS selector that exists when more pages available"
  },
  "confidence": 0.95
}

Rules:
- item_selector MUST match the repeating container for each data item (e.g. each product card, table row, list item)
- field selectors are relative to the item_selector element
- Use "text" as the attribute to get the text content of an element
- Provide fallback selectors for resilience against minor DOM changes
- Set confidence between 0 and 1 based on how certain you are about the selectors
- For tables, item_selector should typically target tbody tr or similar
- If no pagination is detected, set pagination to null
- No markdown fences, no explanation, no text outside the JSON object
- Keep selectors as simple and resilient as possible (prefer class-based over positional)`

// DeriveRequest holds the inputs for strategy derivation.
type DeriveRequest struct {
	SimplifiedHTML string
	URL            string
	FieldNames     []string
	FieldDescs     map[string]string // optional field descriptions
	Model          string
	APIKey         string
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Derive calls the Anthropic API to derive an extraction strategy from HTML.
func Derive(ctx context.Context, req DeriveRequest) (*ExtractionStrategy, error) {
	userMsg := buildDerivePrompt(req)

	body := anthropicRequest{
		Model:     req.Model,
		MaxTokens: deriveMaxTokens,
		System:    systemPrompt,
		Messages:  []anthropicMessage{{Role: "user", Content: userMsg}},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpCtx, cancel := context.WithTimeout(ctx, deriveTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(httpCtx, http.MethodPost, anthropicMessagesURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("x-api-key", req.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("Anthropic API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Anthropic API")
	}

	return parseStrategyFromResponse(apiResp.Content[0].Text)
}

func buildDerivePrompt(req DeriveRequest) string {
	var sb strings.Builder

	sb.WriteString("Target URL: ")
	sb.WriteString(req.URL)
	sb.WriteString("\n\n")

	sb.WriteString("Fields to extract:\n")
	for _, name := range req.FieldNames {
		sb.WriteString("- ")
		sb.WriteString(name)
		if desc, ok := req.FieldDescs[name]; ok && desc != "" {
			sb.WriteString(": ")
			sb.WriteString(desc)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nSimplified HTML:\n```html\n")
	// Truncate HTML to avoid exceeding token limits
	html := req.SimplifiedHTML
	if len(html) > 50000 {
		html = html[:50000] + "\n... (truncated)"
	}
	sb.WriteString(html)
	sb.WriteString("\n```\n\nDerive the extraction strategy for this page.")

	return sb.String()
}

func extractJSON(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		lines := strings.SplitN(text, "\n", 2)
		if len(lines) == 2 {
			text = lines[1]
		}
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end < start {
		return text
	}
	return text[start : end+1]
}

func parseStrategyFromResponse(text string) (*ExtractionStrategy, error) {
	jsonStr := extractJSON(text)

	var s ExtractionStrategy
	if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
		return nil, fmt.Errorf("parsing strategy JSON: %w\nraw: %s", err, jsonStr)
	}

	if s.ItemSelector == "" {
		return nil, fmt.Errorf("strategy has empty item_selector")
	}
	if len(s.Fields) == 0 {
		return nil, fmt.Errorf("strategy has no field mappings")
	}

	return &s, nil
}
