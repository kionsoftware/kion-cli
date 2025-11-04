package kion

import (
	"bytes"
	"context"
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
	"time"

	"github.com/fatih/color"
	saml2 "github.com/russellhaering/gosaml2"
	samlTypes "github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
)

var (
	// SAMLLocalAuthPort is the port to use to accept back the access token from SAML
	SAMLLocalAuthPort = "8400"
	AuthPage          = `
		<!doctype html>
		<html lang="en">
		  <head>
        <meta charset="utf-8">
        <title>Kion-CLI</title>
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
          #wrapper {
            text-align: center;
            font-family: monospace, monospace;
          }
        </style>
		  </head>
		  <body>
        <div id="wrapper">
          <svg class="kion_logo_mark" viewBox="0 0 500.00001 499.99998" version="1.1" width="150" height="150" xmlns="http://www.w3.org/2000/svg" xmlns:svg="http://www.w3.org/2000/svg">
            <path id="logoMark" d="m 99.882574,277.61145 -57.26164,71.71925 -7.378755,-19.96374 a 228.4366,228.4366 0 0 1 -8.809416,-30.09222 l -1.227632,-5.59757 32.199414,-40.32547 a 3.7941326,3.7941326 0 0 0 0.01752,-4.71222 L 25,207.40537 l 1.18086,-5.51016 a 228.0104,228.0104 0 0 1 8.737594,-30.39825 l 7.395922,-20.26924 57.785764,73.49185 a 41.908883,41.908883 0 0 1 -0.217566,52.89188 z M 350.42408,252.5466 a 9.7816414,9.7816414 0 0 1 0.0175,-6.9699 L 411.27297,87.263147 405.28196,81.733373 A 231.43333,231.43333 0 0 0 384.39067,64.61169 L 371.72774,55.418289 305.32087,228.24236 a 58.091098,58.091098 0 0 0 -0.10371,41.41155 l 66.25377,175.08822 12.72442,-9.21548 a 230.66081,230.66081 0 0 0 20.93859,-17.12659 l 5.96806,-5.49911 -60.67792,-160.35313 z m 92.26509,-5.157 L 475,206.92118 l -1.20766,-5.57917 a 228.10814,228.10814 0 0 0 -8.73777,-30.17859 l -7.35283,-20.04081 -57.4913,72.00601 a 41.902051,41.902051 0 0 0 -0.22002,52.89399 l 57.56049,73.20281 7.42588,-20.18989 a 228.3357,228.3357 0 0 0 8.80171,-30.31802 l 1.19838,-5.5275 -32.30645,-41.08678 a 3.7946582,3.7946582 0 0 1 0.0175,-4.71363 z M 237.23179,21.415791 l -11.3535,0.62748 V 477.95476 l 11.3535,0.6273 c 4.35767,0.24104 8.6684,0.36332 12.81341,0.36332 4.14501,0 8.45591,-0.12263 12.81358,-0.36332 l 11.35349,-0.6273 V 22.043271 l -11.35349,-0.62748 a 227.47839,227.47839 0 0 0 -25.62699,0 z M 128.39244,55.397443 115.66276,64.640069 A 230.8761,230.8761 0 0 0 94.739412,81.801341 L 88.786063,87.300109 149.66684,248.1926 a 9.7721819,9.7721819 0 0 1 -0.0175,6.972 l -60.623967,157.77734 6.00853,5.52837 a 231.25886,231.25886 0 0 0 20.901277,17.08717 l 12.65785,9.16625 66.17459,-172.22251 a 58.03837,58.03837 0 0 0 0.10615,-41.41348 z" style="fill:#61d7ac;stroke-width:1.75176" />
          </svg>
          <p>YOU MAY CLOSE THIS WINDOW</p>
          <script type="text/javascript">
            window.close()
          </script>
        </div>
		  </body>
		</html>
    `
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

func callExternalAuth(sp *saml2.SAMLServiceProvider, tokenChan chan SamlCallbackResult, printURL bool) (*AuthData, error) {
	authURL, err := sp.BuildAuthURL("")
	if err != nil {
		log.Fatalf("The login info is invalid.\n %v", err)
	}

	if printURL {
		// print the authentication URL for the user to copy
		color.Cyan("Please copy the following URL into your browser to authenticate:")
		fmt.Printf("\n%s\n\n", authURL)
	} else {
		// define a context with 15 second timeout
		var browserCommand *exec.Cmd
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// identify command based on operating system
		switch runtime.GOOS {
		case "windows":
			browserCommand = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", authURL)
		case "darwin":
			browserCommand = exec.CommandContext(ctx, "open", authURL)
		case "linux":
			browserCommand = exec.CommandContext(ctx, "xdg-open", authURL)
		default:
			log.Println("Unsupported operating system:", runtime.GOOS)
			return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
		}

		// run the command to open the browser
		err = browserCommand.Run()
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout reached while trying to open the browser.")
		} else if err != nil {
			log.Println("Error opening browser:", err)
		}
	}

	server := &http.Server{Addr: ":" + SAMLLocalAuthPort}

	// create a timer for the callback
	var timer *time.Timer
	if printURL {
		timer = time.NewTimer(180 * time.Second)
	} else {
		timer = time.NewTimer(60 * time.Second)
	}

	// goroutine to handle timeout and token receipt
	go func() {
		select {
		case tempResult := <-tokenChan:
			// token received, stop the timer
			timer.Stop()

			// shut down the server gracefully
			err := server.Shutdown(context.Background())
			if err != nil {
				tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("error shutting down server: %w", err)}
				return
			}

			// forward the result
			tokenChan <- tempResult

		case <-timer.C:
			// timeout occurred
			log.Println("Authentication timed out after 60 seconds")

			// shut down the server
			err := server.Shutdown(context.Background())
			if err != nil {
				log.Printf("Error shutting down server after timeout: %v", err)
			}

			// send timeout error
			tokenChan <- SamlCallbackResult{
				Data: nil,
				Err:  fmt.Errorf("authentication timed out after 60 seconds"),
			}
		}
	}()

	// start the server
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("The login info is invalid.\n %v", err)
	}

	// wait for result
	samlResult := <-tokenChan

	if samlResult.Err != nil {
		return nil, samlResult.Err
	}

	return samlResult.Data, nil
}

