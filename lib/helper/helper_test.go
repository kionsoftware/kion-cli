package helper

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/kionsoftware/kion-cli/lib/kion"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Resources                                                                 //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

var kionTestProjects = []kion.Project{
	{Archived: false, AutoPay: true, DefaultAwsRegion: "us-east-1", Description: "test description one", ID: 101, Name: "project one", OuID: 201},
	{Archived: false, AutoPay: false, DefaultAwsRegion: "us-west-1", Description: "test description two", ID: 102, Name: "project two", OuID: 202},
	{Archived: true, AutoPay: false, DefaultAwsRegion: "us-east-1", Description: "test description three", ID: 103, Name: "project three", OuID: 203},
	{Archived: false, AutoPay: true, DefaultAwsRegion: "us-west-1", Description: "test description four", ID: 104, Name: "project four", OuID: 204},
	{Archived: true, AutoPay: false, DefaultAwsRegion: "us-east-1", Description: "test description five", ID: 105, Name: "project five", OuID: 205},
	{Archived: false, AutoPay: true, DefaultAwsRegion: "us-east-1", Description: "test description six", ID: 106, Name: "project six", OuID: 206},
}

var kionTestProjectsNames = []string{
	"project one",
	"project two",
	"project three",
	"project four",
	"project five",
	"project six",
}

