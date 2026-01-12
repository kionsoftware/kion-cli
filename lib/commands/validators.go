package commands

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/kionsoftware/kion-cli/lib/styles"

	samlTypes "github.com/russellhaering/gosaml2/types"
	"github.com/urfave/cli/v2"
)

// validationContext holds context for SAML validation.
type validationContext struct {
	styles     *styles.OutputStyles
	httpClient *http.Client
	allPassed  bool
}

// newValidationContext creates a new validation context.
func newValidationContext() *validationContext {
	return &validationContext{
		styles: styles.NewOutputStyles(),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		allPassed: true,
	}
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Validation Helpers                                                        //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// checkBasicConfig validates basic SAML configuration parameters
func (c *Cmd) checkBasicConfig(ctx *validationContext) error {
	// Check 1: Kion URL is configured
	if c.config.Kion.URL == "" {
		fmt.Println(ctx.styles.RenderCheck("Kion URL configured", false))
		fmt.Println(ctx.styles.RenderError("Kion URL is not configured"))
		fmt.Println(ctx.styles.RenderFix("Set 'url' in ~/.kion.yml or use --url flag"))
		ctx.allPassed = false
	} else {
		fmt.Println(ctx.styles.RenderCheck("Kion URL configured", true))
		fmt.Println(ctx.styles.RenderDetail("URL: " + c.config.Kion.URL))
	}
	fmt.Println()

	// Check 2: SAML Metadata File/URL is configured
	if c.config.Kion.SamlMetadataFile == "" {
		fmt.Println(ctx.styles.RenderCheck("SAML Metadata configured", false))
		fmt.Println(ctx.styles.RenderError("SAML Metadata File/URL is not configured"))
		fmt.Println(ctx.styles.RenderFix("Set 'saml_metadata_file' in ~/.kion.yml or use --saml-metadata-file flag"))
		ctx.allPassed = false
	} else {
		fmt.Println(ctx.styles.RenderCheck("SAML Metadata configured", true))
		fmt.Println(ctx.styles.RenderDetail("Source: " + c.config.Kion.SamlMetadataFile))
	}
	fmt.Println()

	// Check 3: SAML Service Provider Issuer is configured
	if c.config.Kion.SamlIssuer == "" {
		fmt.Println(ctx.styles.RenderCheck("SAML SP Issuer configured", false))
		fmt.Println(ctx.styles.RenderError("SAML Service Provider Issuer is not configured"))
		fmt.Println(ctx.styles.RenderFix("Set 'saml_sp_issuer' in ~/.kion.yml or use --saml-sp-issuer flag"))
		ctx.allPassed = false
	} else {
		fmt.Println(ctx.styles.RenderCheck("SAML SP Issuer configured", true))
		fmt.Println(ctx.styles.RenderDetail("Issuer: " + c.config.Kion.SamlIssuer))
	}
	fmt.Println()

	// Check 3b: SAML Service Provider Issuer format is valid
	if c.config.Kion.SamlIssuer != "" {
		issuerURL, err := url.Parse(c.config.Kion.SamlIssuer)
		isValidURL := err == nil && (issuerURL.Scheme == "http" || issuerURL.Scheme == "https") && issuerURL.Host != ""
		isValidURN := strings.HasPrefix(c.config.Kion.SamlIssuer, "urn:")

		if isValidURL || isValidURN {
			fmt.Println(ctx.styles.RenderCheck("SAML SP Issuer format is valid", true))
			if isValidURL {
				fmt.Println(ctx.styles.RenderDetail("Format: URL"))
			} else {
				fmt.Println(ctx.styles.RenderDetail("Format: URN"))
			}
		} else {
			fmt.Println(ctx.styles.RenderCheck("SAML SP Issuer format is valid", false))
			fmt.Println(ctx.styles.RenderWarning("SP Issuer should be a valid URL or URN"))
			fmt.Println(ctx.styles.RenderFix("Common formats: https://your-kion-url.com or urn:your:issuer:id"))
			fmt.Println(ctx.styles.RenderNote("This may still work, but could cause issues with some IDPs"))
		}
		fmt.Println()
	}

	// If basic config is missing, stop here
	if !ctx.allPassed {
		fmt.Println()
		fmt.Println(ctx.styles.RenderWarning("Please configure the missing parameters before continuing."))
		return fmt.Errorf("SAML configuration is incomplete")
	}

	return nil
}

// checkPortAvailability verifies port 8400 is available
func (c *Cmd) checkPortAvailability(ctx *validationContext) {
	listener, err := net.Listen("tcp", ":8400")
	if err != nil {
		fmt.Println(ctx.styles.RenderCheck("Port 8400 is available for callback", false))
		fmt.Println(ctx.styles.RenderError("Port 8400 is already in use"))
		fmt.Println(ctx.styles.RenderNote("The SAML callback server needs port 8400 to be available"))
		fmt.Println(ctx.styles.RenderFix("Please stop any process using this port and try again"))
		ctx.allPassed = false
	} else {
		listener.Close()
		fmt.Println(ctx.styles.RenderCheck("Port 8400 is available for callback", true))
		fmt.Println(ctx.styles.RenderDetail("Port is available for SAML callback"))
	}
	fmt.Println()
}

// checkKionConnectivity verifies Kion server and CSRF endpoint are accessible
func (c *Cmd) checkKionConnectivity(ctx *validationContext) bool {
	kionAccessible := false
	resp, err := ctx.httpClient.Get(c.config.Kion.URL)
	if err != nil {
		fmt.Println(ctx.styles.RenderCheck("Kion server is accessible", false))
		fmt.Println(ctx.styles.RenderError(err.Error()))
		ctx.allPassed = false
	} else {
		resp.Body.Close()
		if resp.StatusCode < 500 {
			fmt.Println(ctx.styles.RenderCheck("Kion server is accessible", true))
			fmt.Println(ctx.styles.RenderDetail(fmt.Sprintf("Status: HTTP %d", resp.StatusCode)))
			kionAccessible = true
		} else {
			fmt.Println(ctx.styles.RenderCheck("Kion server is accessible", false))
			fmt.Println(ctx.styles.RenderError(fmt.Sprintf("HTTP %d - Server error", resp.StatusCode)))
			ctx.allPassed = false
		}
	}
	fmt.Println()

	return kionAccessible
}

// checkCSRFEndpoint verifies Kion CSRF endpoint is accessible
func (c *Cmd) checkCSRFEndpoint(ctx *validationContext) {
	csrfResp, err := ctx.httpClient.Get(c.config.Kion.URL + "/api/v2/csrf-token")
	if err != nil {
		fmt.Println(ctx.styles.RenderCheck("Kion CSRF endpoint is accessible", false))
		fmt.Println(ctx.styles.RenderError(err.Error()))
		fmt.Println(ctx.styles.RenderNote("This is required for SAML authentication"))
		ctx.allPassed = false
	} else {
		csrfResp.Body.Close()
		if csrfResp.StatusCode == http.StatusOK {
			fmt.Println(ctx.styles.RenderCheck("Kion CSRF endpoint is accessible", true))
			fmt.Println(ctx.styles.RenderDetail(fmt.Sprintf("Status: HTTP %d", csrfResp.StatusCode)))
		} else {
			fmt.Println(ctx.styles.RenderCheck("Kion CSRF endpoint is accessible", false))
			fmt.Println(ctx.styles.RenderError(fmt.Sprintf("HTTP %d", csrfResp.StatusCode)))
			ctx.allPassed = false
		}
	}
	fmt.Println()
}

// loadMetadata loads and validates SAML metadata
func (c *Cmd) loadMetadata(ctx *validationContext) (*samlTypes.EntityDescriptor, error) {
	var metadata *samlTypes.EntityDescriptor
	var err error
	if strings.HasPrefix(c.config.Kion.SamlMetadataFile, "http") {
		metadata, err = kion.DownloadSAMLMetadata(c.config.Kion.SamlMetadataFile)
	} else {
		metadata, err = kion.ReadSAMLMetadataFile(c.config.Kion.SamlMetadataFile)
	}

	if err != nil {
		fmt.Println(ctx.styles.RenderCheck("SAML Metadata is accessible", false))
		fmt.Println(ctx.styles.RenderError(err.Error()))
		ctx.allPassed = false
		fmt.Println()
		return nil, err
	}

	fmt.Println(ctx.styles.RenderCheck("SAML Metadata is accessible", true))
	return metadata, nil
}

// validateMetadataStructure validates the structure of SAML metadata
func (c *Cmd) validateMetadataStructure(ctx *validationContext, metadata *samlTypes.EntityDescriptor) bool {
	fmt.Println()

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
		fmt.Println(ctx.styles.RenderCheck("SAML Metadata structure is valid", false))
		for _, errMsg := range validationErrors {
			fmt.Println(ctx.styles.RenderError(errMsg))
		}
		ctx.allPassed = false
		return false
	}

	fmt.Println(ctx.styles.RenderCheck("SAML Metadata structure is valid", true))
	return true
}

