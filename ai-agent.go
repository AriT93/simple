package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	textInput  textinput.Model
	viewport   viewport.Model
	messages   []string
	err        error
	spinner    spinner.Model
	processing bool
}

func initialModel() model {
	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to AI Assistant!\nType 'help' for instructions or start typing your request.\n")

	ti := textinput.New()
	ti.Placeholder = "Ask for a joke..."
	ti.Focus()
	ti.Width = 80

	return model{
		messages:   []string{"Welcome to AI Assistant!", "Type 'help' for instructions or start typing your request."},
		viewport:   vp,
		textInput:  ti,
		spinner:    spinner.New(),
		processing: false,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

type jokeResponseMsg struct {
	joke string
}

type errorResponseMsg struct {
	err error
}

const helpMessage = `
AI Assistant Help:
-----------------
This is a natural language interface for fetching jokes.

Commands:
- Type a request like "Tell me a joke about programming"
- Add parameters like "category:programming" or "type:twopart" 
- Type "quit" or press ESC to exit
- Type "help" to show this message

Parameters:
- category: [programming, misc, dark, pun, spooky, christmas]
- type: [single, twopart]
- blacklist: [nsfw, religious, political, racist, sexist, explicit]

Examples:
"Tell me a programming joke"
"Give me a twopart joke about christmas"
"Tell me a joke but nothing nsfw or political"
`

// JokeResponse struct to hold the API response
type JokeResponse struct {
	Error    bool   `json:"error"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Setup    string `json:"setup"`
	Delivery string `json:"delivery"`
	Joke     string `json:"joke"`
	Flags    struct {
		Nsfw      bool `json:"nsfw"`
		Religious bool `json:"religious"`
		Political bool `json:"political"`
		Racist    bool `json:"racist"`
		Sexist    bool `json:"sexist"`
		Explicit  bool `json:"explicit"`
	} `json:"flags"`
	ID   int    `json:"id"`
	Safe bool   `json:"safe"`
	Lang string `json:"lang"`
}

// Command to fetch a joke asynchronously
func fetchJokeCmd(input string) tea.Cmd {
	return func() tea.Msg {
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
		url := fmt.Sprintf("https://v2.jokeapi.dev/joke/%s", jokeCategory)

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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create a new request with the context
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return errorResponseMsg{err: fmt.Errorf("failed to create request: %w", err)}
		}

		// Make the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return errorResponseMsg{err: fmt.Errorf("API request timed out after 5 seconds")}
			}
			return errorResponseMsg{err: fmt.Errorf("API request failed: %w", err)}
		}
		defer resp.Body.Close()

		// Check the response status code
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errorResponseMsg{err: fmt.Errorf("API request failed with status code: %d", resp.StatusCode)}
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorResponseMsg{err: fmt.Errorf("failed to read response body: %w", err)}
		}

		// Unmarshal the JSON response
		var joke JokeResponse
		err = json.Unmarshal(body, &joke)
		if err != nil {
			return errorResponseMsg{err: fmt.Errorf("failed to unmarshal JSON: %w", err)}
		}

		// Format the joke based on its type
		var formattedJoke string
		if joke.Type == "single" {
			formattedJoke = joke.Joke
		} else if joke.Type == "twopart" {
			formattedJoke = fmt.Sprintf("%s\n\n%s", joke.Setup, joke.Delivery)
		} else {
			return errorResponseMsg{err: fmt.Errorf("unknown joke type: %s", joke.Type)}
		}

		return jokeResponseMsg{joke: formattedJoke}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			input := m.textInput.Value()
			if input == "" {
				return m, nil
			}

			// Handle quit command
			if input == "quit" || input == "exit" {
				return m, tea.Quit
			}

			// Reset input
			m.textInput.Reset()

			// Add user message to history
			m.messages = append(m.messages, "You: "+input)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			// Handle help command
			if input == "help" {
				m.messages = append(m.messages, helpMessage)
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.viewport.GotoBottom()
				return m, nil
			}

			// Initiate joke fetching
			m.processing = true
			return m, tea.Batch(fetchJokeCmd(input), m.spinner.Tick)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case jokeResponseMsg:
		m.processing = false
		m.messages = append(m.messages, "AI: "+wordWrap(msg.joke, 70))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case errorResponseMsg:
		m.processing = false
		m.messages = append(m.messages, "Error: "+msg.err.Error())
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func wordWrap(text string, lineWidth int) string {
	var result string
	words := strings.Fields(text)

	if len(words) == 0 {
		return ""
	}

	line := words[0]

	for _, word := range words[1:] {
		if len(line)+len(word)+1 > lineWidth {
			result += line + "\n"
			line = word
		} else {
			line += " " + word
		}
	}
	result += line
	return result
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	var status string
	if m.processing {
		status = m.spinner.View() + " Getting response..."
	} else {
		status = ""
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"AI Agent Chat",
		"------------",
		m.viewport.View(),
		status,
		m.textInput.View(),
		"Type 'help' for instructions or 'quit' to exit.",
		"Press Ctrl+C to quit.",
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
