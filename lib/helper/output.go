package helper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kionsoftware/kion-cli/lib/kion"
)

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
