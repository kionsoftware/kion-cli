package commands

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/kionsoftware/kion-cli/lib/kion"
	samlTypes "github.com/russellhaering/gosaml2/types"
	"github.com/urfave/cli/v2"
)

// FlushCache clears the Kion CLI cache.
func (c *Cmd) FlushCache(cCtx *cli.Context) error {
	return c.cache.FlushCache()
}

// PushFavorites pushes the local favorites to a target instance of Kion.
func (c *Cmd) PushFavorites(cCtx *cli.Context) error {
	// identify target instance of Kion
	// check version of target instance of Kion
	// pull existing upstream favorites
	// deconflict with local version of favorites, id to be pushed
	// prompt user to confirm what will be pushed where
	// perform push or bail
	return nil
}

type validationContext struct {
	checkMark  string
	xMark      string
	httpClient *http.Client
	allPassed  bool
}

func newValidationContext() *validationContext {
	return &validationContext{
		checkMark: color.GreenString("✓"),
		xMark:     color.RedString("✗"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		allPassed: true,
	}
}

// checkBasicConfig validates basic SAML configuration parameters
func (c *Cmd) checkBasicConfig(ctx *validationContext) error {
	// Check 1: Kion URL is configured
	fmt.Printf("Kion URL configured ........................... ")
	if c.config.Kion.Url == "" {
		fmt.Printf("%s\n", ctx.xMark)
		color.Red("  Error: Kion URL is not configured")
		color.Yellow("  Fix: Set 'url' in ~/.kion.yml or use --url flag")
		ctx.allPassed = false
	} else {
		fmt.Printf("%s\n", ctx.checkMark)
		fmt.Printf("  URL: %s\n", c.config.Kion.Url)
	}
	fmt.Println()

	// Check 2: SAML Metadata File/URL is configured
	fmt.Printf("SAML Metadata configured ...................... ")
	if c.config.Kion.SamlMetadataFile == "" {
		fmt.Printf("%s\n", ctx.xMark)
		color.Red("  Error: SAML Metadata File/URL is not configured")
		color.Yellow("  Fix: Set 'saml_metadata_file' in ~/.kion.yml or use --saml-metadata-file flag")
		ctx.allPassed = false
	} else {
		fmt.Printf("%s\n", ctx.checkMark)
		fmt.Printf("  Source: %s\n", c.config.Kion.SamlMetadataFile)
	}
	fmt.Println()

	// Check 3: SAML Service Provider Issuer is configured
	fmt.Printf("SAML SP Issuer configured ..................... ")
	if c.config.Kion.SamlIssuer == "" {
		fmt.Printf("%s\n", ctx.xMark)
		color.Red("  Error: SAML Service Provider Issuer is not configured")
		color.Yellow("  Fix: Set 'saml_sp_issuer' in ~/.kion.yml or use --saml-sp-issuer flag")
		ctx.allPassed = false
	} else {
		fmt.Printf("%s\n", ctx.checkMark)
		fmt.Printf("  Issuer: %s\n", c.config.Kion.SamlIssuer)
	}
	fmt.Println()

	// Check 3b: SAML Service Provider Issuer format is valid
	if c.config.Kion.SamlIssuer != "" {
		fmt.Printf("SAML SP Issuer format is valid ................ ")
		issuerURL, err := url.Parse(c.config.Kion.SamlIssuer)
		isValidURL := err == nil && (issuerURL.Scheme == "http" || issuerURL.Scheme == "https") && issuerURL.Host != ""
		isValidURN := strings.HasPrefix(c.config.Kion.SamlIssuer, "urn:")

		if isValidURL || isValidURN {
			fmt.Printf("%s\n", ctx.checkMark)
			if isValidURL {
				fmt.Println("  Format: URL")
			} else {
				fmt.Println("  Format: URN")
			}
		} else {
			fmt.Printf("%s\n", ctx.xMark)
			color.Red("  Warning: SP Issuer should be a valid URL or URN")
			color.Yellow("  Common formats: https://your-kion-url.com or urn:your:issuer:id")
			color.Yellow("  Note: This may still work, but could cause issues with some IDPs")
		}
		fmt.Println()
	}

	// If basic config is missing, stop here
	if !ctx.allPassed {
		fmt.Println()
		color.Yellow("Please configure the missing parameters before continuing.")
		return fmt.Errorf("SAML configuration is incomplete")
	}

	return nil
}

// checkPortAvailability verifies port 8400 is available
func (c *Cmd) checkPortAvailability(ctx *validationContext) {
	fmt.Printf("Port 8400 is available for callback ........... ")
	listener, err := net.Listen("tcp", ":8400")
	if err != nil {
		fmt.Printf("%s\n", ctx.xMark)
		color.Red("  Error: Port 8400 is already in use")
		color.Yellow("  The SAML callback server needs port 8400 to be available")
		color.Yellow("  Please stop any process using this port and try again")
		ctx.allPassed = false
	} else {
		listener.Close()
		fmt.Printf("%s\n", ctx.checkMark)
		fmt.Println("  Port is available for SAML callback")
	}
	fmt.Println()
}

// checkKionConnectivity verifies Kion server and CSRF endpoint are accessible
func (c *Cmd) checkKionConnectivity(ctx *validationContext) bool {
	fmt.Printf("Kion server is accessible ..................... ")
	kionAccessible := false
	resp, err := ctx.httpClient.Get(c.config.Kion.Url)
	if err != nil {
		fmt.Printf("%s\n", ctx.xMark)
		color.Red("  Error: %v", err)
		ctx.allPassed = false
	} else {
		resp.Body.Close()
		if resp.StatusCode < 500 {
			fmt.Printf("%s\n", ctx.checkMark)
			fmt.Printf("  Status: HTTP %d\n", resp.StatusCode)
			kionAccessible = true
		} else {
			fmt.Printf("%s\n", ctx.xMark)
			color.Red("  Error: HTTP %d - Server error", resp.StatusCode)
			ctx.allPassed = false
		}
	}
	fmt.Println()

	return kionAccessible
}

// checkCSRFEndpoint verifies Kion CSRF endpoint is accessible
func (c *Cmd) checkCSRFEndpoint(ctx *validationContext) {
	fmt.Printf("Kion CSRF endpoint is accessible .............. ")
	csrfResp, err := ctx.httpClient.Get(c.config.Kion.Url + "/api/v2/csrf-token")
	if err != nil {
		fmt.Printf("%s\n", ctx.xMark)
		color.Red("  Error: %v", err)
		color.Yellow("  Note: This is required for SAML authentication")
		ctx.allPassed = false
	} else {
		csrfResp.Body.Close()
		if csrfResp.StatusCode == http.StatusOK {
			fmt.Printf("%s\n", ctx.checkMark)
			fmt.Printf("  Status: HTTP %d\n", csrfResp.StatusCode)
		} else {
			fmt.Printf("%s\n", ctx.xMark)
			color.Red("  Error: HTTP %d", csrfResp.StatusCode)
			ctx.allPassed = false
		}
	}
	fmt.Println()
}

// loadMetadata loads and validates SAML metadata
func (c *Cmd) loadMetadata(ctx *validationContext) (*samlTypes.EntityDescriptor, error) {
	fmt.Printf("SAML Metadata is accessible ................... ")
	var metadata *samlTypes.EntityDescriptor
	var err error
	if strings.HasPrefix(c.config.Kion.SamlMetadataFile, "http") {
		metadata, err = kion.DownloadSAMLMetadata(c.config.Kion.SamlMetadataFile)
	} else {
		metadata, err = kion.ReadSAMLMetadataFile(c.config.Kion.SamlMetadataFile)
	}

	if err != nil {
		fmt.Printf("%s\n", ctx.xMark)
		color.Red("  Error: %v", err)
		ctx.allPassed = false
		fmt.Println()
		return nil, err
	}

	fmt.Printf("%s\n", ctx.checkMark)
	return metadata, nil
}

// validateMetadataStructure validates the structure of SAML metadata
func (c *Cmd) validateMetadataStructure(ctx *validationContext, metadata *samlTypes.EntityDescriptor) bool {
	fmt.Println()
	fmt.Printf("SAML Metadata structure is valid .............. ")

	validationErrors := []string{}
	if metadata == nil {
		validationErrors = append(validationErrors, "Metadata is nil")
	} else {
		if metadata.EntityID == "" {
			validationErrors = append(validationErrors, "Missing EntityID")
		}
		if metadata.IDPSSODescriptor == nil {
			validationErrors = append(validationErrors, "Missing IDPSSODescriptor (may be SP metadata instead of IDP)")
		} else {
			if len(metadata.IDPSSODescriptor.SingleSignOnServices) == 0 {
				validationErrors = append(validationErrors, "No SingleSignOnServices defined")
			}
			if len(metadata.IDPSSODescriptor.KeyDescriptors) == 0 {
				validationErrors = append(validationErrors, "No KeyDescriptors (certificates) found")
			}
		}
	}

	if len(validationErrors) > 0 {
		fmt.Printf("%s\n", ctx.xMark)
		for _, errMsg := range validationErrors {
			color.Red("  Error: %s", errMsg)
		}
		ctx.allPassed = false
		return false
	}

	fmt.Printf("%s\n", ctx.checkMark)
	return true
}

// printMetadataDetails prints detailed information about the SAML metadata
func (c *Cmd) printMetadataDetails(metadata *samlTypes.EntityDescriptor) {
	fmt.Println()
	color.Cyan("SAML Metadata Details:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Entity ID: %s\n", metadata.EntityID)
	if metadata.IDPSSODescriptor != nil {
		if len(metadata.IDPSSODescriptor.SingleSignOnServices) > 0 {
			fmt.Printf("  SSO URL: %s\n", metadata.IDPSSODescriptor.SingleSignOnServices[0].Location)
			fmt.Printf("  SSO Binding: %s\n", metadata.IDPSSODescriptor.SingleSignOnServices[0].Binding)
		}
		fmt.Printf("  Certificates: %d key descriptor(s) found\n", len(metadata.IDPSSODescriptor.KeyDescriptors))
	}
}

// validateCertificates validates IDP certificates
func (c *Cmd) validateCertificates(ctx *validationContext, metadata *samlTypes.EntityDescriptor) {
	fmt.Println()
	fmt.Printf("IDP certificates are valid .................... ")
	if metadata.IDPSSODescriptor == nil || len(metadata.IDPSSODescriptor.KeyDescriptors) == 0 {
		return
	}

	expiredCerts := 0
	expiringSoonCerts := 0
	validCerts := 0
	certErrors := []string{}

	for idx, kd := range metadata.IDPSSODescriptor.KeyDescriptors {
		for _, xcert := range kd.KeyInfo.X509Data.X509Certificates {
			if xcert.Data == "" {
				continue
			}
			certData, err := base64.StdEncoding.DecodeString(xcert.Data)
			if err != nil {
				certErrors = append(certErrors, fmt.Sprintf("Certificate %d: Failed to decode (%v)", idx, err))
				continue
			}

			cert, err := x509.ParseCertificate(certData)
			if err != nil {
				certErrors = append(certErrors, fmt.Sprintf("Certificate %d: Failed to parse (%v)", idx, err))
				continue
			}

			now := time.Now()
			if now.After(cert.NotAfter) {
				expiredCerts++
				certErrors = append(certErrors, fmt.Sprintf("Certificate %d: EXPIRED on %s", idx, cert.NotAfter.Format("2006-01-02")))
			} else if now.Add(30 * 24 * time.Hour).After(cert.NotAfter) {
				expiringSoonCerts++
				certErrors = append(certErrors, fmt.Sprintf("Certificate %d: Expires soon on %s", idx, cert.NotAfter.Format("2006-01-02")))
			} else {
				validCerts++
			}
		}
	}

	if expiredCerts > 0 {
		fmt.Printf("%s\n", ctx.xMark)
		for _, errMsg := range certErrors {
			if strings.Contains(errMsg, "EXPIRED") {
				color.Red("  Error: %s", errMsg)
			}
		}
		ctx.allPassed = false
	} else if expiringSoonCerts > 0 {
		fmt.Printf("%s\n", ctx.checkMark)
		fmt.Printf("  Valid certificates: %d\n", validCerts)
		for _, errMsg := range certErrors {
			if strings.Contains(errMsg, "Expires soon") {
				color.Yellow("  Warning: %s", errMsg)
			}
		}
	} else if len(certErrors) > 0 {
		fmt.Printf("%s\n", ctx.xMark)
		for _, errMsg := range certErrors {
			color.Red("  Error: %s", errMsg)
		}
		ctx.allPassed = false
	} else {
		fmt.Printf("%s\n", ctx.checkMark)
		fmt.Printf("  All %d certificate(s) are valid and not expiring soon\n", validCerts)
	}
}

// checkSSOURLReachability validates that the IDP SSO URL is reachable
func (c *Cmd) checkSSOURLReachability(ctx *validationContext, metadata *samlTypes.EntityDescriptor) {
	fmt.Println()
	fmt.Printf("IDP SSO URL is reachable ...................... ")
	if metadata.IDPSSODescriptor == nil || len(metadata.IDPSSODescriptor.SingleSignOnServices) == 0 {
		return
	}

	ssoURL := metadata.IDPSSODescriptor.SingleSignOnServices[0].Location
	ssoResp, err := ctx.httpClient.Get(ssoURL)
	if err != nil {
		fmt.Printf("%s\n", ctx.xMark)
		color.Red("  Error: %v", err)
		color.Yellow("  Note: The IDP SSO endpoint may not be accessible from your network")
		ctx.allPassed = false
	} else {
		ssoResp.Body.Close()
		// Accept any response code < 500, as auth endpoints often return 302, 401, etc.
		if ssoResp.StatusCode < 500 {
			fmt.Printf("%s\n", ctx.checkMark)
			fmt.Printf("  Status: HTTP %d (IDP is responding)\n", ssoResp.StatusCode)
		} else {
			fmt.Printf("%s\n", ctx.xMark)
			color.Red("  Error: HTTP %d - IDP server error", ssoResp.StatusCode)
			ctx.allPassed = false
		}
	}
}

// ValidateSAML validates SAML configuration and connectivity.
func (c *Cmd) ValidateSAML(cCtx *cli.Context) error {
	fmt.Println("\n" + color.CyanString("SAML Configuration Validation"))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	ctx := newValidationContext()

	// Check basic configuration
	if err := c.checkBasicConfig(ctx); err != nil {
		return err
	}

	// Check port availability
	c.checkPortAvailability(ctx)

	// Check Kion connectivity
	kionAccessible := c.checkKionConnectivity(ctx)

	// Load and validate metadata
	metadata, err := c.loadMetadata(ctx)
	if err == nil {
		// Validate metadata structure
		if c.validateMetadataStructure(ctx, metadata) {
			// Print metadata details
			c.printMetadataDetails(metadata)

			// Validate certificates
			c.validateCertificates(ctx, metadata)

			// Check SSO URL reachability
			c.checkSSOURLReachability(ctx, metadata)
		}
	}
	fmt.Println()

	// Check CSRF endpoint if Kion is accessible
	if kionAccessible {
		c.checkCSRFEndpoint(ctx)
	}

	// Summary
	fmt.Println(strings.Repeat("=", 60))
	if ctx.allPassed {
		color.Green("✓ All validation checks passed!")
		fmt.Println("\nYour SAML configuration appears to be correct.")
		fmt.Println("Try running SAML authentication to complete the flow.")
		return nil
	}

	color.Red("✗ Some validation checks failed.")
	fmt.Println("\nPlease review the errors above and fix the configuration.")
	return fmt.Errorf("SAML validation failed")
}
