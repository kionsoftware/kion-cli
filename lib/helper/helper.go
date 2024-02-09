package helper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Config                                                                    //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// LoadConfig reads in the configuration yaml file located at `configFile`.
func LoadConfig(filename string, config *structs.Configuration) error {
	// read in the file
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// unmarshal to config struct
	return yaml.Unmarshal(bytes, &config)
}

// saveConfig saves the etirety of the current config to the users config file.
func SaveConfig(filename string, config structs.Configuration) error {
	// marshal to yaml
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// write it out
	return os.WriteFile(filename, bytes, 0644)
}

// SaveSession updates the session section only of the users config file.
func SaveSession(filename string, config structs.Configuration) error {
	// load in the current config file
	var newConfig structs.Configuration
	err := LoadConfig(filename, &newConfig)
	if err != nil {
		return err
	}

	// replace just the session
	newConfig.Session = config.Session

	// marshal to yaml
	bytes, err := yaml.Marshal(newConfig)
	if err != nil {
		return err
	}

	// write it out
	return os.WriteFile(filename, bytes, 0644)
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Output                                                                    //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// PrintSTAK prints out the short term access keys for AWS auth.
func PrintSTAK(stak kion.STAK, account string, region string) error {
	color.Green("Short-term Access Keys for %v:", account)
	fmt.Printf("export AWS_ACCESS_KEY_ID=%v\n", stak.AccessKey)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY=%v\n", stak.SecretAccessKey)
	fmt.Printf("export AWS_SESSION_TOKEN=%v\n", stak.SessionToken)
	fmt.Printf("export AWS_REGION=%v\n", region)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Browser                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// OpenBrowser opens up a URL in the users system default browser.
func OpenBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Shell                                                                     //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// CreateSubShell creates a subshell containing set variables for AWS short
// term access keys. It attempts to use the users configured shell and rc file
// while overriding the prompt to indicate the authed AWS account.
func CreateSubShell(accountNumber string, accountAlias string, carName string, stak kion.STAK, defaultRegion string) error {
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
		fmt.Fprintf(f, `source $HOME/.zshrc; autoload -U colors && colors; export PS1="%%F{green}[%v|%v]%%b%%f $PS1"`, accountAlias, accountNumber)
		err = f.Sync()
		if err != nil {
			return err
		}
		cmd = fmt.Sprintf(`ZDOTDIR=%v zsh`, zdotdir)
	case "bash":
		cmd = fmt.Sprintf(`bash --rcfile <(echo "source "$HOME/.bashrc; export PS1='[%v|%v] > '")`, accountAlias, accountNumber)
	default:
		cmd = fmt.Sprintf(`bash --rcfile <(echo "source "$HOME/.bashrc; export PS1='[%v|%v] > '")`, accountAlias, accountNumber)
	}

	// init shell
	shell := exec.Command("bash", "-c", cmd)

	// replicate current env vars and add stak
	shell.Env = os.Environ()
	shell.Env = append(shell.Env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", stak.AccessKey))
	shell.Env = append(shell.Env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", stak.SecretAccessKey))
	shell.Env = append(shell.Env, fmt.Sprintf("AWS_SESSION_TOKEN=%s", stak.SessionToken))
	shell.Env = append(shell.Env, fmt.Sprintf("AWS_REGION=%s", defaultRegion))
	shell.Env = append(shell.Env, fmt.Sprintf("KION_ACCOUNT_NUM=%s", accountNumber))
	shell.Env = append(shell.Env, fmt.Sprintf("KION_ACCOUNT_ALIAS=%s", accountAlias))
	shell.Env = append(shell.Env, fmt.Sprintf("KION_CAR=%s", carName))

	// configure file handlers
	shell.Stdin = os.Stdin
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr

	// run the shell
	color.Green("Starting session for %v (%v)", accountAlias, accountNumber)
	err := shell.Run()
	color.Green("Shutting down session for %v (%v)", accountAlias, accountNumber)

	return err
}

// RunCommand executes a one time command with AWS credentials set within the
// environment. Command output is sent dirctly to stdout / stderr.
func RunCommand(accountNumber string, accountAlias string, carName string, stak kion.STAK, defaultRegion string, cmd string, args ...string) error {
	// stub out an empty command stack
	newCmd := make([]string, 0)

	// if we can't find a binary, assume it's a shell alias and prep a subshell call, otherwise use the binary path
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
	env = append(env, fmt.Sprintf("AWS_REGION=%s", defaultRegion))
	env = append(env, fmt.Sprintf("KION_ACCOUNT_NUM=%s", accountNumber))
	env = append(env, fmt.Sprintf("KION_ACCOUNT_ALIAS=%s", accountAlias))
	env = append(env, fmt.Sprintf("KION_CAR=%s", carName))

	// moosh it all together
	newCmd = append(newCmd, args...)

	err = syscall.Exec(newCmd[0], newCmd[0:], env)
	return err
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Transform                                                                 //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// MapProjects transofrms a slice of Projects into a slice of their names and a
// map indexed by their names.
func MapProjects(projects []kion.Project) ([]string, map[string]kion.Project) {
	var pNames []string
	pMap := make(map[string]kion.Project)
	for _, project := range projects {
		pNames = append(pNames, project.Name)
		pMap[project.Name] = project
	}
	sort.Strings(pNames)

	return pNames, pMap
}

// MapAccounts transofrms a slice of Accounts into a slice of their names and a
// map indexed by their names.
func MapAccounts(accounts []kion.Account) ([]string, map[string]kion.Account) {
	var aNames []string
	aMap := make(map[string]kion.Account)
	for _, account := range accounts {
		aNames = append(aNames, account.Name)
		aMap[account.Name] = account
	}
	sort.Strings(aNames)

	return aNames, aMap
}

// MapCAR transofrms a slice of CARs into a slice of their names and a map
// indexed by their names.
func MapCAR(cars []kion.CAR) ([]string, map[string]kion.CAR) {
	var cNames []string
	cMap := make(map[string]kion.CAR)
	for _, car := range cars {
		cNames = append(cNames, car.Name)
		cMap[car.Name] = car
	}
	sort.Strings(cNames)

	return cNames, cMap
}

// MapIDMSs transofrms a slice of IDMSs into a slice of their names and a map
// indexd by their names.
func MapIDMSs(idmss []kion.IDMS) ([]string, map[string]kion.IDMS) {
	var iNames []string
	iMap := make(map[string]kion.IDMS)
	for _, idms := range idmss {
		iNames = append(iNames, idms.Name)
		iMap[idms.Name] = idms
	}
	sort.Strings(iNames)

	return iNames, iMap
}

// MapFavs transofrms a slice of Favorites into a slice of their names and a
// map indexd by their names.
func MapFavs(favs []structs.Favorite) ([]string, map[string]structs.Favorite) {
	var fNames []string
	fMap := make(map[string]structs.Favorite)
	for _, fav := range favs {
		fNames = append(fNames, fav.Name)
		fMap[fav.Name] = fav
	}
	sort.Strings(fNames)

	return fNames, fMap
}

// FindCARByName returns a CAR identified by its name.
func FindCARByName(cars []kion.CAR, carName string) (*kion.CAR, error) {
	for _, c := range cars {
		if c.Name == carName {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("cannot find cloud access role with name %v", carName)
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Prompts                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// survey icon and color configs
var surveyFormat = survey.WithIcons(func(icons *survey.IconSet) {
	icons.Question.Text = ""
	icons.Question.Format = "default+hb"
})

// PromptSelect promps the user to select from a slice of options. It requires
// that the selection made be one of the options provided.
func PromptSelect(message string, options []string) (string, error) {
	selection := ""
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &selection, surveyFormat)
	return selection, err
}

// PromptInput prompts the user to provide dynamic input.
func PromptInput(message string) (string, error) {
	var input string
	pi := &survey.Input{
		Message: message,
	}
	err := survey.AskOne(pi, &input, surveyFormat, survey.WithValidator(survey.Required))
	return input, err
}

// PromptPassword prompts the user to provide sensitive dynamic input.
func PromptPassword(message string) (string, error) {
	var input string
	pi := &survey.Password{
		Message: message,
	}
	err := survey.AskOne(pi, &input, surveyFormat, survey.WithValidator(survey.Required))
	return input, err
}
