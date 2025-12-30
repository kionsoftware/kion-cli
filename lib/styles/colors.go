// Package styles provides consistent styling and theming for the Kion CLI.
// It defines brand colors, interactive prompt themes, and output formatting
// styles used throughout the application.
package styles

import "github.com/charmbracelet/lipgloss"

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Brand Colors                                                              //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// Kion brand colors - primary palette
// These use AdaptiveColor to work on both light and dark terminal backgrounds
var (
	// Black is the primary dark color used for backgrounds and contrast
	Black = lipgloss.AdaptiveColor{Light: "#101C21", Dark: "#101C21"}

	// Green is the primary accent color used for highlights and selections
	Green = lipgloss.AdaptiveColor{Light: "#0D9668", Dark: "#61D7AC"}

	// Mint is used for primary text - light on dark terminals, dark on light terminals
	Mint = lipgloss.AdaptiveColor{Light: "#1A202C", Dark: "#F3F7F4"}
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  UI Colors                                                                 //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// Muted colors - secondary palette for less prominent elements
var (
	// MutedMint is a softer color for descriptions and secondary text
	MutedMint = lipgloss.AdaptiveColor{Light: "#4A5568", Dark: "#A8B2A5"}

	// MutedGray is used for placeholders and disabled elements
	MutedGray = lipgloss.AdaptiveColor{Light: "#718096", Dark: "#6B7B70"}

	// SelectionBg is used for selected item backgrounds
	SelectionBg = lipgloss.AdaptiveColor{Light: "#E2E8F0", Dark: "#4A5568"}
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Status Colors                                                             //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// Status colors - semantic colors for feedback
// Slightly adjusted between light/dark for optimal visibility
var (
	// Success indicates successful operations
	Success = lipgloss.AdaptiveColor{Light: "#0D9668", Dark: "#61D7AC"}

	// Error indicates errors and failures
	Error = lipgloss.AdaptiveColor{Light: "#C53030", Dark: "#FF6B6B"}

	// Warning indicates warnings and cautions
	Warning = lipgloss.AdaptiveColor{Light: "#B7791F", Dark: "#FFCC66"}

	// Info indicates informational messages
	Info = lipgloss.AdaptiveColor{Light: "#2B6CB0", Dark: "#66B2FF"}
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  ANSI Fallback Colors                                                      //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// ANSI fallback colors - for terminals with limited color support
// These are used by OutputStyles for CLI output formatting
var (
	// ANSIGreen is ANSI color 10 (bright green)
	ANSIGreen = lipgloss.Color("10")

	// ANSIRed is ANSI color 9 (bright red)
	ANSIRed = lipgloss.Color("9")

	// ANSIYellow is ANSI color 11 (bright yellow)
	ANSIYellow = lipgloss.Color("11")

	// ANSIBlue is ANSI color 12 (bright blue)
	ANSIBlue = lipgloss.Color("12")

	// ANSICyan is ANSI color 14 (bright cyan)
	ANSICyan = lipgloss.Color("14")

	// ANSIGray is ANSI color 245 (medium gray)
	ANSIGray = lipgloss.Color("245")

	// ANSIDarkGray is ANSI color 240 (dark gray)
	ANSIDarkGray = lipgloss.Color("240")
)
