package main

import (
	"encoding/json"
	"fmt"
	"context"
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
	jokeChan   chan tea.Msg
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type your message here..."
	ti.Focus()
	ti.Width = 80

	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to AI Agent! Type a message and press Enter to chat.")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		textInput:  ti,
		viewport:   vp,
		messages:   []string{"Welcome to AI Agent! Type a message and press Enter to chat."},
		spinner:    s,
		processing: false,
		jokeChan:   make(chan tea.Msg),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

type jokeMsg string
type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

// helpMessage provides instructions on how to use the Joke API
func helpMessage() string {
	return `
Joke API Usage:
To get a joke, type a message containing keywords related to the joke you want to hear.
For example: "tell me a joke about cats"

Available flags:
- category=[Category]: Specify the category of the joke (e.g., "Programming", "Christmas").
- blacklistFlags=[flag1,flag2,...]: Blacklist certain flags (e.g., "nsfw", "racist").
- type=[single|twopart]: Specify the joke type.

Example: "tell me a programming joke category=Programming blacklistFlags=nsfw,racist type=single"
`
}

// JokeResponse struct to hold the API response
type JokeResponse struct {
	Error     bool   `json:"error"`
	Category  string `json:"category"`
	Type      string `json:"type"`
	Setup     string `json:"setup"`
	Delivery  string `json:"delivery"`
	Joke      string `json:"joke"`
	Flags     struct {
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
func fetchJoke(input string) (string, error) {
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
		return jokeMsg(helpMessage())
	}
	return jokeMsg("AI: I received your message: \"" + msg + "\"")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q", "quit":
			return m, tea.Quit
		case "enter":
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

				// Update the viewport content
				m.viewport.SetContent(strings.Join(m.messages, "\n"))

				cmd = func() tea.Msg {
					go func() {
						joke, err := fetchJoke(input)
						if err != nil {
							m.jokeChan <- errMsg{err}
							return
						}
						m.jokeChan <- jokeMsg(joke)
					}()
					return m.spinner.Tick
				}
				return m, cmd
			}
		}
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
	case spinner.TickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.Cmd:
		// Ignore other commands
		return m, nil
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
