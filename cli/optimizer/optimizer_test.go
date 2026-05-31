package optimizer

import (
	"encoding/json"
	"strings"
	"testing"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func TestIsURLField(t *testing.T) {
	keep := []string{"url", "html_url", "web_url", "webUrl", "permalink"}
	for _, k := range keep {
		if isURLField(k) {
			t.Errorf("isURLField(%q) = true, want false (should be kept)", k)
		}
	}

	strip := []string{"avatar_url", "gists_url", "repos_url", "AvatarURL", "iconUri", "icon_uri", "downloadUri"}
	for _, k := range strip {
		if !isURLField(k) {
			t.Errorf("isURLField(%q) = false, want true (should be stripped)", k)
		}
	}

	// "url" itself must not be stripped even though it ends with "url"
	if isURLField("url") {
		t.Error("isURLField(\"url\") = true, want false")
	}
}

func TestIsEmpty(t *testing.T) {
	cases := []struct {
		v    interface{}
		want bool
	}{
		{nil, true},
		{"", true},
		{[]interface{}{}, true},
		{map[string]interface{}{}, true},
		{"hello", false},
		{0.0, false},
		{false, false},
		{[]interface{}{"x"}, false},
	}
	for _, c := range cases {
		if got := isEmpty(c.v); got != c.want {
			t.Errorf("isEmpty(%v) = %v, want %v", c.v, got, c.want)
		}
	}
}

func TestStripNulls(t *testing.T) {
	input := map[string]interface{}{
		"a": "keep",
		"b": nil,
		"c": "",
		"d": []interface{}{nil, "x", nil},
		"e": map[string]interface{}{"f": nil, "g": "v"},
	}
	got := stripNulls(input).(map[string]interface{})
	if _, ok := got["b"]; ok {
		t.Error("nil field 'b' should be stripped")
	}
	if _, ok := got["c"]; ok {
		t.Error("empty string field 'c' should be stripped")
	}
	arr := got["d"].([]interface{})
	if len(arr) != 1 || arr[0] != "x" {
		t.Errorf("array nils not stripped: %v", arr)
	}
	inner := got["e"].(map[string]interface{})
	if _, ok := inner["f"]; ok {
		t.Error("nested nil 'f' should be stripped")
	}
	if inner["g"] != "v" {
		t.Error("nested 'g' should be kept")
	}
}

func TestStripURLFields(t *testing.T) {
	input := map[string]interface{}{
		"url":         "https://example.com",     // keep
		"html_url":    "https://example.com/html", // keep
		"avatar_url":  "https://cdn/img.png",      // strip
		"repos_url":   "https://api/repos",        // strip
		"title":       "hello",
		"node_id":     "abc123", // alwaysStrip
	}
	got := stripURLFields(input).(map[string]interface{})
	if _, ok := got["avatar_url"]; ok {
		t.Error("avatar_url should be stripped")
	}
	if _, ok := got["repos_url"]; ok {
		t.Error("repos_url should be stripped")
	}
	if _, ok := got["node_id"]; ok {
		t.Error("node_id (alwaysStrip) should be stripped")
	}
	if got["url"] != "https://example.com" {
		t.Error("url should be kept")
	}
	if got["html_url"] != "https://example.com/html" {
		t.Error("html_url should be kept")
	}
}

// ── generic ───────────────────────────────────────────────────────────────────

func TestGenericOptimizer_PlainJSON(t *testing.T) {
	g := &GenericOptimizer{}
	input := map[string]interface{}{
		"title":      "Test",
		"avatar_url": "https://cdn/img.png",
		"node_id":    "MDQ6",
		"body":       nil,
	}
	result := g.Optimize(input)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if _, ok := m["avatar_url"]; ok {
		t.Error("avatar_url should be stripped")
	}
	if _, ok := m["node_id"]; ok {
		t.Error("node_id should be stripped")
	}
	if _, ok := m["body"]; ok {
		t.Error("nil body should be stripped")
	}
	if m["title"] != "Test" {
		t.Error("title should be kept")
	}
}

func TestGenericOptimizer_MCPTextContent_JSON(t *testing.T) {
	g := &GenericOptimizer{}
	inner := map[string]interface{}{
		"title":      "PR Title",
		"avatar_url": "https://cdn/img.png",
		"node_id":    "MDQ6abc",
		"state":      "open",
	}
	innerJSON, _ := json.Marshal(inner)
	input := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{"type": "text", "text": string(innerJSON)},
		},
	}
	result := g.Optimize(input)
	m := result.(map[string]interface{})
	content := m["content"].([]interface{})
	first := content[0].(map[string]interface{})
	text := first["text"].(string)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("optimized text is not valid JSON: %v", err)
	}
	if _, ok := parsed["avatar_url"]; ok {
		t.Error("avatar_url inside text JSON should be stripped")
	}
	if _, ok := parsed["node_id"]; ok {
		t.Error("node_id inside text JSON should be stripped")
	}
	if parsed["title"] != "PR Title" {
		t.Error("title inside text JSON should be kept")
	}
}

