package styles

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Output Styles                                                             //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// OutputStyles holds all lipgloss styles for CLI output formatting.
// Use NewOutputStyles() to create an instance with terminal-aware dimensions.
type OutputStyles struct {
	// Status indicators
	CheckMark lipgloss.Style
	XMark     lipgloss.Style

	// Text styles
	CheckLabel  lipgloss.Style
	ErrorText   lipgloss.Style
	WarningText lipgloss.Style
	InfoText    lipgloss.Style
	DetailText  lipgloss.Style
	SuccessText lipgloss.Style

	// Headers and sections
	MainHeader    lipgloss.Style
	SectionHeader lipgloss.Style
	Separator     lipgloss.Style

	// Boxes and containers
	DetailsBox lipgloss.Style
	SummaryBox lipgloss.Style

	// Layout dimensions
	TerminalWidth   int
	CheckLabelWidth int
}

// NewOutputStyles creates a new set of output styles with terminal-aware dimensions.
func NewOutputStyles() *OutputStyles {
	// Detect terminal width
	termWidth := 80 // Default fallback
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		termWidth = width
	}

	// Use terminal width - 2 for margins, but cap at 80 and minimum 50
	effectiveWidth := max(min(termWidth-2, 80), 50)

	// Calculate check label width: total width - space (1) - checkmark (1)
	checkLabelWidth := effectiveWidth - 2

	// Box width for content (accounting for borders and padding)
	boxWidth := checkLabelWidth + 4

	return &OutputStyles{
		// Status indicators
		CheckMark: lipgloss.NewStyle().
			Foreground(ANSIGreen).
			Bold(true),

		XMark: lipgloss.NewStyle().
			Foreground(ANSIRed).
			Bold(true),

		// Text styles
		CheckLabel: lipgloss.NewStyle().
			Width(checkLabelWidth).
			Align(lipgloss.Left),

		ErrorText: lipgloss.NewStyle().
			Foreground(ANSIRed).
			PaddingLeft(2),

		WarningText: lipgloss.NewStyle().
			Foreground(ANSIYellow).
			PaddingLeft(2),

		InfoText: lipgloss.NewStyle().
			Foreground(ANSIBlue).
			PaddingLeft(2),

		DetailText: lipgloss.NewStyle().
			Foreground(ANSIGray).
			PaddingLeft(2),

		SuccessText: lipgloss.NewStyle().
			Foreground(ANSIGreen).
			Bold(true),

		// Headers and sections
		MainHeader: lipgloss.NewStyle().
			Foreground(ANSICyan).
			Bold(true).
			Padding(0, 1),

		SectionHeader: lipgloss.NewStyle().
			Foreground(ANSIBlue).
			Bold(true).
			PaddingLeft(0).
			MarginTop(1),

		Separator: lipgloss.NewStyle().
			Foreground(ANSIDarkGray).
			Bold(false),

		// Boxes and containers
		DetailsBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ANSIBlue).
			PaddingLeft(1).
			PaddingRight(1).
			Width(boxWidth - 4).
			MarginTop(1).
			MarginBottom(1),

		SummaryBox: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ANSIGreen).
			PaddingLeft(1).
			PaddingRight(1).
			Width(boxWidth - 4).
			MarginTop(1).
			MarginBottom(1),

		// Dimensions
		TerminalWidth:   termWidth,
		CheckLabelWidth: checkLabelWidth,
	}
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Render Helpers                                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// RenderCheck renders a check result with label and status indicator.
func (s *OutputStyles) RenderCheck(label string, passed bool) string {
	status := s.CheckMark.Render("✓")
	if !passed {
		status = s.XMark.Render("✗")
	}
	return s.CheckLabel.Render(label) + " " + status
}

// RenderDetail renders a detail line (indented gray text).
func (s *OutputStyles) RenderDetail(text string) string {
	return s.DetailText.Render(text)
}

// RenderError renders an error message.
func (s *OutputStyles) RenderError(text string) string {
	return s.ErrorText.Render("Error: " + text)
}

// RenderWarning renders a warning message.
func (s *OutputStyles) RenderWarning(text string) string {
	return s.WarningText.Render("Warning: " + text)
}

// RenderFix renders a fix suggestion.
func (s *OutputStyles) RenderFix(text string) string {
	return s.WarningText.Render("Fix: " + text)
}

// RenderNote renders an informational note.
func (s *OutputStyles) RenderNote(text string) string {
	return s.InfoText.Render("Note: " + text)
}

// RenderSeparator renders a horizontal separator line.
func (s *OutputStyles) RenderSeparator() string {
	width := s.CheckLabelWidth + 2
	line := ""
	for range width {
		line += "─"
	}
	return s.Separator.Render(line)
}

// RenderMainHeader renders the main header text.
func (s *OutputStyles) RenderMainHeader(text string) string {
	return s.MainHeader.Render(text)
}

// RenderSectionHeader renders a section header.
func (s *OutputStyles) RenderSectionHeader(text string) string {
	return s.SectionHeader.Render(text)
}
