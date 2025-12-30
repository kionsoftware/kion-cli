package commands

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
)

func (c *Cmd) getFavorites(cCtx *cli.Context) ([]structs.Favorite, error) {
	// get the combined list of favorites from the CLI config and the Kion API (if compatible)
	useAPI := cCtx.App.Metadata["useFavoritesAPI"].(bool)
	var apiFavorites []structs.Favorite
	var combinedFavorites []structs.Favorite
	var err error
	if useAPI {
		// Authenticate before making API calls
		err = c.setAuthToken(cCtx)
		if err != nil {
			return apiFavorites, err
		}
		apiFavorites, _, err = kion.GetAPIFavorites(c.config.Kion.URL, c.config.Kion.APIKey)
		if err != nil {
			fmt.Printf("Error retrieving favorites from API: %v\n", err)
			return apiFavorites, err
		}
	}
	combinedFavorites, _, err = helper.CombineFavorites(c.config.Favorites, apiFavorites)
	if err != nil {
		fmt.Printf("Error combining favorites: %v\n", err)
		return combinedFavorites, err
	}
	return combinedFavorites, nil
}

// ListFavorites prints out the users stored favorites and favorites from the
// Kion API. Extra information is provided if the verbose flag is set.
func (c *Cmd) ListFavorites(cCtx *cli.Context) error {

	favorites, err := c.getFavorites(cCtx)
	if err != nil {
		return err
	}

	// sort favorites by name
	sort.Slice(favorites, func(i, j int) bool {
		return favorites[i].Name < favorites[j].Name
	})

	// print it out
	if cCtx.Bool("verbose") {
		for _, f := range favorites {
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
		for _, f := range favorites {
			fmt.Printf(" %v\n", f.DescriptiveName)
		}
	}

	return nil
}

// Favorites generates short term access keys or launches the web console
// from stored favorites. If a favorite is found that matches the passed
// argument it is used, otherwise the user is walked through a wizard to make a
// selection.
func (c *Cmd) Favorites(cCtx *cli.Context) error {

	favorites, err := c.getFavorites(cCtx)
	if err != nil {
		return err
	}

	// run favorites through MapFavs
	fNames, fMap := helper.MapFavs(favorites)

	// if arg passed is a valid favorite use it else prompt
	var fav string
	if fMap[cCtx.Args().First()] != (structs.Favorite{}) {
		fav = cCtx.Args().First()
	} else {
		fav, err = helper.PromptSelect("Choose a Favorite:", "Select your favorite from the list below.", fNames)
		if err != nil {
			return err
		}
	}

	// grab the favorite object
	favorite := fMap[fav]

	// override access type if explicitly set
	if cCtx.String("access-type") != "" {
		favorite.AccessType = cCtx.String("access-type")
	}

	// have --web flag take precedence over access-type
	if cCtx.Bool("web") {
		favorite.AccessType = "web"
	}

	// determine favorite action, default to cli unless explicitly set to web
	if favorite.AccessType == "web" {
		// handle auth
		err = c.setAuthToken(cCtx)
		if err != nil {
			return err
		}

		// attempt to find an exact match then fallback to the first match
		car, err := kion.GetCARByNameAndAccount(c.config.Kion.URL, c.config.Kion.APIKey, favorite.CAR, favorite.Account)
		if err != nil {
			car, err = kion.GetCARByName(c.config.Kion.URL, c.config.Kion.APIKey, favorite.CAR)
			if err != nil {
				return err
			}
			car.AccountNumber = favorite.Account
		}

		url, err := kion.GetFederationURL(c.config.Kion.URL, c.config.Kion.APIKey, car)
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
			// NOTE: Do not use os.Stderr here else credentials can be written to logs
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
