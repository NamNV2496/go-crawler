package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/namnv2496/crawler/internal/domain"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/pkg/logging"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/repository/schedulerservice"
	"github.com/namnv2496/crawler/internal/service/mq"
	"github.com/namnv2496/crawler/internal/service/server"
	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
)

const (
	METHOD_ROBOTS string = "ROBOTS"
	METHOD_CURL   string = "CURL"
)

//go:generate mockgen -source=$GOFILE -destination=../../mocks/usecase/$GOFILE.mock.go -package=$GOPACKAGE
type ICrawlerService interface {
	Crawl(ctx context.Context, url entity.CrawlerEvent) error
}

type crawlerService struct {
	isDistributedWorkerPool bool
	maxDepth                int64
	teleService             ITeleService
	resultRepo              repository.IResultRepository
	workerPool              server.IWorkerPool
	distributedWorkerPool   server.IDistributedWorkerPool
	retryProducer           mq.IAsynqProducer
	schedulerServiceClient  schedulerservice.ISchedulerService
}

// crawlSession holds temporary state for a single crawl operation
type crawlSession struct {
	visited map[string]bool
	results map[string]string
	mutex   sync.Mutex
}

// NewCrawler creates a new crawler instance
func NewCrawlerService(
	teleService ITeleService,
	resultRepo repository.IResultRepository,
	workerPool server.IWorkerPool,
	distributedWorkerPool server.IDistributedWorkerPool,
	retryProducer mq.IAsynqProducer,
	schedulerServiceClient schedulerservice.ISchedulerService,
) *crawlerService {
	return &crawlerService{
		isDistributedWorkerPool: false,
		maxDepth:                3,
		teleService:             teleService,
		resultRepo:              resultRepo,
		workerPool:              workerPool,
		distributedWorkerPool:   distributedWorkerPool,
		retryProducer:           retryProducer,
		schedulerServiceClient:  schedulerServiceClient,
	}
}

// Crawl starts crawling from the given URL up to the maximum depth
func (_self *crawlerService) Crawl(ctx context.Context, event entity.CrawlerEvent) error {
	deferFunc := logging.AppendPrefix("Crawl")
	defer deferFunc()
	if !event.IsActive {
		return nil
	}
	// Create a new isolated session for this crawl operation
	session := newCrawlSession()
	if _self.isDistributedWorkerPool {
		_self.distributedWorkerPool.ExecuteOnServerAsync(ctx, &server.ExecutionRequest{}, make(chan *server.ExecutionResponse))
	} else {
		_self.workerPool.Execute(
			func() (any, error) {
				return nil, _self.crawlPage(ctx, event, _self.maxDepth, session)
			},
			_self.maxDepth,
			nil, // stats call back
			func(result any, err error) {
				status := string(entity.StatusSuccessed)
				if err != nil {
					// delay 5m if fail
					if event.Retrytime < 3 {
						event.Retrytime += 1
						_self.retryProducer.EnqueueRetryEvent(ctx, event, time.Now().Add(5*time.Minute))
					}
					status = string(entity.StatusFailed)
					logging.Error(ctx, "crawl failed: %v", err)
				}
				if err := _self.schedulerServiceClient.UpdateSchedulerEvent(ctx, &entity.UpdateSchedulerEventRequest{
					Id: fmt.Sprint(event.Id),
					Event: &entity.SchedulerEvent{
						Id:          fmt.Sprint(event.Id),
						Status:      status,
						Queue:       event.Queue,
						Domain:      event.Domain,
						Url:         event.Url,
						Method:      event.Method,
						SchedulerAt: event.Retrytime,
						NextRunTime: event.Retrytime,
						IsActive:    event.IsActive,
					},
				}); err != nil {
					logging.Error(ctx, "failed to update scheduler event %d: %v", event.Id, err)
				} else {
					logging.Debug(ctx, "successfully updated scheduler event %d with status %s", event.Id, status)
				}
			},
		)
	}
	return nil
}

func (_self *crawlerService) crawlPage(ctx context.Context, url entity.CrawlerEvent, depth int64, session *crawlSession) error {
	deferFunc := logging.AppendPrefix("crawlPage")
	defer deferFunc()

	if depth == 0 {
		return _self.crawlBFS(ctx, url, session)
	}
	if !session.shouldCrawl(url.Url, depth, _self.maxDepth) {
		return nil
	}
	session.markAsVisited(url.Url)
	_, err := _self.fetchPage(ctx, url, depth, session)
	if err != nil {
		logging.Error(ctx, "fetch page error: %v", err)
		return err
	}

	_self.updateCrawlStatus(ctx, url)
	return nil
}

