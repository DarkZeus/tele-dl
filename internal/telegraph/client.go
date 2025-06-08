package telegraph

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client handles Telegraph API communication
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Telegraph API client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    baseURL,
	}
}

// FetchPage retrieves a Telegraph page by path
func (c *Client) FetchPage(pagePath string) (*Page, error) {
	url := fmt.Sprintf("%s%s?return_content=true", c.baseURL, pagePath)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var telegraphResp Response
	if err := json.NewDecoder(resp.Body).Decode(&telegraphResp); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	if !telegraphResp.Ok {
		return nil, fmt.Errorf("Telegraph API returned error")
	}

	content := c.parseContent(telegraphResp.Result.RawContent)
	
	return &Page{
		Title:   telegraphResp.Result.Title,
		Content: content,
	}, nil
}

// parseContent converts mixed RawContent to ContentNodes
func (c *Client) parseContent(rawContent []interface{}) []ContentNode {
	var nodes []ContentNode
	
	for _, item := range rawContent {
		if v, ok := item.(map[string]interface{}); ok {
			if itemBytes, err := json.Marshal(v); err == nil {
				var node ContentNode
				if err := json.Unmarshal(itemBytes, &node); err == nil {
					nodes = append(nodes, node)
				}
			}
		}
		// Skip strings and other types
	}
	
	return nodes
}

// ExtractPagePath extracts the page path from a Telegraph URL
func ExtractPagePath(url string) (string, error) {
	pagePath := strings.TrimPrefix(url, "https://telegra.ph/")
	if pagePath == url {
		return "", fmt.Errorf("invalid Telegraph URL: must start with https://telegra.ph/")
	}
	return pagePath, nil
} 