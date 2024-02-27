package helper

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"

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
func PrintSTAK(w io.Writer, stak kion.STAK) error {
	fmt.Fprintf(w, "export AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\nexport AWS_SESSION_TOKEN=%v\n", stak.AccessKey, stak.SecretAccessKey, stak.SessionToken)
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
func CreateSubShell(accountNumber string, accountAlias string, carName string, stak kion.STAK) error {
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
func RunCommand(accountNumber string, accountAlias string, carName string, stak kion.STAK, cmd string, args ...string) error {
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

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Wizard Flows                                                              //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// CARSelector is a wizard that walks a user through the selection of a
// Project, then associated Accounts, then available Cloud Access Roles,
// returning the user selected Cloud Access Role.
func CARSelector(cCtx *cli.Context) (kion.CAR, error) {
	// get list of projects, then build list of names and lookup map
	projects, err := kion.GetProjects(cCtx.String("endpoint"), cCtx.String("token"))
	if err != nil {
		return kion.CAR{}, err
	}
	pNames, pMap := MapProjects(projects)
	if len(pNames) == 0 {
		return kion.CAR{}, fmt.Errorf("no projects found")
	}

	// prompt user to select a project
	project, err := PromptSelect("Choose a project:", pNames)
	if err != nil {
		return kion.CAR{}, err
	}

	// get list of accounts on project, then build a list of names and lookup map
	accounts, statusCode, err := kion.GetAccountsOnProject(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID)
	if err != nil {
		if statusCode == 403 {
			// if we're getting a 403 work around permissions bug by temp using private api
			return carSelectorPrivateAPI(cCtx, pMap, project)
		} else {
			return kion.CAR{}, err
		}
	}
	aNames, aMap := MapAccounts(accounts)
	if len(aNames) == 0 {
		return kion.CAR{}, fmt.Errorf("no accounts found")
	}

	// prompt user to select an account
	account, err := PromptSelect("Choose an Account:", aNames)
	if err != nil {
		return kion.CAR{}, err
	}

	// get a list of cloud access roles, then build a list of names and lookup map
	cars, err := kion.GetCARSOnProject(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID, aMap[account].ID)
	if err != nil {
		return kion.CAR{}, err
	}
	cNames, cMap := MapCAR(cars)
	if len(cNames) == 0 {
		return kion.CAR{}, fmt.Errorf("no cloud access roles found")
	}

	// prompt user to select a car
	car, err := PromptSelect("Choose a Cloud Access Role:", cNames)
	if err != nil {
		return kion.CAR{}, err
	}

	// inject the account name into the car struct (not returned via api)
	carObj := cMap[car]
	carObj.AccountName = account
	carObj.AccountTypeID = aMap[account].TypeID

	// return the selected car
	return carObj, nil
}

// carSelectorPrivateAPI is a temp shim workaround to address a public api
// permissions issue. CARSelector should be called directly which will the
// forward to this function if needed.
func carSelectorPrivateAPI(cCtx *cli.Context, pMap map[string]kion.Project, project string) (kion.CAR, error) {
	// hit private api endpoint to gather all users cars and their associated accounts
	caCARs, err := kion.GetConsoleAccessCARS(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID)
	if err != nil {
		return kion.CAR{}, err
	}

	// build a consolidated list of accounts from all available CARS and slice of cars per account
	var accounts []kion.Account
	cMap := make(map[string]kion.ConsoleAccessCAR)
	aToCMap := make(map[string][]string)
	for _, car := range caCARs {
		cMap[car.CARName] = car
		for _, account := range car.Accounts {
			aToCMap[account.Name] = append(aToCMap[account.Name], car.CARName)
			found := false
			for _, a := range accounts {
				if a.ID == account.ID {
					found = true
				}
			}
			if !found {
				accounts = append(accounts, account)
			}
		}
	}

	// build a list of names and lookup map
	aNames, aMap := MapAccounts(accounts)
	if len(aNames) == 0 {
		return kion.CAR{}, fmt.Errorf("no accounts found")
	}

	// prompt user to select an account
	account, err := PromptSelect("Choose an Account:", aNames)
	if err != nil {
		return kion.CAR{}, err
	}

	// prompt user to select car
	car, err := PromptSelect("Choose a Cloud Access Role:", aToCMap[account])
	if err != nil {
		return kion.CAR{}, err
	}

	// build enough of a car and return it
	return kion.CAR{
		Name:                car,
		AccountName:         account,
		AccountNumber:       aMap[account].Number,
		AccountID:           aMap[account].ID,
		AwsIamRoleName:      cMap[car].AwsIamRoleName,
		AccountTypeID:       aMap[account].TypeID,
		ID:                  cMap[car].CARID,
		CloudAccessRoleType: cMap[car].CARRoleType,
	}, nil
}
