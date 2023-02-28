package errors

import (
	"fmt"
)

// CrawlError represents an error that occurred while crawling a website.
type CrawlError struct {
	URL     string // The URL that was being crawled when the error occurred.
	Message string // A message describing the error.
}

// NewCrawlError creates a new instance of CrawlError with the given URL and error message.
func NewCrawlError(url string, msg string) *CrawlError {
	return &CrawlError{
		URL:     url,
		Message: msg,
	}
}

// Error returns a string representation of the CrawlError.
func (e *CrawlError) Error() string {
	return fmt.Sprintf("CrawlError: %s (%s)", e.Message, e.URL)
}
