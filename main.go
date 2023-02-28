package main

import (
	"github.com/fcampos1205/mirroring-web-crawler/config"
	"github.com/fcampos1205/mirroring-web-crawler/crawler"
	"go.uber.org/zap"
)

func main() {
	// Load configuration settings
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("failed to start config", zap.Error(err))
	}

	// Initialize crawler
	c := crawler.NewCrawler(cfg)
	c.Start()

	// Log summary of the crawl
	zap.L().Info("Crawling process exit.")
}