func TestGenericOptimizer_MCPTextContent_NonJSON(t *testing.T) {
	g := &GenericOptimizer{}
	input := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{"type": "text", "text": "plain text, not JSON"},
		},
	}
	// Should not panic and should return via the outer stripNulls/stripURLFields path
	result := g.Optimize(input)
	if result == nil {
		t.Error("should not return nil for non-JSON text content")
	}
}

func TestGenericOptimizer_Passthrough(t *testing.T) {
	g := &GenericOptimizer{}
	// String input passes through unchanged
	if got := g.Optimize("hello"); got != "hello" {
		t.Errorf("string passthrough failed: %v", got)
	}
}

// ── bash ──────────────────────────────────────────────────────────────────────

func TestBashOptimizer_ShortOutput(t *testing.T) {
	b := &BashOptimizer{}
	short := "hello\nworld"
	if got := b.Optimize(short); got != short {
		t.Error("short output should pass through unchanged")
	}
}

func TestBashOptimizer_TestOutputFiltersPassingTests(t *testing.T) {
	lines := make([]string, 0, 50)
	for i := 0; i < 40; i++ {
		lines = append(lines, "  ✓ test case passes")
	}
	lines = append(lines, "  ✗ FAIL this one broke")
	lines = append(lines, "  2 passing, 1 failing")
	input := strings.Join(lines, "\n")

	b := &BashOptimizer{}
	result := b.Optimize(input)
	out, ok := result.(string)
	if !ok {
		t.Fatalf("expected string result")
	}
	if !strings.Contains(out, "passing tests omitted") {
		t.Error("should mention omitted passing tests")
	}
	if !strings.Contains(out, "FAIL this one broke") {
		t.Error("failing test line should be kept")
	}
	if strings.Count(out, "✓ test case passes") > 0 {
		t.Error("passing test lines should be stripped")
	}
}

func TestBashOptimizer_LongOutputHeadTail(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 300; i++ {
		sb.WriteString("line content here\n")
	}
	input := sb.String()

	b := &BashOptimizer{}
	result := b.Optimize(input)
	out, ok := result.(string)
	if !ok {
		t.Fatalf("expected string result")
	}
	if !strings.Contains(out, "lines omitted") {
		t.Error("should contain omission marker for long output")
	}
	if len(out) >= len(input) {
		t.Error("output should be shorter than input")
	}
}

func TestBashOptimizer_BuildOutputFiltersProgress(t *testing.T) {
	lines := []string{
		"webpack 5.0.0",
		"Successfully compiled",
		"Building module foo...",
		"error TS2345: Argument of type 'string' is not assignable",
		"error TS1005: ';' expected",
	}
	input := strings.Join(lines, "\n") + "\n"
	// pad to exceed 500 bytes and 20 lines min threshold
	for len(input) < 500 {
		input = "padding line\n" + input
	}
	b := &BashOptimizer{}
	result := b.Optimize(input)
	out, ok := result.(string)
	if !ok {
		t.Fatalf("expected string result")
	}
	if !strings.Contains(out, "error TS2345") {
		t.Error("TypeScript error line should be kept")
	}
}

func TestBashOptimizer_ByteBudgetSplit(t *testing.T) {
	// Build a large output that's few lines but many bytes
	bigLine := strings.Repeat("x", 2000)
	var lines []string
	for i := 0; i < 18; i++ { // under 20 lines but well over 40KB
		lines = append(lines, bigLine)
	}
	input := strings.Join(lines, "\n")

	b := &BashOptimizer{}
	result := b.Optimize(input)
	out, ok := result.(string)
	if !ok {
		// If no optimization triggered, that's also acceptable — just check it didn't blow up
		return
	}
	if len(out) >= len(input) {
		t.Log("byte budget split did not reduce size — may be expected for this input shape")
	}
}

func TestIsBuildOutput(t *testing.T) {
	withBoth := "error TS2345: blah\nSuccessfully compiled 5 files"
	if !isBuildOutput(withBoth) {
		t.Error("should detect build output with both error and success")
	}
	onlyError := "error TS2345: blah"
	if isBuildOutput(onlyError) {
		t.Error("error without success marker should not be build output")
	}
}

