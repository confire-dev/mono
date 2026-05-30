// Worker-only optimizer — NOT registered in NewRegistry.
// The Figma optimizer runs exclusively in the Cloudflare Worker (paid tier).
// This file is kept as a reference for the Go optimizer interface and for
// potential future use in a self-hosted Worker mode.
package optimizer

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Regexes ported from leanmcp/optimizers/figma.js

var (
	// Strip data-node-id="xxx:xxx" attributes from JSX
	reDataNodeID = regexp.MustCompile(`\s*data-node-id="[^"]*"`)

	// Unwrap CSS variables: var(--token,#hex) → #hex
	reCSSVar = regexp.MustCompile(`\bvar\(--[^,)]+,\s*([^)]+)\)`)

	// Strip "SUPER CRITICAL" footer block (dotall mode)
	reSuperCritical = regexp.MustCompile(`(?s)\nSUPER CRITICAL:.*`)

	// imgLine* / imgDivider* const declarations
	reLineAssetDecl = regexp.MustCompile(`(?m)^const (imgLine\d+|imgDivider\w*) = "[^"]+";$`)

	// <img> elements using line assets
	reLineImgElem = regexp.MustCompile(`<img[^>]*src=\{(imgLine\d+|imgDivider\w*)\}[^>]*\/>`)

	// Figma-specific utility with no CSS equivalent
	reContentStretch = regexp.MustCompile(`\bcontent-stretch\s*`)

	// Collapse 3+ blank lines
	reMultiBlank = regexp.MustCompile(`\n{3,}`)

	// Sparse metadata: full-width frames (~1920px wide) at 2-space indent
	reSectionFrame = regexp.MustCompile(`(?m)^  <frame id="([^"]+)" name="([^"]+)"[^>]*?(?:y="(\d+)")?[^>]*width="(19[012]\d)"[^>]*height="(\d+)"`)

	// Asset name patterns
	reAssetName = regexp.MustCompile(`(?i)\b(asset|illustration|logo|hero|visual|graphic|banner|flow|decoration|deal)\b`)

	// Node type counts
	reNodeTypes = regexp.MustCompile(`<(frame|instance|text|vector|group)`)
	reInstance  = regexp.MustCompile(`<instance `)
	reHidden    = regexp.MustCompile(`hidden="true"`)

	// Text labels and asset names within a section
	reTextLabel = regexp.MustCompile(`<text [^>]*name="([^"]+)"`)
	reSectionAsset = regexp.MustCompile(`(?i)name="([^"]*(?:asset|logo|hero|flow|deal|illustration)[^"]*)"`)

	// Line asset variable name detection
	reLineAssetName = regexp.MustCompile(`^const (imgLine\d+|imgDivider\w*) = "`)
)

type FigmaOptimizer struct {
	// matchAll is set when we know we're wrapping a Figma MCP server, so ALL
	// tool names (e.g. "get_design_context") should go through this optimizer.
	matchAll bool
}

func (f *FigmaOptimizer) Matches(toolName string) bool {
	if f.matchAll {
		return true
	}
	// Fallback: tool name contains "figma" (e.g. mcp__claude_ai_Figma__get_design_context)
	return strings.Contains(strings.ToLower(toolName), "figma")
}

func (f *FigmaOptimizer) Optimize(data interface{}) interface{} {
	defer func() { recover() }()

	// Navigate MCP result: {"content": [{"type": "text", "text": "..."}]}
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	content, ok := m["content"].([]interface{})
	if !ok || len(content) == 0 {
		return data
	}

	first, ok := content[0].(map[string]interface{})
	if !ok {
		return data
	}

	text, ok := first["text"].(string)
	if !ok || text == "" {
		return data
	}

	optimized := optimizeFigmaText(text)
	if optimized == "" || optimized == text {
		return data
	}

	// Return modified copy — never mutate input
	newFirst := make(map[string]interface{}, len(first))
	for k, v := range first {
		newFirst[k] = v
	}
	newFirst["text"] = optimized

	newContent := make([]interface{}, len(content))
	copy(newContent, content)
	newContent[0] = newFirst

	newResult := make(map[string]interface{}, len(m))
	for k, v := range m {
		newResult[k] = v
	}
	newResult["content"] = newContent

	return newResult
}

