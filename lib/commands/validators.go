package commands

import (
	"errors"

	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Validators                                                                //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// ValidateCmdStak validates the flags passed to the stak command.
func (c *Cmd) ValidateCmdStak(cCtx *cli.Context) error {
	if (cCtx.String("account") != "" || cCtx.String("alias") != "") && cCtx.String("car") == "" {
		return errors.New("must specify --car parameter when using --account or --alias")
	} else if cCtx.String("car") != "" && cCtx.String("account") == "" && cCtx.String("alias") == "" {
		return errors.New("must specify --account OR --alias parameter when using --car")
	}
	return nil
}

// ValidateCmdConsole validates the flags passed to the console command.
func (c *Cmd) ValidateCmdConsole(cCtx *cli.Context) error {
	if cCtx.String("car") != "" {
		if cCtx.String("account") == "" && cCtx.String("alias") == "" {
			return errors.New("must specify --account or --alias parameter when using --car")
		}
	} else if cCtx.String("account") != "" || cCtx.String("alias") != "" {
		return errors.New("must specify --car parameter when using --account or --alias")
	}
	return nil
}

// ValidateCmdRun validates the flags passed to the run command and sets the
// favorites region as the default region if needed to ensure precedence.
func (c *Cmd) ValidateCmdRun(cCtx *cli.Context) error {
	// Validate that either a favorite is used or both account/alias and car are provided
	if cCtx.String("favorite") == "" && ((cCtx.String("account") == "" && cCtx.String("alias") == "") || cCtx.String("car") == "") {
		return errors.New("must specify either --fav OR --account and --car  OR --alias and --car parameters")
	}

	// Set the favorite region as the default region if a favorite is used
	favName := cCtx.String("favorite")
	_, fMap := helper.MapFavs(c.config.Favorites)
	var fav string
	if fMap[favName] != (structs.Favorite{}) {
		fav = favName
	} else {
		return errors.New("can't find favorite")
	}
	favorite := fMap[fav]
	if favorite.Region != "" {
		c.config.Kion.DefaultRegion = favorite.Region
	}

	return nil
}
