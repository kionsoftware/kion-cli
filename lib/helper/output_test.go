package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
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

const (
	testAccessTypeWeb = AccessTypeWeb
	testAccessTypeCLI = AccessTypeCLI
	testCARName       = "some-car"
	testAccountNum    = "111111111111"
)

func TestMatchExistingFavorite(t *testing.T) {
	car := kion.CAR{AccountNumber: testAccountNum, Name: testCARName}

	tests := []struct {
		description string
		favorites   []structs.Favorite
		accessType  string
		want        string
	}{
		{
			"Nil Favorites",
			nil,
			testAccessTypeCLI,
			"",
		},
		{
			"Empty Favorites",
			[]structs.Favorite{},
			testAccessTypeCLI,
			"",
		},
		{
			"Exact Access Type Match",
			[]structs.Favorite{
				{Name: "web-fav", Account: testAccountNum, CAR: testCARName, AccessType: testAccessTypeWeb},
			},
			testAccessTypeWeb,
			"web-fav",
		},
		{
			"Generic Favorite Matches Any Access Type",
			[]structs.Favorite{
				{Name: "generic-fav", Account: testAccountNum, CAR: testCARName, AccessType: ""},
			},
			testAccessTypeWeb,
			"generic-fav",
		},
		{
			"Exact Match Preferred Over Generic",
			[]structs.Favorite{
				{Name: "generic-fav", Account: testAccountNum, CAR: testCARName, AccessType: ""},
				{Name: "web-fav", Account: testAccountNum, CAR: testCARName, AccessType: testAccessTypeWeb},
			},
			testAccessTypeWeb,
			"web-fav",
		},
		{
			"Exact Match Preferred Regardless Of Order",
			[]structs.Favorite{
				{Name: "web-fav", Account: testAccountNum, CAR: testCARName, AccessType: testAccessTypeWeb},
				{Name: "generic-fav", Account: testAccountNum, CAR: testCARName, AccessType: ""},
			},
			testAccessTypeWeb,
			"web-fav",
		},
		{
			"First Generic Wins",
			[]structs.Favorite{
				{Name: "generic-one", Account: testAccountNum, CAR: testCARName, AccessType: ""},
				{Name: "generic-two", Account: testAccountNum, CAR: testCARName, AccessType: ""},
			},
			testAccessTypeCLI,
			"generic-one",
		},
		{
			"Mismatched Access Type Only",
			[]structs.Favorite{
				{Name: "web-fav", Account: testAccountNum, CAR: testCARName, AccessType: testAccessTypeWeb},
			},
			testAccessTypeCLI,
			"",
		},
		{
			"Mismatched Account",
			[]structs.Favorite{
				{Name: "other-acct", Account: "222222222222", CAR: testCARName, AccessType: testAccessTypeCLI},
			},
			testAccessTypeCLI,
			"",
		},
		{
			"Mismatched CAR",
			[]structs.Favorite{
				{Name: "other-car", Account: testAccountNum, CAR: "other-car", AccessType: testAccessTypeCLI},
			},
			testAccessTypeCLI,
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := MatchExistingFavorite(test.favorites, car, test.accessType)
			if test.want == "" {
				if got != nil {
					t.Errorf("\ngot:\n  %v\nwanted:\n  no match", got.Name)
				}
				return
			}
			if got == nil {
				t.Fatalf("\ngot:\n  no match\nwanted:\n  %v", test.want)
			}
			if got.Name != test.want {
				t.Errorf("\ngot:\n  %v\nwanted:\n  %v", got.Name, test.want)
			}
		})
	}
}

func TestPrintFavoriteConfig(t *testing.T) {
	// pin color output off so assertions do not depend on tty detection
	origNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = origNoColor }()

	car := kion.CAR{AccountNumber: testAccountNum, Name: testCARName}

	tests := []struct {
		description string
		region      string
		accessType  string
		existing    *structs.Favorite
		want        string
	}{
		{
			"No Match Prints Snippet Without Region",
			"",
			testAccessTypeWeb,
			nil,
			"\nTo save your selection as a favorite add the following to\nyour configuration file under the 'favorites:' section:\n  - name: [your favorite alias]\n    account: 111111111111\n    cloud_access_role: some-car\n    access_type: web\n\n",
		},
		{
			"No Match Prints Snippet With Region",
			"us-gov-west-1",
			testAccessTypeCLI,
			nil,
			"\nTo save your selection as a favorite add the following to\nyour configuration file under the 'favorites:' section:\n  - name: [your favorite alias]\n    account: 111111111111\n    cloud_access_role: some-car\n    region: us-gov-west-1\n    access_type: cli\n\n",
		},
		{
			"Existing Exact Web Favorite Omits Flag",
			"",
			testAccessTypeWeb,
			&structs.Favorite{Name: "web-fav", AccessType: testAccessTypeWeb},
			"\nYou already have a favorite for this selection. Next time you can run:\n  kion favorite web-fav\n\n",
		},
		{
			"Existing Generic Favorite Adds Web Flag",
			"",
			testAccessTypeWeb,
			&structs.Favorite{Name: "generic-fav", AccessType: ""},
			"\nYou already have a favorite for this selection. Next time you can run:\n  kion favorite --web generic-fav\n\n",
		},
		{
			"Existing Generic Favorite For Cli Omits Flag",
			"us-gov-west-1",
			testAccessTypeCLI,
			&structs.Favorite{Name: "generic-fav", AccessType: ""},
			"\nYou already have a favorite for this selection. Next time you can run:\n  kion favorite generic-fav\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var output bytes.Buffer
			err := PrintFavoriteConfig(&output, car, test.region, test.accessType, test.existing)
			if err != nil {
				t.Error(err)
			}
			if output.String() != test.want {
				t.Errorf("\ngot:\n  %q\nwanted:\n  %q", output.String(), test.want)
			}
		})
	}
}
