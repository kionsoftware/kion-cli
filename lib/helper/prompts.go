package helper

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/kionsoftware/kion-cli/lib/styles"
	"golang.org/x/term"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Screen Reader Mode                                                        //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// ScreenReaderMode disables all TUI widgets (huh forms, lipgloss colours,
// box-drawing characters) and falls back to plain text I/O.  Set this before
// any prompt is called; typically done in BeforeCommands after profile
// resolution so that --screen-reader and the config-file value are both
// honoured.
var ScreenReaderMode bool

// stdinReader is a shared reader for plain-text prompt fallbacks.
// A single instance is required so that multiple sequential reads
// do not each buffer ahead and silently discard each other's input.
var stdinReader = bufio.NewReader(os.Stdin)

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

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Prompts                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// PromptSelect prompts the user to select from a slice of options.  It
// requires that the selection made be one of the options provided.
//
// In screen reader mode the TUI is replaced with a numbered list printed to
// stdout; the user types the corresponding number and presses Enter.
func PromptSelect(message string, description string, options []string) (string, error) {
	if ScreenReaderMode {
		fmt.Println(message)
		if description != "" {
			fmt.Println(description)
		}
		for i, opt := range options {
			fmt.Printf("  %d. %s\n", i+1, opt)
		}
		for {
			fmt.Printf("Enter number (1-%d): ", len(options))
			line, err := stdinReader.ReadString('\n')
			if err != nil {
				return "", err
			}
			num, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil || num < 1 || num > len(options) {
				fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(options))
				continue
			}
			return options[num-1], nil
		}
	}

	var selection string

	// Convert to huh options
	huhOptions := make([]huh.Option[string], len(options))
	for i, option := range options {
		huhOptions[i] = huh.NewOption(option, option)
	}

	selectField := huh.NewSelect[string]().
		Title(message).
		Description(description).
		Options(huhOptions...).
		Value(&selection)

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
//
// In screen reader mode the TUI is replaced with a plain text prompt written
// to stdout; the user types their answer and presses Enter.
func PromptInput(message string) (string, error) {
	if ScreenReaderMode {
		fmt.Print(message + " ")
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input := strings.TrimSpace(line)
		if input == "" {
			return "", fmt.Errorf("input is required")
		}
		return input, nil
	}

	var input string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				Value(&input).
				Validate(styles.RequiredValidator),
		),
	).WithTheme(styles.FormTheme)

	if err := form.Run(); err != nil {
		return "", err
	}

	return input, nil
}

// PromptPassword prompts the user to provide sensitive dynamic input.
//
// In screen reader mode the TUI is replaced with a plain prompt written to
// stdout; input is read without echo using term.ReadPassword so the password
// is never displayed on screen.
func PromptPassword(message string) (string, error) {
	if ScreenReaderMode {
		fmt.Print(message + " ")
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // move to next line after hidden input
		if err != nil {
			return "", err
		}
		if len(password) == 0 {
			return "", fmt.Errorf("password is required")
		}
		return string(password), nil
	}

	var input string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				EchoMode(huh.EchoModePassword).
				Value(&input).
				Validate(styles.RequiredValidator),
		),
	).WithTheme(styles.FormTheme)

	if err := form.Run(); err != nil {
		return "", err
	}

	return input, nil
}
