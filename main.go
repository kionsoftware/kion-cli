package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/kionsoftware/kion-cli/lib/cache"
	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"

	"github.com/fatih/color"
	samlTypes "github.com/russellhaering/gosaml2/types"
	"github.com/urfave/cli/v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Globals                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

var (
	config     structs.Configuration
	configPath string
	configFile = ".kion.yml"

	c         *cache.Cache
	cachePath string
	cacheKey  = []byte("0123456789abcdef")
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Context Helpers                                                           //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// setEndpoint sets the target Kion installation to interact with. If not
// passed to the tool as an argument, set in the env, or present in the
// configuration dotfile it will prompt the user to provide it.
func setEndpoint(cCtx *cli.Context) error {
	if cCtx.Value("endpoint") == "" {
		kionURL, err := helper.PromptInput("Kion URL:")
		if err != nil {
			return err
		}
		err = cCtx.Set("endpoint", kionURL)
		if err != nil {
			return err
		}
	}
	return nil
}

// AuthUNPW prompts for any missing credentials then auths the users against
// Kion, stores the session data, and sets the context token.
func AuthUNPW(cCtx *cli.Context) error {
	var err error
	un := cCtx.String("user")
	pw := cCtx.String("password")
	idmsID := cCtx.Uint("idms")

	// prompt idms if needed
	if idmsID == 0 {
		idmss, err := kion.GetIDMSs(cCtx.String("endpoint"))
		if err != nil {
			return err
		}
		iNames, iMap := helper.MapIDMSs(idmss)
		if len(iNames) > 1 {
			idms, err := helper.PromptSelect("Select Login IDMS:", iNames)
			if err != nil {
				return err
			}
			idmsID = iMap[idms].ID
		} else {
			idmsID = iMap[iNames[0]].ID
		}
	}

	// prompt username if needed
	if un == "" {
		un, err = helper.PromptInput("Username:")
		if err != nil {
			return err
		}
	}

	// prompt password if needed
	if pw == "" {
		pw, err = helper.PromptPassword("Password:")
		if err != nil {
			return err
		}
	}

	// auth and capture our session
	config.Session, err = kion.Authenticate(cCtx.String("endpoint"), idmsID, un, pw)
	if err != nil {
		return err
	}
	config.Session.IDMSID = idmsID
	config.Session.UserName = un
	err = helper.SaveSession(configPath, config)
	if err != nil {
		return err
	}

	return cCtx.Set("token", config.Session.Access.Token)
}

// AuthSAML directs the user to authenticate via SAML in a web browser.
// The SAML assertion is posted to this app which is forwarded to Kion and
// exchanged for the context token.
func AuthSAML(cCtx *cli.Context) error {
	var err error
	samlMetadataFile := cCtx.String("saml-metadata-file")
	samlServiceProviderIssuer := cCtx.String("saml-sp-issuer")

	// prompt metadata url if needed
	if samlMetadataFile == "" {
		samlMetadataFile, err = helper.PromptInput("SAML Metadata URL:")
		if err != nil {
			return err
		}
	}

	// prompt issuer if needed
	if samlServiceProviderIssuer == "" {
		samlServiceProviderIssuer, err = helper.PromptInput("SAML Service Provider Issuer:")
		if err != nil {
			return err
		}
	}

	var samlMetadata *samlTypes.EntityDescriptor
	if strings.HasPrefix(samlMetadataFile, "http") {
		samlMetadata, err = kion.DownloadSAMLMetadata(samlMetadataFile)
		if err != nil {
			return err
		}
	} else {
		samlMetadata, err = kion.ReadSAMLMetadataFile(samlMetadataFile)
		if err != nil {
			return err
		}
	}

	authData, err := kion.AuthenticateSAML(
		cCtx.String("endpoint"),
		samlMetadata,
		samlServiceProviderIssuer)
	if err != nil {
		return err
	}

	return cCtx.Set("token", authData.AuthToken)
}

// setAuthToken sets the token to be used for querying the Kion API. If not
// passed to the tool as an argument, set in the env, or present in the
// configuration dotfile it will prompt the users to authenticate.
func setAuthToken(cCtx *cli.Context) error {
	if cCtx.Value("token") == "" {
		// if we still have an active session use it
		if config.Session.Access.Expiry != "" {
			timeFormat := "2006-01-02T15:04:05-0700"
			now := time.Now()
			expiration, err := time.Parse(timeFormat, config.Session.Access.Expiry)
			if err != nil {
				return err
			}
			if expiration.After(now) {
				err := cCtx.Set("token", config.Session.Access.Token)
				if err != nil {
					return err
				}
				return nil
			}
			// TODO: uncomment when / if the application supports refresh tokens
			// see if we can use the refresh token
			// refresh_exp, err := time.Parse(timeFormat, config.Session.Refresh.Expiry)
			// if err != nil {
			// 	return err
			// }

			// if refresh_exp.After(now) {
			// 	un := config.Session.UserName
			// 	idmsId := config.Session.IDMSID
			// 	config.Session, err = kion.Authenticate(cCtx.String("endpoint"), idmsId, un, config.Session.Refresh.Token)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	config.Session.UserName = un
			// 	config.Session.IDMSID = idmsId
			// 	err = helper.SaveSession(configPath, config)
			// 	if err != nil {
			// 		return err
			// 	}

			// 	return cCtx.Set("token", config.Session.Access.Token)
			// }
		}

		// check un / pw were set via flags and infer auth method
		if cCtx.String("user") != "" || cCtx.String("password") != "" {
			err := AuthUNPW(cCtx)
			return err
		}

		if cCtx.String("saml-metadata-file") != "" && cCtx.String("saml-sp-issuer") != "" {
			err := AuthSAML(cCtx)
			return err
		}

		// if no token or session found, prompt for desired auth method
		methods := []string{
			"API Key",
			"Password",
			"SAML",
		}
		authMethod, err := helper.PromptSelect("How would you like to authenticate", methods)
		if err != nil {
			return err
		}

		// handle chosen auth method
		switch authMethod {
		case "API Key":
			apiKey, err := helper.PromptPassword("API Key:")
			if err != nil {
				return err
			}
			err = cCtx.Set("token", apiKey)
			if err != nil {
				return err
			}
		case "Password":
			err := AuthUNPW(cCtx)
			if err != nil {
				return err
			}
		case "SAML":
			err := AuthSAML(cCtx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Commands                                                                  //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// beforeCommands run after the context is ready but before any subcommands are
// executed. Currently used to test feature compatibility with targeted Kion.
func beforeCommands(cCtx *cli.Context) error {
	// gather the targeted kion version
	kionVer, err := kion.GetVersion(cCtx.String("endpoint"), cCtx.String("token"))
	if err != nil {
		return err
	}
	curVer, err := version.NewSemver(kionVer)
	if err != nil {
		return err
	}

	// api/v3/me/cloud-access-role fix constraints
	v3mecarC1, _ := version.NewConstraint(">=3.6.29, < 3.7.0")
	v3mecarC2, _ := version.NewConstraint(">=3.7.17, < 3.8.0")
	v3mecarC3, _ := version.NewConstraint(">=3.8.9, < 3.9.0")
	v3mecarC4, _ := version.NewConstraint(">=3.9.0")

	// check constraints and set bool in metadata
	if v3mecarC1.Check(curVer) ||
		v3mecarC2.Check(curVer) ||
		v3mecarC3.Check(curVer) ||
		v3mecarC4.Check(curVer) {
		cCtx.App.Metadata["useUpdatedCloudAccessRoleAPI"] = true
	}

	// initialize the cache
	c = cache.NewCache(cacheKey)
	err = c.LoadFromFile(cachePath)
	if err != nil {
		return err
	}

	return nil
}

// authCommand prompts for authentication as needed and ensures an auth token
// is set.
func authCommand(cCtx *cli.Context) error {
	// run prompts for any missing items
	err := setEndpoint(cCtx)
	if err != nil {
		return err
	}
	err = setAuthToken(cCtx)
	if err != nil {
		return err
	}

	return nil
}

// genStaks generates short term access keys by walking users through an
// interactive prompt. Short term access keys are either printed to stdout or a
// sub-shell is created with them set in the environment.
func genStaks(cCtx *cli.Context) error {
	// handle auth
	err := authCommand(cCtx)
	if err != nil {
		return err
	}

	// stub out placeholders
	var car kion.CAR
	var stak kion.STAK

	// set vars for easier access
	token := cCtx.String("token")
	endpoint := cCtx.String("endpoint")
	carName := cCtx.String("car")
	account := cCtx.String("account")
	cacheKey := fmt.Sprintf("%s-%s", carName, account)
	region := cCtx.String("region")

	// grab the command usage [stak, s, setenv, savecreds, etc]
	cmdUsed := cCtx.Lineage()[1].Args().Slice()[0]

	// determine action and set required cache validity buffer
	var action string
	var buffer time.Duration
	if cCtx.Bool("credential-process") {
		action = "credential-process"
		buffer = 5
	} else if cCtx.Bool("print") || cmdUsed == "setenv" {
		action = "print"
		buffer = 300
	} else if cCtx.Bool("save") || cmdUsed == "savecreds" {
		action = "save"
		buffer = 600
	} else {
		action = "subshell"
		buffer = 300
	}

	// if we have what we need go look stuff up without prompts do it
	if account != "" && carName != "" {
		// determine if we have a valid cached entry
		cachedSTAK, found := c.Get(cacheKey)
		getCar := true
		if found && cachedSTAK.Expiration.After(time.Now().Add(-buffer*time.Second)) {
			// cached stak found and is still valid
			stak = cachedSTAK
			if action != "subshell" {
				getCar = false
			}
		}

		// grab the car if needed
		if getCar {
			car, err = kion.GetCARByNameAndAccount(endpoint, token, carName, account)
			if err != nil {
				return err
			}
		}
	} else {
		// run through the car selector to fill any gaps
		err = helper.CARSelector(cCtx, &car)
		if err != nil {
			return err
		}

		// rebuild cache key and determine if we have a valid cached entry
		cacheKey = fmt.Sprintf("%s-%s", car.Name, car.AccountNumber)
		cachedSTAK, found := c.Get(cacheKey)
		if found && cachedSTAK.Expiration.After(time.Now().Add(-buffer*time.Second)) {
			// cached stak found and is still valid
			stak = cachedSTAK
		}
	}

	// grab a new stak if needed
	if stak == (kion.STAK{}) {
		// generate short term tokens
		stak, err = kion.GetSTAK(endpoint, token, car.Name, car.AccountNumber)
		if err != nil {
			return err
		}

		// store the stak in the cache
		c.Set(cacheKey, stak)
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
		return helper.CreateSubShell(car.AccountNumber, car.AccountName, car.Name, stak, region)
	default:
		return nil
	}
}

// favorites generates short term access keys or launches the web console
// from stored favorites. If a favorite is found that matches the passed
// argument it is used, otherwise the user is walked through a wizard to make a
// selection.
func favorites(cCtx *cli.Context) error {
	// map our favorites for ease of use
	fNames, fMap := helper.MapFavs(config.Favorites)

	// if arg passed is a valid favorite use it else prompt
	var fav string
	var err error
	if fMap[cCtx.Args().First()] != (structs.Favorite{}) {
		fav = cCtx.Args().First()
	} else {
		fav, err = helper.PromptSelect("Choose a Favorite:", fNames)
		if err != nil {
			return err
		}
	}

	// handle auth
	err = authCommand(cCtx)
	if err != nil {
		return err
	}

	// grab the favorite object
	favorite := fMap[fav]

	// determine favorite action, default to cli unless explicitly set to web
	if favorite.AccessType == "web" {
		var car kion.CAR
		// attempt to find exact match then fallback to first match
		car, err = kion.GetCARByNameAndAccount(cCtx.String("endpoint"), cCtx.String("token"), favorite.CAR, favorite.Account)
		if err != nil {
			car, err = kion.GetCARByName(cCtx.String("endpoint"), cCtx.String("token"), favorite.CAR)
			if err != nil {
				return err
			}
			car.AccountNumber = favorite.Account
		}
		url, err := kion.GetFederationURL(cCtx.String("endpoint"), cCtx.String("token"), car)
		if err != nil {
			return err
		}
		fmt.Printf("Federating into %s (%s) via %s\n", favorite.Name, favorite.Account, car.AwsIamRoleName)
		return helper.OpenBrowser(url, car.AccountTypeID)
	} else {
		// placeholder for our stak
		var stak kion.STAK

		// determine action and set required cache validity buffer
		var action string
		var buffer time.Duration
		if cCtx.Bool("credential-process") {
			action = "credential-process"
			buffer = 5
		} else if cCtx.Bool("print") {
			action = "print"
			buffer = 300
		} else {
			action = "subshell"
			buffer = 300
		}

		// check if we have a valid cached stak else grab a new one
		cacheKey := fmt.Sprintf("%s-%s", favorite.CAR, favorite.Account)
		cachedSTAK, found := c.Get(cacheKey)
		if found && cachedSTAK.Expiration.After(time.Now().Add(-buffer*time.Second)) {
			stak = cachedSTAK
		} else {
			stak, err = kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), favorite.CAR, favorite.Account)
			if err != nil {
				return err
			}
		}

		// cred process output, print, or create sub-shell
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

// fedConsole opens the CSP console for the selected account and cloud access
// role in the users default browser.
func fedConsole(cCtx *cli.Context) error {
	// handle auth
	err := authCommand(cCtx)
	if err != nil {
		return err
	}

	// walk user through the prompt workflow to select a car
	var car kion.CAR
	err = helper.CARSelector(cCtx, &car)
	if err != nil {
		return err
	}

	// grab the csp federation url
	url, err := kion.GetFederationURL(cCtx.String("endpoint"), cCtx.String("token"), car)
	if err != nil {
		return err
	}
	return helper.OpenBrowser(url, car.AccountTypeID)
}

// listFavorites prints out the users stored favorites. Extra information is
// provided if the verbose flag is set.
func listFavorites(cCtx *cli.Context) error {
	// map our favorites for ease of use
	fNames, fMap := helper.MapFavs(config.Favorites)

	// print it out
	if cCtx.Bool("verbose") {
		for _, f := range fMap {
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
		for _, f := range fNames {
			fmt.Printf(" %v\n", f)
		}
	}
	return nil
}

// runCommand generates creds for an AWS account then executes the user
// provided command with said credentials set.
func runCommand(cCtx *cli.Context) error {
	if cCtx.String("favorite") != "" {
		// map our favorites for ease of use
		_, fMap := helper.MapFavs(config.Favorites)

		// if arg passed is a valid favorite use it else prompt
		var fav string
		var err error
		if fMap[cCtx.String("fav")] != (structs.Favorite{}) {
			fav = cCtx.String("fav")
		} else {
			return errors.New("can't find fav")
		}

		// handle auth
		err = authCommand(cCtx)
		if err != nil {
			return err
		}

		// generate stak
		favorite := fMap[fav]
		stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), favorite.CAR, favorite.Account)
		if err != nil {
			return err
		}

		// take the region flag over the favorite region
		targetRegion := cCtx.String("region")
		if targetRegion == "" {
			targetRegion = favorite.Region
		}

		err = helper.RunCommand(favorite.Account, favorite.Name, favorite.CAR, stak, targetRegion, cCtx.Args().First(), cCtx.Args().Tail()...)
		if err != nil {
			return err
		}

	} else if cCtx.String("account") != "" && cCtx.String("car") != "" {
		account, statusCode, err := kion.GetAccount(cCtx.String("endpoint"), cCtx.String("token"), cCtx.String("account"))
		if err != nil {
			if statusCode == 403 || statusCode == 401 {
				// try our way prone to collisions of car names
				err := authCommand(cCtx)
				if err != nil {
					return err
				}

				car, err := kion.GetCARByName(cCtx.String("endpoint"), cCtx.String("token"), cCtx.String("car"))
				if err != nil {
					return err
				}

				stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), cCtx.String("car"), cCtx.String("account"))
				if err != nil {
					return err
				}

				err = helper.RunCommand(cCtx.String("account"), car.AccountName, car.Name, stak, cCtx.String("region"), cCtx.Args().First(), cCtx.Args().Tail()...)
				if err != nil {
					return err
				}

				return nil
			}
			return err
		}

		// get a list of cloud access roles, then build a list of names and lookup map
		cars, err := kion.GetCARSOnAccount(cCtx.String("endpoint"), cCtx.String("token"), account.ID)
		if err != nil {
			return err
		}
		car, err := helper.FindCARByName(cars, cCtx.String("car"))
		if err != nil {
			return err
		}

		stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), car.Name, account.Number)
		if err != nil {
			return err
		}

		err = helper.RunCommand(account.Number, account.Name, car.Name, stak, cCtx.String("region"), cCtx.Args().First(), cCtx.Args().Tail()...)
		if err != nil {
			return err
		}
	} else {
		return errors.New("must specify either --fav OR --account and --car parameters")
	}

	return nil
}