// validateSAMLMetadata performs comprehensive validation of SAML metadata
// and returns detailed error messages to help users diagnose configuration issues.
func validateSAMLMetadata(metadata *samlTypes.EntityDescriptor) error {
	if metadata == nil {
		return fmt.Errorf("SAML metadata is nil - the metadata file may be empty or failed to parse")
	}

	if metadata.EntityID == "" {
		return fmt.Errorf("SAML metadata is missing EntityID. The metadata file may be malformed or from an incorrect source")
	}

	if metadata.IDPSSODescriptor == nil {
		return fmt.Errorf("SAML metadata is missing IDPSSODescriptor. This usually means:\n" +
			"  1. You may have downloaded Service Provider (SP) metadata instead of Identity Provider (IDP) metadata\n" +
			"  2. The metadata file is incomplete or malformed\n" +
			"  3. The URL points to the wrong endpoint\n" +
			"Please verify you're using the IDP metadata URL from your SAML identity provider (Okta, Azure AD, etc.)")
	}

	if len(metadata.IDPSSODescriptor.SingleSignOnServices) == 0 {
		return fmt.Errorf("SAML metadata IDPSSODescriptor has no SingleSignOnServices defined. The metadata may be incomplete")
	}

	if metadata.IDPSSODescriptor.SingleSignOnServices[0].Location == "" {
		return fmt.Errorf("SAML metadata SingleSignOnService Location is empty. The metadata may be malformed")
	}

	if len(metadata.IDPSSODescriptor.KeyDescriptors) == 0 {
		return fmt.Errorf("SAML metadata has no KeyDescriptors (signing certificates). The IDP metadata may be incomplete")
	}

	// Validate that at least one key descriptor has valid certificate data
	hasValidCert := false
	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {
		if len(kd.KeyInfo.X509Data.X509Certificates) > 0 {
			for _, cert := range kd.KeyInfo.X509Data.X509Certificates {
				if cert.Data != "" {
					hasValidCert = true
					break
				}
			}
		}
		if hasValidCert {
			break
		}
	}

	if !hasValidCert {
		return fmt.Errorf("SAML metadata KeyDescriptors contain no valid X509 certificates. The metadata may be malformed")
	}

	return nil
}

