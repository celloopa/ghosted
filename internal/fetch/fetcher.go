package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FetchType indicates what kind of content to fetch
type FetchType int

const (
	FetchTypeJobPosting FetchType = iota
	FetchTypeCV
)

// Fetcher handles fetching and saving job postings from URLs
type Fetcher struct {
	Client    *http.Client
	OutputDir string
}

// FetchResult contains the result of a fetch operation
type FetchResult struct {
	URL         string `json:"url"`
	OutputPath  string `json:"output_path"`
	Company     string `json:"company"`
	Position    string `json:"position"`
	ContentSize int    `json:"content_size"`
}

// NewFetcher creates a new Fetcher instance
func NewFetcher(outputDir string) *Fetcher {
	return &Fetcher{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		OutputDir: outputDir,
	}
}

// IsURL checks if the input looks like a URL
func IsURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

// DetectFetchType determines whether the input should be fetched as a CV or job posting.
// Rules:
// - Bare domain (no path after domain) → CV fetch
// - Explicit /cv.json path → CV fetch
// - Everything else → Job posting fetch
func DetectFetchType(input string) FetchType {
	// Normalize input: add https:// if missing
	normalizedURL := input
	if !IsURL(input) {
		normalizedURL = "https://" + input
	}

	parsedURL, err := url.Parse(normalizedURL)
	if err != nil {
		// If we can't parse it, default to job posting
		return FetchTypeJobPosting
	}

	// Get the path, trimming any trailing slashes
	path := strings.TrimRight(parsedURL.Path, "/")

	// Explicit /cv.json path → CV fetch
	if strings.HasSuffix(path, "/cv.json") || path == "/cv.json" {
		return FetchTypeCV
	}

	// Bare domain (empty path or just /) → CV fetch
	if path == "" {
		return FetchTypeCV
	}

	// Everything else → Job posting
	return FetchTypeJobPosting
}

