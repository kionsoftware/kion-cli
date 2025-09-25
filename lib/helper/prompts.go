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

	// Kion brand colors
	kionBlack := lipgloss.AdaptiveColor{Light: "#101C21", Dark: "#101C21"}
	kionGreen := lipgloss.AdaptiveColor{Light: "#61D7AC", Dark: "#61D7AC"}
	kionMint := lipgloss.AdaptiveColor{Light: "#F3F7F4", Dark: "#F3F7F4"}
	mutedMint := lipgloss.Color("#A8B2A5")
	mutedGray := lipgloss.Color("#6B7B70")
	errorRed := lipgloss.Color("#FF6B6B")

	// Helper functions for common style patterns
	greenBold := func() lipgloss.Style {
		return lipgloss.NewStyle().Foreground(kionGreen).Bold(true)
	}

	mintOnBlack := func() lipgloss.Style {
		return lipgloss.NewStyle().Foreground(kionBlack).Background(kionMint).Bold(true)
	}

	muted := func(color lipgloss.TerminalColor) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(color)
	}

	// Apply styles using helpers
	t.Focused.Title = greenBold()
	t.Focused.NoteTitle = greenBold()
	t.Focused.SelectSelector = greenBold()
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(kionGreen)
	t.Focused.TextInput.Prompt = greenBold()
	t.Focused.MultiSelectSelector = muted(kionGreen)
	t.Focused.SelectedPrefix = greenBold()

	t.Focused.SelectedOption = mintOnBlack()
	t.Focused.FocusedButton = mintOnBlack().Padding(0, 2)

	t.Focused.Option = muted(kionMint)
	t.Focused.UnselectedOption = muted(kionMint)
	t.Focused.Description = muted(mutedMint)
	t.Focused.TextInput.Placeholder = muted(mutedGray)
	t.Focused.BlurredButton = muted(mutedGray).Bold(true).Padding(0, 2)
	t.Focused.UnselectedPrefix = muted(mutedGray)

	t.Focused.ErrorMessage = muted(errorRed).Bold(true)
	t.Focused.ErrorIndicator = muted(errorRed)

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
