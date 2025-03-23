package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
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
	spinner   spinner.Model
	processing bool
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
		textInput: ti,
		viewport:  vp,
		messages:  []string{"Welcome to AI Agent! Type a message and press Enter to chat."},
		spinner:   s,
		processing: false,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
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
	ID            int    `json:"id"`
	Safe          bool   `json:"safe"`
	Lang          string `json:"lang"`
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
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() != "" && !m.processing {
				m.processing = true
				m.textInput.Blur()
				userMsg := "You: " + m.textInput.Value()
				m.messages = append(m.messages, userMsg)

				input := m.textInput.Value()
				m.textInput.Reset()

				m.messages = append(m.messages, "AI: "+string(simulateAIResponse(input).(jokeMsg)))

				cmd = func() tea.Msg {
					joke, err := fetchJoke(input)
					if err != nil {
						return errMsg(err)
					}
					return jokeMsg(joke)
				}
				return m, cmd
			}
		}
	case jokeMsg:
		m.messages = append(m.messages, "AI: "+string(msg))
		m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
		m.viewport.GotoBottom()
		m.processing = false
		m.textInput.Focus()
		return m, nil

	case errMsg:
		m.messages = append(m.messages, "Error fetching joke: "+string(msg))
		m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
		m.viewport.GotoBottom()
		m.processing = false
		m.textInput.Focus()
		return m, nil

	case tea.Cmd:
		return m, msg

	case error:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Handle text input updates
	m.textInput, cmd = m.textInput.Update(msg)

	// Handle viewport updates (for scrolling)
	m.viewport, _ = m.viewport.Update(msg)

	return m, cmd
}

// Custom message types for handling joke and error responses
type jokeMsg string
type errMsg string

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
		"Press Ctrl+C to quit.",
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
