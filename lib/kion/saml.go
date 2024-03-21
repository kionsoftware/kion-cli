package kion

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	saml2 "github.com/russellhaering/gosaml2"
	samlTypes "github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
)

var (
	// SAMLLocalAuthPort is the port to use to accept back the access token from SAML
	SAMLLocalAuthPort = "8400"
)

type CSRFResponse struct {
	Data string `json:"data"`
}

type SSOAuthResponse struct {
	Data AccessData `json:"data"`
}

type AccessData struct {
	Access TokenData `json:"access"`
}

type TokenData struct {
	Token string `json:"token"`
}

type AuthData struct {
	AuthToken string
	Cookies   []*http.Cookie
	CSRFToken string
}

type SamlCallbackResult struct {
	Data *AuthData
	Err  error
}

func AuthenticateSAML(appUrl string, metadata *samlTypes.EntityDescriptor, serviceProviderIssuer string) (*AuthData, error) {
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}

	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {
		for idx, xcert := range kd.KeyInfo.X509Data.X509Certificates {
			if xcert.Data == "" {
				return nil, fmt.Errorf("metadata certificate(%d) must not be empty", idx)
			}
			certData, err := base64.StdEncoding.DecodeString(xcert.Data)
			if err != nil {
				return nil, err
			}

			idpCert, err := x509.ParseCertificate(certData)
			if err != nil {
				return nil, err
			}

			certStore.Roots = append(certStore.Roots, idpCert)
		}
	}

	// TODO: Allow importing private key and certificate from Kion application
	// For now we use a generated key/cert to sign the request, which will work
	// unless the customer has set up the IDP to verify our SP cert.
	randomKeyStore := dsig.RandomKeyStoreForTest()

	sp := &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      metadata.IDPSSODescriptor.SingleSignOnServices[0].Location,
		IdentityProviderIssuer:      metadata.EntityID,
		ServiceProviderIssuer:       serviceProviderIssuer,
		AssertionConsumerServiceURL: "http://localhost:" + SAMLLocalAuthPort + "/callback",
		SignAuthnRequests:           false,
		IDPCertificateStore:         &certStore,
		SPKeyStore:                  randomKeyStore,
	}

	tokenChan := make(chan SamlCallbackResult, 1)
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.String(), "/favicon.ico") {
			http.NotFound(rw, req)
			return
		}

		// Ensure we work with private network access check preflight requests
		if req.Method == "OPTIONS" {
			return
		}

		b, err := io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("bad SAML callback request: %w", err)}
			return
		}

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		// get csrf token
		csrfToken, csrfCookie, err := getCSRFToken(appUrl, client)
		if err != nil {
			fmt.Println("error getting csrf token: ", csrfToken)
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("error getting CSRF token: %s", csrfToken)}
			return
		}

		// update the client to use the csrf cookies
		jar, err := cookiejar.New(nil)
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("failed to create an empty cookie jar: %w", err)}
			return
		}
		url, err := url.Parse(appUrl)
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("failed to parse ssl url: %w", err)}
			return
		}
		jar.SetCookies(url, csrfCookie)
		client.Jar = jar

		r, err := http.NewRequest("POST", appUrl+"/api/v1/saml/callback", bytes.NewReader(b))
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("error creating SAML request: %w", err)}
			return
		}
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(r)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("error posting SAML assertion: %w", err)}
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("error reading SAML response body: %w", err)}
			return
		}

		ssoCodeRegexp, err := regexp.Compile(`code=(.+)">`)
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("failed to compile access token regular expression: %w", err)}
			return
		}
		groups := ssoCodeRegexp.FindStringSubmatch(string(body))
		if len(groups) < 2 {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("could not find SSO code in SAML authentication response.  Response: %v", string(body))}
			return
		}
		// parse the sso code from the groups
		ssoCode := groups[1]

		// get auth and refresh token
		authToken, refreshCookie, err := getAuthToken(appUrl, ssoCode, csrfToken, client)
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("failed to get auth token: %w", err)}
			return
		}

		// send auto-close response before returning token
		_, err = rw.Write([]byte(`
		<!doctype html>
		<html lang="en">
		  <head>
			<meta charset="utf-8">
		  </head>
		  <body>
		  	<script type="text/javascript">
			  window.close()
			</script>
			<p>You may close this window</p>
		  </body>
		</html>
		`))
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("failed to send auto-close response: %w", err)}
			return
		}

		tokenChan <- SamlCallbackResult{Data: &AuthData{
			AuthToken: authToken,
			Cookies:   append(refreshCookie, csrfCookie...),
			CSRFToken: csrfToken,
		}, Err: nil}
	})

	authURL, err := sp.BuildAuthURL("")
	if err != nil {
		log.Fatalf("The login info is invalid.\n %v", err)
	}
	var chromeCommand *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		chromeCommand = exec.Command("start", "chrome", authURL)
	case "darwin":
		chromeCommand = exec.Command("open", authURL)
	case "linux":
		chromeCommand = exec.Command("/usr/bin/google-chrome", "--new-window", authURL)
	}
	err = chromeCommand.Run()
	if chromeCommand == nil || err != nil {
		if err != nil {
			println("Error opening Chrome browser: ", err)
		} else {
			println("Could not locate Chrome browser")
		}
		println("Visit this URL To Authenticate:")
		println(authURL)
	}

	server := &http.Server{Addr: ":" + SAMLLocalAuthPort}

	go func() {

		tempResult := <-tokenChan
		err = server.Close()
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: err}
			return
		}
		tokenChan <- tempResult
	}()

	err = server.ListenAndServe()
	if err != nil && !strings.Contains(fmt.Sprintf("%v", err), "Server closed") {
		log.Fatalf("The login info is invalid.\n %v", err)
	}

	samlResult := <-tokenChan

	if samlResult.Err != nil {
		return nil, samlResult.Err
	}

	return samlResult.Data, nil
}

