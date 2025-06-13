package commands

import (
	"os"
	"time"

	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/urfave/cli/v2"
)

// GenStaks generates short term access keys by walking users through an
// interactive prompt. Short term access keys are either printed to stdout or a
// sub-shell is created with them set in the environment.
func (c *Cmd) GenStaks(cCtx *cli.Context) error {
	// stub out placeholders
	var car kion.CAR
	var stak kion.STAK

	// set vars for easier access
	endpoint := c.config.Kion.Url
	carName := cCtx.String("car")
	accNum := cCtx.String("account")
	accAlias := cCtx.String("alias")
	region := cCtx.String("region")

	// get command used and set cache validity buffer
	action, buffer := getActionAndBuffer(cCtx)

	// if we have what we need go look stuff up without prompts do it
	if (accNum != "" || accAlias != "") && carName != "" {
		// determine if we have a valid cached entry
		cachedSTAK, found, err := c.cache.GetStak(carName, accNum, accAlias)
		if err != nil {
			return err
		}
		getCar := true
		if found && cachedSTAK.Expiration.After(time.Now().Add(-buffer*time.Second)) {
			// cached stak found and is still valid
			stak = cachedSTAK
			if action != "subshell" && action != "save" {
				// skip getting the car for everything but subshell and save
				getCar = false
			}
		}

		// grab the car if needed
		if getCar {
			// handle auth
			err := c.setAuthToken(cCtx)
			if err != nil {
				return err
			}

			if accNum != "" {
				car, err = kion.GetCARByNameAndAccount(endpoint, c.config.Kion.ApiKey, carName, accNum)
				if err != nil {
					return err
				}
			} else {
				car, err = kion.GetCARByNameAndAlias(endpoint, c.config.Kion.ApiKey, carName, accAlias)
				if err != nil {
					return err
				}
			}
		}
	} else {
		// handle auth
		err := c.setAuthToken(cCtx)
		if err != nil {
			return err
		}

		// run through the car selector to fill any gaps
		err = helper.CARSelector(cCtx, &car)
		if err != nil {
			return err
		}

		// rebuild cache key and determine if we have a valid cached entry
		cachedSTAK, found, err := c.cache.GetStak(car.Name, car.AccountNumber, "")
		if err != nil {
			return err
		}
		if found && cachedSTAK.Expiration.After(time.Now().Add(-buffer*time.Second)) {
			// cached stak found and is still valid
			stak = cachedSTAK
		}
	}

	// grab a new stak if needed
	if stak == (kion.STAK{}) {
		var err error
		stak, err = c.authStakCache(cCtx, car.Name, car.AccountNumber, car.AccountAlias)
		if err != nil {
			return err
		}
	}

	// run the action
	switch action {
	case "credential-process":
		// NOTE: do not use os.Stderr here else credentials can be written to logs
		return helper.PrintCredentialProcess(os.Stdout, stak)
	case "print":
		return helper.PrintSTAK(os.Stdout, stak, region)
	case "save":
		return helper.SaveAWSCreds(stak, car)
	case "subshell":
		if !c.config.Kion.QuietMode {
			if err := helper.PrintFavoriteConfig(os.Stdout, car, region, "cli"); err != nil {
				return err
			}
		}
		var displayAlais string
		if accAlias != "" {
			displayAlais = accAlias
		} else {
			displayAlais = car.AccountName
		}
		return helper.CreateSubShell(car.AccountNumber, displayAlais, car.Name, stak, region)
	default:
		return nil
	}
}