// printMetadataDetails prints detailed information about the SAML metadata
func (c *Cmd) printMetadataDetails(ctx *validationContext, metadata *samlTypes.EntityDescriptor) {
	var details strings.Builder
	details.WriteString("SAML Metadata Details\n\n")
	details.WriteString(fmt.Sprintf("Entity ID: %s\n", metadata.EntityID))
	if metadata.IDPSSODescriptor != nil {
		if len(metadata.IDPSSODescriptor.SingleSignOnServices) > 0 {
			details.WriteString(fmt.Sprintf("SSO URL: %s\n", metadata.IDPSSODescriptor.SingleSignOnServices[0].Location))
			details.WriteString(fmt.Sprintf("SSO Binding: %s\n", metadata.IDPSSODescriptor.SingleSignOnServices[0].Binding))
		}
		details.WriteString(fmt.Sprintf("Certificates: %d key descriptor(s) found", len(metadata.IDPSSODescriptor.KeyDescriptors)))
	}

	fmt.Println(ctx.styles.DetailsBox.Render(details.String()))
}

// validateCertificates validates IDP certificates
func (c *Cmd) validateCertificates(ctx *validationContext, metadata *samlTypes.EntityDescriptor) {
	fmt.Println()
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
		fmt.Println(ctx.styles.RenderCheck("IDP certificates are valid", false))
		for _, errMsg := range certErrors {
			if strings.Contains(errMsg, "EXPIRED") {
				fmt.Println(ctx.styles.RenderError(errMsg))
			}
		}
		ctx.allPassed = false
	} else if expiringSoonCerts > 0 {
		fmt.Println(ctx.styles.RenderCheck("IDP certificates are valid", true))
		fmt.Println(ctx.styles.RenderDetail(fmt.Sprintf("Valid certificates: %d", validCerts)))
		for _, errMsg := range certErrors {
			if strings.Contains(errMsg, "Expires soon") {
				fmt.Println(ctx.styles.RenderWarning(errMsg))
			}
		}
	} else if len(certErrors) > 0 {
		fmt.Println(ctx.styles.RenderCheck("IDP certificates are valid", false))
		for _, errMsg := range certErrors {
			fmt.Println(ctx.styles.RenderError(errMsg))
		}
		ctx.allPassed = false
	} else {
		fmt.Println(ctx.styles.RenderCheck("IDP certificates are valid", true))
		fmt.Println(ctx.styles.RenderDetail(fmt.Sprintf("All %d certificate(s) are valid and not expiring soon", validCerts)))
	}
}

