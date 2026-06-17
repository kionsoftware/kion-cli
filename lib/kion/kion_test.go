package kion

import "testing"

func TestConvertAccessType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// API to CLI format
		{"console_access to web", "console_access", "web"},
		{"short_term_key_access to cli", "short_term_key_access", "cli"},

		// CLI to API format
		{"web to console_access", "web", "console_access"},
		{"cli to short_term_key_access", "cli", "short_term_key_access"},

		// Passthrough for unknown values
		{"empty string passthrough", "", ""},
		{"unknown value passthrough", "unknown", "unknown"},
		{"random string passthrough", "something_else", "something_else"},

		// Case sensitivity - function is case-sensitive, these should passthrough
		{"uppercase WEB passthrough", "WEB", "WEB"},
		{"uppercase CLI passthrough", "CLI", "CLI"},
		{"mixed case Web passthrough", "Web", "Web"},
		{"uppercase CONSOLE_ACCESS passthrough", "CONSOLE_ACCESS", "CONSOLE_ACCESS"},

		// Whitespace - function does not trim, these should passthrough
		{"leading space passthrough", " web", " web"},
		{"trailing space passthrough", "web ", "web "},
		{"space padded passthrough", " cli ", " cli "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ConvertAccessType(test.input)
			if got != test.want {
				t.Errorf("ConvertAccessType(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestConvertAccessType_Bidirectional(t *testing.T) {
	// Test that conversions are properly bidirectional
	pairs := []struct {
		apiFormat string
		cliFormat string
	}{
		{"console_access", "web"},
		{"short_term_key_access", "cli"},
	}

	for _, pair := range pairs {
		t.Run("api_to_cli_"+pair.apiFormat, func(t *testing.T) {
			got := ConvertAccessType(pair.apiFormat)
			if got != pair.cliFormat {
				t.Errorf("ConvertAccessType(%q) = %q, want %q", pair.apiFormat, got, pair.cliFormat)
			}
		})

		t.Run("cli_to_api_"+pair.cliFormat, func(t *testing.T) {
			got := ConvertAccessType(pair.cliFormat)
			if got != pair.apiFormat {
				t.Errorf("ConvertAccessType(%q) = %q, want %q", pair.cliFormat, got, pair.apiFormat)
			}
		})
	}
}
