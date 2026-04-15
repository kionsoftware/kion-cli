package helper

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/kionsoftware/kion-cli/lib/styles"
	"golang.org/x/term"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Screen Reader Mode                                                        //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// ScreenReaderMode switches all huh forms to WithAccessible(true), which
// replaces TUI widgets (pipe bar, arrow-key navigation, screen redraw) with
// sequential plain-text prompts that screen readers can follow linearly.
// Set this before any prompt is called; typically done in BeforeCommands after
// profile resolution so --screen-reader and the config-file value are both
// honoured.
var ScreenReaderMode bool

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

// newForm wraps huh.NewForm and applies WithAccessible when ScreenReaderMode
// is on, so callers don't need to repeat the check.
func newForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).
		WithTheme(styles.FormTheme).
		WithAccessible(ScreenReaderMode)
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Prompts                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// PromptSelect prompts the user to select from a slice of options. It
// requires that the selection made be one of the options provided.
//
// In screen reader mode huh switches to a numbered list printed sequentially
// to stdout; the user types the number and presses Enter.
func PromptSelect(message string, description string, options []string) (string, error) {
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

	// Apply height limiting only if needed (no-op in accessible mode)
	if shouldLimit, height := shouldLimitHeight(len(options)); shouldLimit {
		selectField = selectField.Height(height)
	}

	if err := newForm(huh.NewGroup(selectField)).Run(); err != nil {
		return "", err
	}

	return selection, nil
}

// PromptInput prompts the user to provide dynamic input.
//
// In screen reader mode huh prints the title and reads a plain line from
// stdin with no TUI decoration.
func PromptInput(message string) (string, error) {
	var input string

	if err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				Value(&input).
				Validate(styles.RequiredValidator),
		),
	).Run(); err != nil {
		return "", err
	}

	return input, nil
}

// PromptPassword prompts the user to provide sensitive dynamic input.
//
// In screen reader mode huh prints the title and reads from stdin without
// echoing characters.
func PromptPassword(message string) (string, error) {
	var input string

	if err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				EchoMode(huh.EchoModePassword).
				Value(&input).
				Validate(styles.RequiredValidator),
		),
	).Run(); err != nil {
		return "", err
	}

	return input, nil
}
