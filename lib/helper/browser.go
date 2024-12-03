package helper

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"github.com/kionsoftware/kion-cli/lib/structs"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Browser                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

var firefoxPathMac = []string{`/Applications/Firefox.app/Contents/MacOS/firefox`}
var firefoxPathLinux = []string{`/usr/bin/firefox`}
var firefoxPathWindows = []string{`\Program Files\Mozilla Firefox\firefox.exe`}

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

// OpenBrowser opens up a URL in the users system default browser. It uses a
// local webserver to host a page that handles logging users out of existing
// sessions then redirecting to the federated login page.
//
// Deprecated: Use OpenBrowserRedirect instead.
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

// OpenBrowserDirect opens up a URL in the users system default browser. It
// uses the redirect_uri query parameter to handle the logout and redirect to
// the federated login page.
func OpenBrowserRedirect(target string, session structs.SessionInfo, config structs.Browser) error {
	var err error
	var logoutURL string
	var replacement string

	switch session.AccountTypeID {
	case 1:
		// commmercial
		logoutURL = "https://signin.aws.amazon.com/oauth?Action=logout&redirect_uri="
		replacement = "://us-east-1.signin"
	case 2:
		// govcloud
		logoutURL = "https://signin.amazonaws-us-gov.com/oauth?Action=logout&redirect_uri="
		replacement = "://us-gov-east-1.signin"
	case 4:
		// c2s
		logoutURL = "http://signin.c2shome.ic.gov/oauth?Action=logout&redirect_uri="
		replacement = "://us-iso-east-1.signin"
	case 5:
		// sc2s
		logoutURL = "http://signin.sc2shome.sgov.gov/oauth?Action=logout&redirect_uri="
		replacement = "://us-isob-east-1.signin"
	}

	// update url to one that supports a redirect uri
	re := regexp.MustCompile(`:\/\/signin`)
	target = re.ReplaceAllString(target, replacement)

	// escape the target url
	encodedUrl := url.QueryEscape(target)

	if config.FirefoxContainers {
		fmt.Printf("Federating into %s (%s) via %s in a new Firefox Container\n", session.AccountName, session.AccountNumber, session.AwsIamRoleName)

		// Generate the target URL with the granted-containers extension
		target = fmt.Sprintf("ext+granted-containers:name=%s&url=%s", session.AccountName, url.QueryEscape(encodedUrl))

		// open the browser using a firefox binary
		if config.CustomBrowserPath != "" {
			fmt.Printf("Using custom browser path: %s\n", config.CustomBrowserPath)
			err = exec.Command(config.CustomBrowserPath, "--new-tab", target).Start()
		} else {
			// Try to infer the path to the Firefox binary based on the OS
			switch runtime.GOOS {
			case "linux":
				err = exec.Command(firefoxPathLinux[0], "--new-tab", target).Start()
			case "windows":
				err = exec.Command(firefoxPathWindows[0], "--new-tab", target).Start()
			case "darwin":
				err = exec.Command(firefoxPathMac[0], "--new-tab", target).Start()
			default:
				err = fmt.Errorf("unsupported platform")
			}
		}
	} else {
		fmt.Printf("Federating into %s (%s) via %s in your default browser\n", session.AccountName, session.AccountNumber, session.AwsIamRoleName)

		// generate the federation link without logout to handle any existing browser sessions
		federationLink := fmt.Sprintf("%s%s", logoutURL, encodedUrl)

		// open the browser
		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", federationLink).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", federationLink).Start()
		case "darwin":
			err = exec.Command("open", federationLink).Start()
		default:
			err = fmt.Errorf("unsupported platform")
		}
	}

	return err
}
