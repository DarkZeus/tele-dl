package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tele-dl/internal/config"
	"tele-dl/internal/downloader"
	"tele-dl/internal/parser"
	"tele-dl/internal/telegraph"
	"tele-dl/internal/utils"

	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Application represents the main application with modern Go features
type Application struct {
	config           *config.Config
	logger           *logrus.Logger
	telegraphClient  *telegraph.Client
	mediaParser      *parser.MediaParser
	downloader       *downloader.Downloader
	progressBar      *progressbar.ProgressBar
}

// Result represents the final application result with detailed information
type Result struct {
	URL           string                    `json:"url"`
	Title         string                    `json:"title"`
	MediaFound    int                       `json:"media_found"`
	Downloaded    int                       `json:"downloaded"`
	Failed        int                       `json:"failed"`
	Skipped       int                       `json:"skipped"`
	TotalSize     int64                     `json:"total_size_bytes"`
	TotalSizeStr  string                    `json:"total_size"`
	Duration      string                    `json:"duration"`
	OutputDir     string                    `json:"output_dir"`
	MediaItems    []telegraph.MediaItem     `json:"media_items,omitempty"`
	Errors        []string                  `json:"errors,omitempty"`
	Statistics    *parser.Stats             `json:"parsing_stats,omitempty"`
}

// New creates a new application instance with modern configuration
func New(cfg *config.Config) *Application {
	logger := logrus.New()
	
	// Configure logger based on settings
	if cfg.Quiet {
		logger.SetLevel(logrus.ErrorLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	
	// Use JSON formatter for structured logging
	if cfg.JSONOutput {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})
	}
	
	return &Application{
		config:          cfg,
		logger:          logger,
		telegraphClient: telegraph.NewClient(cfg.TelegraphAPIBase, cfg.Timeout),
		mediaParser:     parser.New(),
		downloader:      downloader.New(cfg.Workers, cfg.Timeout, cfg.OutputDir, cfg.TelegraphFileBase, cfg.Retries),
	}
}

// RunDownload is the main entry point called by Cobra command
func RunDownload(cmd *cobra.Command, args []string) error {
	// Parse configuration from command flags
	cfg, err := config.FromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	
	// Create application instance
	app := New(cfg)
	
	// Run with context for proper cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle timeout if specified
	if cfg.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}
	
	return app.Run(ctx)
}

// Run executes the main application logic with modern Go patterns
func (a *Application) Run(ctx context.Context) error {
	startTime := time.Now()
	
	a.logger.WithFields(logrus.Fields{
		"url":         a.config.URL,
		"output_dir":  a.config.OutputDir,
		"workers":     a.config.Workers,
		"retries":     a.config.Retries,
	}).Info("Starting download process")
	
	// Create output directory if it doesn't exist
	if err := a.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Step 1: Fetch Telegraph page content
	telegraphData, err := a.fetchTelegraphData(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch Telegraph data: %w", err)
	}
	
	a.logger.WithField("title", telegraphData.Content.Title).Info("Successfully fetched Telegraph page")
	
	// Step 2: Parse media content
	mediaItems, stats := a.mediaParser.ExtractWithStats(telegraphData.Content)
	
	// Validate media items
	if validationErrs := a.mediaParser.ValidateMedia(mediaItems); validationErrs.HasErrors() {
		a.logger.WithError(validationErrs).Warn("Media validation warnings")
	}
	
	// Deduplicate URLs
	mediaItems = a.mediaParser.DeduplicateURLs(mediaItems)
	
	a.logger.WithFields(logrus.Fields{
		"total_nodes":   stats.TotalNodes,
		"media_nodes":   stats.MediaNodes,
		"valid_media":   stats.ValidMedia,
		"unique_urls":   stats.UniqueURLs,
	}).Info("Media parsing completed")
	
	// Check if any media was found
	if len(mediaItems) == 0 {
		a.logger.Warn("No media files found in the Telegraph page")
		return a.outputResult(Result{
			URL:        a.config.URL,
			Title:      telegraphData.Content.Title,
			MediaFound: 0,
			Duration:   time.Since(startTime).String(),
			OutputDir:  a.config.OutputDir,
			Statistics: &stats,
		})
	}
	
	// Step 3: Download media files
	downloadStats, downloadResults := a.downloadMedia(ctx, mediaItems)
	
	// Collect any download errors
	var downloadErrors []string
	for _, result := range downloadResults {
		if result.Error != nil {
			downloadErrors = append(downloadErrors, result.Error.Error())
		}
	}
	
	// Prepare final result
	result := Result{
		URL:          a.config.URL,
		Title:        telegraphData.Content.Title,
		MediaFound:   len(mediaItems),
		Downloaded:   downloadStats.Successful,
		Failed:       downloadStats.Failed,
		Skipped:      downloadStats.Skipped,
		TotalSize:    downloadStats.TotalSize,
		TotalSizeStr: utils.FormatBytes(downloadStats.TotalSize),
		Duration:     downloadStats.Duration.String(),
		OutputDir:    a.config.OutputDir,
		Errors:       downloadErrors,
		Statistics:   &stats,
	}
	
	// Include media items in result if not in quiet mode
	if !a.config.Quiet {
		result.MediaItems = mediaItems
	}
	
	a.logger.WithFields(logrus.Fields{
		"downloaded": downloadStats.Successful,
		"failed":     downloadStats.Failed,
		"skipped":    downloadStats.Skipped,
		"total_size": result.TotalSizeStr,
		"duration":   result.Duration,
	}).Info("Download process completed")
	
	return a.outputResult(result)
}

