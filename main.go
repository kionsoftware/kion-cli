package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	debugMode  bool
)

func init() {
	// Initialize debug mode based on an environment variable
	debugMode = os.Getenv("DEBUG_MODE") == "true"
}

// DebugLog prints debug information if debug mode is enabled
func DebugLog(format string, v ...interface{}) {
	if debugMode {
		log.Printf(format, v...)
	}
}

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

// AuthSAML directs the user to authenticte via SAML in a web browser.
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
// executed.
func beforeCommands(cCtx *cli.Context) error {
	return nil
}

// Prompt for authentication and ensure auth token is set
func authCommand(cCtx *cli.Context) error {
	// run propmts for any missing items
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
// subshell is created with them set in the environment.
func genStaks(cCtx *cli.Context) error {
	DebugLog("Starting genstacks")

	err := authCommand(cCtx)
	if err != nil {
		return err
	}

	// Assume host and token are available from your context or configuration
	userConfig, err := kion.GetUserDefaultRegions(cCtx.String("endpoint"), cCtx.String("token"))
	if err != nil {
		return err
	}

	// get list of projects, then build list of names and lookup map
	projects, err := kion.GetProjects(cCtx.String("endpoint"), cCtx.String("token"))
	if err != nil {
		return err
	}

	pNames, pMap := helper.MapProjects(projects)
	if len(pNames) == 0 {
		return fmt.Errorf("no projects found")
	}

	// prompt user to select a project
	project, err := helper.PromptSelect("Choose a project:", pNames)
	if err != nil {
		return err
	}

	// get list of accounts on project, then build a list of names and lookup map
	accounts, err := kion.GetAccountsOnProject(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID)
	if err != nil {
		return err
	}
	aNames, aMap := helper.MapAccounts(accounts)
	if len(aNames) == 0 {
		return fmt.Errorf("no accounts found")
	}

	// prompt user to select an account
	account, err := helper.PromptSelect("Choose an Account:", aNames)
	if err != nil {
		return err
	}

	selectedAccount := aMap[account]

	var defaultRegion string

	// Now check the TypeID of the selected account
	if selectedAccount.TypeID == 1 {
		defaultRegion = userConfig.Data.AwsDefaultCommercialRegion
	} else if selectedAccount.TypeID == 2 {
		defaultRegion = userConfig.Data.AwsDefaultGovcloudRegion
	} else {
		return fmt.Errorf("unknown genStaks account type")
	}

	DebugLog("default region for stack gen: %s", defaultRegion)

	// get a list of cloud access roles, then build a list of names and lookup map
	cars, err := kion.GetCARSOnProject(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID, aMap[account].ID)
	if err != nil {
		return err
	}
	cNames, _ := helper.MapCAR(cars)
	if len(cNames) == 0 {
		return fmt.Errorf("no cloud access roles found")
	}

	// prompt user to select a car
	car, err := helper.PromptSelect("Choose a Cloud Access Role:", cNames)
	if err != nil {
		return err
	}

	// generate short term tokens
	stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), car, aMap[account].Number, defaultRegion)
	if err != nil {
		return err
	}

	// print or create subshell
	if cCtx.Bool("print") {
		return helper.PrintSTAK(stak, aMap[account].Number, defaultRegion)

	} else {
		return helper.CreateSubShell(aMap[account].Number, account, car, stak, defaultRegion)

	}
}

// genStaksFav generates short term access keys from stored favorites. If a
// favorite is found that matches the passed argument it is used, otherwise the
// user is walked through a wizard to make a selection.
func genStaksFav(cCtx *cli.Context) error {
	DebugLog("Starting genStaksFav")

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

	err = authCommand(cCtx)
	if err != nil {
		return err
	}

	// Assume host and token are available from your context or configuration
	userConfig, err := kion.GetUserDefaultRegions(cCtx.String("endpoint"), cCtx.String("token"))
	if err != nil {
		return err
	}

	// generate stak
	favorite := fMap[fav]

	var defaultRegion string

	if favorite.AccountType == 1 {
		defaultRegion = userConfig.Data.AwsDefaultCommercialRegion
	} else if favorite.AccountType == 2 {
		defaultRegion = userConfig.Data.AwsDefaultGovcloudRegion
	} else {
		return fmt.Errorf("unknown genStaksFav account type: %d", favorite.AccountType)
	}

	DebugLog("default region for fav stak gen: %s", defaultRegion)

	stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), favorite.CAR, favorite.Account, defaultRegion)
	if err != nil {
		return err
	}

	// print or create subshell
	if cCtx.Bool("print") {
		return helper.PrintSTAK(stak, favorite.Account, defaultRegion)
	} else {
		return helper.CreateSubShell(favorite.Account, favorite.Name, favorite.CAR, stak, defaultRegion)
	}
}

// fedConsole opens the AWS console for the selected account and cloud access
// role in the users default browser.
func fedConsole(cCtx *cli.Context) error {
	DebugLog("Starting fedConsole")

	err := authCommand(cCtx)
	if err != nil {
		return err
	}

	// TODO: handle arg if passed else run prompts
	url, err := kion.GetFederationURL(cCtx.String("endpoint"), cCtx.String("token"), kion.CAR{})
	if err != nil {
		return err
	}
	return helper.OpenBrowser(url)
}

