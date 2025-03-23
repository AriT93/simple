package jokeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AriT93/ai-agent/model"
)

// ANSI color codes
const (
	resetColor = "\033[0m"
	boldText   = "\033[1m"
)

// Client represents a joke API client
type Client struct {
	BaseURL   string
	Timeout   time.Duration
	Debug     bool   // Debug flag
	DebugFile string // File to write debug output to
}

// NewClient creates a new joke API client
func NewClient(debug ...bool) *Client {
	// Default debug to false, but allow it to be enabled via optional parameter
	isDebug := false
	if len(debug) > 0 {
		isDebug = debug[0]
	}

	client := &Client{
		BaseURL:   "https://v2.jokeapi.dev/joke",
		Timeout:   5 * time.Second,
		Debug:     isDebug,
		DebugFile: "joke_api_debug.log",
	}

	// Initialize debug writer if debug is enabled
	if isDebug {
		// Create or truncate debug file
		file, err := os.OpenFile(client.DebugFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			// Write header to debug file
			file.WriteString("=== Joke API Debug Log ===\n\n")
			file.Close()
		}
	}

	return client
}

// writeDebug writes debug output to the specified file
func (c *Client) writeDebug(format string, args ...interface{}) {
	if !c.Debug {
		return
	}

	// Open file in append mode
	file, err := os.OpenFile(c.DebugFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	// Write timestamp
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(file, "%s[%s] %s", resetColor, timestamp, boldText)

	// Write formatted message
	fmt.Fprintf(file, format, args...)

	// Add newline if not present
	if !strings.HasSuffix(format, "\n") {
		file.WriteString("\n")
	}
}

// WriteDebug writes debug output to the file with proper formatting
func (c *Client) WriteDebug(format string, args ...interface{}) {
	if !c.Debug {
		return
	}

	c.writeDebug(format, args...)
}

// WriteDebugSeparator adds a visual separator to the debug log
func (c *Client) WriteDebugSeparator(title string) {
	if !c.Debug {
		return
	}

	c.writeDebug("\n%s=== %s ===%s\n", boldText, title, resetColor)
}

// FetchJoke fetches a joke from the API based on the given parameters
func (c *Client) FetchJoke(input string) (string, error) {
	// Parse input to extract category and joke type
	input = strings.ToLower(input)

	// Initialize parameters
	jokeCategory := "Any" // Default category
	jokeType := ""        // No specific type by default
	blacklistFlags := []string{}

	// Check for joke type
	if strings.Contains(input, "twopart") {
		jokeType = "twopart"
	} else if strings.Contains(input, "single") {
		jokeType = "single"
	}

	// Check for categories
	categories := []string{"programming", "misc", "dark", "pun", "spooky", "christmas"}
	for _, category := range categories {
		if strings.Contains(input, category) {
			jokeCategory = strings.Title(category) // Capitalize first letter for API
			break
		}
	}

	// Check for blacklist flags
	blacklistOptions := []string{"religious", "political", "racist", "sexist", "explicit"}
	nsfw := false

	// Check if NSFW is explicitly requested (expand patterns to check)
	if strings.Contains(input, "nsfw") || strings.Contains(input, "dirty") ||
		strings.Contains(input, "adult") || strings.Contains(input, "explicit") {
		nsfw = true
	}

	// Add other blacklist flags, but not NSFW if it was requested
	for _, flag := range blacklistOptions {
		if strings.Contains(input, "no "+flag) || strings.Contains(input, "not "+flag) {
			blacklistFlags = append(blacklistFlags, flag)
		}
	}

	// Only blacklist NSFW if not explicitly requested
	if !nsfw && (strings.Contains(input, "no nsfw") || strings.Contains(input, "not nsfw") || strings.Contains(input, "clean joke")) {
		blacklistFlags = append(blacklistFlags, "nsfw")
	}

	// Construct URL with proper path parameters
	url := fmt.Sprintf("%s/%s", c.BaseURL, jokeCategory)

	// Add query parameters
	params := []string{}

	if jokeType != "" {
		params = append(params, "type="+jokeType)
	}

	if len(blacklistFlags) > 0 {
		params = append(params, "blacklistFlags="+strings.Join(blacklistFlags, ","))
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	// Debug output for request URL
	if c.Debug {
		c.writeDebug("REQUEST: %s\n", url)
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// Create a new request with the context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("API request timed out after %v seconds", c.Timeout.Seconds())
		}
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Debug output for response JSON
	if c.Debug {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, body, "", "  ")
		if err != nil {
			c.writeDebug("RESPONSE (raw): %s\n", string(body))
		} else {
			c.writeDebug("RESPONSE JSON:\n%s\n", prettyJSON.String())
		}
	}

	// Unmarshal the JSON response
	var joke model.JokeResponse
	err = json.Unmarshal(body, &joke)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Format the joke based on its type
	var formattedJoke string
	if joke.Type == "single" {
		formattedJoke = joke.Joke
	} else if joke.Type == "twopart" {
		formattedJoke = fmt.Sprintf("%s\n\n%s", joke.Setup, joke.Delivery)
	} else {
		return "", fmt.Errorf("unknown joke type: %s", joke.Type)
	}

	return formattedJoke, nil
}
