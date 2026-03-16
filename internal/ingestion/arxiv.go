// Package ingestion handles downloading and parsing papers from ArXiv and local PDFs.
package ingestion

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Paper holds everything extracted from a paper source before LLM processing.
type Paper struct {
	ID       string // ArXiv ID or filename stem
	Title    string
	Authors  []string
	Abstract string
	FullText string // may be empty if extraction failed
	PDFPath  string
	Source   string // original URL or file path
}

var httpClient = &http.Client{Timeout: 60 * time.Second}

// arxivIDRe matches bare IDs like 2403.05821 or full URLs.
var arxivIDRe = regexp.MustCompile(`(?:arxiv\.org/(?:abs|pdf)/)?(\d{4}\.\d{4,5}(?:v\d+)?)`)

// FetchArXiv downloads a paper from an ArXiv URL or bare ID.
// It stores the PDF in destDir and returns a Paper with metadata + full text.
func FetchArXiv(ctx context.Context, input, destDir string) (*Paper, error) {
	id := extractArXivID(input)
	if id == "" {
		return nil, fmt.Errorf("could not parse ArXiv ID from %q", input)
	}

	// Fetch metadata from ArXiv API
	paper, err := fetchArXivMetadata(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching ArXiv metadata: %w", err)
	}

	// Download PDF
	pdfURL := fmt.Sprintf("https://arxiv.org/pdf/%s", id)
	pdfPath := filepath.Join(destDir, id+".pdf")

	if err := downloadFile(ctx, pdfURL, pdfPath); err != nil {
		return nil, fmt.Errorf("downloading PDF: %w", err)
	}
	paper.PDFPath = pdfPath
	paper.Source = input

	// Extract text from PDF (best-effort)
	text, err := ExtractTextFromPDF(pdfPath)
	if err != nil {
		// Fall back to abstract only — not fatal
		paper.FullText = ""
	} else {
		paper.FullText = text
	}

	return paper, nil
}

// FetchLocalPDF reads a local PDF file and extracts its text.
func FetchLocalPDF(path string) (*Paper, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("file not found: %s", abs)
	}

	text, err := ExtractTextFromPDF(abs)
	if err != nil {
		return nil, fmt.Errorf("extracting text from PDF: %w", err)
	}

	stem := strings.TrimSuffix(filepath.Base(abs), filepath.Ext(abs))
	return &Paper{
		ID:       stem,
		Title:    stem,
		FullText: text,
		PDFPath:  abs,
		Source:   path,
	}, nil
}

// ─── ArXiv metadata ──────────────────────────────────────────────────────────

type arxivFeed struct {
	Entries []arxivEntry `xml:"entry"`
}

type arxivEntry struct {
	Title   string        `xml:"title"`
	Summary string        `xml:"summary"`
	Authors []arxivAuthor `xml:"author"`
	ID      string        `xml:"id"`
}

type arxivAuthor struct {
	Name string `xml:"name"`
}

func fetchArXivMetadata(ctx context.Context, id string) (*Paper, error) {
	url := fmt.Sprintf("https://export.arxiv.org/api/query?id_list=%s", id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed arxivFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parsing ArXiv API response: %w", err)
	}

	if len(feed.Entries) == 0 {
		return nil, fmt.Errorf("no results for ArXiv ID %q", id)
	}

	entry := feed.Entries[0]
	authors := make([]string, len(entry.Authors))
	for i, a := range entry.Authors {
		authors[i] = strings.TrimSpace(a.Name)
	}

	return &Paper{
		ID:       id,
		Title:    strings.TrimSpace(strings.ReplaceAll(entry.Title, "\n", " ")),
		Abstract: strings.TrimSpace(strings.ReplaceAll(entry.Summary, "\n", " ")),
		Authors:  authors,
	}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func extractArXivID(input string) string {
	m := arxivIDRe.FindStringSubmatch(input)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func downloadFile(ctx context.Context, url, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(dest); err == nil {
		return nil // already downloaded
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "research-loop/0.1 (https://github.com/research-loop/research-loop)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d downloading %s", resp.StatusCode, url)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
