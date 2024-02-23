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
	// handle auth
	err := authCommand(cCtx)
	if err != nil {
		return err
	}

	// walk user through the propt workflow to select a car
	car, err := helper.CARSelector(cCtx)
	if err != nil {
		return err
	}

	// generate short term tokens
	stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), car.Name, car.AccountNumber)
	if err != nil {
		return err
	}

	// print or create subshell
	if cCtx.Bool("print") {
		return helper.PrintSTAK(stak, car.AccountNumber)
	} else {
		return helper.CreateSubShell(car.AccountNumber, car.AccountName, car.Name, stak)
	}
}

// genStaksFav generates short term access keys from stored favorites. If a
// favorite is found that matches the passed argument it is used, otherwise the
// user is walked through a wizard to make a selection.
func genStaksFav(cCtx *cli.Context) error {
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

	// generate stak
	favorite := fMap[fav]
	stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), favorite.CAR, favorite.Account)
	if err != nil {
		return err
	}

	// print or create subshell
	if cCtx.Bool("print") {
		return helper.PrintSTAK(stak, favorite.Account)
	} else {
		return helper.CreateSubShell(favorite.Account, favorite.Name, favorite.CAR, stak)
	}
}

// fedConsole opens the AWS console for the selected account and cloud access
// role in the users default browser.
func fedConsole(cCtx *cli.Context) error {
	// handle auth
	err := authCommand(cCtx)
	if err != nil {
		return err
	}

	// walk user through the prompt workflow to select a car
	car, err := helper.CARSelector(cCtx)
	if err != nil {
		return err
	}

	// TODO: handle arg if passed else run prompts
	url, err := kion.GetFederationURL(cCtx.String("endpoint"), cCtx.String("token"), car)
	if err != nil {
		return err
	}
	return helper.OpenBrowser(url)
}

// listFavorites prints out the users stored favorites. Extra information is
// provided if the verbose flag is set.
func listFavorites(cCtx *cli.Context) error {
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
	if cCtx.String("fav") != "" {
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

		err = helper.RunCommand(favorite.Account, favorite.Name, favorite.CAR, stak, cCtx.Args().First(), cCtx.Args().Tail()...)
		if err != nil {
			return err
		}

	} else if cCtx.String("account") != "" && cCtx.String("car") != "" {
		err := authCommand(cCtx)
		if err != nil {
			return err
		}

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

		stak, err := kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), car.Name, account.Number)
		if err != nil {
			return err
		}

		err = helper.RunCommand(account.Number, account.Name, car.Name, stak, cCtx.Args().First(), cCtx.Args().Tail()...)
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
			// {
			// 	Name:    "console",
			// 	Aliases: []string{"con", "c"},
			// 	Usage:   "Federate into the AWS console",
			// 	Action:  fedConsole,
			// },
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
