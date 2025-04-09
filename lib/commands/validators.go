package commands

import (
	"errors"

	"github.com/urfave/cli/v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Validators                                                                //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// validateCmdStak validates the flags passed to the stak command.
func (c *Cmd) ValidateCmdStak(cCtx *cli.Context) error {
	if (cCtx.String("account") != "" || cCtx.String("alias") != "") && cCtx.String("car") == "" {
		return errors.New("must specify --car parameter when using --account or --alias")
	} else if cCtx.String("car") != "" && cCtx.String("account") == "" && cCtx.String("alias") == "" {
		return errors.New("must specify --account OR --alias parameter when using --car")
	}
	return nil
}

// validateCmdConsole validates the flags passed to the console command.
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

// validateCmdRun validates the flags passed to the run command.
func (c *Cmd) ValidateCmdRun(cCtx *cli.Context) error {
	if cCtx.String("favorite") == "" && ((cCtx.String("account") == "" && cCtx.String("alias") == "") || cCtx.String("car") == "") {
		return errors.New("must specify either --fav OR --account and --car  OR --alias and --car parameters")
	}
	return nil
}
