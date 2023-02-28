package config

import (
	"errors"
	"flag"
	"runtime"

	"github.com/fcampos1205/mirroring-web-crawler/logger"
)

// Config holds the configuration settings for the crawler.
type Config struct {
	StartURL      string
	DirectoryPath string
	NumWorkers    int
}

// Load loads the configuration settings from command line arguments and environment variables.
func Load() (*Config, error) {
	cfg := new(Config)
	targetURL := flag.String("url", "", "enter URL to crawl")
	directoryPath := flag.String("path", "", "enter URL to crawl")
	numWorkers := flag.Int("workers", runtime.GOMAXPROCS(0), "maximum number of concurrent requests")
	logLevel := flag.String("log-level", "info", "logging level (debug, info, warning, error)")
	flag.Parse()

	// Initialize logger
	logger.Start(*logLevel)

	if *targetURL == "" {
		return nil, errors.New("target URL is required")
	}

	if *directoryPath == "" {
		return nil, errors.New("directory path is required")
	}

	// Set default values
	cfg.StartURL = *targetURL
	cfg.DirectoryPath = *directoryPath
	cfg.NumWorkers = *numWorkers

	return cfg, nil
}
