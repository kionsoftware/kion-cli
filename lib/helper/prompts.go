package helper

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/kionsoftware/kion-cli/lib/styles"
	"golang.org/x/term"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// shouldLimitHeight determines if the selection prompt height should be
// limited based on the height of the terminal.
func shouldLimitHeight(optionCount int) (bool, int) {
	_, termHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Conservative fallback - limit if more than 10 options
		return optionCount > 10, 10
	}

	// Reserve space for title, description, padding, and some buffer
	availableLines := termHeight - 8

	// Ensure at least 3 lines are available for options
	availableLines = max(availableLines, 3)

	// Only limit height if options exceed available terminal space
	if optionCount > availableLines {
		return true, availableLines
	}

	return false, 0
}

// toHuhOptions converts a string slice to huh options.
func toHuhOptions(options []string) []huh.Option[string] {
	huhOptions := make([]huh.Option[string], len(options))
	for i, option := range options {
		huhOptions[i] = huh.NewOption(option, option)
	}
	return huhOptions
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Prompts                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// PromptSelect prompts the user to select from a slice of options. It
// requires that the selection made be one of the options provided.
func PromptSelect(message string, description string, options []string) (string, error) {
	var selection string

	selectField := huh.NewSelect[string]().
		Title(message).
		Description(description).
		Options(toHuhOptions(options)...).
		Value(&selection).
		Filtering(false)

	// Apply height limiting only if needed
	if shouldLimit, height := shouldLimitHeight(len(options)); shouldLimit {
		selectField = selectField.Height(height)
	}

	form := huh.NewForm(
		huh.NewGroup(selectField),
	).WithTheme(styles.FormTheme)

	if err := form.Run(); err != nil {
		return "", err
	}

	return selection, nil
}

// PromptInput prompts the user to provide dynamic input.
func PromptInput(message string) (string, error) {
	return promptInput(message, huh.EchoModeNormal)
}

// PromptPassword prompts the user to provide sensitive dynamic input.
func PromptPassword(message string) (string, error) {
	return promptInput(message, huh.EchoModePassword)
}

// promptInput is the shared implementation for text input prompts.
func promptInput(message string, echoMode huh.EchoMode) (string, error) {
	var input string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				EchoMode(echoMode).
				Value(&input).
				Validate(styles.RequiredValidator),
		),
	).WithTheme(styles.FormTheme)

	if err := form.Run(); err != nil {
		return "", err
	}

	return input, nil
}
