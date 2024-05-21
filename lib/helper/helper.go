package helper

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Configuration                                                             //
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

// SaveConfig saves the entirety of the current config to the users config file.
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
func PrintSTAK(w io.Writer, stak kion.STAK, region string) error {
	// handle windows vs linux for exports
	var export string
	if runtime.GOOS == "windows" {
		export = "SET"
	} else {
		export = "export"
	}

	// conditionally print region
	if region != "" {
		fmt.Fprintf(w, "export AWS_REGION=%v\n", region)
	}

	// print the stak
	fmt.Fprintf(w, "%v AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\nexport AWS_SESSION_TOKEN=%v\n", export, stak.AccessKey, stak.SecretAccessKey, stak.SessionToken)

	return nil
}

func PrintCredentialProcess(w io.Writer, stak kion.STAK, duration int) error {
	credentials := map[string]interface{}{
		"Version":         1,
		"AccessKeyId":     stak.AccessKey,
		"SecretAccessKey": stak.SecretAccessKey,
		"SessionToken":    stak.SessionToken,
		"Expiration":      time.Now().Add(time.Duration(duration) * time.Minute).Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(w, string(jsonData))
	return nil
}

// SaveAWSCreds saves the short term access keys for AWS auth to the users AWS
// configuration file.
func SaveAWSCreds(stak kion.STAK, car kion.CAR) error {
	// get the current user home directory.
	user, err := user.Current()
	if err != nil {
		return err
	}

	// derive aws creds paths
	awsCredsDir := filepath.Join(user.HomeDir, ".aws")
	awsCredsFile := filepath.Join(user.HomeDir, ".aws/credentials")

	// if the folder or file does not exist, create them
	if _, err := os.Stat(awsCredsDir); os.IsNotExist(err) {
		// create directory
		errDir := os.MkdirAll(awsCredsDir, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat(awsCredsFile); os.IsNotExist(err) {
		err = os.WriteFile(awsCredsFile, []byte(""), 0600)
		if err != nil {
			return err
		}
	}

	// read in the creds file
	contents, err := os.ReadFile(awsCredsFile)
	if err != nil {
		return err
	}

	// determine if the profile already exists
	profileName := fmt.Sprintf("[%v_%v]", car.AccountNumber, car.AwsIamRoleName)
	index := strings.Index(string(contents), profileName)

	// append the profile if it does not exist, else update it
	if index == -1 {
		f, err := os.OpenFile(awsCredsFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		linebreak := "\n"
		if runtime.GOOS == "windows" {
			linebreak = "\r\n"
		}

		text := ""
		text += fmt.Sprintf(linebreak+"[%v_%v]"+linebreak, car.AccountNumber, car.AwsIamRoleName)
		text += fmt.Sprintf("aws_access_key_id=%v"+linebreak, stak.AccessKey)
		text += fmt.Sprintf("aws_secret_access_key=%v"+linebreak, stak.SecretAccessKey)
		text += fmt.Sprintf("aws_session_token=%v"+linebreak, stak.SessionToken)

		_, err = f.WriteString(text)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}
	} else {
		f, err := os.Open(awsCredsFile)
		if err != nil {
			return err
		}

		started := false

		buf := ""

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			t := scanner.Text()
			if strings.Contains(t, profileName) {
				started = true
				buf += fmt.Sprintln(t)
				continue
			}

			if started {
				if !strings.Contains(t, "=") {
					started = false
					buf += fmt.Sprintln(t)
					continue
				}

				switch true {
				case strings.Contains(t, "aws_access_key_id"):
					buf += fmt.Sprintln("aws_access_key_id=" + stak.AccessKey)
				case strings.Contains(t, "aws_secret_access_key"):
					buf += fmt.Sprintln("aws_secret_access_key=" + stak.SecretAccessKey)
				case strings.Contains(t, "aws_session_token"):
					buf += fmt.Sprintln("aws_session_token=" + stak.SessionToken)
				default:
					return errors.New("there is a problem with the aws credentials file")
				}
				continue
			}
			buf += fmt.Sprintln(t)
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		err = os.WriteFile(awsCredsFile, []byte(buf), 0600)
		if err != nil {
			return err
		}
	}

	fmt.Println("Credentials updated in the file:", awsCredsFile)
	fmt.Printf("You can reference this profile using this flag: --profile %v_%v\n", car.AccountNumber, car.AwsIamRoleName)
	fmt.Printf("Example command: aws s3 ls --profile %v_%v\n", car.AccountNumber, car.AwsIamRoleName)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Browser                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// redirectServer runs a temp go http server to handle logging out any existing
// AWS sessions then redirecting to the federated console login.
func redirectServer(url string, typeID uint) {
	// stub out a new mux
	mux := http.NewServeMux()

	// handles logout from aws and redirection to login
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		redirPage := `
    <!DOCTYPE html>
    <html lang="en">
      <head>
        <meta charset="utf-8">
        <title>Kion-CLI: Redirecting...</title>
        <style>
          html {
            background: #f3f7f4;
          }
          body {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
          }
          iframe {
            display: none;
          }
          .kion_logo_mark {
            animation: bounce 0.5s;
            animation-direction: alternate;
            animation-timing-function: cubic-bezier(.5, 0.05, 1, .5);
            animation-iteration-count: infinite;
          }
          @keyframes bounce {
            from {
              transform: translate3d(0, 0, 0);
            }
            to {
              transform: translate3d(0, 50px, 0);
            }
          }
          /* Prefix Support */
          kion_logo_mark {
            -webkit-animation-name: bounce;
            -webkit-animation-duration: 0.5s;
            -webkit-animation-direction: alternate;
            -webkit-animation-timing-function: cubic-bezier(.5, 0.05, 1, .5);
            -webkit-animation-iteration-count: infinite;
          }
          @-webkit-keyframes bounce {
            from {
              -webkit-transform: translate3d(0, 0, 0);
              transform: translate3d(0, 0, 0);
            }
            to {
              -webkit-transform: translate3d(0, 50px, 0);
              transform: translate3d(0, 50px, 0);
            }
          }
        </style>
        <script>
          function callbackClose() {
            fetch('http://localhost:56092/done')
              .then(data => {
                console.log(data);
              })
              .catch(error => {
                console.error('Error:', error);
              });
          }

          window.onload = function() {
            let redirectURL = '%v'
            let accountTypeID = '%v'
            let agent = navigator.userAgent;
            if (accountTypeID == '1') {
                // commercial
                logoutURL = 'https://signin.aws.amazon.com/oauth?Action=logout';
            } else if (accountTypeID == '2') {
                // govcloud
                logoutURL = 'https://signin.amazonaws-us-gov.com/oauth?Action=logout';
            } else if (accountTypeID == '4') {
                logoutURL = 'http://signin.c2shome.ic.gov/oauth?Action=logout';
            } else if (accountTypeID == '5') {
                logoutURL = 'http://signin.sc2shome.sgov.gov/oauth?Action=logout';
            }
            if (agent.includes('Firefox')) {
              // popup blocked by default, user must allow
              let tab = window.open(logoutURL, '_blank')
              setTimeout(() => {
                callbackClose()
                tab.location.replace(redirectURL);
                window.close()
              }, 500);
            } else {
              const logout_iframe = document.createElement('iframe');
              logout_iframe.height = '0';
              logout_iframe.width = '0';
              logout_iframe.src = logoutURL;
              logout_iframe.onload = () => {
                callbackClose()
                window.location.replace(redirectURL);
              }
              document.body.appendChild(logout_iframe);
            }
          }
        </script>
      </head>
      <body>
        <div>
          <svg class="kion_logo_mark" viewBox="0 0 500.00001 499.99998" version="1.1" width="150" height="150" xmlns="http://www.w3.org/2000/svg" xmlns:svg="http://www.w3.org/2000/svg">
            <path id="logoMark" d="m 99.882574,277.61145 -57.26164,71.71925 -7.378755,-19.96374 a 228.4366,228.4366 0 0 1 -8.809416,-30.09222 l -1.227632,-5.59757 32.199414,-40.32547 a 3.7941326,3.7941326 0 0 0 0.01752,-4.71222 L 25,207.40537 l 1.18086,-5.51016 a 228.0104,228.0104 0 0 1 8.737594,-30.39825 l 7.395922,-20.26924 57.785764,73.49185 a 41.908883,41.908883 0 0 1 -0.217566,52.89188 z M 350.42408,252.5466 a 9.7816414,9.7816414 0 0 1 0.0175,-6.9699 L 411.27297,87.263147 405.28196,81.733373 A 231.43333,231.43333 0 0 0 384.39067,64.61169 L 371.72774,55.418289 305.32087,228.24236 a 58.091098,58.091098 0 0 0 -0.10371,41.41155 l 66.25377,175.08822 12.72442,-9.21548 a 230.66081,230.66081 0 0 0 20.93859,-17.12659 l 5.96806,-5.49911 -60.67792,-160.35313 z m 92.26509,-5.157 L 475,206.92118 l -1.20766,-5.57917 a 228.10814,228.10814 0 0 0 -8.73777,-30.17859 l -7.35283,-20.04081 -57.4913,72.00601 a 41.902051,41.902051 0 0 0 -0.22002,52.89399 l 57.56049,73.20281 7.42588,-20.18989 a 228.3357,228.3357 0 0 0 8.80171,-30.31802 l 1.19838,-5.5275 -32.30645,-41.08678 a 3.7946582,3.7946582 0 0 1 0.0175,-4.71363 z M 237.23179,21.415791 l -11.3535,0.62748 V 477.95476 l 11.3535,0.6273 c 4.35767,0.24104 8.6684,0.36332 12.81341,0.36332 4.14501,0 8.45591,-0.12263 12.81358,-0.36332 l 11.35349,-0.6273 V 22.043271 l -11.35349,-0.62748 a 227.47839,227.47839 0 0 0 -25.62699,0 z M 128.39244,55.397443 115.66276,64.640069 A 230.8761,230.8761 0 0 0 94.739412,81.801341 L 88.786063,87.300109 149.66684,248.1926 a 9.7721819,9.7721819 0 0 1 -0.0175,6.972 l -60.623967,157.77734 6.00853,5.52837 a 231.25886,231.25886 0 0 0 20.901277,17.08717 l 12.65785,9.16625 66.17459,-172.22251 a 58.03837,58.03837 0 0 0 0.10615,-41.41348 z" style="fill:#61d7ac;stroke-width:1.75176" />
          </svg>
        </div>
      </body>
    </html>
    `
		fmt.Fprintf(w, redirPage, url, typeID)
	})

	// handles callback from client when login is complete
	mux.HandleFunc("/done", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
		os.Exit(0)
	})

	// define our server
	server := http.Server{
		Addr:    ":56092",
		Handler: mux,
	}

	// start our server
	log.Fatal(server.ListenAndServe())
}

// OpenBrowser opens up a URL in the users system default browser.
func OpenBrowser(url string, typeID uint) error {
	var err error

	// start our server
	go redirectServer(url, typeID)

	// define our open url
	serverURL := "http://localhost:56092/"

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", serverURL).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", serverURL).Start()
	case "darwin":
		err = exec.Command("open", serverURL).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	// give ourselves up to 5 seconds to complete
	time.Sleep(5 * time.Second)

	return err
}

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
		accountMeta = fmt.Sprintf("%v|%v", accountAlias, accountNumber)
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
	shell := exec.Command("bash", "-c", cmd)

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
func RunCommand(accountNumber string, accountAlias string, carName string, stak kion.STAK, region string, cmd string, args ...string) error {
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
	env = append(env, fmt.Sprintf("KION_ACCOUNT_NUM=%s", accountNumber))
	env = append(env, fmt.Sprintf("KION_ACCOUNT_ALIAS=%s", accountAlias))
	env = append(env, fmt.Sprintf("KION_CAR=%s", carName))

	// set region if one was passed
	if region != "" {
		env = append(env, fmt.Sprintf("AWS_REGION=%s", region))
	}

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

// MapProjects transforms a slice of Projects into a slice of their names and a
// map indexed by their names.
func MapProjects(projects []kion.Project) ([]string, map[string]kion.Project) {
	var pNames []string
	pMap := make(map[string]kion.Project)
	for _, project := range projects {
		name := fmt.Sprintf("%v (%v)", project.Name, project.ID)
		pNames = append(pNames, name)
		pMap[name] = project
	}
	sort.Strings(pNames)

	return pNames, pMap
}

// MapAccounts transforms a slice of Accounts into a slice of their names and a
// map indexed by their names.
func MapAccounts(accounts []kion.Account) ([]string, map[string]kion.Account) {
	var aNames []string
	aMap := make(map[string]kion.Account)
	for _, account := range accounts {
		name := fmt.Sprintf("%v (%v)", account.Name, account.Number)
		aNames = append(aNames, name)
		aMap[name] = account
	}
	sort.Strings(aNames)

	return aNames, aMap
}

// MapAccountsFromCARS transforms a slice of CARs into a slice of account names
// and a map of account numbers indexed by their names. If a project ID is
// passed it will only return accounts in the given project. Note that some
// versions of Kion will not populate account metadata in CAR objects so use
// carefully (see useUpdatedCloudAccessRoleAPI bool).
func MapAccountsFromCARS(cars []kion.CAR, pid uint) ([]string, map[string]string) {
	var aNames []string
	aMap := make(map[string]string)
	for _, car := range cars {
		if pid == 0 || car.ProjectID == pid {
			name := fmt.Sprintf("%v (%v)", car.AccountName, car.AccountNumber)
			if slices.Contains(aNames, name) {
				continue
			}
			aNames = append(aNames, name)
			aMap[name] = car.AccountNumber
		}
	}
	sort.Strings(aNames)

	return aNames, aMap
}

// MapCAR transforms a slice of CARs into a slice of their names and a map
// indexed by their names.
func MapCAR(cars []kion.CAR) ([]string, map[string]kion.CAR) {
	var cNames []string
	cMap := make(map[string]kion.CAR)
	for _, car := range cars {
		name := fmt.Sprintf("%v (%v)", car.Name, car.ID)
		cNames = append(cNames, name)
		cMap[name] = car
	}
	sort.Strings(cNames)

	return cNames, cMap
}

// MapIDMSs transforms a slice of IDMSs into a slice of their names and a map
// indexed by their names.
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

// MapFavs transforms a slice of Favorites into a slice of their names and a
// map indexed by their names.
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
	return &kion.CAR{}, fmt.Errorf("cannot find cloud access role with name %v", carName)
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Prompts                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// surveyFormat sets survey icon and color configs.
var surveyFormat = survey.WithIcons(func(icons *survey.IconSet) {
	icons.Question.Text = ""
	icons.Question.Format = "default+hb"
})

// PromptSelect prompts the user to select from a slice of options. It requires
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
// Project, then associated Accounts, then available Cloud Access Roles, to set
// the user selected Cloud Access Role. Optional account number and or car name
// can be passed via an existing car struct, the flow will dynamically ask what
// is needed to be able to find the full car.
func CARSelector(cCtx *cli.Context, car *kion.CAR) error {
	// get list of projects, then build list of names and lookup map
	projects, err := kion.GetProjects(cCtx.String("endpoint"), cCtx.String("token"))
	if err != nil {
		return err
	}
	pNames, pMap := MapProjects(projects)
	if len(pNames) == 0 {
		return fmt.Errorf("no projects found")
	}

	// prompt user to select a project
	project, err := PromptSelect("Choose a project:", pNames)
	if err != nil {
		return err
	}

	if cCtx.App.Metadata["useUpdatedCloudAccessRoleAPI"] == true {
		// TODO: consolidate on this logic when support for 3.9 drops, that will
		// give us one full support line of buffer

		// get all cars for authed user, works with min permission set
		cars, err := kion.GetCARS(cCtx.String("endpoint"), cCtx.String("token"))
		if err != nil {
			return err
		}
		aNames, aMap := MapAccountsFromCARS(cars, pMap[project].ID)
		if len(aNames) == 0 {
			return fmt.Errorf("no accounts found")
		}

		// prompt user to select an account
		account, err := PromptSelect("Choose an Account:", aNames)
		if err != nil {
			return err
		}

		// narrow it down to just cars associated with the account
		var carsFiltered []kion.CAR
		for _, carObj := range cars {
			if carObj.AccountNumber == aMap[account] {
				carsFiltered = append(carsFiltered, carObj)
			}
		}
		cNames, cMap := MapCAR(carsFiltered)
		if len(cNames) == 0 {
			return fmt.Errorf("you have no cloud access roles assigned")
		}

		// prompt user to select a car
		carname, err := PromptSelect("Choose a Cloud Access Role:", cNames)
		if err != nil {
			return err
		}

		// inject the metadata into the car
		car.Name = cMap[carname].Name
		car.AccountName = cMap[carname].AccountName
		car.AccountNumber = aMap[account]
		car.AccountTypeID = cMap[carname].AccountTypeID
		car.AccountID = cMap[carname].AccountID
		car.AwsIamRoleName = cMap[carname].AwsIamRoleName
		car.ID = cMap[carname].ID
		car.CloudAccessRoleType = cMap[carname].CloudAccessRoleType

		// return nil
		return nil
	} else {
		// get list of accounts on project, then build a list of names and lookup map
		accounts, statusCode, err := kion.GetAccountsOnProject(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID)
		if err != nil {
			if statusCode == 403 {
				// if we're getting a 403 work around permissions bug by temp using private api
				return carSelectorPrivateAPI(cCtx, pMap, project, car)
			} else {
				return err
			}
		}
		aNames, aMap := MapAccounts(accounts)
		if len(aNames) == 0 {
			return fmt.Errorf("no accounts found")
		}

		// prompt user to select an account
		account, err := PromptSelect("Choose an Account:", aNames)
		if err != nil {
			return err
		}

		// get a list of cloud access roles, then build a list of names and lookup map
		cars, err := kion.GetCARSOnProject(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID, aMap[account].ID)
		if err != nil {
			return err
		}
		cNames, cMap := MapCAR(cars)
		if len(cNames) == 0 {
			return fmt.Errorf("no cloud access roles found")
		}

		// prompt user to select a car
		carname, err := PromptSelect("Choose a Cloud Access Role:", cNames)
		if err != nil {
			return err
		}

		// inject the metadata into the car
		car.Name = cMap[carname].Name
		car.AccountName = cMap[carname].AccountName
		car.AccountNumber = aMap[account].Number
		car.AccountTypeID = aMap[account].TypeID
		car.AccountID = aMap[account].ID
		car.AwsIamRoleName = cMap[carname].AwsIamRoleName
		car.ID = cMap[carname].ID
		car.CloudAccessRoleType = cMap[carname].CloudAccessRoleType

		// return nil
		return nil
	}
}

// carSelectorPrivateAPI is a temp shim workaround to address a public API
// permissions issue. CARSelector should be called directly which will the
// forward to this function if needed.
func carSelectorPrivateAPI(cCtx *cli.Context, pMap map[string]kion.Project, project string, car *kion.CAR) error {
	// hit private api endpoint to gather all users cars and their associated accounts
	caCARs, err := kion.GetConsoleAccessCARS(cCtx.String("endpoint"), cCtx.String("token"), pMap[project].ID)
	if err != nil {
		return err
	}

	// build a consolidated list of accounts from all available CARS and slice of cars per account
	var accounts []kion.Account
	cMap := make(map[string]kion.ConsoleAccessCAR)
	aToCMap := make(map[string][]string)
	for _, car := range caCARs {
		cname := fmt.Sprintf("%v (%v)", car.CARName, car.CARID)
		cMap[cname] = car
		for _, account := range car.Accounts {
			name := fmt.Sprintf("%v (%v)", account.Name, account.Number)
			aToCMap[name] = append(aToCMap[account.Name], cname)
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
		return fmt.Errorf("no accounts found")
	}

	// prompt user to select an account
	account, err := PromptSelect("Choose an Account:", aNames)
	if err != nil {
		return err
	}

	// prompt user to select car
	carname, err := PromptSelect("Choose a Cloud Access Role:", aToCMap[account])
	if err != nil {
		return err
	}

	// build enough of a car and return it
	car.Name = cMap[carname].CARName
	car.AccountName = aMap[account].Name
	car.AccountNumber = aMap[account].Number
	car.AccountID = aMap[account].ID
	car.AwsIamRoleName = cMap[carname].AwsIamRoleName
	car.AccountTypeID = aMap[account].TypeID
	car.ID = cMap[carname].CARID
	car.CloudAccessRoleType = cMap[carname].CARRoleType

	return nil
}
