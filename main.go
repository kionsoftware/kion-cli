package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/99designs/keyring"
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

	c cache.Cache
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
	session, err := kion.Authenticate(cCtx.String("endpoint"), idmsID, un, pw)
	if err != nil {
		return err
	}
	session.IDMSID = idmsID
	session.UserName = un
	err = c.SetSession(session)
	if err != nil {
		return err
	}

	return cCtx.Set("token", session.Access.Token)
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

	// cache the session for 9.5 minutes, tokens are valid for 10 minutes
	timeFormat := "2006-01-02T15:04:05-0700"
	session := kion.Session{
		Access: struct {
			Expiry string `json:"expiry"`
			Token  string `json:"token"`
		}{
			Token:  authData.AuthToken,
			Expiry: time.Now().Add(570 * time.Second).Format(timeFormat),
		},
	}
	err = c.SetSession(session)
	if err != nil {
		return err
	}

	return cCtx.Set("token", authData.AuthToken)
}

// setAuthToken sets the token to be used for querying the Kion API. If not
// passed to the tool as an argument, set in the env, or present in the
// configuration dotfile it will prompt the users to authenticate. Auth methods
// are prioritized as follows: api/bearer token -> username/password -> saml.
// If flags are set for multiple methods the highest priority method will be
// used.
func setAuthToken(cCtx *cli.Context) error {
	if cCtx.Value("token") == "" {
		// if we still have an active session use it
		session, found, err := c.GetSession()
		if err != nil {
			return err
		}
		if found && session.Access.Expiry != "" {
			timeFormat := "2006-01-02T15:04:05-0700"
			now := time.Now()
			expiration, err := time.Parse(timeFormat, session.Access.Expiry)
			if err != nil {
				return err
			}
			if expiration.After(now) {
				err := cCtx.Set("token", session.Access.Token)
				if err != nil {
					return err
				}
				return nil
			}

			// TODO: uncomment when / if the application supports refresh tokens

			// // see if we can use the refresh token
			// refreshExp, err := time.Parse(timeFormat, session.Refresh.Expiry)
			// if err != nil {
			// 	return err
			// }

			// if refreshExp.After(now) {
			// 	un := session.UserName
			// 	idmsId := session.IDMSID
			// 	session, err = kion.Authenticate(cCtx.String("endpoint"), idmsId, un, session.Refresh.Token)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	session.UserName = un
			// 	session.IDMSID = idmsId
			// 	err = c.SetSession(session)
			// 	if err != nil {
			// 		return err
			// 	}

			// 	return cCtx.Set("token", session.Access.Token)
			// }
		}

		// check un / pw were set via flags and infer auth method
		if cCtx.String("user") != "" || cCtx.String("password") != "" {
			err := AuthUNPW(cCtx)
			return err
		}

		// check if saml auth flags set and auth with saml if so
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

	// initialize the keyring
	name := "kion-cli"
	ring, err := keyring.Open(keyring.Config{
		ServiceName: name,
		KeyCtlScope: "session",

		// osx
		KeychainName:             "login",
		KeychainTrustApplication: true,
		KeychainSynchronizable:   false,

		// kde wallet
		KWalletAppID:  name,
		KWalletFolder: name,

		// windows
		WinCredPrefix: name,

		// password store
		PassPrefix: name,

		//  encrypted file fallback
		FileDir:          "~/.kion",
		FilePasswordFunc: helper.PromptPassword,
	})
	if err != nil {
		return err
	}

	// initialize the cache
	if cCtx.Bool("disable-cache") {
		c = cache.NewNullCache(ring)
	} else {
		c = cache.NewCache(ring)
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
	// stub out placeholders
	var car kion.CAR
	var stak kion.STAK

	// set vars for easier access
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
		cachedSTAK, found, err := c.GetStak(cacheKey)
		if err != nil {
			return err
		}
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
			// handle auth
			err := authCommand(cCtx)
			if err != nil {
				return err
			}

			car, err = kion.GetCARByNameAndAccount(endpoint, cCtx.String("token"), carName, account)
			if err != nil {
				return err
			}
		}
	} else {
		// handle auth
		err := authCommand(cCtx)
		if err != nil {
			return err
		}

		// run through the car selector to fill any gaps
		err = helper.CARSelector(cCtx, &car)
		if err != nil {
			return err
		}

		// rebuild cache key and determine if we have a valid cached entry
		cacheKey = fmt.Sprintf("%s-%s", car.Name, car.AccountNumber)
		cachedSTAK, found, err := c.GetStak(cacheKey)
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
		// handle auth
		err := authCommand(cCtx)
		if err != nil {
			return err
		}

		// generate short term tokens
		stak, err = kion.GetSTAK(endpoint, cCtx.String("token"), car.Name, car.AccountNumber)
		if err != nil {
			return err
		}

		// store the stak in the cache
		err = c.SetStak(cacheKey, stak)
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

	// grab the favorite object
	favorite := fMap[fav]

	// determine favorite action, default to cli unless explicitly set to web
	if favorite.AccessType == "web" {
		// handle auth
		err = authCommand(cCtx)
		if err != nil {
			return err
		}

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
		return helper.OpenBrowserDirect(url, car.AccountTypeID)
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
		cachedSTAK, found, err := c.GetStak(cacheKey)
		if err != nil {
			return err
		}
		if found && cachedSTAK.Expiration.After(time.Now().Add(-buffer*time.Second)) {
			stak = cachedSTAK
		} else {
			// handle auth
			err = authCommand(cCtx)
			if err != nil {
				return err
			}

			// grab a new stak
			stak, err = kion.GetSTAK(cCtx.String("endpoint"), cCtx.String("token"), favorite.CAR, favorite.Account)
			if err != nil {
				return err
			}

			// store the stak in the cache
			err = c.SetStak(cacheKey, stak)
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
	return helper.OpenBrowserDirect(url, car.AccountTypeID)
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
	// set vars for easier access
	endpoint := cCtx.String("endpoint")
	favName := cCtx.String("favorite")
	accNum := cCtx.String("account")
	carName := cCtx.String("car")
	region := cCtx.String("region")

	// fail fast if we don't have what we need
	if favName == "" && (accNum == "" || carName == "") {
		return errors.New("must specify either --fav OR --account and --car parameters")
	}

	// placeholder for our stak
	var stak kion.STAK

	// prefer favorites if specified, else use account and car
	if favName != "" {
		// map our favorites for ease of use
		_, fMap := helper.MapFavs(config.Favorites)

		// if arg passed is a valid favorite use it else prompt
		var fav string
		var err error
		if fMap[favName] != (structs.Favorite{}) {
			fav = favName
		} else {
			return errors.New("can't find favorite")
		}

		// grab our favorite
		favorite := fMap[fav]

		// check if we have a valid cached stak else grab a new one
		cacheKey := fmt.Sprintf("%s-%s", favorite.CAR, favorite.Account)
		cachedSTAK, found, err := c.GetStak(cacheKey)
		if err != nil {
			return err
		}
		if found && cachedSTAK.Expiration.After(time.Now().Add(-5*time.Second)) {
			stak = cachedSTAK
		} else {
			// handle auth
			err := authCommand(cCtx)
			if err != nil {
				return err
			}

			// grab a new stak
			stak, err = kion.GetSTAK(endpoint, cCtx.String("token"), favorite.CAR, favorite.Account)
			if err != nil {
				return err
			}

			// store the stak in the cache
			err = c.SetStak(cacheKey, stak)
			if err != nil {
				return err
			}
		}

		// take the region flag over the favorite region
		targetRegion := region
		if targetRegion == "" {
			targetRegion = favorite.Region
		}

		// run the command
		err = helper.RunCommand(stak, targetRegion, cCtx.Args().First(), cCtx.Args().Tail()...)
		if err != nil {
			return err
		}
	} else {
		// check if we have a valid cached stak else grab a new one
		cacheKey := fmt.Sprintf("%s-%s", carName, accNum)
		cachedSTAK, found, err := c.GetStak(cacheKey)
		if err != nil {
			return err
		}
		if found && cachedSTAK.Expiration.After(time.Now().Add(-5*time.Second)) {
			stak = cachedSTAK
		} else {
			// handle auth
			err := authCommand(cCtx)
			if err != nil {
				return err
			}

			// grab a new stak
			stak, err = kion.GetSTAK(endpoint, cCtx.String("token"), carName, accNum)
			if err != nil {
				return err
			}

			// store the stak in the cache
			err = c.SetStak(cacheKey, stak)
			if err != nil {
				return err
			}
		}

		err = helper.RunCommand(stak, region, cCtx.Args().First(), cCtx.Args().Tail()...)
		if err != nil {
			return err
		}
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
