package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/kionsoftware/kion-cli/lib/kion"
)

func TestPrintSTAK(t *testing.T) {
	tests := []struct {
		description string
		stak        kion.STAK
		region      string
		want        string
	}{
		{
			"Empty",
			kion.STAK{},
			"",
			"export AWS_ACCESS_KEY_ID=\nexport AWS_SECRET_ACCESS_KEY=\nexport AWS_SESSION_TOKEN=\n",
		},
		// {
		// 	"Panic Condition",
		// 	kion.STAK{},
		// 	"panic",
		// },
		{
			"Partial STAK",
			kion.STAK{
				AccessKey:       "",
				SecretAccessKey: "aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf",
				SessionToken:    "",
			},
			"",
			"export AWS_ACCESS_KEY_ID=\nexport AWS_SECRET_ACCESS_KEY=aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf\nexport AWS_SESSION_TOKEN=\n",
		},
		{
			"Full STAK",
			kion.STAK{
				AccessKey:       "ASIAABCDEFGHIJ1K23LM",
				SecretAccessKey: "aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf",
				SessionToken:    "AbcDEFghIJKlMNoPQrStuVwXYZabcDEfGhI1JklmNoPQRStu2VWXYZaBcd34ef+GH+IJKLmNOPQRSTU5VwxyzABcdeFGHIj6KlMNoPQ7rSTUvW8X9yZAbCD0ef+gHIJkLMnoPqrstUVwxyzAb1CD2e34fgHiJKlMnOPqr56STuvwXyzABcdEfgh7IJK+8LM91No2pqrSTuvWxyz3ABCdEFGH4ijklMNOP5qrs6TUvWxyz789abcDefgH12iJKlM3no4pQRs+5t6UVw7/xy+ZaBcdE+FGhIj8kLmnOpqrstuvw9xyzab1cD/ef23GhIjkLMNoPQrstuv=",
			},
			"",
			"export AWS_ACCESS_KEY_ID=ASIAABCDEFGHIJ1K23LM\nexport AWS_SECRET_ACCESS_KEY=aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf\nexport AWS_SESSION_TOKEN=AbcDEFghIJKlMNoPQrStuVwXYZabcDEfGhI1JklmNoPQRStu2VWXYZaBcd34ef+GH+IJKLmNOPQRSTU5VwxyzABcdeFGHIj6KlMNoPQ7rSTUvW8X9yZAbCD0ef+gHIJkLMnoPqrstUVwxyzAb1CD2e34fgHiJKlMnOPqr56STuvwXyzABcdEfgh7IJK+8LM91No2pqrSTuvWxyz3ABCdEFGH4ijklMNOP5qrs6TUvWxyz789abcDefgH12iJKlM3no4pQRs+5t6UVw7/xy+ZaBcdE+FGhIj8kLmnOpqrstuvw9xyzab1cD/ef23GhIjkLMNoPQrstuv=\n",
		},
		{
			"With Region",
			kion.STAK{
				AccessKey:       "ASIAABCDEFGHIJ1K23LM",
				SecretAccessKey: "aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf",
				SessionToken:    "AbcDEFghIJKlMNoPQrStuVwXYZabcDEfGhI1JklmNoPQRStu2VWXYZaBcd34ef+GH+IJKLmNOPQRSTU5VwxyzABcdeFGHIj6KlMNoPQ7rSTUvW8X9yZAbCD0ef+gHIJkLMnoPqrstUVwxyzAb1CD2e34fgHiJKlMnOPqr56STuvwXyzABcdEfgh7IJK+8LM91No2pqrSTuvWxyz3ABCdEFGH4ijklMNOP5qrs6TUvWxyz789abcDefgH12iJKlM3no4pQRs+5t6UVw7/xy+ZaBcdE+FGhIj8kLmnOpqrstuvw9xyzab1cD/ef23GhIjkLMNoPQrstuv=",
			},
			"us-gov-west-1",
			"export AWS_REGION=us-gov-west-1\nexport AWS_ACCESS_KEY_ID=ASIAABCDEFGHIJ1K23LM\nexport AWS_SECRET_ACCESS_KEY=aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf\nexport AWS_SESSION_TOKEN=AbcDEFghIJKlMNoPQrStuVwXYZabcDEfGhI1JklmNoPQRStu2VWXYZaBcd34ef+GH+IJKLmNOPQRSTU5VwxyzABcdeFGHIj6KlMNoPQ7rSTUvW8X9yZAbCD0ef+gHIJkLMnoPqrstUVwxyzAb1CD2e34fgHiJKlMnOPqr56STuvwXyzABcdEfgh7IJK+8LM91No2pqrSTuvWxyz3ABCdEFGH4ijklMNOP5qrs6TUvWxyz789abcDefgH12iJKlM3no4pQRs+5t6UVw7/xy+ZaBcdE+FGhIj8kLmnOpqrstuvw9xyzab1cD/ef23GhIjkLMNoPQrstuv=\n",
		},
		// TODO: add test that would print SETs for windows
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// defer func to handle panic in test
			defer func() {
				if test.want == "panic" {
					if r := recover(); r == nil {
						t.Errorf("function should panic")
					}
				}
			}()

			var output bytes.Buffer
			err := PrintSTAK(&output, test.stak, test.region)
			if err != nil {
				t.Error(err)
			}
			if test.want != "panic" && test.want != output.String() {
				t.Errorf("\ngot:\n  %v\nwanted:\n  %v", output.String(), test.want)
			}
		})
	}
}

