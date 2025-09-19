package commands

import (
	"errors"
	"time"

	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
)

// RunCommand generates creds for an AWS account then executes the user
// provided command with said credentials set.
func (c *Cmd) RunCommand(cCtx *cli.Context) error {
	// set vars for easier access
	favName := cCtx.String("favorite")
	accNum := cCtx.String("account")
	accAlias := cCtx.String("alias")
	carName := cCtx.String("car")
	region := c.config.Kion.DefaultRegion

	// placeholder for our stak
	var stak kion.STAK

	// prefer favorites if specified, else use account/alias and car
	if favName != "" {
		// map our favorites for ease of use
		_, fMap := helper.MapFavs(c.config.Favorites)

		// if arg passed is a valid favorite use it else error out
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
		cachedSTAK, found, err := c.cache.GetStak(favorite.CAR, favorite.Account, "")
		if err != nil {
			return err
		}
		if found && cachedSTAK.Expiration.After(time.Now().Add(-5*time.Second)) {
			stak = cachedSTAK
		} else {
			stak, err = c.authStakCache(cCtx, favorite.CAR, favorite.Account, "")
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
		cachedSTAK, found, err := c.cache.GetStak(carName, accNum, accAlias)
		if err != nil {
			return err
		}
		if found && cachedSTAK.Expiration.After(time.Now().Add(-5*time.Second)) {
			stak = cachedSTAK
		} else {
			stak, err = c.authStakCache(cCtx, carName, accNum, accAlias)
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