// ensureOutputDir creates the output directory if it doesn't exist
func (a *Application) ensureOutputDir() error {
	if err := os.MkdirAll(a.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", a.config.OutputDir, err)
	}
	
	// Verify directory is writable
	testFile := filepath.Join(a.config.OutputDir, ".write_test")
	if file, err := os.Create(testFile); err != nil {
		return fmt.Errorf("output directory %s is not writable: %w", a.config.OutputDir, err)
	} else {
		file.Close()
		os.Remove(testFile)
	}
	
	return nil
}

// fetchTelegraphData fetches and parses Telegraph page data with context
func (a *Application) fetchTelegraphData(ctx context.Context) (telegraph.ParsedResponse, error) {
	// Extract path from URL
	path, err := telegraph.ExtractPagePath(a.config.URL)
	if err != nil {
		return telegraph.ParsedResponse{}, fmt.Errorf("invalid Telegraph URL: %w", err)
	}
	
	// Fetch page data
	page, err := a.telegraphClient.FetchPage(path)
	if err != nil {
		return telegraph.ParsedResponse{}, fmt.Errorf("failed to fetch page: %w", err)
	}
	
	// Convert old Page structure to new format
	content := telegraph.Content{
		Title: page.Title,
		Nodes: page.Content,
	}
	
	return telegraph.ParsedResponse{
		Content: content,
	}, nil
}

// downloadMedia handles the download process with progress tracking
func (a *Application) downloadMedia(ctx context.Context, mediaItems []telegraph.MediaItem) (*downloader.Stats, []downloader.Result) {
	// Initialize progress bar if enabled
	if a.config.Progress && !a.config.Quiet {
		a.progressBar = progressbar.NewOptions(len(mediaItems),
			progressbar.OptionSetDescription("Downloading media files..."),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("files"),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowElapsedTimeOnFinish(),
			progressbar.OptionSetRenderBlankState(true),
		)
	}
	
	// Create progress callback
	progressCallback := func(completed int) {
		if a.progressBar != nil {
			a.progressBar.Set(completed)
		}
	}
	
	// Download with context support
	stats, results := a.downloader.DownloadAllWithContext(ctx, mediaItems, progressCallback)
	
	// Finish progress bar
	if a.progressBar != nil {
		a.progressBar.Finish()
		fmt.Println() // Add newline after progress bar
	}
	
	return stats, results
}

// outputResult outputs the final result in the appropriate format
func (a *Application) outputResult(result Result) error {
	if a.config.JSONOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
	
	// Human-readable output
	if a.config.Quiet {
		// Minimal output for quiet mode
		if result.Failed > 0 {
			return fmt.Errorf("download completed with %d failures", result.Failed)
		}
		return nil
	}
	
	// Full summary
	fmt.Printf("\n📊 Download Summary\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("🔗 URL: %s\n", result.URL)
	fmt.Printf("📄 Title: %s\n", result.Title)
	fmt.Printf("📁 Output: %s\n", result.OutputDir)
	fmt.Printf("📊 Media found: %d\n", result.MediaFound)
	fmt.Printf("✅ Downloaded: %d\n", result.Downloaded)
	fmt.Printf("⏭️  Skipped: %d\n", result.Skipped)
	if result.Failed > 0 {
		fmt.Printf("❌ Failed: %d\n", result.Failed)
	}
	fmt.Printf("📦 Total size: %s\n", result.TotalSizeStr)
	fmt.Printf("⏱️  Duration: %s\n", result.Duration)
	
	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Errors encountered:\n")
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
	}
	
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	
	// Return error if there were failures and user wants to know
	if result.Failed > 0 {
		return fmt.Errorf("completed with %d failed downloads", result.Failed)
	}
	
	return nil
}

 