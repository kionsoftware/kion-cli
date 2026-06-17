package styles

import (
	"strings"
	"testing"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  OutputStyles Tests                                                        //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// TestNewOutputStylesScreenReader verifies that screen reader mode produces
// plain ASCII output free of Unicode symbols and ANSI colour codes.
func TestNewOutputStylesScreenReader(t *testing.T) {
	s := NewOutputStyles(true)

	if !s.ScreenReader {
		t.Error("ScreenReader field should be true")
	}
}

func TestRenderCheckScreenReader(t *testing.T) {
	s := NewOutputStyles(true)

	pass := s.RenderCheck("Configuration valid", true)
	if !strings.HasPrefix(pass, "[PASS]") {
		t.Errorf("expected pass result to start with [PASS], got: %q", pass)
	}
	if strings.Contains(pass, "✓") {
		t.Errorf("screen reader mode must not emit ✓, got: %q", pass)
	}

	fail := s.RenderCheck("Configuration valid", false)
	if !strings.HasPrefix(fail, "[FAIL]") {
		t.Errorf("expected fail result to start with [FAIL], got: %q", fail)
	}
	if strings.Contains(fail, "✗") {
		t.Errorf("screen reader mode must not emit ✗, got: %q", fail)
	}
}

func TestRenderCheckNormal(t *testing.T) {
	s := NewOutputStyles(false)

	pass := s.RenderCheck("Configuration valid", true)
	if !strings.Contains(pass, "✓") {
		t.Errorf("normal mode should contain ✓, got: %q", pass)
	}

	fail := s.RenderCheck("Configuration valid", false)
	if !strings.Contains(fail, "✗") {
		t.Errorf("normal mode should contain ✗, got: %q", fail)
	}
}

func TestRenderSeparatorScreenReader(t *testing.T) {
	s := NewOutputStyles(true)

	sep := s.RenderSeparator()
	if strings.Contains(sep, "─") {
		t.Errorf("screen reader mode must not emit box-drawing characters, got: %q", sep)
	}
	if sep != "---" {
		t.Errorf("expected separator to be \"---\", got: %q", sep)
	}
}

func TestPassStrFailStr(t *testing.T) {
	sr := NewOutputStyles(true)
	if sr.PassStr() != "[PASS]" {
		t.Errorf("screen reader PassStr should be [PASS], got: %q", sr.PassStr())
	}
	if sr.FailStr() != "[FAIL]" {
		t.Errorf("screen reader FailStr should be [FAIL], got: %q", sr.FailStr())
	}

	normal := NewOutputStyles(false)
	if normal.PassStr() != "✓" {
		t.Errorf("normal PassStr should be ✓, got: %q", normal.PassStr())
	}
	if normal.FailStr() != "✗" {
		t.Errorf("normal FailStr should be ✗, got: %q", normal.FailStr())
	}
}

func TestRenderMessagesScreenReader(t *testing.T) {
	s := NewOutputStyles(true)

	tests := []struct {
		name   string
		got    string
		prefix string
	}{
		{"RenderError", s.RenderError("something broke"), "Error:"},
		{"RenderWarning", s.RenderWarning("watch out"), "Warning:"},
		{"RenderFix", s.RenderFix("do this instead"), "Fix:"},
		{"RenderNote", s.RenderNote("for your info"), "Note:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trimmed := strings.TrimSpace(tt.got)
			if !strings.HasPrefix(trimmed, tt.prefix) {
				t.Errorf("expected output to start with %q, got: %q", tt.prefix, trimmed)
			}
		})
	}
}
