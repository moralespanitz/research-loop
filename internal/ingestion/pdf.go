package ingestion

import (
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractTextFromPDF reads a PDF file and returns its plain text content.
// Returns an error only for unreadable files; partial extraction is returned
// alongside an error describing what was lost.
func ExtractTextFromPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening PDF %q: %w", path, err)
	}
	defer f.Close()

	var sb strings.Builder
	totalPages := r.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Non-fatal: skip the page, continue extracting
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("no extractable text found in %q (may be a scanned PDF)", path)
	}
	return result, nil
}

// TruncateText truncates text to approximately maxChars characters,
// breaking at a word boundary. Used to stay within LLM context limits.
func TruncateText(text string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}
	// Find the last space before maxChars
	truncated := text[:maxChars]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "\n\n[... text truncated for context limit ...]"
}