func TestPrintFavoriteConfig(t *testing.T) {
	tests := []struct {
		name       string
		car        kion.CAR
		region     string
		accessType string
		wantParts  []string // Parts that must appear in output
	}{
		{
			name: "with region",
			car: kion.CAR{
				AccountNumber: "123456789012",
				Name:          "AdminRole",
			},
			region:     "us-east-1",
			accessType: "cli",
			wantParts: []string{
				"account: 123456789012",
				"cloud_access_role: AdminRole",
				"region: us-east-1",
				"access_type: cli",
				"[your favorite alias]",
			},
		},
		{
			name: "without region",
			car: kion.CAR{
				AccountNumber: "987654321098",
				Name:          "ReadOnlyRole",
			},
			region:     "",
			accessType: "web",
			wantParts: []string{
				"account: 987654321098",
				"cloud_access_role: ReadOnlyRole",
				"access_type: web",
			},
		},
		{
			name: "empty car values",
			car: kion.CAR{
				AccountNumber: "",
				Name:          "",
			},
			region:     "",
			accessType: "",
			wantParts: []string{
				"account:",
				"cloud_access_role:",
				"access_type:",
			},
		},
		{
			name: "gov cloud region",
			car: kion.CAR{
				AccountNumber: "111122223333",
				Name:          "GovRole",
			},
			region:     "us-gov-west-1",
			accessType: "cli",
			wantParts: []string{
				"account: 111122223333",
				"cloud_access_role: GovRole",
				"region: us-gov-west-1",
				"access_type: cli",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintFavoriteConfig(&buf, test.car, test.region, test.accessType)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, part := range test.wantParts {
				if !containsString(output, part) {
					t.Errorf("output missing expected part %q\noutput: %s", part, output)
				}
			}

			// If region is empty, verify "region:" line is not present
			if test.region == "" && containsString(output, "region:") {
				t.Errorf("output should not contain 'region:' when region is empty\noutput: %s", output)
			}
		})
	}
}

func TestPrintFavoriteConfig_AlwaysReturnsNil(t *testing.T) {
	var buf bytes.Buffer
	err := PrintFavoriteConfig(&buf, kion.CAR{}, "", "")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// containsString checks if s contains substr, stripping ANSI codes first
func containsString(s, substr string) bool {
	// Strip ANSI escape codes for comparison
	stripped := stripANSI(s)
	return contains(stripped, substr)
}

func stripANSI(s string) string {
	var result []byte
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result = append(result, s[i])
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPrintCredentialProcess(t *testing.T) {
	tests := []struct {
		description string
		stak        kion.STAK
	}{
		{
			"Partial STAK",
			kion.STAK{
				AccessKey:       "",
				SecretAccessKey: "aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf",
				SessionToken:    "",
				Duration:        43200,
				Expiration:      time.Now().Add(43200 * time.Second),
			},
		},
		{
			"Full STAK",
			kion.STAK{
				AccessKey:       "ASIAABCDEFGHIJ1K23LM",
				SecretAccessKey: "aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf",
				SessionToken:    "AbcDEFghIJKlMNoPQrStuVwXYZabcDEfGhI1JklmNoPQRStu2VWXYZaBcd34ef+GH+IJKLmNOPQRSTU5VwxyzABcdeFGHIj6KlMNoPQ7rSTUvW8X9yZAbCD0ef+gHIJkLMnoPqrstUVwxyzAb1CD2e34fgHiJKlMnOPqr56STuvwXyzABcdEfgh7IJK+8LM91No2pqrSTuvWxyz3ABCdEFGH4ijklMNOP5qrs6TUvWxyz789abcDefgH12iJKlM3no4pQRs+5t6UVw7/xy+ZaBcdE+FGhIj8kLmnOpqrstuvw9xyzab1cD/ef23GhIjkLMNoPQrstuv=",
				Duration:        3600,
				Expiration:      time.Now().Add(3600 * time.Second),
			},
		},
		// TODO: add a test that would cause the json marshaling to fail
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintCredentialProcess(&buf, test.stak)
			if err != nil {
				t.Fatalf("PrintCredentialProcess returned an error: %v", err)
			}

			// unmarshal the output into a map
			var output map[string]any
			err = json.Unmarshal(buf.Bytes(), &output)
			if err != nil {
				t.Fatalf("Failed to unmarshal the output: %v", err)
			}

			// parse the expiration field
			expiration, err := time.Parse(time.RFC3339, output["Expiration"].(string))
			if err != nil {
				t.Fatalf("Failed to parse the Expiration field: %v", err)
			}

			// check if the expiration field is within a 1-second tolerance
			now := time.Now()
			duration := test.stak.Duration
			if duration == 0 {
				duration = 900
			}
			hardExpiration := now.Add(time.Duration(duration) * time.Second)
			if expiration.Before(hardExpiration.Add(-1*time.Second)) || expiration.After(hardExpiration.Add(1*time.Second)) {
				t.Fatalf("The 'Expiration' field is not within the expected range")
			}

			expected := fmt.Sprintf("{\n  \"Version\": 1,\n  \"AccessKeyId\": \"%v\",\n  \"SecretAccessKey\": \"%v\",\n  \"SessionToken\": \"%v\",\n  \"Expiration\": \"%v\"\n}\n", test.stak.AccessKey, test.stak.SecretAccessKey, test.stak.SessionToken, output["Expiration"].(string))

			if buf.String() != expected {
				t.Fatalf("Expected %s, but got %s", expected, buf.String())
			}
		})
	}
}