func AuthenticateSAML(appURL string, metadata *samlTypes.EntityDescriptor, serviceProviderIssuer string, printURL bool) (*AuthData, error) {
	// Validate parameters
	if appURL == "" {
		return nil, fmt.Errorf("appUrl (Kion URL) is required but was empty")
	}
	if serviceProviderIssuer == "" {
		return nil, fmt.Errorf("serviceProviderIssuer (SAML SP Issuer) is required but was empty. This should be configured in your Kion settings")
	}

	// Validate metadata structure
	if err := validateSAMLMetadata(metadata); err != nil {
		return nil, fmt.Errorf("SAML metadata validation failed: %w", err)
	}

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
		csrfToken, csrfCookie, err := getCSRFToken(appURL, client)
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
		url, err := url.Parse(appURL)
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("failed to parse ssl url: %w", err)}
			return
		}
		jar.SetCookies(url, csrfCookie)
		client.Jar = jar

		r, err := http.NewRequest("POST", appURL+"/api/v1/saml/callback", bytes.NewReader(b))
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
		authToken, refreshCookie, err := getAuthToken(appURL, ssoCode, csrfToken, client)
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("failed to get auth token: %w", err)}
			return
		}

		// send auto-close response before returning token
		_, err = rw.Write([]byte(AuthPage))
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

	return callExternalAuth(sp, tokenChan, printURL)
}

// AuthenticateSAMLOld is the old version of AuthenticateSAML that does not use a cookie-based exchange.
func AuthenticateSAMLOld(appURL string, metadata *samlTypes.EntityDescriptor, serviceProviderIssuer string, printURL bool) (*AuthData, error) {
	// Validate parameters
	if appURL == "" {
		return nil, fmt.Errorf("appUrl (Kion URL) is required but was empty")
	}
	if serviceProviderIssuer == "" {
		return nil, fmt.Errorf("serviceProviderIssuer (SAML SP Issuer) is required but was empty. This should be configured in your Kion settings")
	}

	// Validate metadata structure
	if err := validateSAMLMetadata(metadata); err != nil {
		return nil, fmt.Errorf("SAML metadata validation failed: %w", err)
	}

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

		r, err := http.NewRequest("POST", appURL+"/api/v1/saml/callback", bytes.NewReader(b))
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

		ssoCodeRegexp, err := regexp.Compile(`token: '(.+)',`)
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

		// send auto-close response before returning token
		_, err = rw.Write([]byte(AuthPage))
		if err != nil {
			tokenChan <- SamlCallbackResult{Data: nil, Err: fmt.Errorf("failed to send auto-close response: %w", err)}
			return
		}

		tokenChan <- SamlCallbackResult{Data: &AuthData{
			AuthToken: ssoCode,
		}, Err: nil}
	})

	return callExternalAuth(sp, tokenChan, printURL)
}

func DownloadSAMLMetadata(metadataURL string) (*samlTypes.EntityDescriptor, error) {
	if metadataURL == "" {
		return nil, fmt.Errorf("SAML metadata URL is empty. Please provide a valid URL to your Identity Provider's metadata")
	}

	res, err := http.Get(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download SAML metadata from %q: %w\nPlease verify:\n  1. The URL is correct\n  2. The URL is accessible from your network\n  3. The Identity Provider is online", metadataURL, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download SAML metadata from %q: received HTTP status %d (%s)\nPlease verify the URL is correct and points to valid IDP metadata", metadataURL, res.StatusCode, res.Status)
	}

	rawMetadata, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading SAML metadata response from %q: %w", metadataURL, err)
	}

	if len(rawMetadata) == 0 {
		return nil, fmt.Errorf("SAML metadata from %q is empty. The URL may be incorrect or the server returned no data", metadataURL)
	}

	metadata := &samlTypes.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
		return nil, fmt.Errorf("error parsing SAML metadata XML from %q: %w\nThe response may not be valid XML or may not be SAML metadata. First 200 chars of response:\n%s", metadataURL, err, truncateString(string(rawMetadata), 200))
	}

	return metadata, nil
}

