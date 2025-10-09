package commands

import (
	"fmt"
	"os"

	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
)

// FedConsole opens the CSP console for the selected account and cloud access
// role in the user's default browser.
func (c *Cmd) FedConsole(cCtx *cli.Context) error {
	// handle auth
	err := c.setAuthToken(cCtx)
	if err != nil {
		return err
	}

	// retrieve the account number, account alias, and CAR name from the context
	accNum := cCtx.String("account")
	accountAlias := cCtx.String("alias")
	carName := cCtx.String("car")

	var car kion.CAR
	if carName != "" && (accNum != "" || accountAlias != "") {
		// fetch the car directly using account number or alias and car name
		if accNum != "" {
			car, err = kion.GetCARByNameAndAccount(c.config.Kion.Url, c.config.Kion.ApiKey, carName, accNum)
			if err != nil {
				return fmt.Errorf("failed to get CAR for account %s and CAR %s: %v", accNum, carName, err)
			}
		} else {
			car, err = kion.GetCARByNameAndAlias(c.config.Kion.Url, c.config.Kion.ApiKey, carName, accountAlias)
			if err != nil {
				return fmt.Errorf("failed to get CAR for alias %s and CAR %s: %v", accountAlias, carName, err)
			}
		}
	} else {
		// walk user through the prompt workflow to select a car
		err = helper.CARSelector(cCtx, &car)
		if err != nil {
			return err
		}
	}

	// grab the csp federation url
	url, err := kion.GetFederationURL(c.config.Kion.Url, c.config.Kion.ApiKey, car)
	if err != nil {
		return err
	}

	// grab the second argument, used as a redirect parameter
	redirect := getSecondArgument(cCtx)

	// print out how to store as a favorite
	if !c.config.Kion.QuietMode {
		if err := helper.PrintFavoriteConfig(os.Stdout, car, "", "web"); err != nil {
			return err
		}
	}

	session := structs.SessionInfo{
		AccountName:    car.AccountName,
		AccountNumber:  car.AccountNumber,
		AccountTypeID:  car.AccountTypeID,
		AwsIamRoleName: car.AwsIamRoleName,
	}
	return helper.OpenBrowserRedirect(url, session, c.config.Browser, redirect, "")
}