// checkSSOURLReachability validates that the IDP SSO URL is reachable
func (c *Cmd) checkSSOURLReachability(ctx *validationContext, metadata *samlTypes.EntityDescriptor) {
	fmt.Println()
	if metadata.IDPSSODescriptor == nil || len(metadata.IDPSSODescriptor.SingleSignOnServices) == 0 {
		return
	}

	ssoURL := metadata.IDPSSODescriptor.SingleSignOnServices[0].Location
	ssoResp, err := ctx.httpClient.Get(ssoURL)
	if err != nil {
		fmt.Println(ctx.styles.RenderCheck("IDP SSO URL is reachable", false))
		fmt.Println(ctx.styles.RenderError(err.Error()))
		fmt.Println(ctx.styles.RenderNote("The IDP SSO endpoint may not be accessible from your network"))
		ctx.allPassed = false
	} else {
		ssoResp.Body.Close()
		// Accept any response code < 500, as auth endpoints often return 302, 401, etc.
		if ssoResp.StatusCode < 500 {
			fmt.Println(ctx.styles.RenderCheck("IDP SSO URL is reachable", true))
			fmt.Println(ctx.styles.RenderDetail(fmt.Sprintf("Status: HTTP %d (IDP is responding)", ssoResp.StatusCode)))
		} else {
			fmt.Println(ctx.styles.RenderCheck("IDP SSO URL is reachable", false))
			fmt.Println(ctx.styles.RenderError(fmt.Sprintf("HTTP %d - IDP server error", ssoResp.StatusCode)))
			ctx.allPassed = false
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Commands                                                                  //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// ValidateSAML validates SAML configuration and connectivity.
func (c *Cmd) ValidateSAML(cCtx *cli.Context) error {
	ctx := newValidationContext()

	// Header
	fmt.Println()
	fmt.Println(ctx.styles.RenderMainHeader("SAML Configuration Validation"))
	fmt.Println(ctx.styles.RenderSeparator())
	fmt.Println()

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
	fmt.Println(ctx.styles.RenderSeparator())
	if ctx.allPassed {
		var summary strings.Builder
		summary.WriteString("✓ All validation checks passed!\n\n")
		summary.WriteString("Your SAML configuration appears to be correct.\n")
		summary.WriteString("Try running SAML authentication to complete the flow.")

		successBox := ctx.styles.SummaryBox.BorderForeground(ctx.styles.CheckMark.GetForeground())
		fmt.Println(successBox.Render(summary.String()))

		// Print metadata details after success message
		if metadata != nil {
			c.printMetadataDetails(ctx, metadata)
		}
		return nil
	}

	var summary strings.Builder
	summary.WriteString("✗ Some validation checks failed.\n\n")
	summary.WriteString("Please review the errors above and fix the configuration.")

	failBox := ctx.styles.SummaryBox.BorderForeground(ctx.styles.XMark.GetForeground())
	fmt.Println(failBox.Render(summary.String()))
	return fmt.Errorf("SAML validation failed")
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Validators                                                                //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// ValidateCmdStak validates the flags passed to the stak command.
func (c *Cmd) ValidateCmdStak(cCtx *cli.Context) error {
	if (cCtx.String("account") != "" || cCtx.String("alias") != "") && cCtx.String("car") == "" {
		return errors.New("must specify --car parameter when using --account or --alias")
	} else if cCtx.String("car") != "" && cCtx.String("account") == "" && cCtx.String("alias") == "" {
		return errors.New("must specify --account OR --alias parameter when using --car")
	}
	return nil
}

// ValidateCmdConsole validates the flags passed to the console command.
func (c *Cmd) ValidateCmdConsole(cCtx *cli.Context) error {
	if cCtx.String("car") != "" {
		if cCtx.String("account") == "" && cCtx.String("alias") == "" {
			return errors.New("must specify --account or --alias parameter when using --car")
		}
	} else if cCtx.String("account") != "" || cCtx.String("alias") != "" {
		return errors.New("must specify --car parameter when using --account or --alias")
	}
	return nil
}

// ValidateCmdRun validates the flags passed to the run command and sets the
// favorites region as the default region if needed to ensure precedence.
func (c *Cmd) ValidateCmdRun(cCtx *cli.Context) error {
	favName := cCtx.String("favorite")
	account := cCtx.String("account")
	alias := cCtx.String("alias")
	car := cCtx.String("car")

	// Validate that either a favorite is used or both account/alias and car are provided
	hasAccountOrAlias := account != "" || alias != ""
	hasCar := car != ""
	hasFavorite := favName != ""

	if !hasFavorite && (!hasAccountOrAlias || !hasCar) {
		return errors.New("must specify either --fav OR --account and --car  OR --alias and --car parameters")
	}

	// If using account/alias + car directly, no favorite lookup needed
	if !hasFavorite {
		return nil
	}

	// Set the favorite region as the default region if a favorite is used
	_, fMap := helper.MapFavs(c.config.Favorites)
	if fMap[favName] == (structs.Favorite{}) {
		return errors.New("can't find favorite")
	}
	favorite := fMap[favName]
	if favorite.Region != "" {
		c.config.Kion.DefaultRegion = favorite.Region
	}

	return nil
}