// afterCommands run after any subcommands are executed.
func afterCommands(cCtx *cli.Context) error {
	// save our cache to disk
	err := c.SaveToFile(cachePath)
	if err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Main                                                                      //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// main defines the command line utilities API. This should probably be broken
// out into its own function some day.
func main() {

	// get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// set global for config path
	configPath = filepath.Join(home, configFile)

	// set global for cache path
	cachePath = filepath.Join(home, "/.kion/kion.cache")

	// load configuration file
	err = helper.LoadConfig(configPath, &config)
	if err != nil {
		log.Fatal(err)
	}

	// prep default text for password
	passwordDefaultText := ""
	if config.Kion.Password != "" {
		passwordDefaultText = "*****"
	}

	// prep default text for api key
	apiKeyDefaultText := ""
	if config.Kion.ApiKey != "" {
		apiKeyDefaultText = "*****"
	}

	// convert relative path specified in config file to absolute path
	samlMetadataFile := config.Kion.SamlMetadataFile
	if samlMetadataFile != "" && !strings.HasPrefix(samlMetadataFile, "http") {
		if !filepath.IsAbs(samlMetadataFile) {
			// Resolve the file path relative to the config path, which is the home directory
			samlMetadataFile = filepath.Join(filepath.Dir(configPath), samlMetadataFile)
		}
	}

	// define app configuration
	app := &cli.App{

		////////////////
		//  Metadata  //
		////////////////

		Name:                 "Kion CLI",
		Version:              "v0.2.0",
		Usage:                "Kion federation on the command line!",
		EnableBashCompletion: true,
		Before:               beforeCommands,
		After:                afterCommands,
		Metadata: map[string]interface{}{
			"useUpdatedCloudAccessRoleAPI": false,
		},

		////////////////////
		//  Global Flags  //
		////////////////////

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "endpoint",
				Aliases:     []string{"url", "e"},
				Value:       config.Kion.Url,
				EnvVars:     []string{"KION_URL"},
				Usage:       "Kion `URL`",
				Destination: &config.Kion.Url,
			},
			&cli.StringFlag{
				Name:        "user",
				Aliases:     []string{"username", "u"},
				Value:       config.Kion.Username,
				EnvVars:     []string{"KION_USERNAME", "CTKEY_USERNAME"},
				Usage:       "`USERNAME` for authentication",
				Destination: &config.Kion.Username,
			},
			&cli.StringFlag{
				Name:        "password",
				Aliases:     []string{"p"},
				Value:       config.Kion.Password,
				EnvVars:     []string{"KION_PASSWORD", "CTKEY_PASSWORD"},
				Usage:       "`PASSWORD` for authentication",
				Destination: &config.Kion.Password,
				DefaultText: passwordDefaultText,
			},
			&cli.StringFlag{
				Name:        "idms",
				Aliases:     []string{"i"},
				Value:       config.Kion.IDMS,
				EnvVars:     []string{"KION_IDMS_ID"},
				Usage:       "`IDMSID` for authentication",
				Destination: &config.Kion.IDMS,
			},
			&cli.StringFlag{
				Name:        "saml-metadata-file",
				Value:       samlMetadataFile,
				EnvVars:     []string{"KION_SAML_METADATA_FILE"},
				Usage:       "SAML metadata `FILE` or URL",
				Destination: &config.Kion.SamlMetadataFile,
			},
			&cli.StringFlag{
				Name:        "saml-sp-issuer",
				Value:       config.Kion.SamlIssuer,
				EnvVars:     []string{"KION_SAML_SP_ISSUER"},
				Usage:       "SAML Service Provider `ISSUER`",
				Destination: &config.Kion.SamlIssuer,
			},
			&cli.StringFlag{
				Name:        "token",
				Aliases:     []string{"t"},
				Value:       config.Kion.ApiKey,
				EnvVars:     []string{"KION_API_KEY", "CTKEY_APPAPIKEY"},
				Usage:       "`TOKEN` for authentication",
				Destination: &config.Kion.ApiKey,
				DefaultText: apiKeyDefaultText,
			},
		},

		////////////////
		//  Commands  //
		////////////////

		Commands: []*cli.Command{
			{
				Name:    "stak",
				Aliases: []string{"setenv", "savecreds", "s"},
				Usage:   "Generate short-term access keys",
				Action:  genStaks,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "print",
						Aliases: []string{"p"},
						Usage:   "print stak only",
					},
					&cli.StringFlag{
						Name:    "account",
						Aliases: []string{"acc", "a"},
						Usage:   "target account number, must be passed with car",
					},
					&cli.StringFlag{
						Name:    "car",
						Aliases: []string{"cloud-access-role", "c"},
						Usage:   "target cloud access role, must be passed with account",
					},
					&cli.StringFlag{
						Name:    "region",
						Aliases: []string{"r"},
						Usage:   "target region",
					},
					&cli.BoolFlag{
						Name:    "save",
						Aliases: []string{"s"},
						Usage:   "save short-term keys as aws credentials profile",
					},
					&cli.BoolFlag{
						Name:  "credential-process",
						Usage: "print stak json as AWS credential process",
					},
				},
			},
			{
				Name:    "console",
				Aliases: []string{"con", "c"},
				Usage:   "Federate into the web console",
				Action:  fedConsole,
			},
			{
				Name:      "favorite",
				Aliases:   []string{"fav", "f"},
				Usage:     "Access favorites via web console or a stak for CLI usage",
				ArgsUsage: "[FAVORITE_NAME]",
				Action:    favorites,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "print",
						Aliases: []string{"p"},
						Usage:   "print stak only",
					},
					&cli.BoolFlag{
						Name:  "credential-process",
						Usage: "print stak json as AWS credential process",
					},
				},
				BashComplete: func(cCtx *cli.Context) {
					// complete if no args are passed
					if cCtx.NArg() > 0 {
						return
					}
					// else pass favorites
					fNames, _ := helper.MapFavs(config.Favorites)
					for _, f := range fNames {
						fmt.Println(f)
					}
				},
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "list favorites",
						Action: listFavorites,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:    "verbose",
								Aliases: []string{"v"},
								Usage:   "show full favorite details",
							},
						},
					},
				},
			},
			{
				Name:      "run",
				Usage:     "Run a command with short-term access keys",
				ArgsUsage: "[COMMAND]",
				Action:    runCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "favorite",
						Aliases: []string{"fav", "f"},
						Usage:   "favorite name",
					},
					&cli.StringFlag{
						Name:    "account",
						Aliases: []string{"acc", "a"},
						Usage:   "account number",
					},
					&cli.StringFlag{
						Name:    "car",
						Aliases: []string{"c"},
						Usage:   "CAR name",
					},
					&cli.StringFlag{
						Name:    "region",
						Aliases: []string{"r"},
						Usage:   "target region",
					},
				},
			},
		},
	}

	// TODO: extend help output to include examples

	// run the app
	if err := app.Run(os.Args); err != nil {
		color.Red(" Error: %v", err)
	}
}
