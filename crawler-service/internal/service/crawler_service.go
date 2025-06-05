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

	"github.com/namnv2496/crawler/internal/domain"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/repository"
	"golang.org/x/net/html"
)

type ICrawlerService interface {
	Crawl(ctx context.Context, url entity.Url) error
}

type Crawler struct {
	maxDepth    int
	visited     map[string]bool
	mutex       sync.Mutex
	results     map[string]string
	teleService ITeleService
	resultRepo  repository.IResultRepository
	workerPool  IWorkerPool
}

// NewCrawler creates a new crawler instance
func NewCrawlerService(
	teleService ITeleService,
	resultRepo repository.IResultRepository,
	workerPool IWorkerPool,
) *Crawler {
	return &Crawler{
		maxDepth:    3,
		visited:     make(map[string]bool),
		results:     make(map[string]string),
		teleService: teleService,
		resultRepo:  resultRepo,
		workerPool:  workerPool,
	}
}

// Crawl starts crawling from the given URL up to the maximum depth
func (c *Crawler) Crawl(ctx context.Context, url entity.Url) error {
	err := c.crawlPage(ctx, url, c.maxDepth)
	if err != nil {
		return err
	}
	return nil
}

func (c *Crawler) crawlPage(ctx context.Context, url entity.Url, depth int) error {
	if depth > c.maxDepth {
		return nil
	}
	c.mutex.Lock()
	if c.visited[url.Url] {
		c.mutex.Unlock()
		return nil
	}
	c.visited[url.Url] = true
	c.mutex.Unlock()
	switch url.Method {
	case "GET":
		c.crawlGET(ctx, url, depth)
	case "POST":
		c.crawlPOST(ctx, url, depth)
	case "CURL":
		c.crawlCurl(ctx, url, depth)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", url.Method)
	}

	return nil
}

func (c *Crawler) crawlGET(ctx context.Context, url entity.Url, depth int) (string, error) {
	resp, err := http.Get(url.Url)
	if err != nil {
		return "", fmt.Errorf("error fetching %s: %v", url.Url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d for %s", resp.StatusCode, url.Url)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %v", err)
	}
	title := extractTitle(doc)
	c.mutex.Lock()
	c.results[title] = doc.Data
	c.mutex.Unlock()
	return doc.Data, nil
}

func (c *Crawler) crawlPOST(ctx context.Context, url entity.Url, depth int) (string, error) {
	resp, err := http.Post(url.Url, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return "", fmt.Errorf("error posting %s: %v", url.Url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d for %s", resp.StatusCode, url.Url)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %v", err)
	}
	title := extractTitle(doc)
	c.mutex.Lock()
	c.results[title] = doc.Data
	c.mutex.Unlock()
	return doc.Data, nil
}

func (c *Crawler) crawlCurl(ctx context.Context, url entity.Url, depth int) (string, error) {
	// Parse the curl command string
	parts := strings.Split(url.Url, "--")
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
	var output []byte
	var err error
	c.workerPool.Crawl(
		func() (any, error) {
			var cmdOutput []byte
			cmdOutput, err = cmd.Output()
			output = cmdOutput // Assign to outer variable
			return cmdOutput, err
		},
		depth,
		nil,
		func(result any, cmdErr error) {
			if cmdErr != nil {
				err = cmdErr // Propagate error to outer scope
				return
			}
			if err = c.teleService.SendMessage(ExtractGoldPrice(output), "text"); err != nil {
				log.Printf("send price error: %s", err.Error())
			}
			// write result to db
			if err = c.resultRepo.CreateResult(ctx, &domain.Result{
				Url:    url.Url,
				Method: url.Method,
				Queue:  url.Queue,
				Domain: url.Domain,
				Result: string(output),
			}); err != nil {
				log.Printf("create result error: %s", err.Error())
			}
			log.Printf("send message to Telegram: %v\n", string(output))
			log.Println("=======================================")
			c.results["test"] = string(output)
		})
	if err != nil {
		return "", fmt.Errorf("error executing curl command: %v", err)
	}
	resp := string(output)
	io.NopCloser(bytes.NewReader(output))
	return resp, nil
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
