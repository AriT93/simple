package main

import (
	"fmt"
	"strings"

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

				// This is where you would integrate with an actual AI service
				aiResponse := simulateAIResponse(m.textInput.Value())
				m.messages = append(m.messages, aiResponse)

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
