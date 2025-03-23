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
	// "github.com/yourusername/yourproject/utils" // Removed or replace with the correct package
)

type model struct {
	textInput  textinput.Model
	viewport   viewport.Model
	messages   []string
	err        error
	spinner    spinner.Model
	processing bool
	jokeChan   chan tea.Msg
}

func initialModel() model {
	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to AI Assistant!\nType 'help' for instructions or start typing your request.\n")

	ti := textinput.New()
	ti.Placeholder = "Ask for a joke..."
	ti.Focus() // Ensure text input is focused from the start
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
	// Return both commands to ensure input is ready
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

type jokeMsg string
type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
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

// fetchJoke fetches a joke from the JokeAPI and accepts flags
func fetchJoke(keywords string, flags map[string]string) (string, error) {
	// Construct URL
	url := "https://v2.jokeapi.dev/joke/Any"
	params := []string{}

	if keywords != "" {
		params = append(params, "contains="+keywords)
	}

	for key, value := range flags {
		params = append(params, key+"="+value)
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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
			return "", fmt.Errorf("API request timed out after 2 seconds")
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
	var joke JokeResponse
	err = json.Unmarshal(body, &joke)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Format the joke based on its type
	var formattedJoke string
	if joke.Type == "single" {
		formattedJoke = "Joke: " + joke.Joke
	} else if joke.Type == "twopart" {
		formattedJoke = "Setup: " + joke.Setup + "\nDelivery: " + joke.Delivery
	} else {
		return "", fmt.Errorf("unknown joke type: %s", joke.Type)
	}

	return fmt.Sprintf("Status Code: %d\n%s", resp.StatusCode, formattedJoke), nil
}

// simulateAIResponse is a placeholder for the help command
func simulateAIResponse(msg string) tea.Msg {
	if msg == "help" {
		return jokeMsg(helpMessage)
	}
	return jokeMsg("AI: I received your message: \"" + msg + "\"")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.textInput.Value() != "" && !m.processing {
				input := m.textInput.Value()
				m.textInput.Reset()

				// Handle help command
				if input == "help" {
					m.processing = true
					m.textInput.Blur()
					m.messages = append(m.messages, "You: "+input)
					go func() {
						helpText := simulateAIResponse(input).(jokeMsg)
						m.jokeChan <- helpText
					}()
					return m, m.spinner.Tick
				}

				m.processing = true
				m.textInput.Blur()

				// Append the user's message to the messages
				m.messages = append(m.messages, "You: "+input)

				// Simulate AI response
				m.messages = append(m.messages, "AI: Processing your request...")

				// Extract keywords and flags
				keywords := ""
				flags := make(map[string]string)
				parts := strings.Split(input, " ")

				for _, part := range parts {
					if strings.Contains(part, "=") {
						// Parse flag
						flagParts := strings.SplitN(part, "=", 2)
						if len(flagParts) == 2 {
							flags[flagParts[0]] = flagParts[1]
						}
					} else {
						// Treat as keyword
						if keywords == "" {
							keywords = part
						} else {
							keywords += " " + part
						}
					}
				}

				// Update the viewport content
				m.viewport.SetContent(strings.Join(m.messages, "\n"))

				// Fetch joke in a goroutine
				go func() {
					// Construct URL and fetch joke
					joke, err := fetchJoke(keywords, flags)
					if err != nil {
						m.jokeChan <- errMsg{err}
						return
					}
					m.jokeChan <- jokeMsg(joke)
				}()

				return m, m.spinner.Tick
			}
		}

	case spinner.TickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case jokeMsg:
		return m.handleJokeResponse(msg)

	case errMsg:
		return m.handleJokeResponse(msg)
	}

	m.textInput, cmd = m.textInput.Update(msg)

	select {
	case res := <-m.jokeChan:
		return m.handleJokeResponse(res)
	default:
		return m, cmd
	}
}

func (m model) handleJokeResponse(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case jokeMsg:
		m.processing = false
		m.textInput.Focus()
		m.messages = append(m.messages, string(msg))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		return m, m.spinner.Tick
	case errMsg:
		m.processing = false
		m.textInput.Focus()
		m.messages = append(m.messages, "Error: "+msg.Error())
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		return m, m.spinner.Tick
	}
	return m, nil
}

func wordWrap(text string, lineWidth int) string {
	var result string
	words := strings.Fields(text)
	line := ""

	for _, word := range words {
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

func (m model) displayJoke(joke JokeResponse) {
	var jokeText string

	if joke.Type == "single" {
		jokeText = joke.Joke
	} else if joke.Type == "twopart" {
		jokeText = fmt.Sprintf("%s\n\n%s", joke.Setup, joke.Delivery)
	}

	// Properly format the joke with word wrapping
	wrappedJoke := wordWrap(jokeText, 70)

	m.messages = append(m.messages, "AI: "+wrappedJoke)

	// Update viewport content with proper formatting
	content := strings.Join(m.messages, "\n\n") // Add extra newline between messages
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
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
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
