package main

import (
	"bufio"
	"context"
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

func main() {
	rootCmd := &cobra.Command{
		Use:   "suggest [description]",
		Short: "Get CLI command suggestions using AI",
		Args:  cobra.MinimumNArgs(1),
		Run:   runSuggest,
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

	fmt.Printf("%s %s\n\n", successStyle.Render("Suggested command:"), commandStyle.Render(suggestion))
	fmt.Print(infoStyle.Render("Press Enter to run, 'y' to copy to clipboard, or any other key to exit: "))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "":
		fmt.Printf("\n%s %s\n", infoStyle.Render("Running:"), commandStyle.Render(suggestion))
		runCommand(suggestion)
	case "y", "Y":
		err := clipboard.WriteAll(suggestion)
		if err != nil {
			fmt.Printf("\n%s\n", errorStyle.Render(fmt.Sprintf("Error copying to clipboard: %v", err)))
		} else {
			fmt.Printf("\n%s\n", successStyle.Render("Command copied to clipboard!"))
		}
	default:
		fmt.Printf("\n%s\n", infoStyle.Render("Exiting..."))
	}
}

type spinnerModel struct {
	spinner     spinner.Model
	description string
	suggestion  string
	err         error
	done        bool
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.generateCommand,
	)
}

func (m spinnerModel) generateCommand() tea.Msg {
	ctx := context.Background()
	suggestion, err := llm.GenerateCommand(m.description, ctx)
	return suggestionMsg{suggestion: suggestion, err: err}
}

type suggestionMsg struct {
	suggestion string
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

func getSuggestion(description string) (string, error) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))

	m := spinnerModel{
		spinner:     s,
		description: description,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	final := finalModel.(spinnerModel)
	return final.suggestion, final.err
}

func runCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error running command: %v", err)))
	}
}
