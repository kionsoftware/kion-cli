package helper

import (
	"fmt"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/urfave/cli/v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// getAccountOptions retrieves account options for a given project.
func getAccountOptions(cCtx *cli.Context, pMap map[string]kion.Project,
	project string) ([]string, error) {

	cars, err := kion.GetCARS(cCtx.String("endpoint"), cCtx.String("token"), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	aNames, _ := MapAccountsFromCARS(cars, pMap[project].ID)
	if len(aNames) == 0 {
		return nil, fmt.Errorf("no accounts found")
	}

	return aNames, nil
}

// getCAROptions retrieves CAR options for a given project and account.
func getCAROptions(cCtx *cli.Context, pMap map[string]kion.Project,
	project, account string) ([]string, error) {

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
}

// populateCARFromSelections populates the CAR object based on user selections.
func populateCARFromSelections(cCtx *cli.Context, car *kion.CAR, pMap map[string]kion.Project,
	results map[string]string) error {

	project := results["project"]
	account := results["account"]
	carName := results["car"]

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

	// Define the dynamic selection steps
	steps := []DynamicStep{
		{
			Title:         "Choose a project:",
			Description:   "Select the project you want to work with.",
			StaticOptions: pNames,
			Key:           "project",
		},
		{
			Title:       "Choose an Account:",
			Description: "Select the account for this project.",
			DynamicOptionsFunc: func(selections map[string]string) ([]string, error) {
				project := selections["project"]
				if project == "" {
					return []string{}, nil
				}

				return getAccountOptions(cCtx, pMap, project)
			},
			Dependencies: []string{"project"},
			Key:          "account",
		},
		{
			Title:       "Choose a Cloud Access Role:",
			Description: "Select your cloud access role.",
			DynamicOptionsFunc: func(selections map[string]string) ([]string, error) {
				project := selections["project"]
				account := selections["account"]
				if project == "" || account == "" {
					return []string{}, nil
				}

				return getCAROptions(cCtx, pMap, project, account)
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

	// Populate the CAR object with the results
	return populateCARFromSelections(cCtx, car, pMap, results)
}