func DownloadSAMLMetadata(metadataUrl string) (*samlTypes.EntityDescriptor, error) {
	res, err := http.Get(metadataUrl)
	if err != nil {
		return nil, fmt.Errorf("error downloading SAML metadata file from %v: %w", metadataUrl, err)
	}
	defer res.Body.Close()

	rawMetadata, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error downloading SAML metadata file from %v: %w", metadataUrl, err)
	}

	metadata := &samlTypes.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
		return nil, fmt.Errorf("error parsing SAML metadata file from %v: %w", metadataUrl, err)
	}

	return metadata, nil
}

func ReadSAMLMetadataFile(metadataFile string) (*samlTypes.EntityDescriptor, error) {
	rawMetadata, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("error reading SAML metadata file %v: %w", metadataFile, err)
	}

	metadata := &samlTypes.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
		return nil, fmt.Errorf("error parsing SAML metadata file %v: %w", metadataFile, err)
	}

	return metadata, nil
}

func getCSRFToken(appUrl string, client *http.Client) (string, []*http.Cookie, error) {
	csrfReq, err := http.NewRequest("GET", appUrl+"/api/v2/csrf-token", nil)
	if err != nil {
		return "", nil, err
	}
	csrfResp, err := client.Do(csrfReq)
	if err != nil {
		return "", nil, err
	}
	defer csrfResp.Body.Close()
	csrfBody, err := io.ReadAll(csrfResp.Body)
	csrfCookie := csrfResp.Cookies()
	if err != nil {
		return "", nil, err
	}
	var csrfData CSRFResponse
	err = json.Unmarshal(csrfBody, &csrfData)
	if err != nil {
		return "", nil, err
	}
	return csrfData.Data, csrfCookie, nil
}

func getAuthToken(appUrl string, ssoCode string, csrfToken string, client *http.Client) (string, []*http.Cookie, error) {
	authReq, err := http.NewRequest("GET", appUrl+"/api/v2/login/sso-provider?code="+ssoCode, nil)
	authReq.Header.Set("X-Csrf-Token", csrfToken)
	if err != nil {
		return "", nil, err
	}
	authResp, err := client.Do(authReq)
	if err != nil {
		return "", nil, err
	}
	defer authResp.Body.Close()
	authBody, err := io.ReadAll(authResp.Body)
	if err != nil {
		return "", nil, err
	}

	var authData SSOAuthResponse
	err = json.Unmarshal(authBody, &authData)
	if err != nil {
		return "", nil, err
	}
	return authData.Data.Access.Token, authResp.Cookies(), nil
}
