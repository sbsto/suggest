package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"suggest/llm"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const version = "1.0.5"

type CommandSuggestion struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "suggest [description]",
		Short:   "Get CLI command suggestions using AI",
		Version: version,
		Args:    cobra.MinimumNArgs(1),
		Run:     runSuggest,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true).Padding(1, 1)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Bold(true).Padding(1, 1)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#45B7D1")).Padding(1, 1)
	commandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#96CEB4")).Bold(true).Padding(1, 1)
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Padding(1, 1)
)

func runSuggest(cmd *cobra.Command, args []string) {
	description := strings.Join(args, " ")

	suggestion, err := getSuggestion(description)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error getting suggestion: %v", err)))
		os.Exit(1)
	}

	menu := menuModel{
		suggestion: suggestion,
		choices:    []string{"Run command", "Copy to clipboard", "Exit"},
		cursor:     0,
		selected:   -1,
	}

	p := tea.NewProgram(menu)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error running menu: %v", err)))
		os.Exit(1)
	}

	final := finalModel.(menuModel)
	switch final.selected {
	case 0:
		fmt.Printf("%s %s\n", infoStyle.Render("Running:"), commandStyle.Render(suggestion.Command))
		if err := runCommand(suggestion.Command); err != nil {
			handleCommandFailure(suggestion.Command, err, description)
		}
	case 1:
		err := clipboard.WriteAll(suggestion.Command)
		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render(fmt.Sprintf("Error copying to clipboard: %v", err)))
		} else {
			fmt.Printf("%s\n", successStyle.Render("Command copied to clipboard!"))
		}
	case 2:
		fmt.Printf("%s\n", infoStyle.Render("Exiting..."))
	}
}

type spinnerModel struct {
	spinner      spinner.Model
	description  string
	errorContext string
	suggestion   CommandSuggestion
	err          error
	done         bool
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.generateCommand,
	)
}

func (m spinnerModel) generateCommand() tea.Msg {
	ctx := context.Background()
	var response string
	var err error
	
	if m.errorContext != "" {
		response, err = llm.GenerateCommandWithContext(m.description, m.errorContext, ctx)
	} else {
		response, err = llm.GenerateCommand(m.description, ctx)
	}
	
	if err != nil {
		return suggestionMsg{suggestion: CommandSuggestion{}, err: err}
	}
	
	// Clean up the response to handle potential markdown or extra text
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		if idx := strings.Index(response, "```"); idx != -1 {
			response = response[:idx]
		}
	}
	response = strings.TrimSpace(response)
	
	var suggestion CommandSuggestion
	if err := json.Unmarshal([]byte(response), &suggestion); err != nil {
		// Fallback to treating response as just a command
		suggestion = CommandSuggestion{Command: response, Description: ""}
	}
	return suggestionMsg{suggestion: suggestion, err: nil}
}

type suggestionMsg struct {
	suggestion CommandSuggestion
	err        error
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case suggestionMsg:
		m.suggestion = msg.suggestion
		m.err = msg.err
		m.done = true
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	thinkingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45B7D1"))
	containerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45B7D1")).Padding(1, 1)
	content := fmt.Sprintf("%s %s", m.spinner.View(), thinkingStyle.Render("Thinking..."))
	return containerStyle.Render(content)
}

func getSuggestion(description string) (CommandSuggestion, error) {
	return getSuggestionWithContext(description, "")
}

func getSuggestionWithContext(description, errorContext string) (CommandSuggestion, error) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))

	m := spinnerModel{
		spinner:      s,
		description:  description,
		errorContext: errorContext,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return CommandSuggestion{}, err
	}

	final := finalModel.(spinnerModel)
	return final.suggestion, final.err
}