func ReadSAMLMetadataFile(metadataFile string) (*samlTypes.EntityDescriptor, error) {
	if metadataFile == "" {
		return nil, fmt.Errorf("SAML metadata file path is empty. Please provide a valid file path to your Identity Provider's metadata")
	}

	rawMetadata, err := os.ReadFile(metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("SAML metadata file %q does not exist. Please verify the file path is correct", metadataFile)
		}
		return nil, fmt.Errorf("error reading SAML metadata file %q: %w", metadataFile, err)
	}

	if len(rawMetadata) == 0 {
		return nil, fmt.Errorf("SAML metadata file %q is empty. Please ensure the file contains valid IDP metadata", metadataFile)
	}

	metadata := &samlTypes.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
		return nil, fmt.Errorf("error parsing SAML metadata XML from file %q: %w\nThe file may not contain valid XML or may not be SAML metadata. First 200 chars of file:\n%s", metadataFile, err, truncateString(string(rawMetadata), 200))
	}

	return metadata, nil
}

// truncateString truncates a string to the specified length and adds "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func getCSRFToken(appURL string, client *http.Client) (string, []*http.Cookie, error) {
	csrfReq, err := http.NewRequest("GET", appURL+"/api/v2/csrf-token", nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create CSRF token request: %w", err)
	}
	csrfResp, err := client.Do(csrfReq)
	if err != nil {
		return "", nil, fmt.Errorf("failed to request CSRF token from %s/api/v2/csrf-token: %w\nPlease verify the Kion URL is correct and accessible", appURL, err)
	}
	defer csrfResp.Body.Close()

	if csrfResp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("CSRF token request failed with HTTP status %d (%s). The Kion server may be unreachable or misconfigured", csrfResp.StatusCode, csrfResp.Status)
	}

	csrfBody, err := io.ReadAll(csrfResp.Body)
	csrfCookie := csrfResp.Cookies()
	if err != nil {
		return "", nil, fmt.Errorf("failed to read CSRF token response: %w", err)
	}
	var csrfData CSRFResponse
	err = json.Unmarshal(csrfBody, &csrfData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse CSRF token response as JSON: %w\nResponse: %s", err, truncateString(string(csrfBody), 200))
	}
	if csrfData.Data == "" {
		return "", nil, fmt.Errorf("CSRF token response contained empty token. Response: %s", truncateString(string(csrfBody), 200))
	}
	return csrfData.Data, csrfCookie, nil
}

func getAuthToken(appURL string, ssoCode string, csrfToken string, client *http.Client) (string, []*http.Cookie, error) {
	authReq, err := http.NewRequest("GET", appURL+"/api/v2/login/sso-provider?code="+ssoCode, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create auth token request: %w", err)
	}
	authReq.Header.Set("X-Csrf-Token", csrfToken)

	authResp, err := client.Do(authReq)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange SSO code for auth token: %w\nThis may indicate:\n  1. Network connectivity issues to Kion\n  2. Invalid SSO code from SAML response\n  3. Kion server issues", err)
	}
	defer authResp.Body.Close()

	authBody, err := io.ReadAll(authResp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read auth token response: %w", err)
	}

	if authResp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("auth token request failed with HTTP status %d (%s)\nResponse: %s\nThis may indicate:\n  1. The SAML Service Provider Issuer doesn't match Kion's configuration\n  2. The SAML SSO integration is not properly configured in Kion\n  3. The SSO code has expired or is invalid",
			authResp.StatusCode, authResp.Status, truncateString(string(authBody), 300))
	}

	var authData SSOAuthResponse
	err = json.Unmarshal(authBody, &authData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse auth token response as JSON: %w\nResponse: %s", err, truncateString(string(authBody), 200))
	}

	if authData.Data.Access.Token == "" {
		return "", nil, fmt.Errorf("auth token response contained empty token. Response: %s", truncateString(string(authBody), 200))
	}

	return authData.Data.Access.Token, authResp.Cookies(), nil
}
