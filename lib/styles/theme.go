package styles

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Interactive Prompt Theme                                                  //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// FormTheme is the cached huh theme for interactive prompts.
// It is initialized once at package load time for efficiency.
var FormTheme = newFormTheme()

// newFormTheme creates the Kion-branded theme for huh forms.
func newFormTheme() *huh.Theme {
	t := huh.ThemeBase16()

	// Reusable style builders
	greenBold := lipgloss.NewStyle().Foreground(Green).Bold(true)
	selected := lipgloss.NewStyle().Foreground(Mint).Background(SelectionBg).Bold(true)

	// Title and header styles
	t.Focused.Title = greenBold
	t.Focused.NoteTitle = greenBold

	// Selection styles
	t.Focused.SelectSelector = greenBold
	t.Focused.SelectedOption = selected
	t.Focused.SelectedPrefix = greenBold
	t.Focused.MultiSelectSelector = lipgloss.NewStyle().Foreground(Green)

	// Option styles
	t.Focused.Option = lipgloss.NewStyle().Foreground(Mint)
	t.Focused.UnselectedOption = lipgloss.NewStyle().Foreground(Mint)
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(MutedGray)

	// Text input styles
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(Green)
	t.Focused.TextInput.Prompt = greenBold
	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().Foreground(MutedGray)

	// Description and help text
	t.Focused.Description = lipgloss.NewStyle().Foreground(MutedMint)

	// Button styles
	t.Focused.FocusedButton = selected.Padding(0, 2)
	t.Focused.BlurredButton = lipgloss.NewStyle().Foreground(MutedGray).Bold(true).Padding(0, 2)

	// Error styles
	t.Focused.ErrorMessage = lipgloss.NewStyle().Foreground(Error).Bold(true)
	t.Focused.ErrorIndicator = lipgloss.NewStyle().Foreground(Error)

	return t
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Validation Helpers                                                        //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// RequiredValidator is a reusable validator that ensures input is not empty.
var RequiredValidator = huh.ValidateNotEmpty()
