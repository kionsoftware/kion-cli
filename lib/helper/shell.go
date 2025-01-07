package helper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/kionsoftware/kion-cli/lib/kion"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Shell                                                                     //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// CreateSubShell creates a sub-shell containing set variables for AWS short
// term access keys. It attempts to use the users configured shell and rc file
// while overriding the prompt to indicate the authed AWS account.
func CreateSubShell(accountNumber string, accountAlias string, carName string, stak kion.STAK, region string) error {
	// check if we know the account name
	var accountMeta string
	var accountMetaSentence string
	if accountAlias == "" {
		accountMeta = accountNumber
		accountMetaSentence = accountNumber
	} else {
		if runtime.GOOS == "windows" {
			accountMeta = fmt.Sprintf("%v^|%v", accountAlias, accountNumber)
		} else {
			accountMeta = fmt.Sprintf("%v|%v", accountAlias, accountNumber)
		}
		accountMetaSentence = fmt.Sprintf("%v (%v)", accountAlias, accountNumber)
	}

	// get users shell information
	usrShellPath := os.Getenv("SHELL")
	usrShellName := filepath.Base(usrShellPath)

	// create command based on the users shell and set prompt
	var cmd string
	switch usrShellName {
	case "zsh":
		zdotdir, err := os.MkdirTemp("", "kionzrootdir")
		if err != nil {
			return err
		}
		defer os.RemoveAll(zdotdir)
		f, err := os.Create(zdotdir + "/.zshrc")
		if err != nil {
			return err
		}
		fmt.Fprintf(f, `source $HOME/.zshrc; autoload -U colors && colors; export PS1="%%F{green}[%v]%%b%%f $PS1"`, accountMeta)
		err = f.Sync()
		if err != nil {
			return err
		}
		cmd = fmt.Sprintf(`ZDOTDIR=%v zsh`, zdotdir)
	case "bash":
		cmd = fmt.Sprintf(`bash --rcfile <(echo "source \"$HOME/.bashrc\"; export PS1='[%v] > '")`, accountMeta)
	default:
		cmd = fmt.Sprintf(`bash --rcfile <(echo "source \"$HOME/.bashrc\"; export PS1='[%v] > '")`, accountMeta)
	}

	// init shell
	var shell *exec.Cmd
	if runtime.GOOS == "windows" && usrShellName == "" {
		cmdPath := "C:\\Windows\\System32\\cmd.exe"
		shell = exec.Command(cmdPath, "/K", fmt.Sprintf(`PROMPT $E[32m[%s]$E[0m$G`, accountMeta))
	} else {
		shell = exec.Command("bash", "-c", cmd)
	}

	// replicate current env vars and add stak
	shell.Env = os.Environ()
	shell.Env = append(shell.Env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", stak.AccessKey))
	shell.Env = append(shell.Env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", stak.SecretAccessKey))
	shell.Env = append(shell.Env, fmt.Sprintf("AWS_SESSION_TOKEN=%s", stak.SessionToken))
	shell.Env = append(shell.Env, fmt.Sprintf("KION_ACCOUNT_NUM=%s", accountNumber))
	shell.Env = append(shell.Env, fmt.Sprintf("KION_ACCOUNT_ALIAS=%s", accountAlias))
	shell.Env = append(shell.Env, fmt.Sprintf("KION_CAR=%s", carName))

	// set region if one was passed
	if region != "" {
		shell.Env = append(shell.Env, fmt.Sprintf("AWS_REGION=%s", region))
	}

	// configure file handlers
	shell.Stdin = os.Stdin
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr

	// run the shell
	color.Green("Starting session for %v", accountMetaSentence)
	err := shell.Run()
	color.Green("Shutting down session for %v", accountMetaSentence)

	return err
}

// RunCommand executes a one time command with AWS credentials set within the
// environment. Command output is sent directly to stdout / stderr.
func RunCommand(stak kion.STAK, region string, cmd string, args ...string) error {
	// stub out an empty command stack
	newCmd := make([]string, 0)

	// if we can't find a binary, assume it's a shell alias and prep a sub-shell call, otherwise use the binary path
	binary, err := exec.LookPath(cmd)
	if len(binary) < 1 || err != nil {
		sh := os.Getenv("SHELL")
		if strings.HasSuffix(sh, "/bash") || strings.HasSuffix(sh, "/fish") || strings.HasSuffix(sh, "/zsh") || strings.HasSuffix(sh, "/ksh") {
			newCmd = append(newCmd, sh, "-i", "-c", cmd)
		}
	} else {
		newCmd = append(newCmd, binary)
	}

	// replicate current env vars and add stak
	env := os.Environ()
	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", stak.AccessKey))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", stak.SecretAccessKey))
	env = append(env, fmt.Sprintf("AWS_SESSION_TOKEN=%s", stak.SessionToken))

	// set region if one was passed
	if region != "" {
		env = append(env, fmt.Sprintf("AWS_REGION=%s", region))
	}

	// moosh it all together
	newCmd = append(newCmd, args...)

	err = syscall.Exec(newCmd[0], newCmd[0:], env)
	return err
}