var kionTestAccounts = []kion.Account{
	{Email: "test1@kion.io", Name: "account one", Number: "111111111111", TypeID: 1, ID: 101, IncludeLinkedAccountSpend: true, LinkedAccountNumber: "", LinkedRole: "", PayerID: 101, ProjectID: 101, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test2@kion.io", Name: "account two", Number: "121212121212", TypeID: 2, ID: 102, IncludeLinkedAccountSpend: false, LinkedAccountNumber: "", LinkedRole: "", PayerID: 102, ProjectID: 102, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test3@kion.io", Name: "account three", Number: "131313131313", TypeID: 3, ID: 103, IncludeLinkedAccountSpend: true, LinkedAccountNumber: "000000000000", LinkedRole: "", PayerID: 103, ProjectID: 103, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test4@kion.io", Name: "account four", Number: "141414141414", TypeID: 4, ID: 104, IncludeLinkedAccountSpend: false, LinkedAccountNumber: "", LinkedRole: "", PayerID: 104, ProjectID: 104, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test5@kion.io", Name: "account five", Number: "151515151515", TypeID: 5, ID: 105, IncludeLinkedAccountSpend: false, LinkedAccountNumber: "", LinkedRole: "", PayerID: 105, ProjectID: 105, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test6@kion.io", Name: "account six", Number: "161616161616", TypeID: 6, ID: 106, IncludeLinkedAccountSpend: false, LinkedAccountNumber: "", LinkedRole: "", PayerID: 106, ProjectID: 106, SkipAccessChecking: true, UseOrgAccountInfo: true},
}

var kionTestAccountsNames = []string{
	"account one",
	"account two",
	"account three",
	"account four",
	"account five",
	"account six",
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Tests                                                                     //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

func TestPrintSTAK(t *testing.T) {
	tests := []struct {
		description string
		stak        kion.STAK
		want        string
	}{
		{
			"Empty",
			kion.STAK{},
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
			"export AWS_ACCESS_KEY_ID=\nexport AWS_SECRET_ACCESS_KEY=aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf\nexport AWS_SESSION_TOKEN=\n",
		},
		{
			"Full STAK",
			kion.STAK{
				AccessKey:       "ASIAABCDEFGHIJ1K23LM",
				SecretAccessKey: "aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf",
				SessionToken:    "AbcDEFghIJKlMNoPQrStuVwXYZabcDEfGhI1JklmNoPQRStu2VWXYZaBcd34ef+GH+IJKLmNOPQRSTU5VwxyzABcdeFGHIj6KlMNoPQ7rSTUvW8X9yZAbCD0ef+gHIJkLMnoPqrstUVwxyzAb1CD2e34fgHiJKlMnOPqr56STuvwXyzABcdEfgh7IJK+8LM91No2pqrSTuvWxyz3ABCdEFGH4ijklMNOP5qrs6TUvWxyz789abcDefgH12iJKlM3no4pQRs+5t6UVw7/xy+ZaBcdE+FGhIj8kLmnOpqrstuvw9xyzab1cD/ef23GhIjkLMNoPQrstuv=",
			},
			"export AWS_ACCESS_KEY_ID=ASIAABCDEFGHIJ1K23LM\nexport AWS_SECRET_ACCESS_KEY=aBCDeFg1hijkl2m3NOPqr4StUvWxY56z7abc8DEf\nexport AWS_SESSION_TOKEN=AbcDEFghIJKlMNoPQrStuVwXYZabcDEfGhI1JklmNoPQRStu2VWXYZaBcd34ef+GH+IJKLmNOPQRSTU5VwxyzABcdeFGHIj6KlMNoPQ7rSTUvW8X9yZAbCD0ef+gHIJkLMnoPqrstUVwxyzAb1CD2e34fgHiJKlMnOPqr56STuvwXyzABcdEfgh7IJK+8LM91No2pqrSTuvWxyz3ABCdEFGH4ijklMNOP5qrs6TUvWxyz789abcDefgH12iJKlM3no4pQRs+5t6UVw7/xy+ZaBcdE+FGhIj8kLmnOpqrstuvw9xyzab1cD/ef23GhIjkLMNoPQrstuv=\n",
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
			err := PrintSTAK(&output, test.stak)
			if err != nil {
				t.Error(err)
			}
			if test.want != "panic" && test.want != output.String() {
				t.Errorf("\ngot:\n  %v\nwanted:\n  %v", output.String(), test.want)
			}
		})
	}
}

func TestMapProjects(t *testing.T) {
	tests := []struct {
		name     string
		projects []kion.Project
		wantOne  []string
		wantTwo  map[string]kion.Project
	}{
		{
			"Basic",
			kionTestProjects,
			[]string{
				kionTestProjectsNames[4],
				kionTestProjectsNames[3],
				kionTestProjectsNames[0],
				kionTestProjectsNames[5],
				kionTestProjectsNames[2],
				kionTestProjectsNames[1],
			},
			map[string]kion.Project{
				kionTestProjectsNames[0]: kionTestProjects[0],
				kionTestProjectsNames[1]: kionTestProjects[1],
				kionTestProjectsNames[2]: kionTestProjects[2],
				kionTestProjectsNames[3]: kionTestProjects[3],
				kionTestProjectsNames[4]: kionTestProjects[4],
				kionTestProjectsNames[5]: kionTestProjects[5],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			one, two := MapProjects(test.projects)
			if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", one, two, test.wantOne, test.wantTwo)
			}
		})
	}
}

func TestMapAccounts(t *testing.T) {
	tests := []struct {
		name     string
		accounts []kion.Account
		wantOne  []string
		wantTwo  map[string]kion.Account
	}{
		{
			"Basic",
			kionTestAccounts,
			[]string{
				kionTestAccountsNames[4],
				kionTestAccountsNames[3],
				kionTestAccountsNames[0],
				kionTestAccountsNames[5],
				kionTestAccountsNames[2],
				kionTestAccountsNames[1],
			},
			map[string]kion.Account{
				kionTestAccountsNames[0]: kionTestAccounts[0],
				kionTestAccountsNames[1]: kionTestAccounts[1],
				kionTestAccountsNames[2]: kionTestAccounts[2],
				kionTestAccountsNames[3]: kionTestAccounts[3],
				kionTestAccountsNames[4]: kionTestAccounts[4],
				kionTestAccountsNames[5]: kionTestAccounts[5],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			one, two := MapAccounts(test.accounts)
			if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", one, two, test.wantOne, test.wantTwo)
			}
		})
	}
}
