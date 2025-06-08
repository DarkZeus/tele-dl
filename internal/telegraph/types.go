package telegraph

import (
	"encoding/json"
	"errors"
)

// Response represents the Telegraph API response structure
type Response struct {
	Ok     bool `json:"ok"`
	Result struct {
		Path        string        `json:"path"`
		URL         string        `json:"url"`
		Title       string        `json:"title"`
		Description string        `json:"description"`
		AuthorName  string        `json:"author_name"`
		AuthorURL   string        `json:"author_url"`
		ImageURL    string        `json:"image_url"`
		RawContent  []interface{} `json:"content"`
	} `json:"result"`
	Error string `json:"error,omitempty"`
}

// ContentNode represents a parsed content node
type ContentNode struct {
	Tag      string                 `json:"tag,omitempty"`
	Attrs    map[string]interface{} `json:"attrs,omitempty"`
	Children []ContentNode          `json:"children,omitempty"`
}

// MediaItem represents a downloadable media item
type MediaItem struct {
	URL      string
	Filename string
}

// Generic Result type for operations that can succeed or fail
type Result[T any] struct {
	Value T
	Err   error
}

// NewResult creates a successful result
func NewResult[T any](value T) Result[T] {
	return Result[T]{Value: value}
}

// NewErrorResult creates a failed result
func NewErrorResult[T any](err error) Result[T] {
	return Result[T]{Err: err}
}

// IsOk returns true if the result contains no error
func (r Result[T]) IsOk() bool {
	return r.Err == nil
}

// Unwrap returns the value or panics if there's an error
func (r Result[T]) Unwrap() T {
	if r.Err != nil {
		panic(r.Err)
	}
	return r.Value
}

// UnwrapOr returns the value or the provided default if there's an error
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.Err != nil {
		return defaultValue
	}
	return r.Value
}

// Content holds the parsed content tree
type Content struct {
	Title string        `json:"title"`
	Nodes []ContentNode `json:"nodes"`
}

// Page represents a Telegraph page (legacy compatibility)
type Page struct {
	Title   string        `json:"title"`
	Content []ContentNode `json:"content"`
}

// ParsedResponse wraps the API response with parsed content
type ParsedResponse struct {
	Response
	Content Content
}

// UnmarshalJSON implements custom JSON unmarshaling for ContentNode
// This handles Telegraph's inconsistent JSON structure where content can be mixed types
func (c *ContentNode) UnmarshalJSON(data []byte) error {
	// First try to unmarshal as string (for text content)
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		// This is a text node, skip it as we only want media elements
		return nil
	}

	// Try to unmarshal as object
	var obj struct {
		Tag      string                 `json:"tag,omitempty"`
		Attrs    map[string]interface{} `json:"attrs,omitempty"`
		Children []interface{}          `json:"children,omitempty"`
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	c.Tag = obj.Tag
	c.Attrs = obj.Attrs

	// Process children recursively
	for _, child := range obj.Children {
		childData, err := json.Marshal(child)
		if err != nil {
			continue
		}

		var childNode ContentNode
		if err := json.Unmarshal(childData, &childNode); err == nil {
			c.Children = append(c.Children, childNode)
		}
	}

	return nil
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return "validation error: " + e.Field + " " + e.Message
}

// MultiError holds multiple errors that can be joined
type MultiError struct {
	Errors []error
}

func (m MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}
	// Use Go 1.20+ errors.Join for modern error handling
	return errors.Join(m.Errors...).Error()
}

// AddError adds an error to the collection
func (m *MultiError) AddError(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// HasErrors returns true if there are any errors
func (m MultiError) HasErrors() bool {
	return len(m.Errors) > 0
} 