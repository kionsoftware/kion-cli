package styles

import (
	"os"
	"strings"

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

	// Accessibility
	ScreenReader bool
}

// NewOutputStyles creates a new set of output styles with terminal-aware dimensions.
// When screenReader is true, all styling is replaced with plain ASCII text so that
// screen readers receive clean, unambiguous output.
func NewOutputStyles(screenReader bool) *OutputStyles {
	if screenReader {
		return &OutputStyles{
			CheckMark:       lipgloss.NewStyle(),
			XMark:           lipgloss.NewStyle(),
			CheckLabel:      lipgloss.NewStyle(),
			ErrorText:       lipgloss.NewStyle().PaddingLeft(2),
			WarningText:     lipgloss.NewStyle().PaddingLeft(2),
			InfoText:        lipgloss.NewStyle().PaddingLeft(2),
			DetailText:      lipgloss.NewStyle().PaddingLeft(2),
			SuccessText:     lipgloss.NewStyle(),
			MainHeader:      lipgloss.NewStyle(),
			SectionHeader:   lipgloss.NewStyle(),
			Separator:       lipgloss.NewStyle(),
			DetailsBox:      lipgloss.NewStyle(),
			SummaryBox:      lipgloss.NewStyle(),
			TerminalWidth:   80,
			CheckLabelWidth: 78,
			ScreenReader:    true,
		}
	}

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
// In screen reader mode it uses "[PASS]" / "[FAIL]" instead of Unicode symbols.
func (s *OutputStyles) RenderCheck(label string, passed bool) string {
	if s.ScreenReader {
		if passed {
			return "[PASS] " + label
		}
		return "[FAIL] " + label
	}
	status := s.CheckMark.Render("✓")
	if !passed {
		status = s.XMark.Render("✗")
	}
	return s.CheckLabel.Render(label) + " " + status
}

// RenderDetail renders a detail line (indented text).
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
// In screen reader mode it uses plain hyphens instead of box-drawing characters.
func (s *OutputStyles) RenderSeparator() string {
	if s.ScreenReader {
		return "---"
	}
	width := s.CheckLabelWidth + 2
	line := strings.Repeat("─", width)
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

// PassStr returns the appropriate pass symbol for embedding in strings.
// In screen reader mode returns "[PASS]"; otherwise returns "✓".
func (s *OutputStyles) PassStr() string {
	if s.ScreenReader {
		return "[PASS]"
	}
	return "✓"
}

// FailStr returns the appropriate fail symbol for embedding in strings.
// In screen reader mode returns "[FAIL]"; otherwise returns "✗".
func (s *OutputStyles) FailStr() string {
	if s.ScreenReader {
		return "[FAIL]"
	}
	return "✗"
}
