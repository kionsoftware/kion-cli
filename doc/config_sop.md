Config SOP
==========

When adding new configuration options there are a few things to keep in mind in
order to support several config features and established operational behavior.
This SOP ensures the appropriate precedence is followed and users can
intuitively use new options in ways they have established with existing ones.
This SOP assumes you are adding an option to the configuration file that will
also be a global flag / option.

1. Add the option to the configuration struct `lib/structs/configuraton-structs.go`
2. Add the option to the defaults override file `lib/defaults/defaults.yml` only if it is non-sensitive in nature
3. Set the option as a flag in `main.go`, this is what handles precedence as documented in the repo `README.md` file
4. If a non string var add a manual re-setting of it for when profiles are switched in `lib/commands/commands.go`
5. Test to ensure precedence is being followed as expected:
  1. No `defaults.yml` entry, no config file
  2. Then add an override in `defaults.yml`
  3. Then add an override entry in your `~/.kion.yml`
  4. Then add an override in the environment variable
  5. Then add an override in the flag option