// Fetch downloads a job posting from a URL and saves it to the output directory
func (f *Fetcher) Fetch(rawURL string, outputName string) (*FetchResult, error) {
	// Validate URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("URL must be http or https")
	}

	// Fetch the page
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a browser-like user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	htmlContent := string(body)

	// Detect the job board and extract content
	content, company, position := f.ExtractJobPosting(htmlContent, parsedURL)

	// Generate output filename
	if outputName == "" {
		outputName = f.GenerateFilename(company, position, parsedURL)
	}
	if !strings.HasSuffix(outputName, ".md") {
		outputName += ".md"
	}

	outputPath := filepath.Join(f.OutputDir, outputName)

	// Ensure output directory exists
	if err := os.MkdirAll(f.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Add metadata header to the content
	finalContent := f.FormatOutput(content, rawURL, company, position)

	// Write to file
	if err := os.WriteFile(outputPath, []byte(finalContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &FetchResult{
		URL:         rawURL,
		OutputPath:  outputPath,
		Company:     company,
		Position:    position,
		ContentSize: len(finalContent),
	}, nil
}

// ExtractJobPosting extracts job posting content from HTML based on the job board
func (f *Fetcher) ExtractJobPosting(html string, parsedURL *url.URL) (content, company, position string) {
	host := strings.ToLower(parsedURL.Host)

	// Try to detect job board and use specialized extraction
	switch {
	case strings.Contains(host, "lever.co"):
		return f.extractLever(html)
	case strings.Contains(host, "greenhouse.io"):
		return f.extractGreenhouse(html)
	case strings.Contains(host, "workday.com"):
		return f.extractWorkday(html)
	case strings.Contains(host, "linkedin.com"):
		return f.extractLinkedIn(html)
	case strings.Contains(host, "ashbyhq.com"):
		return f.extractAshby(html)
	case strings.Contains(host, "careers.microsoft.com"):
		return f.extractMicrosoft(html)
	default:
		return f.extractGeneric(html)
	}
}

// extractLever extracts job posting from Lever pages
func (f *Fetcher) extractLever(html string) (content, company, position string) {
	// Extract title
	position = extractBetween(html, `<h2>`, `</h2>`)
	if position == "" {
		position = extractMetaContent(html, "og:title")
	}

	// Extract company from URL or page
	company = extractBetween(html, `<div class="main-header-content">`, `</div>`)
	if company == "" {
		company = extractMetaContent(html, "og:site_name")
	}

	// Extract job description
	content = extractBetween(html, `<div class="section-wrapper page-full-width">`, `<div class="section last">`)
	if content == "" {
		content = extractBetween(html, `<div class="content">`, `</div>`)
	}

	content = cleanHTML(content)
	company = cleanText(company)
	position = cleanText(position)

	return content, company, position
}

// extractGreenhouse extracts job posting from Greenhouse pages
func (f *Fetcher) extractGreenhouse(html string) (content, company, position string) {
	position = extractMetaContent(html, "og:title")
	company = extractBetween(html, `<span class="company-name">`, `</span>`)
	if company == "" {
		company = extractMetaContent(html, "og:site_name")
	}

	content = extractBetween(html, `<div id="content">`, `<div id="application">`)
	if content == "" {
		content = extractBetween(html, `<div class="content">`, `</div>`)
	}

	content = cleanHTML(content)
	company = cleanText(company)
	position = cleanText(position)

	return content, company, position
}

// extractWorkday extracts job posting from Workday pages
func (f *Fetcher) extractWorkday(html string) (content, company, position string) {
	position = extractMetaContent(html, "og:title")
	company = extractMetaContent(html, "og:site_name")

	// Workday uses dynamic content, so we get what we can
	content = extractBetween(html, `<div data-automation-id="jobPostingDescription">`, `</div>`)

	content = cleanHTML(content)
	company = cleanText(company)
	position = cleanText(position)

	return content, company, position
}

// extractLinkedIn extracts job posting from LinkedIn pages
func (f *Fetcher) extractLinkedIn(html string) (content, company, position string) {
	position = extractMetaContent(html, "og:title")
	company = extractBetween(html, `<a class="topcard__org-name-link"`, `</a>`)
	if company == "" {
		// Try alternate selector
		company = extractBetween(html, `"companyName":"`, `"`)
	}

	content = extractBetween(html, `<div class="description__text">`, `</div>`)
	if content == "" {
		content = extractMetaContent(html, "og:description")
	}

	content = cleanHTML(content)
	company = cleanText(company)
	position = cleanText(position)

	return content, company, position
}

// extractAshby extracts job posting from Ashby pages
func (f *Fetcher) extractAshby(html string) (content, company, position string) {
	position = extractMetaContent(html, "og:title")
	company = extractMetaContent(html, "og:site_name")

	content = extractBetween(html, `<div class="ashby-job-posting-description">`, `</div>`)

	content = cleanHTML(content)
	company = cleanText(company)
	position = cleanText(position)

	return content, company, position
}

// extractMicrosoft extracts job posting from Microsoft Careers pages
// Microsoft uses Next.js with job data in <script id="__NEXT_DATA__"> JSON
func (f *Fetcher) extractMicrosoft(html string) (content, company, position string) {
	// Default company name
	company = "Microsoft"

	// Try to extract from __NEXT_DATA__ JSON
	nextDataStart := strings.Index(html, `<script id="__NEXT_DATA__"`)
	if nextDataStart != -1 {
		// Find the start of the JSON content
		jsonStart := strings.Index(html[nextDataStart:], ">")
		if jsonStart != -1 {
			jsonStart += nextDataStart + 1
			jsonEnd := strings.Index(html[jsonStart:], "</script>")
			if jsonEnd != -1 {
				jsonData := html[jsonStart : jsonStart+jsonEnd]
				content, position = f.parseMicrosoftNextData(jsonData)
			}
		}
	}

	// Fallback to meta tags if __NEXT_DATA__ parsing failed
	if position == "" {
		position = extractMetaContent(html, "og:title")
		// Clean up Microsoft title format: "Job Title | Microsoft Careers"
		if idx := strings.Index(position, " | "); idx != -1 {
			position = position[:idx]
		}
	}

	if content == "" {
		content = extractMetaContent(html, "og:description")
	}

	// Validate extracted data
	content, company, position = f.validateMicrosoftExtraction(content, company, position)

	content = cleanHTML(content)
	company = cleanText(company)
	position = cleanText(position)

	return content, company, position
}

// parseMicrosoftNextData parses the __NEXT_DATA__ JSON and extracts job details
func (f *Fetcher) parseMicrosoftNextData(jsonData string) (content, position string) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return "", ""
	}

	// Navigate to props.pageProps where job data typically lives
	props, ok := data["props"].(map[string]interface{})
	if !ok {
		return "", ""
	}

	pageProps, ok := props["pageProps"].(map[string]interface{})
	if !ok {
		return "", ""
	}

	// Try to find job data - Microsoft uses various structures
	// Common paths: pageProps.job, pageProps.jobDetail, pageProps.data

	// Try pageProps.job first
	if job, ok := pageProps["job"].(map[string]interface{}); ok {
		return f.extractMicrosoftJobData(job)
	}

	// Try pageProps.jobDetail
	if jobDetail, ok := pageProps["jobDetail"].(map[string]interface{}); ok {
		return f.extractMicrosoftJobData(jobDetail)
	}

	// Try pageProps.data
	if dataObj, ok := pageProps["data"].(map[string]interface{}); ok {
		return f.extractMicrosoftJobData(dataObj)
	}

	// Try pageProps directly (sometimes job data is at this level)
	return f.extractMicrosoftJobData(pageProps)
}

