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
	"github.com/pterm/pterm"
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

func runSuggest(cmd *cobra.Command, args []string) {
	description := strings.Join(args, " ")

	suggestion, err := getSuggestion(description)
	if err != nil {
		pterm.Error.Printf("Error getting suggestion: %v\n", err)
		os.Exit(1)
	}

	pterm.Success.Printf("Suggested command: %s\n", pterm.LightCyan(suggestion))
	pterm.Info.Print("Press Enter to run, 'y' to copy to clipboard, or any other key to exit: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "":
		pterm.Info.Printf("Running: %s\n", pterm.LightCyan(suggestion))
		runCommand(suggestion)
	case "y", "Y":
		err := clipboard.WriteAll(suggestion)
		if err != nil {
			pterm.Error.Printf("Error copying to clipboard: %v\n", err)
		} else {
			pterm.Success.Println("Command copied to clipboard!")
		}
	default:
		pterm.Info.Println("Exiting...")
	}
}

func getSuggestion(description string) (string, error) {
	ctx := context.Background()

	spinner, _ := pterm.DefaultSpinner.Start("Phinking...")

	suggestion, err := llm.GenerateCommand(description, ctx)

	spinner.Stop()

	return suggestion, err
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
		pterm.Error.Printf("Error running command: %v\n", err)
	}
}
