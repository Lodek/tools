package parse

import "strings"

// extractFrontmatter splits content into body (without frontmatter) and the
// description value from frontmatter. No YAML library needed — we only care
// about the description field.
func extractFrontmatter(content string) (body string, description string) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return content, ""
	}

	// Find closing ---
	closeIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closeIdx = i
			break
		}
	}
	if closeIdx == -1 {
		return content, ""
	}

	// Extract description from frontmatter lines
	for _, line := range lines[1:closeIdx] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "description:") {
			description = strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
			break
		}
	}

	body = strings.Join(lines[closeIdx+1:], "\n")
	return body, description
}