// extractMicrosoftJobData extracts content and position from a job data object
func (f *Fetcher) extractMicrosoftJobData(job map[string]interface{}) (content, position string) {
	// Extract position/title
	for _, key := range []string{"title", "jobTitle", "name", "positionTitle"} {
		if val, ok := job[key].(string); ok && val != "" {
			position = val
			break
		}
	}

	// Extract description/content
	var descParts []string

	// Try various description fields
	for _, key := range []string{"description", "jobDescription", "fullDescription", "summary"} {
		if val, ok := job[key].(string); ok && val != "" {
			descParts = append(descParts, val)
		}
	}

	// Try qualifications
	if quals, ok := job["qualifications"].(string); ok && quals != "" {
		descParts = append(descParts, "\n\n## Qualifications\n\n"+quals)
	} else if quals, ok := job["qualifications"].([]interface{}); ok {
		descParts = append(descParts, "\n\n## Qualifications\n")
		for _, q := range quals {
			if qs, ok := q.(string); ok {
				descParts = append(descParts, "- "+qs)
			}
		}
	}

	// Try responsibilities
	if resp, ok := job["responsibilities"].(string); ok && resp != "" {
		descParts = append(descParts, "\n\n## Responsibilities\n\n"+resp)
	} else if resp, ok := job["responsibilities"].([]interface{}); ok {
		descParts = append(descParts, "\n\n## Responsibilities\n")
		for _, r := range resp {
			if rs, ok := r.(string); ok {
				descParts = append(descParts, "- "+rs)
			}
		}
	}

	// Try location info
	if loc, ok := job["location"].(string); ok && loc != "" {
		descParts = append(descParts, "\n\n**Location:** "+loc)
	} else if loc, ok := job["primaryLocation"].(string); ok && loc != "" {
		descParts = append(descParts, "\n\n**Location:** "+loc)
	}

	// Try employment type
	if empType, ok := job["employmentType"].(string); ok && empType != "" {
		descParts = append(descParts, "\n\n**Employment Type:** "+empType)
	}

	content = strings.Join(descParts, "\n")
	return content, position
}

// validateMicrosoftExtraction validates and cleans up extracted Microsoft data
func (f *Fetcher) validateMicrosoftExtraction(content, company, position string) (string, string, string) {
	// Reject numeric company names (indicates parsing error)
	if isNumeric(company) {
		company = "Microsoft"
	}

	// Reject empty or very short positions
	if len(strings.TrimSpace(position)) < 3 {
		position = ""
	}

	// Reject positions that are just numbers
	if isNumeric(position) {
		position = ""
	}

	return content, company, position
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// extractGeneric extracts job posting from any HTML page
func (f *Fetcher) extractGeneric(html string) (content, company, position string) {
	// Try common meta tags
	position = extractMetaContent(html, "og:title")
	if position == "" {
		position = extractBetween(html, `<title>`, `</title>`)
	}

	company = extractMetaContent(html, "og:site_name")

	// Try to find main content
	content = extractMetaContent(html, "og:description")

	// Look for common job posting containers
	containers := []string{
		`<div class="job-description">`,
		`<div id="job-description">`,
		`<div class="description">`,
		`<article>`,
		`<main>`,
	}

	for _, container := range containers {
		endTag := "</" + strings.TrimPrefix(strings.Split(container, " ")[0], "<")
		if extracted := extractBetween(html, container, endTag); extracted != "" {
			content = extracted
			break
		}
	}

	content = cleanHTML(content)
	company = cleanText(company)
	position = cleanText(position)

	return content, company, position
}

// GenerateFilename creates a filename from company and position
func (f *Fetcher) GenerateFilename(company, position string, parsedURL *url.URL) string {
	// Use company and position if available
	if company != "" && position != "" {
		return sanitizeFilename(company) + "-" + sanitizeFilename(position) + "-posting"
	}

	// Fall back to using the URL path
	path := parsedURL.Path
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) > 0 {
		// Use last meaningful part of path
		name := parts[len(parts)-1]
		if name == "" && len(parts) > 1 {
			name = parts[len(parts)-2]
		}
		return sanitizeFilename(name) + "-posting"
	}

	// Last resort: use host
	return sanitizeFilename(parsedURL.Host) + "-posting"
}

