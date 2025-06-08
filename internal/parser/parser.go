package parser

import (
	"fmt"
	"iter"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"tele-dl/internal/telegraph"
)

// MediaParser extracts media URLs from Telegraph content using modern Go features
type MediaParser struct {
	supportedTags    []string
	fileExtensions   []string
	urlFilter        func(string) bool
}

// Config holds parser configuration with modern defaults
type Config struct {
	SupportedTags  []string
	FileExtensions []string
	URLFilter      func(string) bool
}

// DefaultConfig returns sensible defaults for media parsing
func DefaultConfig() Config {
	return Config{
		SupportedTags:  []string{"img", "video", "audio", "source"},
		FileExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".mp4", ".mov", ".avi", ".mp3", ".wav"},
		URLFilter:      nil, // Accept all URLs by default
	}
}

// New creates a new media parser with default settings
func New() *MediaParser {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig creates a parser with custom configuration
func NewWithConfig(cfg Config) *MediaParser {
	return &MediaParser{
		supportedTags:  cfg.SupportedTags,
		fileExtensions: cfg.FileExtensions,
		urlFilter:      cfg.URLFilter,
	}
}

// ParsedMedia represents a successfully parsed media item with metadata
type ParsedMedia struct {
	telegraph.MediaItem
	Tag        string
	Alt        string
	Title      string
	NodeIndex  int
}

// ExtractMedia extracts all media URLs from the Telegraph page content
func (p *MediaParser) ExtractMedia(content telegraph.Content) []telegraph.MediaItem {
	var mediaItems []telegraph.MediaItem
	
	// Use modern iterator pattern to process nodes
	for item := range p.iterateMediaItems(content.Nodes) {
		mediaItems = append(mediaItems, item.MediaItem)
	}
	
	return mediaItems
}

// ExtractMediaDetailed returns detailed media information including metadata
func (p *MediaParser) ExtractMediaDetailed(content telegraph.Content) []ParsedMedia {
	var mediaItems []ParsedMedia
	
	for item := range p.iterateMediaItems(content.Nodes) {
		mediaItems = append(mediaItems, item)
	}
	
	return mediaItems
}

// iterateMediaItems uses Go 1.23+ iterator pattern to lazily process media items
func (p *MediaParser) iterateMediaItems(nodes []telegraph.ContentNode) iter.Seq[ParsedMedia] {
	return func(yield func(ParsedMedia) bool) {
		mediaIndex := 0
		p.walkNodesForMedia(nodes, &mediaIndex, yield)
	}
}

// walkNodesForMedia recursively walks content nodes and assigns sequential indices to media items only
func (p *MediaParser) walkNodesForMedia(nodes []telegraph.ContentNode, mediaIndex *int, yield func(ParsedMedia) bool) {
	for _, node := range nodes {
		if p.isMediaTag(node.Tag) {
			if item, ok := p.parseMediaNode(node, *mediaIndex); ok {
				if !yield(item) {
					return
				}
				*mediaIndex++ // Only increment for actual media items
			}
		}
		
		// Recursively process children
		if len(node.Children) > 0 {
			p.walkNodesForMedia(node.Children, mediaIndex, yield)
		}
	}
}

// parseMediaNode extracts media information from a content node
func (p *MediaParser) parseMediaNode(node telegraph.ContentNode, index int) (ParsedMedia, bool) {
	if node.Attrs == nil {
		return ParsedMedia{}, false
	}
	
	src, ok := node.Attrs["src"].(string)
	if !ok || src == "" {
		return ParsedMedia{}, false
	}
	
	// Apply URL filter if configured
	if p.urlFilter != nil && !p.urlFilter(src) {
		return ParsedMedia{}, false
	}
	
	// Extract additional metadata using modern string processing
	alt := p.extractStringAttr(node.Attrs, "alt")
	title := p.extractStringAttr(node.Attrs, "title")
	
	filename := p.generateFilename(src, index)
	
	return ParsedMedia{
		MediaItem: telegraph.MediaItem{
			URL:      src,
			Filename: filename,
		},
		Tag:       node.Tag,
		Alt:       alt,
		Title:     title,
		NodeIndex: index,
	}, true
}

// extractStringAttr safely extracts string attributes
func (p *MediaParser) extractStringAttr(attrs map[string]interface{}, key string) string {
	if val, ok := attrs[key].(string); ok {
		return val
	}
	return ""
}

// generateFilename creates a unique filename for the media item
func (p *MediaParser) generateFilename(url string, index int) string {
	// Extract filename from URL
	var baseFilename string
	
	if strings.Contains(url, "/") {
		parts := strings.Split(url, "/")
		baseFilename = parts[len(parts)-1]
	} else {
		baseFilename = url
	}
	
	// Remove query parameters and fragments
	if idx := strings.IndexAny(baseFilename, "?#"); idx != -1 {
		baseFilename = baseFilename[:idx]
	}
	
	// If no extension found, try to determine from URL patterns
	if filepath.Ext(baseFilename) == "" {
		baseFilename = p.addExtensionFromURL(baseFilename, url)
	}
	
	// Add index prefix for uniqueness
	return fmt.Sprintf("%d_%s", index, baseFilename)
}

// addExtensionFromURL attempts to add appropriate file extension
func (p *MediaParser) addExtensionFromURL(filename, url string) string {
	// Common patterns for Telegraph files
	if strings.Contains(url, "/file/") {
		// Telegraph files often don't have extensions in URL
		return filename + ".jpg" // Default assumption for Telegraph
	}
	
	// For external URLs, try to guess from URL patterns
	lowerURL := strings.ToLower(url)
	for _, ext := range p.fileExtensions {
		if strings.Contains(lowerURL, ext[1:]) { // Remove the dot for searching
			return filename + ext
		}
	}
	
	return filename + ".jpg" // Default fallback
}

// isMediaTag checks if a tag represents media content using modern slice operations
func (p *MediaParser) isMediaTag(tag string) bool {
	return slices.Contains(p.supportedTags, tag)
}

// FilterByExtension filters media items by file extension using modern Go patterns
func (p *MediaParser) FilterByExtension(items []telegraph.MediaItem, extensions ...string) []telegraph.MediaItem {
	if len(extensions) == 0 {
		return items
	}
	
	return slices.DeleteFunc(slices.Clone(items), func(item telegraph.MediaItem) bool {
		ext := strings.ToLower(filepath.Ext(item.Filename))
		return !slices.Contains(extensions, ext)
	})
}

// GroupByExtension groups media items by file extension using modern Go generics
func (p *MediaParser) GroupByExtension(items []telegraph.MediaItem) map[string][]telegraph.MediaItem {
	groups := make(map[string][]telegraph.MediaItem)
	
	for _, item := range items {
		ext := strings.ToLower(filepath.Ext(item.Filename))
		if ext == "" {
			ext = "unknown"
		}
		groups[ext] = append(groups[ext], item)
	}
	
	return groups
}

// Stats represents parsing statistics with modern error handling
type Stats struct {
	TotalNodes     int
	MediaNodes     int
	ValidMedia     int
	InvalidMedia   int
	UniqueURLs     int
	Errors         telegraph.MultiError
}

// ExtractWithStats extracts media and returns detailed statistics
func (p *MediaParser) ExtractWithStats(content telegraph.Content) ([]telegraph.MediaItem, Stats) {
	stats := Stats{}
	var mediaItems []telegraph.MediaItem
	urlSet := make(map[string]bool)
	
	for item := range p.iterateMediaItems(content.Nodes) {
		stats.MediaNodes++
		
		if item.URL == "" {
			stats.InvalidMedia++
			stats.Errors.AddError(fmt.Errorf("empty URL in node %d", item.NodeIndex))
			continue
		}
		
		stats.ValidMedia++
		mediaItems = append(mediaItems, item.MediaItem)
		
		if !urlSet[item.URL] {
			urlSet[item.URL] = true
		}
	}
	
	stats.TotalNodes = p.countTotalNodes(content.Nodes)
	stats.UniqueURLs = len(urlSet)
	
	return mediaItems, stats
}

// countTotalNodes counts all nodes in the content tree
func (p *MediaParser) countTotalNodes(nodes []telegraph.ContentNode) int {
	count := len(nodes)
	for _, node := range nodes {
		count += p.countTotalNodes(node.Children)
	}
	return count
}

// ValidateMedia validates media items against various criteria
func (p *MediaParser) ValidateMedia(items []telegraph.MediaItem) telegraph.MultiError {
	var errors telegraph.MultiError
	
	for i, item := range items {
		if item.URL == "" {
			errors.AddError(telegraph.ValidationError{
				Field:   fmt.Sprintf("items[%d].URL", i),
				Message: "cannot be empty",
			})
		}
		
		if item.Filename == "" {
			errors.AddError(telegraph.ValidationError{
				Field:   fmt.Sprintf("items[%d].Filename", i),
				Message: "cannot be empty",
			})
		}
		
		// Check for suspicious filenames
		if strings.Contains(item.Filename, "..") {
			errors.AddError(telegraph.ValidationError{
				Field:   fmt.Sprintf("items[%d].Filename", i),
				Message: "contains path traversal",
			})
		}
	}
	
	return errors
}

// DeduplicateURLs removes duplicate URLs while preserving order using modern Go features
func (p *MediaParser) DeduplicateURLs(items []telegraph.MediaItem) []telegraph.MediaItem {
	seen := make(map[string]bool)
	return slices.DeleteFunc(slices.Clone(items), func(item telegraph.MediaItem) bool {
		if seen[item.URL] {
			return true
		}
		seen[item.URL] = true
		return false
	})
}

// SortByFilename sorts media items by filename using modern comparison
func (p *MediaParser) SortByFilename(items []telegraph.MediaItem) {
	slices.SortFunc(items, func(a, b telegraph.MediaItem) int {
		return strings.Compare(a.Filename, b.Filename)
	})
}

// SortByIndex sorts media items by extracting the index from filename
func (p *MediaParser) SortByIndex(items []telegraph.MediaItem) {
	slices.SortFunc(items, func(a, b telegraph.MediaItem) int {
		indexA := p.extractIndexFromFilename(a.Filename)
		indexB := p.extractIndexFromFilename(b.Filename)
		return indexA - indexB
	})
}

// extractIndexFromFilename extracts the numeric index from a filename like "5_image.jpg"
func (p *MediaParser) extractIndexFromFilename(filename string) int {
	if idx := strings.Index(filename, "_"); idx != -1 {
		if num, err := strconv.Atoi(filename[:idx]); err == nil {
			return num
		}
	}
	return 0
} 