package helper

import (
	"fmt"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/urfave/cli/v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Wizards                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// CARSelector is a wizard that walks a user through the selection of a
// Project, then associated Accounts, then available Cloud Access Roles, to set
// the user selected Cloud Access Role. Optional account number and or car name
// can be passed via an existing car struct, the flow will dynamically ask what
// is needed to be able to find the full car.
func CARSelector(cCtx *cli.Context, car *kion.CAR) error {
	// get list of projects, then build list of names and lookup map
	projects, err := kion.GetProjects(cCtx.String("endpoint"), cCtx.String("token"))
	if err != nil {
		return err
	}
	pNames, pMap := MapProjects(projects)
	if len(pNames) == 0 {
		return fmt.Errorf("no projects found")
	}

	// prompt user to select a project
	project, err := PromptSelect("Choose a project:", pNames)
	if err != nil {
		return err
	}

	if cCtx.App.Metadata["useUpdatedCloudAccessRoleAPI"] == true {
		// TODO: consolidate on this logic when support for 3.9 drops, that will
		// give us one full support line of buffer

		// get all cars for authed user, works with min permission set
		cars, err := kion.GetCARS(cCtx.String("endpoint"), cCtx.String("token"))
		if err != nil {
			return err
		}
		aNames, aMap := MapAccountsFromCARS(cars, pMap[project].ID)
		if len(aNames) == 0 {
			return fmt.Errorf("no accounts found")
		}

		// prompt user to select an account
		account, err := PromptSelect("Choose an Account:", aNames)
		if err != nil {
			return err
		}

		// narrow it down to just cars associated with the account
		var carsFiltered []kion.CAR
		for _, carObj := range cars {
			if carObj.AccountNumber == aMap[account] {
				carsFiltered = append(carsFiltered, carObj)
			}
		}
		cNames, cMap := MapCAR(carsFiltered)
		if len(cNames) == 0 {
			return fmt.Errorf("you have no cloud access roles assigned")
		}

		// prompt user to select a car
		carname, err := PromptSelect("Choose a Cloud Access Role:", cNames)
		if err != nil {
			return err
		}

		// inject the metadata into the car
		car.Name = cMap[carname].Name
		car.AccountName = cMap[carname].AccountName
		car.AccountNumber = aMap[account]
		car.AccountTypeID = cMap[carname].AccountTypeID
		car.AccountID = cMap[carname].AccountID
		car.AwsIamRoleName = cMap[carname].AwsIamRoleName
		car.ID = cMap[carname].ID
		car.CloudAccessRoleType = cMap[carname].CloudAccessRoleType

		// return nil
		return nil
	} else {
		// get list of accounts on project, then build a list of names and lookup map
		accounts, statusCode, err := kion.GetAccountsOnProject(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID)
		if err != nil {
			if statusCode == 403 {
				// if we're getting a 403 work around permissions bug by temp using private api
				return carSelectorPrivateAPI(cCtx, pMap, project, car)
			} else {
				return err
			}
		}
		aNames, aMap := MapAccounts(accounts)
		if len(aNames) == 0 {
			return fmt.Errorf("no accounts found")
		}

		// prompt user to select an account
		account, err := PromptSelect("Choose an Account:", aNames)
		if err != nil {
			return err
		}

		// get a list of cloud access roles, then build a list of names and lookup map
		cars, err := kion.GetCARSOnProject(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID, aMap[account].ID)
		if err != nil {
			return err
		}
		cNames, cMap := MapCAR(cars)
		if len(cNames) == 0 {
			return fmt.Errorf("no cloud access roles found")
		}

		// prompt user to select a car
		carname, err := PromptSelect("Choose a Cloud Access Role:", cNames)
		if err != nil {
			return err
		}

		// inject the metadata into the car
		car.Name = cMap[carname].Name
		car.AccountName = cMap[carname].AccountName
		car.AccountNumber = aMap[account].Number
		car.AccountTypeID = aMap[account].TypeID
		car.AccountID = aMap[account].ID
		car.AwsIamRoleName = cMap[carname].AwsIamRoleName
		car.ID = cMap[carname].ID
		car.CloudAccessRoleType = cMap[carname].CloudAccessRoleType

		// return nil
		return nil
	}
}

// carSelectorPrivateAPI is a temp shim workaround to address a public API
// permissions issue. CARSelector should be called directly which will the
// forward to this function if needed.
func carSelectorPrivateAPI(cCtx *cli.Context, pMap map[string]kion.Project, project string, car *kion.CAR) error {
	// hit private api endpoint to gather all users cars and their associated accounts
	caCARs, err := kion.GetConsoleAccessCARS(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID)
	if err != nil {
		return err
	}

	// build a consolidated list of accounts from all available CARS and slice of cars per account
	var accounts []kion.Account
	cMap := make(map[string]kion.ConsoleAccessCAR)
	aToCMap := make(map[string][]string)
	for _, car := range caCARs {
		cname := fmt.Sprintf("%v (%v)", car.CARName, car.CARID)
		cMap[cname] = car
		for _, account := range car.Accounts {
			name := fmt.Sprintf("%v (%v)", account.Name, account.Number)
			aToCMap[name] = append(aToCMap[account.Name], cname)
			found := false
			for _, a := range accounts {
				if a.ID == account.ID {
					found = true
				}
			}
			if !found {
				accounts = append(accounts, account)
			}
		}
	}

	// build a list of names and lookup map
	aNames, aMap := MapAccounts(accounts)
	if len(aNames) == 0 {
		return fmt.Errorf("no accounts found")
	}

	// prompt user to select an account
	account, err := PromptSelect("Choose an Account:", aNames)
	if err != nil {
		return err
	}

	// prompt user to select car
	carname, err := PromptSelect("Choose a Cloud Access Role:", aToCMap[account])
	if err != nil {
		return err
	}

	// build enough of a car and return it
	car.Name = cMap[carname].CARName
	car.AccountName = aMap[account].Name
	car.AccountNumber = aMap[account].Number
	car.AccountID = aMap[account].ID
	car.AwsIamRoleName = cMap[carname].AwsIamRoleName
	car.AccountTypeID = aMap[account].TypeID
	car.ID = cMap[carname].CARID
	car.CloudAccessRoleType = cMap[carname].CARRoleType

	return nil
}