// FormatOutput creates the final markdown output with metadata
func (f *Fetcher) FormatOutput(content, sourceURL, company, position string) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("source: %s\n", sourceURL))
	sb.WriteString(fmt.Sprintf("fetched: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	if company != "" {
		sb.WriteString(fmt.Sprintf("company: %s\n", company))
	}
	if position != "" {
		sb.WriteString(fmt.Sprintf("position: %s\n", position))
	}
	sb.WriteString("---\n\n")

	if position != "" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", position))
	}
	if company != "" {
		sb.WriteString(fmt.Sprintf("**Company:** %s\n\n", company))
	}

	sb.WriteString("## Job Description\n\n")
	sb.WriteString(content)
	sb.WriteString("\n")

	return sb.String()
}

// Helper functions

func extractBetween(html, start, end string) string {
	startIdx := strings.Index(html, start)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(start)

	endIdx := strings.Index(html[startIdx:], end)
	if endIdx == -1 {
		return ""
	}

	return html[startIdx : startIdx+endIdx]
}

func extractMetaContent(html, property string) string {
	// Try og: prefix
	pattern := fmt.Sprintf(`<meta[^>]*property="%s"[^>]*content="([^"]*)"`, property)
	re := regexp.MustCompile(pattern)
	if matches := re.FindStringSubmatch(html); len(matches) > 1 {
		return matches[1]
	}

	// Try name attribute
	pattern = fmt.Sprintf(`<meta[^>]*name="%s"[^>]*content="([^"]*)"`, property)
	re = regexp.MustCompile(pattern)
	if matches := re.FindStringSubmatch(html); len(matches) > 1 {
		return matches[1]
	}

	// Try reversed order (content before property)
	pattern = fmt.Sprintf(`<meta[^>]*content="([^"]*)"[^>]*property="%s"`, property)
	re = regexp.MustCompile(pattern)
	if matches := re.FindStringSubmatch(html); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func cleanHTML(html string) string {
	// Remove script and style tags
	scriptRe := regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	html = scriptRe.ReplaceAllString(html, "")

	styleRe := regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	html = styleRe.ReplaceAllString(html, "")

	// Convert common HTML elements to markdown
	html = regexp.MustCompile(`<h1[^>]*>`).ReplaceAllString(html, "\n# ")
	html = regexp.MustCompile(`</h1>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`<h2[^>]*>`).ReplaceAllString(html, "\n## ")
	html = regexp.MustCompile(`</h2>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`<h3[^>]*>`).ReplaceAllString(html, "\n### ")
	html = regexp.MustCompile(`</h3>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`<h4[^>]*>`).ReplaceAllString(html, "\n#### ")
	html = regexp.MustCompile(`</h4>`).ReplaceAllString(html, "\n")

	html = regexp.MustCompile(`<li[^>]*>`).ReplaceAllString(html, "\n- ")
	html = regexp.MustCompile(`</li>`).ReplaceAllString(html, "")

	html = regexp.MustCompile(`<p[^>]*>`).ReplaceAllString(html, "\n\n")
	html = regexp.MustCompile(`</p>`).ReplaceAllString(html, "\n")

	html = regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`<hr\s*/?>`).ReplaceAllString(html, "\n---\n")

	html = regexp.MustCompile(`<strong[^>]*>`).ReplaceAllString(html, "**")
	html = regexp.MustCompile(`</strong>`).ReplaceAllString(html, "**")
	html = regexp.MustCompile(`<b[^>]*>`).ReplaceAllString(html, "**")
	html = regexp.MustCompile(`</b>`).ReplaceAllString(html, "**")

	html = regexp.MustCompile(`<em[^>]*>`).ReplaceAllString(html, "*")
	html = regexp.MustCompile(`</em>`).ReplaceAllString(html, "*")
	html = regexp.MustCompile(`<i[^>]*>`).ReplaceAllString(html, "*")
	html = regexp.MustCompile(`</i>`).ReplaceAllString(html, "*")

	// Remove remaining HTML tags
	tagRe := regexp.MustCompile(`<[^>]+>`)
	html = tagRe.ReplaceAllString(html, "")

	// Decode HTML entities
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&#39;", "'")
	html = strings.ReplaceAll(html, "&apos;", "'")

	// Clean up whitespace
	html = regexp.MustCompile(`\n{3,}`).ReplaceAllString(html, "\n\n")
	html = regexp.MustCompile(`[ \t]+`).ReplaceAllString(html, " ")

	return strings.TrimSpace(html)
}

func cleanText(text string) string {
	// Remove HTML tags
	tagRe := regexp.MustCompile(`<[^>]+>`)
	text = tagRe.ReplaceAllString(text, "")

	// Decode entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	return strings.TrimSpace(text)
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove characters that are invalid in filenames
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "'", ",", ".", "(", ")", "&", "@", "#", "$", "%", "^", "+", "=", "[", "]", "{", "}"}
	for _, char := range invalid {
		s = strings.ReplaceAll(s, char, "")
	}
	// Remove multiple dashes
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