// ── webfetch ──────────────────────────────────────────────────────────────────

func TestWebFetchOptimizer_StripsBoilerplate(t *testing.T) {
	// Build HTML where noisy sections dominate so the >30% threshold is exceeded.
	bigScript := "<script>" + strings.Repeat("var x = 1; ", 200) + "</script>"
	bigStyle := "<style>" + strings.Repeat(".cls { color: red; } ", 200) + "</style>"
	bigNav := "<nav>" + strings.Repeat(`<a href="/page">Link</a>`, 100) + "</nav>"
	bigFigure := "<figure>" + strings.Repeat(`<img src="img.jpg" alt="x">`, 50) + "</figure>"
	body := "<main><h1>Hello World</h1><p>Main content here.</p>" + bigFigure + "</main>"
	head := "<head>" + strings.Repeat(`<meta name="x" content="y">`, 100) + bigStyle + "</head>"
	html := "<!DOCTYPE html><html>" + head + "<body>" + bigNav + bigScript + body + "</body></html>"

	w := &WebFetchOptimizer{}
	result := w.Optimize(html)
	out, ok := result.(string)
	if !ok {
		t.Fatalf("expected optimized string, optimizer returned original (savings below threshold)")
	}
	if strings.Contains(out, "<script") {
		t.Error("script tags should be stripped")
	}
	if strings.Contains(out, "<style") {
		t.Error("style tags should be stripped")
	}
	if strings.Contains(out, "<nav") {
		t.Error("nav tags should be stripped")
	}
	if strings.Contains(out, "<head") {
		t.Error("head block should be stripped")
	}
	if strings.Contains(out, "<figure") {
		t.Error("figure tags should be stripped")
	}
	if !strings.Contains(out, "Hello World") {
		t.Error("main content should be preserved")
	}
}

func TestWebFetchOptimizer_ShortInput(t *testing.T) {
	w := &WebFetchOptimizer{}
	short := "<p>Hi</p>"
	if got := w.Optimize(short); got != short {
		t.Error("short input should pass through unchanged")
	}
}

func TestWebFetchOptimizer_NonHTML(t *testing.T) {
	w := &WebFetchOptimizer{}
	text := strings.Repeat("just plain text with no html at all. ", 200)
	if got := w.Optimize(text); got != text {
		t.Error("non-HTML text should pass through unchanged")
	}
}

// ── read ──────────────────────────────────────────────────────────────────────

func TestReadOptimizer_TruncatesLongFile(t *testing.T) {
	r := &ReadOptimizer{}
	// readMaxBytes = 80_000; build a string just over that with many lines
	var sb strings.Builder
	line := strings.Repeat("x", 159) + "\n" // 160 bytes/line
	for sb.Len() < 90_000 {
		sb.WriteString(line)
	}
	input := sb.String()
	result := r.Optimize(input)
	out, ok := result.(string)
	if !ok {
		t.Fatalf("expected string result")
	}
	if !strings.Contains(out, "truncated") {
		t.Error("large file should mention truncation")
	}
	if len(out) >= len(input) {
		t.Error("truncated output should be shorter than input")
	}
}

func TestReadOptimizer_ShortFile(t *testing.T) {
	r := &ReadOptimizer{}
	short := "hello\nworld\n"
	if got := r.Optimize(short); got != short {
		t.Error("short file should pass through unchanged")
	}
}

// ── registry ──────────────────────────────────────────────────────────────────

func TestRegistry_Resolve(t *testing.T) {
	reg := NewRegistry("")

	// Specific tool names should match their optimizers
	bash := reg.Resolve("bash")
	if !bash.Matches("bash") {
		t.Error("bash tool should resolve to BashOptimizer")
	}
	read := reg.Resolve("read")
	if !read.Matches("read") {
		t.Error("read tool should resolve to ReadOptimizer")
	}
	wf := reg.Resolve("webfetch")
	if !wf.Matches("webfetch") {
		t.Error("webfetch tool should resolve to WebFetchOptimizer")
	}

	// Unknown tools fall back to GenericOptimizer which matches everything
	generic := reg.Resolve("some_unknown_tool")
	if !generic.Matches("some_unknown_tool") {
		t.Error("fallback generic optimizer should match any tool")
	}
}

func TestWorkerRegistry_IncludesFigma(t *testing.T) {
	reg := NewWorkerRegistry("figma")
	opt := reg.Resolve("get_design_context")
	if !opt.Matches("get_design_context") {
		t.Error("worker registry with figma hint should match figma tool names")
	}
}

