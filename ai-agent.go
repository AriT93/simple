package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"github.com/yourusername/ai-agent/jokeclient"
	"github.com/yourusername/ai-agent/utils"
)

type model struct {
	textInput  textinput.Model
	viewport   viewport.Model
	messages   []string
	err        error
	spinner    spinner.Model
	processing bool
	jokeClient *jokeclient.Client
}

func initialModel() model {
	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to AI Assistant!\nType 'help' for instructions or start typing your request.\n")

	ti := textinput.New()
	ti.Placeholder = "Ask for a joke..."
	ti.Focus()
	ti.Width = 80

	// Create a colorful spinner
	s := spinner.New()
	s.Spinner = spinner.Monkey
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")) // Red

	return model{
		messages:   []string{"Welcome to AI Assistant!", "Type 'help' for instructions or start typing your request."},
		viewport:   vp,
		textInput:  ti,
		spinner:    s,
		processing: false,
		jokeClient: jokeclient.NewClient(),
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

// Command to fetch a joke asynchronously
func fetchJokeCmd(client *jokeclient.Client, input string) tea.Cmd {
	return func() tea.Msg {
		joke, err := client.FetchJoke(input)
		if err != nil {
			return errorResponseMsg{err: err}
		}
		return jokeResponseMsg{joke: joke}
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
			return m, tea.Batch(fetchJokeCmd(m.jokeClient, input), m.spinner.Tick)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case jokeResponseMsg:
		m.processing = false
		m.messages = append(m.messages, "AI: "+utils.WordWrap(msg.joke, 70))
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