type menuModel struct {
	suggestion CommandSuggestion
	choices    []string
	cursor     int
	selected   int
	done       bool
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.selected = 2 // Exit
			m.done = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m menuModel) View() string {
	if m.done {
		return ""
	}

	commandHighlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#2A2A2A")).
		Foreground(lipgloss.Color("#96CEB4")).
		Bold(true).
		Padding(1, 2).
		Margin(0, 1, 0, 1)

	descriptionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Padding(0, 1)

	s := "\n" + commandHighlightStyle.Render(m.suggestion.Command)
	if m.suggestion.Description != "" {
		s += "\n" + descriptionStyle.Render(m.suggestion.Description)
	}
	s += "\n\n"

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = " >"
			choice = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Bold(true).Render(choice)
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 1, 1, 1)
	s += "\n" + helpStyle.Render("Use ↑/↓ arrows or j/k to navigate, Enter to select, q to quit")

	return s
}

type CommandError struct {
	Command string
	Err     error
	Stderr  string
	Stdout  string
}

func (ce CommandError) Error() string {
	result := fmt.Sprintf("Command '%s' failed: %v", ce.Command, ce.Err)
	if ce.Stderr != "" {
		result += fmt.Sprintf("\nStderr: %s", ce.Stderr)
	}
	return result
}

func runCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	
	// Style for command output
	outputStyle := lipgloss.NewStyle().
		Padding(0, 1, 1, 1).
		Margin(0, 1, 0, 1)
	
	// Display stdout if there's any
	if stdout.Len() > 0 {
		output := strings.TrimRight(stdout.String(), "\n")
		if output != "" {
			fmt.Print(outputStyle.Render(output) + "\n")
		}
	}
	
	// Display stderr if there's any
	if stderr.Len() > 0 {
		stderrStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Padding(0, 1, 1, 1).
			Margin(0, 1, 0, 1)
		stderrOutput := strings.TrimRight(stderr.String(), "\n")
		if stderrOutput != "" {
			fmt.Print(stderrStyle.Render(stderrOutput) + "\n")
		}
	}
	
	// Return error with details if command failed
	if err != nil {
		return CommandError{
			Command: command,
			Err:     err,
			Stderr:  strings.TrimSpace(stderr.String()),
			Stdout:  strings.TrimSpace(stdout.String()),
		}
	}
	
	return nil
}

func handleCommandFailure(originalCommand string, cmdErr error, originalDescription string) {
	fmt.Printf("\n%s\n", errorStyle.Render(fmt.Sprintf("Command failed: %v", cmdErr)))

	menu := menuModel{
		suggestion: CommandSuggestion{Command: originalCommand, Description: "Failed command"},
		choices:    []string{"Suggest new command", "Exit"},
		cursor:     0,
		selected:   -1,
	}

	p := tea.NewProgram(menu)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error running menu: %v", err)))
		return
	}

	final := finalModel.(menuModel)
	switch final.selected {
	case 0:
		// Suggest new command with error context
		errorContext := fmt.Sprintf("Previous command '%s' failed with error: %v", originalCommand, cmdErr)
		newSuggestion, err := getSuggestionWithContext(originalDescription, errorContext)
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error getting new suggestion: %v", err)))
			return
		}

		menu := menuModel{
			suggestion: newSuggestion,
			choices:    []string{"Run command", "Copy to clipboard", "Exit"},
			cursor:     0,
			selected:   -1,
		}

		p := tea.NewProgram(menu)
		finalModel, err := p.Run()
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error running menu: %v", err)))
			return
		}

		final := finalModel.(menuModel)
		switch final.selected {
		case 0:
			fmt.Printf("%s %s\n", infoStyle.Render("Running:"), commandStyle.Render(newSuggestion.Command))
			if err := runCommand(newSuggestion.Command); err != nil {
				handleCommandFailure(newSuggestion.Command, err, originalDescription)
			}
		case 1:
			err := clipboard.WriteAll(newSuggestion.Command)
			if err != nil {
				fmt.Printf("%s\n", errorStyle.Render(fmt.Sprintf("Error copying to clipboard: %v", err)))
			} else {
				fmt.Printf("%s\n", successStyle.Render("Command copied to clipboard!"))
			}
		case 2:
			fmt.Printf("%s\n", infoStyle.Render("Exiting..."))
		}
	case 1:
		fmt.Printf("%s\n", infoStyle.Render("Exiting..."))
	}
}
