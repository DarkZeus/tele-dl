package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tele-dl/internal/telegraph"
	"tele-dl/internal/utils"
)

// Result represents the outcome of a download operation
type Result struct {
	Item     telegraph.MediaItem
	Size     int64
	Error    error
	Skipped  bool
}

// Stats holds download statistics
type Stats struct {
	Total       int
	Successful  int
	Failed      int
	Skipped     int
	TotalSize   int64
	Duration    time.Duration
}

// ProgressCallback is called when a download completes
type ProgressCallback func(completed int)

// Downloader handles concurrent media downloads with modern Go features
type Downloader struct {
	client            *http.Client
	maxWorkers        int
	outputDir         string
	telegraphFileBase string
	retries           int
}



// New creates a new downloader with retry support
func New(maxWorkers int, timeout time.Duration, outputDir, telegraphFileBase string, retries int) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10,
				MaxIdleConns:        100,
			},
		},
		maxWorkers:        maxWorkers,
		outputDir:         outputDir,
		telegraphFileBase: telegraphFileBase,
		retries:           retries,
	}
}

// DownloadAll downloads with context support for cancellation
func (d *Downloader) DownloadAllWithContext(ctx context.Context, items []telegraph.MediaItem, progressCallback ProgressCallback) (*Stats, []Result) {
	startTime := time.Now()
	
	jobs := make(chan telegraph.MediaItem, len(items))
	results := make(chan Result, len(items))
	
	// Start worker pool with context
	var wg sync.WaitGroup
	for i := 0; i < d.maxWorkers; i++ {
		wg.Add(1)
		go d.workerWithContext(ctx, jobs, results, &wg)
	}
	
	// Send jobs
	go func() {
		defer close(jobs)
		for _, item := range items {
			select {
			case jobs <- item:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	// Collect results with progress tracking
	go func() {
		wg.Wait()
		close(results)
	}()
	
	var allResults []Result
	stats := &Stats{Total: len(items)}
	completed := 0
	
	for result := range results {
		allResults = append(allResults, result)
		completed++
		
		if result.Error != nil {
			stats.Failed++
		} else if result.Skipped {
			stats.Skipped++
		} else {
			stats.Successful++
			stats.TotalSize += result.Size
		}
		
		// Call progress callback if provided
		if progressCallback != nil {
			progressCallback(completed)
		}
	}
	
	stats.Duration = time.Since(startTime)
	return stats, allResults
}

// workerWithContext processes download jobs with context support
func (d *Downloader) workerWithContext(ctx context.Context, jobs <-chan telegraph.MediaItem, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for {
		select {
		case item, ok := <-jobs:
			if !ok {
				return
			}
			result := d.downloadFileWithRetryAndContext(ctx, item)
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}



// downloadFileWithRetryAndContext downloads a single file with retry logic and context
func (d *Downloader) downloadFileWithRetryAndContext(ctx context.Context, item telegraph.MediaItem) Result {
	var errors telegraph.MultiError
	
	for attempt := 0; attempt <= d.retries; attempt++ {
		select {
		case <-ctx.Done():
			return Result{Item: item, Error: ctx.Err()}
		default:
		}
		
		result := d.downloadFileWithContext(ctx, item)
		
		// If successful or skipped, return immediately
		if result.Error == nil {
			return result
		}
		
		errors.AddError(fmt.Errorf("attempt %d: %w", attempt+1, result.Error))
		
		// Don't retry on certain errors (like 404)
		if strings.Contains(result.Error.Error(), "404") {
			break
		}
		
		// Wait before retry (exponential backoff)
		if attempt < d.retries {
			waitTime := time.Duration(attempt+1) * time.Second
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return Result{Item: item, Error: ctx.Err()}
			}
		}
	}
	
	return Result{Item: item, Error: errors}
}



// downloadFileWithContext downloads a single media file with context support
func (d *Downloader) downloadFileWithContext(ctx context.Context, item telegraph.MediaItem) Result {
	filePath := filepath.Join(d.outputDir, item.Filename)
	
	// Check if file already exists
	if d.fileExists(filePath) {
		if info, err := os.Stat(filePath); err == nil && info.Size() > 0 {
			fmt.Printf("[skip] %s already exists (%s)\n", item.Filename, utils.FormatBytes(info.Size()))
			return Result{
				Item:    item,
				Size:    info.Size(),
				Skipped: true,
			}
		}
	}
	
	// Build full URL
	fullURL := d.buildURL(item.URL)
	
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return Result{Item: item, Error: fmt.Errorf("failed to create request: %w", err)}
	}
	
	// Download file
	resp, err := d.client.Do(req)
	if err != nil {
		return Result{Item: item, Error: fmt.Errorf("HTTP request failed: %w", err)}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return Result{Item: item, Error: fmt.Errorf("HTTP %d", resp.StatusCode)}
	}
	
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return Result{Item: item, Error: fmt.Errorf("failed to create file: %w", err)}
	}
	defer file.Close()
	
	// Copy with context-aware reader
	size, err := d.copyWithContext(ctx, file, resp.Body)
	if err != nil {
		// Clean up partial file on error
		os.Remove(filePath)
		return Result{Item: item, Error: fmt.Errorf("failed to copy file: %w", err)}
	}
	
	return Result{Item: item, Size: size}
}



// copyWithContext copies data from src to dst with context cancellation support
func (d *Downloader) copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	// Use a reasonable buffer size for copying
	buf := make([]byte, 32*1024)
	var written int64
	
	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}
		
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return written, er
			}
			break
		}
	}
	return written, nil
}

// buildURL constructs the full URL for downloading
func (d *Downloader) buildURL(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return d.telegraphFileBase + url
}

// fileExists checks if a file exists
func (d *Downloader) fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

 