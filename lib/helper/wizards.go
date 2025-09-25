package helper

import (
	"fmt"
	"time"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/urfave/cli/v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// executeWithRetry executes the provided operation function with retry logic.
// It retries the operation up to maxRetries times with a delay added between
// attempts.
func executeWithRetry[T any](operation func() (T, error), maxRetries int,
	delay time.Duration) (T, error) {

	var result T
	var err error

	for attempt := range maxRetries {
		result, err = operation()
		if err == nil {
			return result, nil
		}

		if attempt == maxRetries-1 {
			return result, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
		}

		time.Sleep(delay)
	}

	return result, err
}

// getAccountsUpdatedAPI retrieves accounts using the updated API with retry
// logic.
func getAccountsUpdatedAPI(cCtx *cli.Context, pMap map[string]kion.Project,
	project string) ([]string, error) {

	return executeWithRetry(func() ([]string, error) {
		cars, err := kion.GetCARS(cCtx.String("endpoint"), cCtx.String("token"), "")
		if err != nil {
			return nil, fmt.Errorf("failed to get accounts: %w", err)
		}

		aNames, _ := MapAccountsFromCARS(cars, pMap[project].ID)
		if len(aNames) == 0 {
			return nil, fmt.Errorf("no accounts found")
		}

		return aNames, nil
	}, 3, 500*time.Millisecond)
}

// getAccountsLegacyAPI retrieves accounts using the legacy API with retry
// logic.
func getAccountsLegacyAPI(cCtx *cli.Context, pMap map[string]kion.Project,
	project string) ([]string, error) {

	return executeWithRetry(func() ([]string, error) {
		accounts, statusCode, err := kion.GetAccountsOnProject(
			cCtx.String("endpoint"),
			cCtx.String("token"),
			pMap[project].ID,
		)
		if err != nil {
			if statusCode == 403 {
				return nil, fmt.Errorf("permission error - will use fallback")
			}
			return nil, fmt.Errorf("failed to get accounts: %w", err)
		}

		aNames, _ := MapAccounts(accounts)
		if len(aNames) == 0 {
			return nil, fmt.Errorf("no accounts found")
		}

		return aNames, nil
	}, 3, 500*time.Millisecond)
}

// getCARsUpdatedAPI retrieves cloud access roles using the updated API with
// retry logic.
func getCARsUpdatedAPI(cCtx *cli.Context, pMap map[string]kion.Project,
	project, account string) ([]string, error) {

	return executeWithRetry(func() ([]string, error) {
		cars, err := kion.GetCARS(cCtx.String("endpoint"), cCtx.String("token"), "")
		if err != nil {
			return nil, fmt.Errorf("failed to get CARs: %w", err)
		}

		_, aMap := MapAccountsFromCARS(cars, pMap[project].ID)

		// Filter cars for the selected account
		var carsFiltered []kion.CAR
		for _, carObj := range cars {
			if carObj.AccountNumber == aMap[account] {
				carsFiltered = append(carsFiltered, carObj)
			}
		}

		cNames, _ := MapCAR(carsFiltered)
		if len(cNames) == 0 {
			return nil, fmt.Errorf("no cloud access roles found")
		}

		return cNames, nil
	}, 3, 500*time.Millisecond)
}

// getCARsLegacyAPI retrieves cloud access roles using the legacy API with
// retry logic.
func getCARsLegacyAPI(cCtx *cli.Context, pMap map[string]kion.Project,
	project, account string) ([]string, error) {

	return executeWithRetry(func() ([]string, error) {
		accounts, _, err := kion.GetAccountsOnProject(
			cCtx.String("endpoint"),
			cCtx.String("token"),
			pMap[project].ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get accounts: %w", err)
		}

		_, aMap := MapAccounts(accounts)

		// Find the account ID for the selected account
		var accountID uint
		for name, acct := range aMap {
			if name == account {
				accountID = acct.ID
				break
			}
		}

		cars, err := kion.GetCARSOnProject(
			cCtx.String("endpoint"),
			cCtx.String("token"),
			pMap[project].ID,
			accountID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get CARs: %w", err)
		}

		cNames, _ := MapCAR(cars)
		if len(cNames) == 0 {
			return nil, fmt.Errorf("no cloud access roles found")
		}

		return cNames, nil
	}, 3, 500*time.Millisecond)
}

// populateCARFromSelections populates the CAR object based on user selections
// and the API version being used.
func populateCARFromSelections(cCtx *cli.Context, car *kion.CAR, pMap map[string]kion.Project,
	results map[string]string, useUpdatedAPI bool) error {

	project := results["project"]
	account := results["account"]
	carName := results["car"]

	if useUpdatedAPI {
		cars, err := kion.GetCARS(cCtx.String("endpoint"), cCtx.String("token"), "")
		if err != nil {
			return err
		}

		_, aMap := MapAccountsFromCARS(cars, pMap[project].ID)

		// Filter cars for the selected account
		var carsFiltered []kion.CAR
		for _, carObj := range cars {
			if carObj.AccountNumber == aMap[account] {
				carsFiltered = append(carsFiltered, carObj)
			}
		}

		_, cMap := MapCAR(carsFiltered)

		// Populate the car
		car.Name = cMap[carName].Name
		car.AccountName = cMap[carName].AccountName
		car.AccountNumber = aMap[account]
		car.AccountTypeID = cMap[carName].AccountTypeID
		car.AccountID = cMap[carName].AccountID
		car.AwsIamRoleName = cMap[carName].AwsIamRoleName
		car.ID = cMap[carName].ID
		car.CloudAccessRoleType = cMap[carName].CloudAccessRoleType

	} else {
		accounts, _, err := kion.GetAccountsOnProject(
			cCtx.String("endpoint"),
			cCtx.String("token"),
			pMap[project].ID,
		)
		if err != nil {
			return err
		}

		_, aMap := MapAccounts(accounts)

		cars, err := kion.GetCARSOnProject(
			cCtx.String("endpoint"),
			cCtx.String("token"),
			pMap[project].ID,
			aMap[account].ID,
		)
		if err != nil {
			return err
		}

		_, cMap := MapCAR(cars)

		// Populate the car
		car.Name = cMap[carName].Name
		car.AccountName = cMap[carName].AccountName
		car.AccountNumber = aMap[account].Number
		car.AccountTypeID = aMap[account].TypeID
		car.AccountID = aMap[account].ID
		car.AwsIamRoleName = cMap[carName].AwsIamRoleName
		car.ID = cMap[carName].ID
		car.CloudAccessRoleType = cMap[carName].CloudAccessRoleType
	}

	return nil
}

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
	// Get list of projects, then build list of names and lookup map
	projects, err := kion.GetProjects(cCtx.String("endpoint"), cCtx.String("token"))
	if err != nil {
		return err
	}
	pNames, pMap := MapProjects(projects)
	if len(pNames) == 0 {
		return fmt.Errorf("no projects found")
	}

	useUpdatedAPI := cCtx.App.Metadata["useUpdatedCloudAccessRoleAPI"] == true

	// Define the dynamic selection steps
	steps := []DynamicStep{
		{
			Title:         "Choose a project:",
			Description:   "Select the project you want to work with",
			StaticOptions: pNames,
			Key:           "project",
		},
		{
			Title:       "Choose an Account:",
			Description: "Select the account for this project",
			DynamicOptionsFunc: func(selections map[string]string) ([]string, error) {
				project := selections["project"]
				if project == "" {
					return []string{}, nil
				}

				if useUpdatedAPI {
					return getAccountsUpdatedAPI(cCtx, pMap, project)
				} else {
					return getAccountsLegacyAPI(cCtx, pMap, project)
				}
			},
			Dependencies: []string{"project"},
			Key:          "account",
		},
		{
			Title:       "Choose a Cloud Access Role:",
			Description: "Select your cloud access role",
			DynamicOptionsFunc: func(selections map[string]string) ([]string, error) {
				project := selections["project"]
				account := selections["account"]
				if project == "" || account == "" {
					return []string{}, nil
				}

				if useUpdatedAPI {
					return getCARsUpdatedAPI(cCtx, pMap, project, account)
				} else {
					return getCARsLegacyAPI(cCtx, pMap, project, account)
				}
			},
			Dependencies: []string{"project", "account"},
			Key:          "car",
		},
	}

	// Run the dynamic selection
	results, err := PromptSelectDynamic(steps)
	if err != nil {
		return err
	}

	// Handle the 403 fallback case for legacy API
	if !useUpdatedAPI {
		// Check if we need to use the private API fallback
		_, statusCode, err := kion.GetAccountsOnProject(
			cCtx.String("endpoint"),
			cCtx.String("token"),
			pMap[results["project"]].ID,
		)
		if err != nil && statusCode == 403 {
			return carSelectorPrivateAPI(cCtx, pMap, results["project"], car)
		}
	}

	// Populate the CAR object with the results
	return populateCARFromSelections(cCtx, car, pMap, results, useUpdatedAPI)
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
	account, err := PromptSelect("Choose an Account:", "Use arrow keys to navigate, / to filter, press enter to select", aNames)
	if err != nil {
		return err
	}

	// prompt user to select car
	carname, err := PromptSelect("Choose a Cloud Access Role:", "Use arrow keys to navigate, / to filter, press enter to select", aToCMap[account])
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
