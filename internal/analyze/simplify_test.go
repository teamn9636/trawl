package analyze

import (
	"strings"
	"testing"
)

func TestSimplifyHTML(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
  <title>Test</title>
  <script>var x = 1;</script>
  <style>.foo { color: red; }</style>
</head>
<body>
  <div class="content" data-tracking="abc" onclick="track()">
    <h1>Hello</h1>
    <table>
      <tr><th>Name</th><th>Price</th></tr>
      <tr><td>Widget</td><td>$9.99</td></tr>
    </table>
  </div>
  <noscript>Enable JS</noscript>
</body>
</html>`)

	result, err := SimplifyHTML(html)
	if err != nil {
		t.Fatalf("SimplifyHTML error: %v", err)
	}

	// Scripts, styles, noscript should be removed
	if strings.Contains(result, "var x = 1") {
		t.Error("script content should be removed")
	}
	if strings.Contains(result, "color: red") {
		t.Error("style content should be removed")
	}
	if strings.Contains(result, "Enable JS") {
		t.Error("noscript content should be removed")
	}

	// Data attributes and event handlers should be stripped
	if strings.Contains(result, "data-tracking") {
		t.Error("data attributes should be stripped")
	}
	if strings.Contains(result, "onclick") {
		t.Error("event handlers should be stripped")
	}

	// Structural content should remain
	if !strings.Contains(result, "Hello") {
		t.Error("heading content should be preserved")
	}
	if !strings.Contains(result, "Widget") {
		t.Error("table content should be preserved")
	}
}

func TestFingerprint(t *testing.T) {
	html1 := []byte(`<html><body><div class="grid"><div class="card"><h2>A</h2></div></div></body></html>`)
	html2 := []byte(`<html><body><div class="grid"><div class="card"><h2>B</h2></div></div></body></html>`)
	html3 := []byte(`<html><body><ul><li>A</li><li>B</li></ul></body></html>`)

	fp1, err := Fingerprint(html1)
	if err != nil {
		t.Fatalf("Fingerprint error: %v", err)
	}
	fp2, err := Fingerprint(html2)
	if err != nil {
		t.Fatalf("Fingerprint error: %v", err)
	}
	fp3, err := Fingerprint(html3)
	if err != nil {
		t.Fatalf("Fingerprint error: %v", err)
	}

	// Same structure, different content => same fingerprint
	if fp1 != fp2 {
		t.Errorf("same structure should produce same fingerprint: %s vs %s", fp1, fp2)
	}

	// Different structure => different fingerprint
	if fp1 == fp3 {
		t.Error("different structure should produce different fingerprint")
	}
}
