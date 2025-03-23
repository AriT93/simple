package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AriT93/ai-agent/jokeclient"
	"github.com/AriT93/ai-agent/utils"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

// Add LangChain components to the model
type model struct {
	textInput   textinput.Model
	viewport    viewport.Model
	messages    []string
	err         error
	spinner     spinner.Model
	processing  bool
	jokeClient  *jokeclient.Client
	llm         llms.Model       // LangChain
	parser      *chains.LLMChain // Chain for parsing input
	enhancer    *chains.LLMChain // Chain for enhancing output
	showingHelp bool
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
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

	// Enable debug mode based on environment variable or flag
	debug := os.Getenv("DEBUG") == "true"
	jokeClient := jokeclient.NewClient(debug)

	// Initialize LangChain components (with error handling)
	var llm llms.Model
	var parser *chains.LLMChain
	var enhancer *chains.LLMChain
	var initError error

	// Check if OPENAI_API_KEY is set
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		// Initialize OpenAI LLM using the correct function name
		llm, initError = openai.New(
			openai.WithToken(apiKey),
			openai.WithModel("gpt-3.5-turbo"),
		)

		if initError == nil {
			// Initialize parser and enhancer chains
			parserPrompt := prompts.NewPromptTemplate(
				`Parse this user request for a joke API: {{.input}}
    
    Extract these parameters:
    - category: [programming, misc, dark, pun, spooky, christmas]
    - type: [single, twopart]
    - blacklist flags: [nsfw, religious, political, racist, sexist, explicit]
    
    For NSFW content, if user specifically requests NSFW jokes, do NOT include nsfw in blacklist.
    
    Format your response EXACTLY like this example (one line, no spaces around =):
    category=programming&type=single&blacklist=religious,political
    
    Only include parameters that are specified or implied in the request.`,
				[]string{"input"},
			)
			parser = chains.NewLLMChain(llm, parserPrompt)

			enhancerPrompt := prompts.NewPromptTemplate(
				"Make this joke more entertaining: {{.output}}",
				[]string{"output"},
			)
			enhancer = chains.NewLLMChain(llm, enhancerPrompt)
		}
	}

	return model{
		messages:   []string{"Welcome to AI Assistant!", "Type 'help' for instructions or start typing your request."},
		viewport:   vp,
		textInput:  ti,
		spinner:    s,
		processing: false,
		jokeClient: jokeClient,
		llm:        llm,
		parser:     parser,
		enhancer:   enhancer,
		err:        initError,
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
This is a natural language interface for fetching jokes, enhanced with LangChain.

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
"I'd like something funny about computers, but keep it clean"

LangChain features:
- Improved natural language understanding
- Enhanced joke presentation
- Better parameter extraction from complex requests
`

// Command to fetch a joke asynchronously
func fetchJokeCmd(client *jokeclient.Client, input string, model model) tea.Cmd {
	return func() tea.Msg {
		var parsedInput string
		var err error

		// Use LangChain to parse input if available
		if model.parser != nil && model.llm != nil {
			// Log original input to debug
			if client.Debug {
				client.WriteDebug("========== LangChain Processing ==========\n")
				client.WriteDebug("ORIGINAL INPUT: %s\n", input)
			}

			// Call LangChain parser
			ctx := context.Background()
			result, parseErr := chains.Call(ctx, model.parser, map[string]any{
				"input": input,
			})

			if parseErr != nil {
				// Log parsing error but continue with original input
				if client.Debug {
					client.WriteDebug("LANGCHAIN ERROR: %v\n", parseErr)
				}
				parsedInput = input
			} else {
				// Extract parsed text from LangChain result
				if text, ok := result["text"].(string); ok {
					parsedInput = strings.TrimSpace(text)

					// Log the parsed result
					if client.Debug {
						client.WriteDebug("LANGCHAIN PARSED: %s\n", parsedInput)
						client.WriteDebug("PARAMETERS EXTRACTED:\n")
						params := strings.Split(parsedInput, "&")
						for _, param := range params {
							client.WriteDebug("  %s\n", param)
						}
					}
				} else {
					parsedInput = input
					if client.Debug {
						client.WriteDebug("LANGCHAIN OUTPUT FORMAT ERROR: Unable to extract text\n")
					}
				}
			}
		} else {
			// Use original input if LangChain is not available
			parsedInput = input
		}

		// Fetch joke with parsed input
		joke, err := client.FetchJoke(parsedInput)
		if err != nil {
			return errorResponseMsg{err: err}
		}

		// Use LangChain to enhance joke if available
		if model.enhancer != nil && model.llm != nil {
			if client.Debug {
				client.WriteDebug("ORIGINAL JOKE: %s\n", joke)
			}

			// Call LangChain enhancer
			ctx := context.Background()
			result, enhanceErr := chains.Call(ctx, model.enhancer, map[string]any{
				"output": joke,
			})

			if enhanceErr == nil {
				// Extract enhanced joke
				if text, ok := result["text"].(string); ok {
					enhancedJoke := strings.TrimSpace(text)

					// Log enhanced joke
					if client.Debug {
						client.WriteDebug("ENHANCED JOKE: %s\n", enhancedJoke)
					}

					// Use enhanced joke
					joke = enhancedJoke
				}
			}
		}

		return jokeResponseMsg{joke: joke}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If showing help, any key returns to normal view
		if m.showingHelp {
			if msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc {
				m.showingHelp = false
				return m, nil
			}
			return m, nil
		}

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

			// Handle help command
			if input == "help" {
				m.textInput.Reset()
				m.showingHelp = true
				return m, nil
			}

			// Reset input
			m.textInput.Reset()

			// Add user message to history
			m.messages = append(m.messages, "You: "+input)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			// Initiate joke fetching
			m.processing = true
			return m, tea.Batch(fetchJokeCmd(m.jokeClient, input, m), m.spinner.Tick)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case jokeResponseMsg:
		m.processing = false

		// Wrap the joke at 72 characters for better display
		wrappedJoke := utils.WordWrap(msg.joke, 72)
		m.messages = append(m.messages, "AI: "+wrappedJoke)

		// Add extra newline for better separation between messages
		m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
		m.viewport.GotoBottom()
		return m, nil

	case errorResponseMsg:
		m.processing = false

		// Wrap error message
		wrappedError := utils.WordWrap(msg.err.Error(), 72)
		m.messages = append(m.messages, "Error: "+wrappedError)

		// Add extra newline for better separation
		m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
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

	if m.showingHelp {
		// Format the help message with proper spacing and structure
		helpStyle := lipgloss.NewStyle().Width(72)

		// Split the help message into sections and format each separately
		sections := strings.Split(helpMessage, "\n\n")
		formattedSections := make([]string, len(sections))

		for i, section := range sections {
			formattedSections[i] = helpStyle.Render(section)
		}

		// Join the sections with proper spacing
		formattedHelp := strings.Join(formattedSections, "\n\n")

		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("AI Assistant Help"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("---------------"),
			formattedHelp,
			"",
			lipgloss.NewStyle().Italic(true).Render("Press ESC or Enter to return to chat"),
		)
	}

	var status string
	if m.processing {
		status = m.spinner.View() + " Getting response..."
	} else {
		status = ""
	}

	// Add LangChain status indicator
	var langchainStatus string
	if m.llm != nil {
		langchainStatus = "LangChain: Active ✓"
	} else {
		langchainStatus = "LangChain: Inactive ✗ (Set OPENAI_API_KEY to enable)"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("AI Agent Chat"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("------------"),
		m.viewport.View(),
		status,
		m.textInput.View(),
		"Type 'help' for instructions or 'quit' to exit.",
		langchainStatus,
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
