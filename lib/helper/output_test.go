package helper

import (
	"bytes"
	"testing"

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
