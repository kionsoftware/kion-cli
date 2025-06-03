package commands

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
)

// ListFavorites prints out the users stored favorites and favorites from the
// Kion API. Extra information is provided if the verbose flag is set.
func (c *Cmd) ListFavorites(cCtx *cli.Context) error {

	// get the combined list of favorites from the CLI config and the Kion API (if compatible)
	useApi := cCtx.App.Metadata["useFavoritesAPI"].(bool)
	var apiFavorites []structs.Favorite
	var err error
	if useApi {
		apiFavorites, _, err = kion.GetAPIFavorites(c.config.Kion.Url, c.config.Kion.ApiKey)
		if err != nil {
			fmt.Printf("Error retrieving favorites from API: %v\n", err)
			return err
		}
	}
	result, err := helper.CombineFavorites(c.config.Favorites, apiFavorites, c.config.Kion.DefaultRegion)
	if err != nil {
		fmt.Printf("Error combining favorites: %v\n", err)
		return err
	}

	// sort favorites by name
	sort.Slice(result.All, func(i, j int) bool {
		return result.All[i].Name < result.All[j].Name
	})

	// print it out
	if cCtx.Bool("verbose") {
		for _, f := range result.All {
			accessType := f.AccessType
			if accessType == "" {
				accessType = "cli (Default)"
			}
			region := f.Region
			if region == "" {
				region = "[unset]"
			}
			fmt.Printf(" %v:\n   account number: %v\n   cloud access role: %v\n   access type: %v\n   region: %v\n", f.Name, f.Account, f.CAR, accessType, region)
		}
	} else {
		// match the generated alias name format: <normalized acct name>_<CAR>_(web|cli)_<region>
		generatedNameRe := `\w+_\w+_(web|cli)_\w+-\w+-\d`
		re, err := regexp.Compile(generatedNameRe)
		if err != nil {
			fmt.Printf("Error compiling regex for favorite alias matching: %v\n", err)
			return nil
		}

		for _, f := range result.All {
			matched := re.MatchString(f.Name)

			// if match, just print the name, which has all of the extra information
			// else, print the name with all of the extra information
			if matched {
				fmt.Printf(" %v\n", f.Name)
			} else {
				fmt.Printf(" %v (%v %v %v %v)\n", f.Name, f.Account, f.CAR, f.AccessType, f.Region)
			}
		}
	}

	return nil
}

// Favorites generates short term access keys or launches the web console
// from stored favorites. If a favorite is found that matches the passed
// argument it is used, otherwise the user is walked through a wizard to make a
// selection.
func (c *Cmd) Favorites(cCtx *cli.Context) error {

	// get the combined list of favorites from the CLI config and the Kion API (if compatible)
	useApi := cCtx.App.Metadata["useFavoritesAPI"].(bool)
	var apiFavorites []structs.Favorite
	var err error
	if useApi {
		apiFavorites, _, err = kion.GetAPIFavorites(c.config.Kion.Url, c.config.Kion.ApiKey)
		if err != nil {
			fmt.Printf("Error retrieving favorites from API: %v\n", err)
			return nil
		}
	}
	result, err := helper.CombineFavorites(c.config.Favorites, apiFavorites, c.config.Kion.DefaultRegion)
	if err != nil {
		fmt.Printf("Error combining favorites: %v\n", err)
		return nil
	}

	// run favorites through MapFavs
	fNames, fMap := helper.MapFavs(result.All)

	// if arg passed is a valid favorite use it else prompt
	var fav string
	if fMap[cCtx.Args().First()] != (structs.Favorite{}) {
		fav = cCtx.Args().First()
	} else {
		fav, err = helper.PromptSelect("Choose a Favorite:", fNames)
		if err != nil {
			return err
		}
	}

	// grab the favorite object
	favorite := fMap[fav]

	// determine favorite action, default to cli unless explicitly set to web
	if favorite.AccessType == "web" {
		// handle auth
		err = c.setAuthToken(cCtx)
		if err != nil {
			return err
		}

		// attempt to find an exact match then fallback to the first match
		car, err := kion.GetCARByNameAndAccount(c.config.Kion.Url, c.config.Kion.ApiKey, favorite.CAR, favorite.Account)
		if err != nil {
			car, err = kion.GetCARByName(c.config.Kion.Url, c.config.Kion.ApiKey, favorite.CAR)
			if err != nil {
				return err
			}
			car.AccountNumber = favorite.Account
		}

		url, err := kion.GetFederationURL(c.config.Kion.Url, c.config.Kion.ApiKey, car)
		if err != nil {
			return err
		}
		fmt.Printf("Federating into %s (%s) via %s\n", favorite.Name, favorite.Account, car.AwsIamRoleName)
		session := structs.SessionInfo{
			AccountName:    favorite.Name,
			AccountNumber:  car.AccountNumber,
			AccountTypeID:  car.AccountTypeID,
			AwsIamRoleName: car.AwsIamRoleName,
			Region:         favorite.Region,
		}
		return helper.OpenBrowserRedirect(url, session, c.config.Browser, favorite.Service, favorite.FirefoxContainerName)
	} else {
		// placeholder for our stak
		var stak kion.STAK

		// determine action and set required cache validity buffer
		action, buffer := getActionAndBuffer(cCtx)

		// check if we have a valid cached stak else grab a new one
		cachedSTAK, found, err := c.cache.GetStak(favorite.CAR, favorite.Account, "")
		if err != nil {
			return err
		}
		if found && cachedSTAK.Expiration.After(time.Now().Add(-buffer*time.Second)) {
			stak = cachedSTAK
		} else {
			stak, err = c.authStakCache(cCtx, favorite.CAR, favorite.Account, "")
			if err != nil {
				return err
			}
		}

		// credential process output, print, or create sub-shell
		switch action {
		case "credential-process":
			// NOTE: do not use os.Stderr here else credentials can be written to logs
			return helper.PrintCredentialProcess(os.Stdout, stak)
		case "print":
			return helper.PrintSTAK(os.Stdout, stak, favorite.Region)
		case "subshell":
			return helper.CreateSubShell(favorite.Account, favorite.Name, favorite.CAR, stak, favorite.Region)
		default:
			return nil
		}
	}
}