// listFavorites prints out the users stored favorites. Extra information is
// provided if the verbose flag is set.
func listFavorites(cCtx *cli.Context) error {
	DebugLog("Listing Favorites")

	// map our favorites for ease of use
	fNames, fMap := helper.MapFavs(config.Favorites)

	// print it out
	if cCtx.Bool("verbose") {
		for _, f := range fMap {
			fmt.Printf(" %v:\n   account number: %v\n   cloud access role: %v\n", f.Name, f.Account, f.CAR)
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
	DebugLog("Starting runCommand")

	if cCtx.String("fav") != "" {
		DebugLog("Favorite is not empty: %s", cCtx.String("fav"))
		// map our favorites for ease of use
		_, fMap := helper.MapFavs(config.Favorites)
		DebugLog("Favorites mapped")

		// if arg passed is a valid favorite use it else prompt
		var fav string
		var err error
		if fMap[cCtx.String("fav")] != (structs.Favorite{}) {
			fav = cCtx.String("fav")
			DebugLog("Using favorite: %s", fav)
		} else {
			DebugLog("Favorite not found: %s", cCtx.String("fav"))
			return errors.New("can't find fav")
		}

		userConfig, err := kion.GetUserDefaultRegions(cCtx.String("endpoint"), cCtx.String("token"))
		if err != nil {
			DebugLog("Error getting user default regions: %v", err)
			return err
		}
		DebugLog("User default regions obtained")

		err = authCommand(cCtx)
		if err != nil {
			DebugLog("Error in authCommand: %v", err)
			return err
		}
		DebugLog("authCommand successful")

		// generate stak
		favorite := fMap[fav]
		DebugLog("Favorite details retrieved: %+v", favorite)

		selectedAccount, ok := fMap[favorite.Name]
		if !ok {
			DebugLog("Account from favorite not found: %s", favorite.Name)
			return fmt.Errorf("account from favorite not found")
		}
		DebugLog("Selected account: %+v", selectedAccount)

		var defaultRegion string

		if selectedAccount.AccountType == 1 {
			defaultRegion = userConfig.Data.AwsDefaultCommercialRegion
			DebugLog("Using AWS Default Commercial Region: %s", defaultRegion)
		} else if selectedAccount.AccountType == 2 {
			defaultRegion = userConfig.Data.AwsDefaultGovcloudRegion
			DebugLog("Using AWS Default Govcloud Region: %s", defaultRegion)
		} else {
			DebugLog("Unknown favorite account type: %d", selectedAccount.AccountType)
			return fmt.Errorf("unknown fav account type")
		}

		DebugLog("default region for run command: %s", defaultRegion)

		stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), favorite.CAR, favorite.Account, defaultRegion)
		if err != nil {
			DebugLog("Error getting STAK: %v", err)
			return err
		}
		DebugLog("STAK obtained")

		err = helper.RunCommand(favorite.Account, favorite.Name, favorite.CAR, stak, defaultRegion, cCtx.Args().First(), cCtx.Args().Tail()...)
		if err != nil {
			DebugLog("Error running command: %v", err)
			return err
		}
		DebugLog("Command executed successfully")

	} else if cCtx.String("account") != "" && cCtx.String("car") != "" {
		err := authCommand(cCtx)
		if err != nil {
			DebugLog("Starting authCommand")

			return err
		}

		var defaultRegion string

		account, err := kion.GetAccount(cCtx.String("endpoint"), cCtx.String("token"), cCtx.String("account"))
		if err != nil {
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

		stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), car.Name, account.Number, defaultRegion)
		if err != nil {
			return err
		}

		err = helper.RunCommand(account.Number, account.Name, car.Name, stak, defaultRegion, cCtx.Args().First(), cCtx.Args().Tail()...)

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
	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Main                                                                      //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// main defines the command line utilities api. This should probably be broken
// out into its own function some day.
func main() {

	DebugLog("Debug Log Turned On")

	// get home directory
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	// set global for config path
	configPath = filepath.Join(home, configFile)

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
		Version:              "v0.0.1",
		Usage:                "Kion federation on the command line!",
		EnableBashCompletion: true,
		Before:               beforeCommands,
		After:                afterCommands,

		////////////////////
		//  Global Flags  //
		////////////////////

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "endpoint",
				Aliases:     []string{"e"},
				Value:       config.Kion.Url,
				EnvVars:     []string{"KION_URL"},
				Usage:       "Kion `URL`",
				Destination: &config.Kion.Url,
			},
			&cli.StringFlag{
				Name:        "user",
				Aliases:     []string{"u"},
				Value:       config.Kion.Username,
				EnvVars:     []string{"KION_USERNAME"},
				Usage:       "`USERNAME` for authentication",
				Destination: &config.Kion.Username,
			},
			&cli.StringFlag{
				Name:        "password",
				Aliases:     []string{"p"},
				Value:       config.Kion.Password,
				EnvVars:     []string{"KION_PASSWORD"},
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
				EnvVars:     []string{"KION_API_KEY"},
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
				Aliases: []string{"s"},
				Usage:   "Generate short-term access keys",
				Action:  genStaks,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "print",
						Aliases: []string{"p"},
						Usage:   "print stak only",
					},
				},
			},
			{
				Name:    "console",
				Aliases: []string{"con", "c"},
				Usage:   "Federate into the AWS console",
				Action:  fedConsole,
			},
			{
				Name:      "favorite",
				Aliases:   []string{"fav", "f"},
				Usage:     "Quickly access a favorite",
				ArgsUsage: "[FAVORITE_NAME]",
				Action:    genStaksFav,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "print",
						Aliases: []string{"p"},
						Usage:   "print stak only",
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
						Name:  "fav",
						Usage: "favorite name",
					},
					&cli.StringFlag{
						Name:  "account",
						Usage: "account number",
					},
					&cli.StringFlag{
						Name:  "car",
						Usage: "CAR name",
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
