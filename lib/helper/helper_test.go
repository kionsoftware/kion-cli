package helper

import (
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
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
	{Email: "test1@kion.io", Name: "account one", Alias: "acct-one-alias", Number: "111111111111", TypeID: 1, ID: 101, IncludeLinkedAccountSpend: true, LinkedAccountNumber: "", LinkedRole: "", PayerID: 101, ProjectID: 101, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test2@kion.io", Name: "account two", Alias: "acct-two-alias", Number: "121212121212", TypeID: 2, ID: 102, IncludeLinkedAccountSpend: false, LinkedAccountNumber: "", LinkedRole: "", PayerID: 102, ProjectID: 102, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test3@kion.io", Name: "account three", Alias: "acct-three-alias", Number: "131313131313", TypeID: 3, ID: 103, IncludeLinkedAccountSpend: true, LinkedAccountNumber: "000000000000", LinkedRole: "", PayerID: 103, ProjectID: 103, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test4@kion.io", Name: "account four", Alias: "acct-four-alias", Number: "141414141414", TypeID: 4, ID: 104, IncludeLinkedAccountSpend: false, LinkedAccountNumber: "", LinkedRole: "", PayerID: 104, ProjectID: 104, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test5@kion.io", Name: "account five", Alias: "acct-five-alias", Number: "151515151515", TypeID: 5, ID: 105, IncludeLinkedAccountSpend: false, LinkedAccountNumber: "", LinkedRole: "", PayerID: 105, ProjectID: 105, SkipAccessChecking: true, UseOrgAccountInfo: false},
	{Email: "test6@kion.io", Name: "account six", Alias: "acct-six-alias", Number: "161616161616", TypeID: 6, ID: 106, IncludeLinkedAccountSpend: false, LinkedAccountNumber: "", LinkedRole: "", PayerID: 106, ProjectID: 106, SkipAccessChecking: true, UseOrgAccountInfo: true},
}

var kionTestAccountsNames = []string{
	"account one",
	"account two",
	"account three",
	"account four",
	"account five",
	"account six",
}

var kionTestCARs = []kion.CAR{
	{AccountID: 101, AccountNumber: "111111111111", AccountAlias: "acct-one-alias", AccountType: "aws", AccountTypeID: 1, AccountName: "account one", ApplyToAllAccounts: true, AwsIamPath: "some path", AwsIamRoleName: "role one", CloudAccessRoleType: "type", FutureAccounts: true, ID: 101, LongTermAccessKeys: false, Name: "car one", ProjectID: 101, ShortTermAccessKeys: true, WebAccess: true},
	{AccountID: 102, AccountNumber: "121212121212", AccountAlias: "acct-two-alias", AccountType: "aws", AccountTypeID: 2, AccountName: "account two", ApplyToAllAccounts: true, AwsIamPath: "some path", AwsIamRoleName: "role two", CloudAccessRoleType: "type", FutureAccounts: true, ID: 102, LongTermAccessKeys: false, Name: "car two", ProjectID: 102, ShortTermAccessKeys: true, WebAccess: true},
	{AccountID: 103, AccountNumber: "131313131313", AccountAlias: "acct-three-alias", AccountType: "aws", AccountTypeID: 3, AccountName: "account three", ApplyToAllAccounts: true, AwsIamPath: "some path", AwsIamRoleName: "role three", CloudAccessRoleType: "type", FutureAccounts: true, ID: 103, LongTermAccessKeys: false, Name: "car three", ProjectID: 103, ShortTermAccessKeys: true, WebAccess: true},
	{AccountID: 104, AccountNumber: "141414141414", AccountAlias: "acct-four-alias", AccountType: "aws", AccountTypeID: 4, AccountName: "account four", ApplyToAllAccounts: true, AwsIamPath: "some path", AwsIamRoleName: "role four", CloudAccessRoleType: "type", FutureAccounts: true, ID: 104, LongTermAccessKeys: false, Name: "car four", ProjectID: 104, ShortTermAccessKeys: true, WebAccess: true},
	{AccountID: 105, AccountNumber: "151515151515", AccountAlias: "acct-five-alias", AccountType: "aws", AccountTypeID: 5, AccountName: "account five", ApplyToAllAccounts: true, AwsIamPath: "some path", AwsIamRoleName: "role five", CloudAccessRoleType: "type", FutureAccounts: true, ID: 105, LongTermAccessKeys: false, Name: "car five", ProjectID: 105, ShortTermAccessKeys: true, WebAccess: true},
	{AccountID: 106, AccountNumber: "161616161616", AccountAlias: "acct-six-alias", AccountType: "aws", AccountTypeID: 6, AccountName: "account six", ApplyToAllAccounts: true, AwsIamPath: "some path", AwsIamRoleName: "role six", CloudAccessRoleType: "type", FutureAccounts: true, ID: 106, LongTermAccessKeys: false, Name: "car six", ProjectID: 106, ShortTermAccessKeys: true, WebAccess: true},
}

var kionTestCARsNames = []string{
	"car one",
	"car two",
	"car three",
	"car four",
	"car five",
	"car six",
}

var kionTestIDMSs = []kion.IDMS{
	{ID: 101, IdmsTypeID: 1, Name: "idms one"},
	{ID: 102, IdmsTypeID: 2, Name: "idms two"},
	{ID: 103, IdmsTypeID: 3, Name: "idms three"},
	{ID: 104, IdmsTypeID: 4, Name: "idms four"},
	{ID: 105, IdmsTypeID: 5, Name: "idms five"},
	{ID: 106, IdmsTypeID: 6, Name: "idms six"},
}

var kionTestIDMSsNames = []string{
	"idms one",
	"idms two",
	"idms three",
	"idms four",
	"idms five",
	"idms six",
}

var kionTestFavorites = []structs.Favorite{
	{
		Name:            "fav one",
		Account:         "111111111111",
		CAR:             "car one",
		AccessType:      "web",
		DescriptiveName: "fav one      [local] (111111111111 car one web)",
	},
	{
		Name:            "fav two",
		Account:         "121212121212",
		CAR:             "car two",
		AccessType:      "web",
		DescriptiveName: "fav two      [local] (121212121212 car two web)",
	},
	{
		Name:            "fav three",
		Account:         "131313131313",
		CAR:             "car three",
		AccessType:      "web",
		DescriptiveName: "fav three    [local] (131313131313 car three web)",
	},
	{
		Name:            "fav four",
		Account:         "141414141414",
		CAR:             "car four",
		AccessType:      "web",
		DescriptiveName: "fav four     [local] (141414141414 car four web)",
	},
	{
		Name:            "fav five",
		Account:         "151515151515",
		CAR:             "car five",
		AccessType:      "web",
		DescriptiveName: "fav five     [local] (151515151515 car five web)",
	},
	{
		Name:            "fav six",
		Account:         "161616161616",
		CAR:             "car six",
		AccessType:      "web",
		DescriptiveName: "fav six      [local] (161616161616 car six web)",
	},
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////
