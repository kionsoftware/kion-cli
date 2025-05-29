package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kionsoftware/kion-cli/lib/commands"
	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/structs"

	"github.com/fatih/color"
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

	kionCliVersion string
)

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

	// allow config file to be overridden by an env var, else use default
	userConfigFile := os.Getenv("KION_CONFIG")
	if userConfigFile != "" {
		configPath = filepath.Clean(userConfigFile)
	} else {
		configPath = filepath.Join(home, configFile)
	}

	// load configuration file
	err = helper.LoadConfig(configPath, &config)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		color.Red(" Error: %v", err)
		os.Exit(1)
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
			// resolve the file path relative to the config path, which is the home directory
			samlMetadataFile = filepath.Join(filepath.Dir(configPath), samlMetadataFile)
		}
	}

	// instantiate commands, populate with config
	cmd := commands.NewCommands(&config)

	// define app configuration
	app := &cli.App{

		////////////////
		//  Metadata  //
		////////////////

		Name:                 "Kion CLI",
		Version:              kionCliVersion,
		Usage:                "Kion federation on the command line!",
		EnableBashCompletion: true,
		Before:               cmd.BeforeCommands,
		After:                cmd.AfterCommands,
		Metadata: map[string]interface{}{
			"useUpdatedCloudAccessRoleAPI": false,
			"useOldSAML":                   false,
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
			&cli.BoolFlag{
				Name:        "saml-print-url",
				Value:       config.Kion.SamlPrintUrl,
				EnvVars:     []string{"KION_SAML_PRINT_URL"},
				Usage:       "print SAML URL instead of opening browser",
				Destination: &config.Kion.SamlPrintUrl,
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
			&cli.StringFlag{
				Name:    "profile",
				EnvVars: []string{"KION_PROFILE"},
				Usage:   "configuration `PROFILE` to use",
			},
			&cli.BoolFlag{
				Name:        "disable-cache",
				Value:       config.Kion.DisableCache,
				Usage:       "disable the use of caching",
				Destination: &config.Kion.DisableCache,
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
				Before:  cmd.ValidateCmdStak,
				Action:  cmd.GenStaks,
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
						Name:    "alias",
						Aliases: []string{"aka", "l"},
						Usage:   "account alias, must be passed with car",
					},
					&cli.StringFlag{
						Name:    "car",
						Aliases: []string{"cloud-access-role", "c"},
						Usage:   "target cloud access role, must be passed with account or alias",
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
				Before:  cmd.ValidateCmdConsole,
				Action:  cmd.FedConsole,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "account",
						Aliases: []string{"acc", "a"},
						Usage:   "target account number, must be passed with car",
					},
					&cli.StringFlag{
						Name:    "alias",
						Aliases: []string{"aka", "l"},
						Usage:   "account alias, must be passed with car",
					},
					&cli.StringFlag{
						Name:    "car",
						Aliases: []string{"cloud-access-role", "c"},
						Usage:   "target cloud access role, must be passed with account or alias",
					},
				},
			},
			{
				Name:      "favorite",
				Aliases:   []string{"fav", "f"},
				Usage:     "Access favorites via web console or a stak for CLI usage",
				ArgsUsage: "[FAVORITE_NAME]",
				Action:    cmd.Favorites,
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
						Action: cmd.ListFavorites,
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
				Before:    cmd.ValidateCmdRun,
				Action:    cmd.RunCommand,
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
						Name:    "alias",
						Aliases: []string{"aka", "l"},
						Usage:   "account alias",
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
			{
				Name:  "util",
				Usage: "Utility commands",
				Subcommands: []*cli.Command{
					{
						Name:   "flush-cache",
						Usage:  "Flush the Kion CLI cache",
						Action: cmd.FlushCache,
					},
					{
						Name:   "push-favorites",
						Usage:  "Push configured favorites to Kion",
						Action: cmd.PushFavorites(configPath),
					},
				},
			},
		},
	}

	// TODO: extend help output to include examples

	// run the app
	if err := app.Run(os.Args); err != nil {
		color.Red(" Error: %v", err)
		os.Exit(1)
	}
}