func (_self *crawlerService) crawlBFS(ctx context.Context, startEvent entity.CrawlerEvent, session *crawlSession) error {
	deferFunc := logging.AppendPrefix("crawlBFS")
	defer deferFunc()
	type queueItem struct {
		event entity.CrawlerEvent
		depth int64
	}
	currentLevel := []queueItem{{event: startEvent, depth: 0}}
	for currentDepth := int64(0); currentDepth <= _self.maxDepth && len(currentLevel) > 0; currentDepth++ {
		logging.Debug(ctx, "BFS: processing depth %d with %d URLs", currentDepth, len(currentLevel))
		var nextLevel []queueItem
		var mu sync.Mutex
		var wg sync.WaitGroup
		for _, item := range currentLevel {
			if !session.shouldCrawl(item.event.Url, item.depth, _self.maxDepth) {
				continue
			}
			wg.Add(1)
			go func(itm queueItem) {
				defer wg.Done()
				session.markAsVisited(itm.event.Url)
				doc, err := _self.fetchPage(ctx, itm.event, itm.depth, session)
				if err != nil {
					logging.Error(ctx, "fetch page error at depth %d for %s: %v", itm.depth, itm.event.Url, err)
					return
				}
				_self.updateCrawlStatus(ctx, itm.event)
				if doc != nil && itm.depth < _self.maxDepth {
					links := extractLinks(doc, itm.event.Url)
					logging.Debug(ctx, "BFS: extracted %d links from %s at depth %d", len(links), itm.event.Url, itm.depth)
					for _, link := range links {
						if session.shouldQueueLink(link) {
							newEvent := entity.CrawlerEvent{
								Url:    link,
								Method: http.MethodGet,
								Domain: itm.event.Domain,
								Queue:  itm.event.Queue,
							}

							mu.Lock()
							nextLevel = append(nextLevel, queueItem{
								event: newEvent,
								depth: itm.depth + 1,
							})
							mu.Unlock()

							logging.Debug(ctx, "BFS: queued link %s for depth %d", link, itm.depth+1)
						}
					}
				}
			}(item)
		}
		wg.Wait()
		logging.Debug(ctx, "BFS: completed depth %d, queued %d URLs for depth %d", currentDepth, len(nextLevel), currentDepth+1)
		currentLevel = nextLevel
	}

	logging.Debug(ctx, "BFS: crawling completed, max depth: %d", _self.maxDepth)
	return nil
}

func (_self *crawlerService) fetchPage(ctx context.Context, url entity.CrawlerEvent, _ int64, session *crawlSession) (*html.Node, error) {
	switch url.Method {
	case http.MethodGet:
		doc, err := _self.crawlGET(ctx, url, session)
		if err != nil {
			logging.Error(ctx, "crawlGET error: %v", err)
			return nil, err
		}
		return doc, nil
	case http.MethodPost:
		doc, err := _self.crawlPOST(ctx, url, session)
		if err != nil {
			logging.Error(ctx, "crawlPOST error: %v", err)
			return nil, err
		}
		return doc, nil
	case METHOD_CURL:
		_, err := _self.crawlCurl(ctx, url)
		return nil, err
	case METHOD_ROBOTS:
		_, err := _self.crawlRobotFile(ctx, url)
		return nil, err
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", url.Method)
	}
}

