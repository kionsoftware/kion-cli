package helper

import (
	"crypto/md5"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// DynamicStep represents a single step in a dynamic multi-step selection
type DynamicStep struct {
	Title       string
	Description string
	// For static options (first step typically)
	StaticOptions []string
	// For dynamic options that depend on previous selections
	DynamicOptionsFunc func(selections map[string]string) ([]string, error)
	// Dependencies - which previous step names this step depends on
	Dependencies []string
	// The key to store this selection under
	Key string
	// CacheKey generates a cache key from selections for expensive operations
	CacheKey func(selections map[string]string) string
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Caching                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// optionsCache is a simple thread-safe in-memory cache for dynamic options.
type optionsCache struct {
	mu    sync.RWMutex
	items map[string][]string
}

// globalCache is a package-level cache instance.
var globalCache = &optionsCache{
	items: make(map[string][]string),
}

// get retrieves a cached value if it exists.
func (c *optionsCache) get(key string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists := c.items[key]
	return val, exists
}

// set stores a value in the cache.
func (c *optionsCache) set(key string, value []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

// generateDefaultCacheKey creates a simple cache key from selections
func generateDefaultCacheKey(stepKey string, selections map[string]string,
	dependencies []string) string {

	// Sort dependencies for consistent cache keys
	sortedDeps := make([]string, len(dependencies))
	copy(sortedDeps, dependencies)
	sort.Strings(sortedDeps)

	var keyParts []string
	keyParts = append(keyParts, stepKey)

	for _, dep := range sortedDeps {
		if val, exists := selections[dep]; exists && val != "" {
			keyParts = append(keyParts, fmt.Sprintf("%s:%s", dep, val))
		}
	}

	combined := strings.Join(keyParts, "|")
	return fmt.Sprintf("%x", md5.Sum([]byte(combined)))
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// buildCurrentSelections creates a map from current step values
func buildCurrentSelections(stepValues []*string, steps []DynamicStep) map[string]string {
	selections := make(map[string]string)
	for i, stepValue := range stepValues {
		if i < len(steps) && stepValue != nil && *stepValue != "" {
			selections[steps[i].Key] = *stepValue
		}
	}
	return selections
}

// createDependencyBindings creates binding pointers for step dependencies
func createDependencyBindings(step DynamicStep, stepValues []*string,
	steps []DynamicStep) []*string {

	var bindings []*string
	for _, dep := range step.Dependencies {
		for i, s := range steps {
			if s.Key == dep && i < len(stepValues) {
				bindings = append(bindings, stepValues[i])
				break
			}
		}
	}
	return bindings
}

// executeWithCache handles caching logic for dynamic options
func executeWithCache(step DynamicStep, selections map[string]string) ([]string, error) {
	// Generate cache key
	var cacheKey string
	if step.CacheKey != nil {
		cacheKey = step.CacheKey(selections)
	} else {
		cacheKey = generateDefaultCacheKey(step.Key, selections, step.Dependencies)
	}

	// Check cache first
	if cached, exists := globalCache.get(cacheKey); exists {
		return cached, nil
	}

	// Execute function and cache result
	result, err := step.DynamicOptionsFunc(selections)
	if err != nil {
		return result, err
	}

	// Cache successful results
	if len(result) > 0 {
		globalCache.set(cacheKey, result)
	}

	return result, nil
}

// calculateOptimalHeight determines an optimal height for selection prompts.
func calculateOptimalHeight() int {
	// Get terminal size
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Fallback to conservative height if we can't detect terminal size
		return 5
	}

	// Reserve space for title, description, and padding
	availableHeight := height - 5

	// Use a reasonable range: minimum 3, maximum 15
	if availableHeight < 3 {
		return 3
	}
	if availableHeight > 30 {
		return 30
	}

	return availableHeight
}

// kionBrandTheme creates a custom theme using Kion's brand colors
func kionBrandTheme() *huh.Theme {
	t := huh.ThemeBase16()

	// Kion brand colors with automatic fallback support
	kionBlack := lipgloss.AdaptiveColor{Light: "#101C21", Dark: "#101C21"}
	kionGreen := lipgloss.AdaptiveColor{Light: "#61D7AC", Dark: "#61D7AC"}
	kionMint := lipgloss.AdaptiveColor{Light: "#F3F7F4", Dark: "#F3F7F4"}

	// Title and headers in brand green
	t.Focused.Title = lipgloss.NewStyle().
		Foreground(kionGreen).
		Bold(true)

	t.Focused.NoteTitle = lipgloss.NewStyle().
		Foreground(kionGreen).
		Bold(true)

	// Selector arrow in green for brand consistency
	t.Focused.SelectSelector = lipgloss.NewStyle().
		Foreground(kionGreen).
		Bold(true)

	// Selected option highlighted in mint with dark text for contrast
	t.Focused.SelectedOption = lipgloss.NewStyle().
		Foreground(kionBlack).
		Background(kionMint).
		Bold(true)

	// Regular options in mint for readability on dark backgrounds
	t.Focused.Option = lipgloss.NewStyle().
		Foreground(kionMint)

	// Description text in muted mint
	t.Focused.Description = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A8B2A5")) // Muted mint tone

	// Error messages in standard red for visibility
	t.Focused.ErrorMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	t.Focused.ErrorIndicator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	// Input field styling
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(kionGreen)

	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7B70")) // Muted between black and mint

	t.Focused.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(kionGreen).
		Bold(true)

	// Buttons - mint background with black text for brand consistency
	t.Focused.FocusedButton = lipgloss.NewStyle().
		Foreground(kionBlack).
		Background(kionMint).
		Bold(true).
		Padding(0, 2)

	t.Focused.BlurredButton = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7B70")).
		Bold(true).
		Padding(0, 2)

	// Multi-select styling
	t.Focused.MultiSelectSelector = lipgloss.NewStyle().
		Foreground(kionGreen)

	t.Focused.SelectedPrefix = lipgloss.NewStyle().
		Foreground(kionGreen).
		Bold(true)

	t.Focused.UnselectedPrefix = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7B70"))

	t.Focused.UnselectedOption = lipgloss.NewStyle().
		Foreground(kionMint)

	return t
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Prompts                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// PromptSelectDynamic presents a series of dynamic selection prompts to the
// user based on the provided steps. Each step can have static options or
// dynamic options that depend on previous selections. The function returns a
// map of step keys to the selected values.
func PromptSelectDynamic(steps []DynamicStep) (map[string]string, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("no selection steps provided")
	}

	// Create result storage - map of step key to selected value
	results := make(map[string]string)

	// Create individual string variables for each step to work around map
	// addressability
	stepValues := make([]*string, len(steps))
	for i := range steps {
		stepValues[i] = new(string)
	}

	// Create groups for each step
	var groups []*huh.Group

	for i, step := range steps {
		var selectField *huh.Select[string]

		if len(step.StaticOptions) > 0 {
			// Static options (typically first step)
			huhOptions := make([]huh.Option[string], len(step.StaticOptions))
			for j, option := range step.StaticOptions {
				huhOptions[j] = huh.NewOption(option, option)
			}

			selectField = huh.NewSelect[string]().
				Title(step.Title).
				Description(step.Description).
				Options(huhOptions...).
				Value(stepValues[i]).
				Height(calculateOptimalHeight())

		} else if step.DynamicOptionsFunc != nil {
			// For dynamic options, create bindings for dependencies
			bindings := createDependencyBindings(step, stepValues, steps)

			selectField = huh.NewSelect[string]().
				Title(step.Title).
				Description(step.Description).
				OptionsFunc(func() []huh.Option[string] {
					// Build current selections from step values
					currentSelections := buildCurrentSelections(stepValues, steps)

					// Check if all dependencies are satisfied
					for _, dep := range step.Dependencies {
						if currentSelections[dep] == "" {
							return []huh.Option[string]{}
						}
					}

					// Execute with caching
					options, err := executeWithCache(step, currentSelections)
					if err != nil {
						return []huh.Option[string]{
							huh.NewOption("Error loading options: "+err.Error(), ""),
						}
					}

					if len(options) == 0 {
						return []huh.Option[string]{
							huh.NewOption("No options available", ""),
						}
					}

					huhOptions := make([]huh.Option[string], len(options))
					for j, option := range options {
						huhOptions[j] = huh.NewOption(option, option)
					}
					return huhOptions
				}, bindings).
				Value(stepValues[i]).
				Height(calculateOptimalHeight())
		} else {
			return nil, fmt.Errorf("step %d must have either StaticOptions or DynamicOptionsFunc", i)
		}

		group := huh.NewGroup(selectField)
		groups = append(groups, group)
	}

	form := huh.NewForm(groups...).WithTheme(kionBrandTheme())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Copy final values from step variables to results map
	for i, stepValue := range stepValues {
		if stepValue != nil && *stepValue != "" {
			results[steps[i].Key] = *stepValue
		}
	}

	return results, nil
}

// PromptSelect prompts the user to select from a slice of options. It
// requires that the selection made be one of the options provided.
func PromptSelect(message string, description string, options []string) (string, error) {
	var selection string

	// Convert options to huh options
	huhOptions := make([]huh.Option[string], len(options))
	for i, option := range options {
		huhOptions[i] = huh.NewOption(option, option)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(message).
				Description(description).
				Options(huhOptions...).
				Value(&selection).
				Filtering(false).
				Height(calculateOptimalHeight()),
		),
	).WithTheme(kionBrandTheme())

	err := form.Run()
	if err != nil {
		return "", err
	}

	return selection, nil
}

// PromptInput prompts the user to provide dynamic input.
func PromptInput(message string) (string, error) {
	var input string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				Value(&input).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("input is required")
					}
					return nil
				}),
		),
	).WithTheme(kionBrandTheme())

	err := form.Run()
	if err != nil {
		return "", err
	}

	return input, nil
}

// PromptPassword prompts the user to provide sensitive dynamic input.
func PromptPassword(message string) (string, error) {
	var input string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				EchoMode(huh.EchoModePassword).
				Value(&input).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("password is required")
					}
					return nil
				}),
		),
	).WithTheme(kionBrandTheme())

	err := form.Run()
	if err != nil {
		return "", err
	}

	return input, nil
}
