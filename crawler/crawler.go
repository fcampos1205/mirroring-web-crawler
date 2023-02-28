package crawler

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"github.com/fcampos1205/mirroring-web-crawler/config"
	"github.com/fcampos1205/mirroring-web-crawler/errors"
	"github.com/fcampos1205/mirroring-web-crawler/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

// Crawler represents a web crawler.
type Crawler struct {
	cfg         *config.Config
	store       storage.Storage
	visitedURLs map[string]bool
	mutex       *sync.Mutex
}

// NewCrawler creates a new Crawler instance.
func NewCrawler(cfg *config.Config) *Crawler {
	return &Crawler{
		cfg:         cfg,
		store:       storage.NewDiskStorage(cfg.DirectoryPath),
		visitedURLs: make(map[string]bool),
		mutex:       &sync.Mutex{},
	}
}

// Crawl starts the crawling process from the given start URLs using the given worker pool.
func (c *Crawler) Start() {
	c.visitedURLs[c.cfg.StartURL] = false

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-quit:
			cancel()
			zap.L().Warn("stopping crawler flow")
			return
		default:
			var nextURLs = make([]string, 0)
			for nextURL, ok := range c.visitedURLs {
				if ok {
					continue
				}

				nextURLs = append(nextURLs, nextURL)
			}

			if len(nextURLs) == 0 {
				zap.L().Info("finished URL to be crawled")
				return
			}

			c.crawlWorker(ctx, nextURLs)
		}
	}
}

// crawlWorker is a worker function that crawls URLs from the `urlsToCrawl` channel.
func (c *Crawler) crawlWorker(ctx context.Context, nextURLs []string) {
	sem := semaphore.NewWeighted(int64(c.cfg.NumWorkers))

	for _, u := range nextURLs {
		if err := sem.Acquire(ctx, 1); err != nil {
			zap.L().Error("failed to acquire semaphore", zap.Error(err))
			break
		}

		go func(ctx context.Context, targetURL string) {
			defer sem.Release(1)

			if err := c.worker(targetURL); err != nil {
				zap.L().Error("failed to crawl", zap.Error(err))
			}
		}(ctx, u)
	}

	if err := sem.Acquire(ctx, int64(c.cfg.NumWorkers)); err != nil {
		zap.L().Error("failed to acquire semaphore", zap.Error(err))
	}
}

func (c *Crawler) worker(urlToCrawl string) error {
	filePath, err := c.createFilePath(urlToCrawl)
	if err != nil {
		return err
	}

	body, err := c.crawlURL(urlToCrawl, filePath)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	c.visitedURLs[urlToCrawl] = true
	c.mutex.Unlock()

	childLinks, err := c.parseLinks(body)
	if err != nil {
		return err
	}

	for _, childURL := range childLinks {
		zap.L().Info("new URL to be added", zap.String("URL", childURL))

		if isVisited, ok := c.visitedURLs[childURL]; !ok && !isVisited {
			c.mutex.Lock()
			c.visitedURLs[childURL] = false
			c.mutex.Unlock()
		}
	}

	return nil
}

func (c *Crawler) createFilePath(urlToCrawl string) (string, error) {
	zap.L().Info("crawling ...", zap.String("URL", urlToCrawl))

	if isVisited, ok := c.visitedURLs[urlToCrawl]; ok {
		if isVisited {
			return "", errors.NewCrawlError(urlToCrawl, "URL was already crawled")
		}
	}

	parsedURL, err := url.Parse(urlToCrawl)
	if err != nil {
		return "", err
	}

	urlPath := parsedURL.Path
	if urlToCrawl == c.cfg.StartURL {
		urlPath = "index.html"
	}

	filePath := filepath.Join(c.cfg.DirectoryPath, urlPath)
	if filepath.Ext(urlPath) == "" {
		filePath += ".html"
	}

	return filePath, nil
}

// crawlURL crawls the given URL and returns the response body and links on the page.
func (c *Crawler) crawlURL(urlToCrawl, filePath string) ([]byte, error) {
	// logger.Debug("Crawling URL %s", urlToCrawl)

	// Send HTTP GET request
	resp, err := http.Get(urlToCrawl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewCrawlError(urlToCrawl, fmt.Sprintf("received statuscode %d", resp.StatusCode))
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fileRetrived, err := c.store.Retrieve(filePath)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(body, fileRetrived) {
		zap.L().Warn("it was crawled", zap.String("URL", urlToCrawl), zap.String("filePath", filePath))
		return body, nil
	}

	if err := c.store.Store(filePath, body); err != nil {
		return nil, err
	}

	return body, nil
}

// parseLinks parses the links on the page.
func (c *Crawler) parseLinks(body []byte) ([]string, error) {
	newURLs := make([]string, 0)

	var attribute string = `a href`
	regex := regexp.MustCompile(attribute + `="(.*?)"`)
	matches := regex.FindAllString(string(body), -1)

	for _, match := range matches {
		value := strings.TrimPrefix(match, attribute+`="`)[:(len(match)-len(attribute+`="`))-1]

		for _, invalidPrefix := range []string{"#popup:", "font-family:", "'"} {
			if strings.Contains(value, invalidPrefix) {
				continue
			}
		}

		// check if is from the same domain
		if strings.HasPrefix(value, "http") && !strings.HasPrefix(value, c.cfg.StartURL) {
			continue
		}

		value = strings.TrimPrefix(value, c.cfg.StartURL)

		if _, ok := c.visitedURLs[value]; ok {
			continue
		}

		// if the URL is relative, make it absolute
		if strings.HasPrefix(value, "/") {
			value = c.cfg.StartURL + value
		}

		newURLs = append(newURLs, value)
	}

	return newURLs, nil
}
