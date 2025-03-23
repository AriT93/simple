package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	textInput textinput.Model
}

func initialModel() model {
	return model{
		textInput: textinput.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Hey there! 👋\n\n%s", m.textInput.View())
}

func main() {
	p := tea.NewProgram(initialModel())
	if err, _ := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
