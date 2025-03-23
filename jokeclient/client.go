package jokeclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AriT93/ai-agent/model"
)

// Client represents a joke API client
type Client struct {
	BaseURL string
	Timeout time.Duration
}

// NewClient creates a new joke API client
func NewClient() *Client {
	return &Client{
		BaseURL: "https://v2.jokeapi.dev/joke",
		Timeout: 5 * time.Second,
	}
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
	blacklistOptions := []string{"nsfw", "religious", "political", "racist", "sexist", "explicit"}
	for _, flag := range blacklistOptions {
		if strings.Contains(input, "no "+flag) || strings.Contains(input, "not "+flag) {
			blacklistFlags = append(blacklistFlags, flag)
		}
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
