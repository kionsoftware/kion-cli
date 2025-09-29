package commands

import (
	"fmt"
	"time"

	"github.com/99designs/keyring"
	"github.com/hashicorp/go-version"
	"github.com/kionsoftware/kion-cli/lib/cache"
	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"

	"github.com/urfave/cli/v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Commands Object                                                           //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// Cmd is the main command object for the Kion CLI. It makes the configuration
// and cache available to all command actions.
type Cmd struct {
	config *structs.Configuration
	cache  cache.Cache
}

// NewCommands stands up a new instance of commands with the provided
// configuration.
func NewCommands(cfg *structs.Configuration) *Cmd {
	return &Cmd{
		config: cfg,
	}
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// getSecondArgument returns the second positional argument passed to the cli.
func getSecondArgument(cCtx *cli.Context) string {
	if cCtx.Args().Len() > 0 {
		return cCtx.Args().Get(0)
	}
	return ""
}

// getThirdArgument returns the third positional argument passed to the cli.
func getThirdArgument(cCtx *cli.Context) string {
	if cCtx.Args().Len() > 0 {
		return cCtx.Args().Get(1)
	}
	return ""
}

// setEndpoint sets the target Kion installation to interact with. If not
// passed to the tool as an argument, set in the env, or present in the
// configuration dotfile it will prompt the user to provide it.
func (c *Cmd) setEndpoint() error {
	if c.config.Kion.Url == "" {
		kionURL, err := helper.PromptInput("Kion URL:")
		if err != nil {
			return err
		}
		c.config.Kion.Url = kionURL
	}
	return nil
}

// getActionAndBuffer determines the action based on the passed flags and sets
// a buffer for the associated action used to determine the cache validity.
func getActionAndBuffer(cCtx *cli.Context) (string, time.Duration) {
	// grab the command usage [stak, s, setenv, savecreds, etc]
	cmdUsed := cCtx.Lineage()[1].Args().Slice()[0]

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

	return action, buffer
}

// authStakCache handles the common pattern of authenticating the user,
// grabbing a STAK, and caching it. Used to dry up code in various commands.
func (c *Cmd) authStakCache(cCtx *cli.Context, carName string, accNum string, accAlias string) (kion.STAK, error) {
	// handle auth
	err := c.setAuthToken(cCtx)
	if err != nil {
		return kion.STAK{}, err
	}

	// generate short term tokens
	stak, err := kion.GetSTAK(c.config.Kion.Url, c.config.Kion.ApiKey, carName, accNum, accAlias)
	if err != nil {
		return kion.STAK{}, err
	}

	// store the stak in the cache
	err = c.cache.SetStak(carName, accNum, accAlias, stak)
	if err != nil {
		return kion.STAK{}, err
	}

	return stak, err
}

// initCache initializes the cache based on the configuration. If the cache
// is disabled or the user has requested to flush the cache, a null cache is
// used. Otherwise, a real cache is initialized using the keyring library.
func (c *Cmd) initCache(cCtx *cli.Context) error {
	// if the cache is not disabled, or if the user has requested to flush the
	// cache, we initialize the real cache. Otherwise, we use a null cache.
	if !c.config.Kion.DisableCache || getThirdArgument(cCtx) == "flush-cache" {
		if c.config.Kion.DebugMode {
			keyring.Debug = true
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

			// gnome wallet (libsecret)
			LibSecretCollectionName: "login",

			// windows
			WinCredPrefix: name,

			// password store
			PassPrefix: name,

			// encrypted file fallback
			FileDir:          "~/.kion",
			FilePasswordFunc: helper.PromptPassword,
		})
		if err != nil {
			return err
		}
		c.cache = cache.NewCache(ring)
	} else {
		c.cache = cache.NewNullCache()
	}

	return nil
}

// handleProfile checks if a profile is specified and if so, it overrides the
// default configuration with the profile's values. It also honors any global
// flags that were set in the CLI context, allowing them to take precedence over
// the profile's values.
func (c *Cmd) handleProfile(profileName string, cCtx *cli.Context) error {
	if profileName != "" {
		// grab all manually set global flags so we can honor them over the chosen
		// profiles values
		// bool values need to be explicitly handled here since we're not iterating
		setStrings := make(map[string]string)

		var disableCacheFlagged bool
		var debugFlagged bool
		var quietFlagged bool

		setGlobalFlags := cCtx.FlagNames()
		for _, flag := range setGlobalFlags {
			switch flag {
			case "endpoint":
				setStrings["endpoint"] = c.config.Kion.Url
			case "user":
				setStrings["user"] = c.config.Kion.Username
			case "password":
				setStrings["password"] = c.config.Kion.Password
			case "idms":
				setStrings["idms"] = c.config.Kion.IDMS
			case "saml-metadata-file":
				setStrings["saml-metadata-file"] = c.config.Kion.SamlMetadataFile
			case "saml-sp-issuer":
				setStrings["saml-sp-issuer"] = c.config.Kion.SamlIssuer
			case "token":
				setStrings["token"] = c.config.Kion.ApiKey
			// non-string flags
			case "disable-cache":
				disableCacheFlagged = true
			case "debug":
				debugFlagged = true
			case "quiet":
				quietFlagged = true
			}
		}

		// grab the profile and if found and not empty override the default config
		profile, found := c.config.Profiles[profileName]
		if found {
			c.config.Kion = profile.Kion
			c.config.Favorites = profile.Favorites
		} else {
			return fmt.Errorf("profile not found: %s", profileName)
		}

		// honor any global flags that were set to maintain precedence
		for key, value := range setStrings {
			err := cCtx.Set(key, value)
			if err != nil {
				return err
			}
		}

		// handle non-string flags
		if disableCacheFlagged {
			c.config.Kion.DisableCache = true
		}
		if debugFlagged {
			c.config.Kion.DebugMode = true
		}
		if quietFlagged {
			c.config.Kion.QuietMode = true
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Before & After Commands                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// BeforeCommands run after the context is ready but before any subcommands are
// executed. Currently used to test feature compatibility with targeted Kion.
func (c *Cmd) BeforeCommands(cCtx *cli.Context) error {
	// skip before bits if we don't need them (ie we're just printing help)
	args := cCtx.Args().Slice()
	if len(args) == 0 || args[0] == "help" || args[0] == "h" {
		return nil
	}

	// switch profiles if specified
	profileName := cCtx.String("profile")
	err := c.handleProfile(profileName, cCtx)
	if err != nil {
		return err
	}

	// grab the Kion url if not already set
	err = c.setEndpoint()
	if err != nil {
		return err
	}

	// gather the targeted Kion version
	kionVer, err := kion.GetVersion(c.config.Kion.Url)
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

	// SAML metadata file handling
	newSaml, _ := version.NewConstraint(">=3.8.0")
	if !newSaml.Check(curVer) {
		cCtx.App.Metadata["useOldSAML"] = true
	}

	// favorites API check
	favoritesAPI1, _ := version.NewConstraint(">=3.13.5, < 3.14.0")
	favoritesAPI2, _ := version.NewConstraint(">=3.14.1")
	if favoritesAPI1.Check(curVer) || favoritesAPI2.Check(curVer) {
		cCtx.App.Metadata["useFavoritesAPI"] = true
	}

	// initialize the cache
	err = c.initCache(cCtx)
	if err != nil {
		return err
	}

	return nil
}

// AfterCommands run after any subcommands are executed.
func (c *Cmd) AfterCommands(cCtx *cli.Context) error {
	return nil
}
