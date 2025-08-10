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
	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
)

const (
	METHOD_ROBOTS string = "ROBOTS"
	METHOD_CURL   string = "CURL"
)

type ICrawlerService interface {
	Crawl(ctx context.Context, url entity.CrawlerEvent) error
}

type CrawlerService struct {
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
) *CrawlerService {
	return &CrawlerService{
		maxDepth:    3,
		visited:     make(map[string]bool),
		results:     make(map[string]string),
		teleService: teleService,
		resultRepo:  resultRepo,
		workerPool:  workerPool,
	}
}

// Crawl starts crawling from the given URL up to the maximum depth
func (_self *CrawlerService) Crawl(ctx context.Context, url entity.CrawlerEvent) error {
	if !url.IsActive {
		return nil
	}
	err := _self.crawlPage(ctx, url, _self.maxDepth)
	if err != nil {
		return err
	}
	return nil
}

func (_self *CrawlerService) crawlPage(ctx context.Context, url entity.CrawlerEvent, depth int) error {
	if depth > _self.maxDepth {
		return nil
	}
	_self.mutex.Lock()
	if _self.visited[url.Url] {
		_self.mutex.Unlock()
		return nil
	}
	_self.visited[url.Url] = true
	_self.mutex.Unlock()
	switch url.Method {
	case http.MethodGet:
		_self.crawlGET(ctx, url, depth)
	case http.MethodPost:
		_self.crawlPOST(ctx, url, depth)
	case METHOD_CURL:
		_self.crawlCurl(ctx, url, depth)
	case METHOD_ROBOTS:
		_self.crawlRobotFile(ctx, url, depth)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", url.Method)
	}

	// update status
	go func() {
		reqBody := fmt.Sprintf(`{"id":%d,"status":"successed"}`, url.Id)
		resp, err := http.Post("http://localhost:8080/api/v1/event/status", "application/json", strings.NewReader(reqBody))
		if err != nil {
			log.Printf("failed to update status: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("status update returned non-200: %d", resp.StatusCode)
		}
	}()
	return nil
}

func (_self *CrawlerService) crawlRobotFile(ctx context.Context, url entity.CrawlerEvent, depth int) (string, error) {
	if !isValidURL(url.Url) {
		return "", nil
	}
	resp, err := http.Get(url.Url)
	if err != nil {
		return "", fmt.Errorf("error fetching %s: %v", url.Url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d for %s", resp.StatusCode, url.Url)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading response body:", err)
	}
	body := string(bodyBytes)
	fmt.Println(body)
	// parsing robot.txt
	if err := _self.handleRobotFile(bodyBytes); err != nil {
		return "", err
	}
	return "", nil
}

func (_self *CrawlerService) crawlGET(ctx context.Context, url entity.CrawlerEvent, depth int) (string, error) {
	if !isValidURL(url.Url) {
		return "", nil
	}
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
	_self.mutex.Lock()
	_self.results[title] = doc.Data
	_self.mutex.Unlock()
	return doc.Data, nil
}

func (_self *CrawlerService) crawlPOST(ctx context.Context, url entity.CrawlerEvent, depth int) (string, error) {
	if !isValidURL(url.Url) {
		return "", nil
	}
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
	_self.mutex.Lock()
	_self.results[title] = doc.Data
	_self.mutex.Unlock()
	return doc.Data, nil
}

func (_self *CrawlerService) crawlCurl(ctx context.Context, url entity.CrawlerEvent, depth int) (string, error) {
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
	_self.workerPool.Crawl(
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
			if err = _self.teleService.SendMessage(ExtractGoldPrice(output), "text"); err != nil {
				log.Printf("send price error: %s", err.Error())
			}
			// write result to db
			if err = _self.resultRepo.CreateResult(ctx, &domain.Result{
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
			_self.results["test"] = string(output)
		})
	if err != nil {
		return "", fmt.Errorf("error executing curl command: %v", err)
	}
	resp := string(output)
	io.NopCloser(bytes.NewReader(output))
	return resp, nil
}

func (_self *CrawlerService) handleRobotFile(bodyBytes []byte) error {
	// parsing robot.txt
	robots, err := robotstxt.FromBytes(bodyBytes)
	if err != nil {
		panic(err)
	}
	// Replace "Googlebot" with your user-agent string
	group := robots.FindGroup("Googlebot")
	testUrl := "/san-pham/iphone-15.html"

	allowed := group.Test(testUrl)
	fmt.Printf("Googlebot can fetch %s: %v\n", testUrl, allowed)
	return nil
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
