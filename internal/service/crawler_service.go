package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type ICrawlerService interface {
	Crawl(ctx context.Context, startURL, method string) error
}

type Crawler struct {
	maxDepth    int
	visited     map[string]bool
	mutex       sync.Mutex
	results     map[string]string
	teleService ITeleService
}

// NewCrawler creates a new crawler instance
func NewCrawler(
	teleService ITeleService,
) *Crawler {
	return &Crawler{
		maxDepth:    3,
		visited:     make(map[string]bool),
		results:     make(map[string]string),
		teleService: teleService,
	}
}

// Crawl starts crawling from the given URL up to the maximum depth
func (c *Crawler) Crawl(ctx context.Context, startURL, method string) error {
	err := c.crawlPage(startURL, method, 0)
	if err != nil {
		return err
	}
	return nil
}

func (c *Crawler) crawlPage(pageURL, method string, depth int) error {
	if depth > c.maxDepth {
		return nil
	}
	c.mutex.Lock()
	if c.visited[pageURL] {
		c.mutex.Unlock()
		return nil
	}
	c.visited[pageURL] = true
	c.mutex.Unlock()
	switch method {
	case "GET":
		c.crawlGET(pageURL)

	case "POST":
		c.crawlPOST(pageURL)
	case "CURL":
		c.crawlCurl(pageURL)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}
	return nil
}

func (c *Crawler) crawlGET(pageURL string) error {
	resp, err := http.Get(pageURL)
	if err != nil {
		return fmt.Errorf("error fetching %s: %v", pageURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status code: %d for %s", resp.StatusCode, pageURL)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing HTML: %v", err)
	}
	title := extractTitle(doc)
	c.mutex.Lock()
	c.results[title] = title
	c.mutex.Unlock()
	return nil
}

func (c *Crawler) crawlPOST(pageURL string) error {
	resp, err := http.Post(pageURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("error posting %s: %v", pageURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status code: %d for %s", resp.StatusCode, pageURL)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing HTML: %v", err)
	}
	title := extractTitle(doc)
	c.mutex.Lock()
	c.results[title] = title
	c.mutex.Unlock()
	return nil
}

func (c *Crawler) crawlCurl(pageURL string) (io.ReadCloser, error) {
	// Parse the curl command string
	parts := strings.Split(pageURL, "--")
	var args []string

	// Skip the first part as it's the 'curl' command itself
	for _, part := range parts[1:] {
		// Trim spaces
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by first space to separate flag from value
		flagAndValue := strings.SplitN(part, " ", 2)
		if len(flagAndValue) == 2 {
			// Add the flag with '--' prefix
			args = append(args, "--"+flagAndValue[0])
			// Remove surrounding quotes if present and add the value
			value := strings.Trim(strings.TrimSpace(flagAndValue[1]), "'`")
			args = append(args, value)
		}
	}

	// Create and execute command
	cmd := exec.Command("curl", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing curl command: %v", err)
	}
	c.mutex.Lock()
	log.Printf("output: %v\n", string(output))
	log.Println("=======================================")
	if err := c.teleService.SendMessage(ExtractGoldPrice(output), "text"); err != nil {
		log.Printf("send price error: %s", err.Error())
	}
	c.results["test"] = string(output)
	c.mutex.Unlock()
	return io.NopCloser(bytes.NewReader(output)), nil
}

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

func extractLinks(n *html.Node, baseURL string) []string {
	var links []string
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				if isValidURL(baseURL) {
					links = append(links, baseURL)
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

func isValidURL(urlStr string) bool {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return false
	}
	return true
}
