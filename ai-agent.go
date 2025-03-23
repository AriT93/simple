package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	textInput textinput.Model
	viewport  viewport.Model
	messages  []string
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type your message here..."
	ti.Focus()
	ti.Width = 80

	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to AI Agent! Type a message and press Enter to chat.")

	return model{
		textInput: ti,
		viewport:  vp,
		messages:  []string{"Welcome to AI Agent! Type a message and press Enter to chat."},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// simulateAIResponse is a placeholder for actual AI integration
func simulateAIResponse(msg string) string {
	return "AI: I received your message: \"" + msg + "\""
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
	ID            int    `json:"id"`
	Safe          bool   `json:"safe"`
	Lang          string `json:"lang"`
}

// fetchJoke fetches a joke from the JokeAPI
func fetchJoke(query string) (string, error) {
	// Construct the API URL with the query
	url := fmt.Sprintf("https://v2.jokeapi.dev/joke/Any?contains=%s", query)

	// Make the HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal the JSON response
	var joke JokeResponse
	err = json.Unmarshal(body, &joke)
	if err != nil {
		return "", err
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

	return formattedJoke, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() != "" {
				userMsg := "You: " + m.textInput.Value()
				m.messages = append(m.messages, userMsg)

				// Fetch joke from the API
				joke, err := fetchJoke(m.textInput.Value())
				if err != nil {
					m.messages = append(m.messages, "Error fetching joke: "+err.Error())
				} else {
					m.messages = append(m.messages, "AI: "+joke)
				}

				// Update viewport content
				m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
				m.viewport.GotoBottom()

				// Clear input
				m.textInput.Reset()

				return m, nil
			}
		}
	case error:
		m.err = msg
		return m, nil
	}

	// Handle text input updates
	m.textInput, cmd = m.textInput.Update(msg)

	// Handle viewport updates (for scrolling)
	m.viewport, _ = m.viewport.Update(msg)

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"AI Agent Chat",
		"------------",
		m.viewport.View(),
		"",
		m.textInput.View(),
		"Press Ctrl+C to quit.",
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
