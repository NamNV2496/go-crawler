package service

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type ICrawlerService interface {
	Crawl(startURL string) (map[string]string, error)
}

type Crawler struct {
	maxDepth int
	visited  map[string]bool
	mutex    sync.Mutex
	results  map[string]string // URL -> Title
}

// NewCrawler creates a new crawler instance
func NewCrawler(maxDepth int) *Crawler {
	return &Crawler{
		maxDepth: maxDepth,
		visited:  make(map[string]bool),
		results:  make(map[string]string),
	}
}

// Crawl starts crawling from the given URL up to the maximum depth
func (c *Crawler) Crawl(startURL string) (map[string]string, error) {
	err := c.crawlPage(startURL, 0)
	if err != nil {
		return nil, err
	}
	return c.results, nil
}

// crawlPage crawls a single page and its links recursively
func (c *Crawler) crawlPage(pageURL string, depth int) error {
	if depth > c.maxDepth {
		return nil
	}

	// Normalize URL
	pageURL = normalizeURL(pageURL)

	// Check if already visited
	c.mutex.Lock()
	if c.visited[pageURL] {
		c.mutex.Unlock()
		return nil
	}
	c.visited[pageURL] = true
	c.mutex.Unlock()

	// Fetch the page
	resp, err := http.Get(pageURL)
	if err != nil {
		return fmt.Errorf("error fetching %s: %v", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status code: %d for %s", resp.StatusCode, pageURL)
	}

	// Parse HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing HTML: %v", err)
	}

	// Extract title
	title := extractTitle(doc)

	// Store result
	c.mutex.Lock()
	c.results[pageURL] = title
	c.mutex.Unlock()

	// Extract links and crawl them
	links := extractLinks(doc, pageURL)
	for _, link := range links {
		err := c.crawlPage(link, depth+1)
		if err != nil {
			// Just log the error and continue with other links
			fmt.Printf("Error crawling %s: %v\n", link, err)
		}
	}

	return nil
}

// extractTitle extracts the title from HTML
func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return n.FirstChild.Data
		}
		return ""
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := extractTitle(c); title != "" {
			return title
		}
	}

	return ""
}

// extractLinks extracts all links from HTML
func extractLinks(n *html.Node, baseURL string) []string {
	var links []string

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				absURL, err := resolveURL(baseURL, attr.Val)
				if err == nil && isValidURL(absURL) {
					links = append(links, absURL)
				}
				break
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = append(links, extractLinks(c, baseURL)...)
	}

	return links
}

// resolveURL resolves a relative URL to an absolute URL
func resolveURL(baseURL, relURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	rel, err := url.Parse(relURL)
	if err != nil {
		return "", err
	}

	absURL := base.ResolveReference(rel)
	return absURL.String(), nil
}

// isValidURL checks if a URL is valid for crawling
func isValidURL(urlStr string) bool {
	// Skip non-HTTP/HTTPS URLs
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return false
	}

	// Add more filters as needed (e.g., file extensions, domains)
	return true
}

// normalizeURL normalizes a URL by removing fragments and some query parameters
func normalizeURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	// Remove fragment
	parsedURL.Fragment = ""

	// You can add more normalization rules here

	return parsedURL.String()
}