func optimizeFigmaText(raw string) string {
	if raw == "" {
		return ""
	}

	// Case 1: sparse XML metadata (too large for JSX output)
	isSparse := strings.Contains(raw, "<frame id=") && strings.Contains(raw, "IMPORTANT:")
	isXMLTree := strings.HasPrefix(strings.TrimSpace(raw), "<frame id=")
	if isSparse || isXMLTree {
		return optimizeSparseMetadata(raw)
	}

	// Case 2: normal JSX output
	isFigmaJSX := strings.Contains(raw, "data-node-id=") ||
		strings.Contains(raw, "export default function") ||
		strings.Contains(raw, "figma.com/api/mcp/asset")
	if !isFigmaJSX {
		return ""
	}

	code := raw

	// Strip data-node-id="..." attributes
	code = reDataNodeID.ReplaceAllString(code, "")

	// Unwrap CSS variables — keep fallback value, trim whitespace
	code = reCSSVar.ReplaceAllStringFunc(code, func(match string) string {
		sub := reCSSVar.FindStringSubmatch(match)
		if len(sub) > 1 {
			return strings.TrimSpace(sub[1])
		}
		return match
	})

	// Strip content-stretch utility
	code = reContentStretch.ReplaceAllString(code, "")

	// Strip SUPER CRITICAL footer
	code = reSuperCritical.ReplaceAllString(code, "")

	// Replace imgLine* <img> elements and drop their declarations
	lineAssets := findLineAssets(code)
	if len(lineAssets) > 0 {
		code = reLineImgElem.ReplaceAllStringFunc(code, func(match string) string {
			sub := reLineImgElem.FindStringSubmatch(match)
			if len(sub) > 1 && lineAssets[sub[1]] {
				return `{/* confire: replace with CSS text-decoration or border-bottom */}`
			}
			return match
		})
		code = reLineAssetDecl.ReplaceAllString(code, "")
		code = "// confire: imgLine* → CSS text-decoration/border-bottom, not <img>\n" + code
	}

	// Collapse multiple blank lines
	code = reMultiBlank.ReplaceAllString(code, "\n\n")
	code = strings.TrimSpace(code)

	return code
}

func findLineAssets(code string) map[string]bool {
	assets := make(map[string]bool)
	for _, line := range strings.Split(code, "\n") {
		if m := reLineAssetName.FindStringSubmatch(line); len(m) > 1 {
			assets[m[1]] = true
		}
	}
	return assets
}

type sectionInfo struct {
	id   string
	name string
	y    int
	h    int
}

func optimizeSparseMetadata(xml string) string {
	matches := reSectionFrame.FindAllStringSubmatch(xml, -1)

	var sections []sectionInfo
	for _, m := range matches {
		// m[1]=id, m[2]=name, m[3]=y (optional), m[4]=width, m[5]=height
		name := m[2]
		if regexp.MustCompile(`^Frame \d+$`).MatchString(name) {
			continue
		}
		y, _ := strconv.Atoi(m[3])
		h, _ := strconv.Atoi(m[5])
		sections = append(sections, sectionInfo{id: m[1], name: name, y: y, h: h})
	}

	// Sort by y position
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].y < sections[j].y
	})

	totalNodes := len(reNodeTypes.FindAllString(xml, -1))
	hiddenNodes := len(reHidden.FindAllString(xml, -1))
	instances := len(reInstance.FindAllString(xml, -1))
	assetPatterns := len(reAssetName.FindAllString(xml, -1))

	var sb strings.Builder
	sb.WriteString("## Figma — Sparse Metadata (design too large for full code output)\n")
	fmt.Fprintf(&sb, "Total nodes: %d | Hidden: %d | Component instances: %d | Detected assets: %d\n",
		totalNodes, hiddenNodes, instances, assetPatterns)
	sb.WriteString("\n## Page Sections (fetch these individually with get_design_context)\n\n")

	for i, s := range sections {
		fmt.Fprintf(&sb, "### %s\n", s.name)
		fmt.Fprintf(&sb, "  node-id: %s  |  y: %dpx  |  height: %dpx\n", s.id, s.y, s.h)

		// Extract XML slice for this section
		sectionStart := strings.Index(xml, fmt.Sprintf(`id="%s"`, s.id))
		sectionEnd := len(xml)
		if i+1 < len(sections) {
			next := strings.Index(xml, fmt.Sprintf(`id="%s"`, sections[i+1].id))
			if next > sectionStart {
				sectionEnd = next
			}
		}
		sectionXML := xml[sectionStart:sectionEnd]

		sNodes := len(reNodeTypes.FindAllString(sectionXML, -1))
		sInst := len(reInstance.FindAllString(sectionXML, -1))

		textMatches := reTextLabel.FindAllStringSubmatch(sectionXML, -1)
		var textLabels []string
		for j, tm := range textMatches {
			if j >= 3 {
				break
			}
			textLabels = append(textLabels, tm[1])
		}

		assetMatches := reSectionAsset.FindAllStringSubmatch(sectionXML, -1)
		var assetNames []string
		for j, am := range assetMatches {
			if j >= 4 {
				break
			}
			assetNames = append(assetNames, am[1])
		}

		line := fmt.Sprintf("  %d nodes | %d components", sNodes, sInst)
		if len(assetNames) > 0 {
			line += " | assets: " + strings.Join(assetNames, ", ")
		}
		sb.WriteString(line + "\n")
		if len(textLabels) > 0 {
			sb.WriteString("  text labels: " + strings.Join(textLabels, " · ") + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Next steps\n")
	sb.WriteString("Call get_design_context on each section node-id above.\n")
	sb.WriteString("Recommended order (largest/most complex first):\n")

	sorted := make([]sectionInfo, len(sections))
	copy(sorted, sections)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].h > sorted[j].h })
	limit := 5
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for _, s := range sorted[:limit] {
		fmt.Fprintf(&sb, "  - %s: %s\n", s.name, s.id)
	}

	return sb.String()
}