func (_self *crawlerService) updateCrawlStatus(ctx context.Context, url entity.CrawlerEvent) {
	go func() {
		reqBody := fmt.Sprintf(`{"id":%d,"status":"successed"}`, url.Id)
		resp, err := http.Post("http://localhost:8080/api/v1/event/status", "application/json", strings.NewReader(reqBody))
		if err != nil {
			logging.Error(ctx, "failed to update status: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logging.Debug(ctx, "status update returned non-200: %d", resp.StatusCode)
		}
	}()
}

func (_self *crawlerService) crawlGET(ctx context.Context, url entity.CrawlerEvent, session *crawlSession) (*html.Node, error) {
	if !isValidURL(url.Url) {
		return nil, nil
	}
	doc, err := _self.fetchPageInfo(url.Url, http.MethodGet, "")
	if err != nil {
		return nil, err
	}
	_self.extractAndSaveData(ctx, doc, url, session)
	return doc, nil
}

func (_self *crawlerService) crawlPOST(ctx context.Context, url entity.CrawlerEvent, session *crawlSession) (*html.Node, error) {
	if !isValidURL(url.Url) {
		return nil, nil
	}
	doc, err := _self.fetchPageInfo(url.Url, http.MethodPost, "application/x-www-form-urlencoded")
	if err != nil {
		return nil, err
	}
	_self.extractAndSaveData(ctx, doc, url, session)
	return doc, nil
}

func (_self *crawlerService) crawlCurl(ctx context.Context, url entity.CrawlerEvent) (string, error) {
	deferFunc := logging.AppendPrefix("crawlCurl")
	defer deferFunc()

	args := _self.parseCurlCommand(url.Url)
	output, err := _self.executeCurlCommand(ctx, args)
	if err != nil {
		return "", err
	}
	_self.processGoldResult(ctx, url, output)
	return string(output), nil
}

func (_self *crawlerService) crawlRobotFile(_ context.Context, url entity.CrawlerEvent) (string, error) {
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

func (_self *crawlerService) handleRobotFile(bodyBytes []byte) error {
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

func (_self *crawlerService) fetchPageInfo(url, method, contentType string) (*html.Node, error) {
	var resp *http.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = http.Get(url)
	case http.MethodPost:
		resp, err = http.Post(url, contentType, nil)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching %s: %v", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("non-200 status code: %d for %s", resp.StatusCode, url)
	}
	// parse response data
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML from %s: %v", url, err)
	}
	return doc, nil
}

func (_self *crawlerService) extractAndSaveData(ctx context.Context, doc *html.Node, url entity.CrawlerEvent, session *crawlSession) {
	title := extractTitle(doc)
	session.mutex.Lock()
	session.results[title] = doc.Data
	session.mutex.Unlock()
	// Save to database
	if err := _self.resultRepo.CreateResult(ctx, &domain.Result{
		Url:    url.Url,
		Method: url.Method,
		Queue:  url.Queue,
		Domain: url.Domain,
		Result: title,
	}); err != nil {
		logging.Error(ctx, "create result error: %s", err.Error())
	}
	logging.Debug(ctx, "extracted data from %s: title=%s", url.Url, title)
}

func (_self *crawlerService) parseCurlCommand(curlCmd string) []string {
	parts := strings.Split(curlCmd, "--")
	var args []string

	// Skip the first part as it's the 'curl' command itself
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by first space to separate flag from value
		flagAndValue := strings.SplitN(part, " ", 2)
		if len(flagAndValue) == 2 {
			args = append(args, "--"+flagAndValue[0])
			value := strings.Trim(strings.TrimSpace(flagAndValue[1]), "'`")
			args = append(args, value)
		}
	}

	return args
}

func (_self *crawlerService) executeCurlCommand(ctx context.Context, args []string) ([]byte, error) {
	cmd := exec.Command("curl", args...)
	output, err := cmd.Output()
	if err != nil {
		logging.Error(ctx, "curl execution error: %v", err)
		return nil, fmt.Errorf("error executing curl command: %v", err)
	}
	return output, nil
}

func (_self *crawlerService) processGoldResult(ctx context.Context, url entity.CrawlerEvent, output []byte) {
	// Send to Telegram
	if err := _self.teleService.SendMessage(entity.ExtractGoldPrice(output), "text"); err != nil {
		logging.Error(ctx, "send price error: %s", err.Error())
	}

	// Save to database
	if err := _self.resultRepo.CreateResult(ctx, &domain.Result{
		Url:    url.Url,
		Method: url.Method,
		Queue:  url.Queue,
		Domain: url.Domain,
		Result: string(output),
	}); err != nil {
		logging.Error(ctx, "create result error: %s", err.Error())
	}

	logging.Debug(ctx, "send message to Telegram: %v", string(output))
	logging.Debug(ctx, "=======================================")
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
				link := attr.Val
				// Handle relative URLs
				if strings.HasPrefix(link, "/") {
					// Extract base domain from baseURL
					if strings.HasPrefix(baseURL, "http://") || strings.HasPrefix(baseURL, "https://") {
						parts := strings.SplitN(baseURL, "//", 2)
						if len(parts) == 2 {
							domain := strings.Split(parts[1], "/")[0]
							link = parts[0] + "//" + domain + link
						}
					}
				} else if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
					// Skip relative links that don't start with /
					break
				}
				if isValidURL(link) {
					links = append(links, link)
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

func newCrawlSession() *crawlSession {
	return &crawlSession{
		visited: make(map[string]bool),
		results: make(map[string]string),
	}
}

func (s *crawlSession) shouldCrawl(url string, depth, maxDepth int64) bool {
	if depth > maxDepth {
		return false
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.visited[url] {
		return false
	}
	return true
}

func (s *crawlSession) markAsVisited(url string) {
	s.mutex.Lock()
	s.visited[url] = true
	s.mutex.Unlock()
}

func (s *crawlSession) isAlreadyVisited(url string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.visited[url]
}

func (s *crawlSession) shouldQueueLink(link string) bool {
	return !s.isAlreadyVisited(link) && isValidURL(link)
}
