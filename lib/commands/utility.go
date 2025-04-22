package commands

import "github.com/urfave/cli/v2"

// FlushCache clears the Kion CLI cache.
func (c *Cmd) FlushCache(cCtx *cli.Context) error {
	return c.cache.FlushCache()
}

// PushFavorites pushes the local favorites to a target instance of Kion.
func (c *Cmd) PushFavorites(cCtx *cli.Context) error {
	// identify target instance of Kion
	// check version of target instance of Kion
	// pull existing upstream favorites
	// deconflict with local version of favorites, id to be pushed
	// prompt user to confirm what will be pushed where
	// perform push or bail
	return nil
}
