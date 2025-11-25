package commands

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"

	"github.com/charmbracelet/lipgloss"
	samlTypes "github.com/russellhaering/gosaml2/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

// validationContext holds context for SAML validation.
type validationContext struct {
	styles     *validationStyles
	httpClient *http.Client
	allPassed  bool
}

// validationStyles holds all the Lipgloss styles for SAML validation output
type validationStyles struct {
	// Status indicators
	checkMark lipgloss.Style
	xMark     lipgloss.Style

	// Text styles
	checkLabel  lipgloss.Style
	errorText   lipgloss.Style
	warningText lipgloss.Style
	infoText    lipgloss.Style
	detailText  lipgloss.Style

	// Headers and sections
	mainHeader    lipgloss.Style
	sectionHeader lipgloss.Style
	separator     lipgloss.Style

	// Boxes and containers
	detailsBox lipgloss.Style
	summaryBox lipgloss.Style

	// Layout dimensions
	terminalWidth   int
	checkLabelWidth int
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Validation Helpers                                                        //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// newValidationStyles creates a new set of validation styles
func newValidationStyles() *validationStyles {
	// Detect terminal width
	termWidth := 80 // Default fallback
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		termWidth = width
	}

	// Use terminal width - 2 for margins, but cap at 80 and minimum 50
	effectiveWidth := max(min(termWidth-2, 80), 50)

	// Calculate check label width: total width - space (1) - checkmark (1)
	checkLabelWidth := effectiveWidth - 2

	// Box width should match the separator width (checkLabelWidth + space + checkmark)
	// The Width() in lipgloss refers to content width, excluding borders and padding
	// Total width = content width + left border (1) + right border (1) + left padding (1) + right padding (1)
	// So content width = total desired width - 4
	// But we want the outer box edge to align with separator, so we need to match separator width
	boxWidth := checkLabelWidth + 4

	return &validationStyles{
		checkMark: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // Green
			Bold(true),

		xMark: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")). // Red
			Bold(true),

		checkLabel: lipgloss.NewStyle().
			Width(checkLabelWidth).
			Align(lipgloss.Left),

		errorText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")). // Red
			PaddingLeft(2),

		warningText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")). // Yellow
			PaddingLeft(2),

		infoText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")). // Blue
			PaddingLeft(2),

		detailText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")). // Gray
			PaddingLeft(2),

		mainHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")). // Cyan
			Bold(true).
			Padding(0, 1),

		sectionHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")). // Blue
			Bold(true).
			PaddingLeft(0).
			MarginTop(1),

		separator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Dark gray
			Bold(false),

		detailsBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")). // Blue
			PaddingLeft(1).
			PaddingRight(1).
			Width(boxWidth - 4). // Subtract borders (2) and padding (2)
			MarginTop(1).
			MarginBottom(1),

		summaryBox: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("10")). // Green
			PaddingLeft(1).
			PaddingRight(1).
			Width(boxWidth - 4). // Subtract borders (2) and padding (2)
			MarginTop(1).
			MarginBottom(1),

		terminalWidth:   termWidth,
		checkLabelWidth: checkLabelWidth,
	}
}

// renderCheck renders a check result with label and status
func (s *validationStyles) renderCheck(label string, passed bool) string {
	status := s.checkMark.Render("✓")
	if !passed {
		status = s.xMark.Render("✗")
	}
	return s.checkLabel.Render(label) + " " + status
}

// renderDetail renders a detail line (indented)
func (s *validationStyles) renderDetail(text string) string {
	return s.detailText.Render(text)
}

// renderError renders an error message
func (s *validationStyles) renderError(text string) string {
	return s.errorText.Render("Error: " + text)
}

// renderWarning renders a warning message
func (s *validationStyles) renderWarning(text string) string {
	return s.warningText.Render("Warning: " + text)
}

// renderFix renders a fix suggestion
func (s *validationStyles) renderFix(text string) string {
	return s.warningText.Render("Fix: " + text)
}

// renderNote renders a note
func (s *validationStyles) renderNote(text string) string {
	return s.infoText.Render("Note: " + text)
}

// renderSeparator renders a separator line that aligns with check labels
func (s *validationStyles) renderSeparator() string {
	// Separator width = checkLabelWidth + 1 space + 1 checkmark
	width := s.checkLabelWidth + 2
	line := ""
	for range width {
		line += "─"
	}
	return s.separator.Render(line)
}

// renderMainHeader renders the main header
func (s *validationStyles) renderMainHeader(text string) string {
	return s.mainHeader.Render(text)
}

// newValidationContext creates a new validation context.
func newValidationContext() *validationContext {
	return &validationContext{
		styles: newValidationStyles(),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		allPassed: true,
	}
}

// checkBasicConfig validates basic SAML configuration parameters
func (c *Cmd) checkBasicConfig(ctx *validationContext) error {
	// Check 1: Kion URL is configured
	if c.config.Kion.URL == "" {
		fmt.Println(ctx.styles.renderCheck("Kion URL configured", false))
		fmt.Println(ctx.styles.renderError("Kion URL is not configured"))
		fmt.Println(ctx.styles.renderFix("Set 'url' in ~/.kion.yml or use --url flag"))
		ctx.allPassed = false
	} else {
		fmt.Println(ctx.styles.renderCheck("Kion URL configured", true))
		fmt.Println(ctx.styles.renderDetail("URL: " + c.config.Kion.URL))
	}
	fmt.Println()

	// Check 2: SAML Metadata File/URL is configured
	if c.config.Kion.SamlMetadataFile == "" {
		fmt.Println(ctx.styles.renderCheck("SAML Metadata configured", false))
		fmt.Println(ctx.styles.renderError("SAML Metadata File/URL is not configured"))
		fmt.Println(ctx.styles.renderFix("Set 'saml_metadata_file' in ~/.kion.yml or use --saml-metadata-file flag"))
		ctx.allPassed = false
	} else {
		fmt.Println(ctx.styles.renderCheck("SAML Metadata configured", true))
		fmt.Println(ctx.styles.renderDetail("Source: " + c.config.Kion.SamlMetadataFile))
	}
	fmt.Println()

	// Check 3: SAML Service Provider Issuer is configured
	if c.config.Kion.SamlIssuer == "" {
		fmt.Println(ctx.styles.renderCheck("SAML SP Issuer configured", false))
		fmt.Println(ctx.styles.renderError("SAML Service Provider Issuer is not configured"))
		fmt.Println(ctx.styles.renderFix("Set 'saml_sp_issuer' in ~/.kion.yml or use --saml-sp-issuer flag"))
		ctx.allPassed = false
	} else {
		fmt.Println(ctx.styles.renderCheck("SAML SP Issuer configured", true))
		fmt.Println(ctx.styles.renderDetail("Issuer: " + c.config.Kion.SamlIssuer))
	}
	fmt.Println()

	// Check 3b: SAML Service Provider Issuer format is valid
	if c.config.Kion.SamlIssuer != "" {
		issuerURL, err := url.Parse(c.config.Kion.SamlIssuer)
		isValidURL := err == nil && (issuerURL.Scheme == "http" || issuerURL.Scheme == "https") && issuerURL.Host != ""
		isValidURN := strings.HasPrefix(c.config.Kion.SamlIssuer, "urn:")

		if isValidURL || isValidURN {
			fmt.Println(ctx.styles.renderCheck("SAML SP Issuer format is valid", true))
			if isValidURL {
				fmt.Println(ctx.styles.renderDetail("Format: URL"))
			} else {
				fmt.Println(ctx.styles.renderDetail("Format: URN"))
			}
		} else {
			fmt.Println(ctx.styles.renderCheck("SAML SP Issuer format is valid", false))
			fmt.Println(ctx.styles.renderWarning("SP Issuer should be a valid URL or URN"))
			fmt.Println(ctx.styles.renderFix("Common formats: https://your-kion-url.com or urn:your:issuer:id"))
			fmt.Println(ctx.styles.renderNote("This may still work, but could cause issues with some IDPs"))
		}
		fmt.Println()
	}

	// If basic config is missing, stop here
	if !ctx.allPassed {
		fmt.Println()
		fmt.Println(ctx.styles.renderWarning("Please configure the missing parameters before continuing."))
		return fmt.Errorf("SAML configuration is incomplete")
	}

	return nil
}

// checkPortAvailability verifies port 8400 is available
func (c *Cmd) checkPortAvailability(ctx *validationContext) {
	listener, err := net.Listen("tcp", ":8400")
	if err != nil {
		fmt.Println(ctx.styles.renderCheck("Port 8400 is available for callback", false))
		fmt.Println(ctx.styles.renderError("Port 8400 is already in use"))
		fmt.Println(ctx.styles.renderNote("The SAML callback server needs port 8400 to be available"))
		fmt.Println(ctx.styles.renderFix("Please stop any process using this port and try again"))
		ctx.allPassed = false
	} else {
		listener.Close()
		fmt.Println(ctx.styles.renderCheck("Port 8400 is available for callback", true))
		fmt.Println(ctx.styles.renderDetail("Port is available for SAML callback"))
	}
	fmt.Println()
}

// checkKionConnectivity verifies Kion server and CSRF endpoint are accessible
func (c *Cmd) checkKionConnectivity(ctx *validationContext) bool {
	kionAccessible := false
	resp, err := ctx.httpClient.Get(c.config.Kion.URL)
	if err != nil {
		fmt.Println(ctx.styles.renderCheck("Kion server is accessible", false))
		fmt.Println(ctx.styles.renderError(err.Error()))
		ctx.allPassed = false
	} else {
		resp.Body.Close()
		if resp.StatusCode < 500 {
			fmt.Println(ctx.styles.renderCheck("Kion server is accessible", true))
			fmt.Println(ctx.styles.renderDetail(fmt.Sprintf("Status: HTTP %d", resp.StatusCode)))
			kionAccessible = true
		} else {
			fmt.Println(ctx.styles.renderCheck("Kion server is accessible", false))
			fmt.Println(ctx.styles.renderError(fmt.Sprintf("HTTP %d - Server error", resp.StatusCode)))
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
		fmt.Println(ctx.styles.renderCheck("Kion CSRF endpoint is accessible", false))
		fmt.Println(ctx.styles.renderError(err.Error()))
		fmt.Println(ctx.styles.renderNote("This is required for SAML authentication"))
		ctx.allPassed = false
	} else {
		csrfResp.Body.Close()
		if csrfResp.StatusCode == http.StatusOK {
			fmt.Println(ctx.styles.renderCheck("Kion CSRF endpoint is accessible", true))
			fmt.Println(ctx.styles.renderDetail(fmt.Sprintf("Status: HTTP %d", csrfResp.StatusCode)))
		} else {
			fmt.Println(ctx.styles.renderCheck("Kion CSRF endpoint is accessible", false))
			fmt.Println(ctx.styles.renderError(fmt.Sprintf("HTTP %d", csrfResp.StatusCode)))
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
		fmt.Println(ctx.styles.renderCheck("SAML Metadata is accessible", false))
		fmt.Println(ctx.styles.renderError(err.Error()))
		ctx.allPassed = false
		fmt.Println()
		return nil, err
	}

	fmt.Println(ctx.styles.renderCheck("SAML Metadata is accessible", true))
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
		fmt.Println(ctx.styles.renderCheck("SAML Metadata structure is valid", false))
		for _, errMsg := range validationErrors {
			fmt.Println(ctx.styles.renderError(errMsg))
		}
		ctx.allPassed = false
		return false
	}

	fmt.Println(ctx.styles.renderCheck("SAML Metadata structure is valid", true))
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

	fmt.Println(ctx.styles.detailsBox.Render(details.String()))
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
		fmt.Println(ctx.styles.renderCheck("IDP certificates are valid", false))
		for _, errMsg := range certErrors {
			if strings.Contains(errMsg, "EXPIRED") {
				fmt.Println(ctx.styles.renderError(errMsg))
			}
		}
		ctx.allPassed = false
	} else if expiringSoonCerts > 0 {
		fmt.Println(ctx.styles.renderCheck("IDP certificates are valid", true))
		fmt.Println(ctx.styles.renderDetail(fmt.Sprintf("Valid certificates: %d", validCerts)))
		for _, errMsg := range certErrors {
			if strings.Contains(errMsg, "Expires soon") {
				fmt.Println(ctx.styles.renderWarning(errMsg))
			}
		}
	} else if len(certErrors) > 0 {
		fmt.Println(ctx.styles.renderCheck("IDP certificates are valid", false))
		for _, errMsg := range certErrors {
			fmt.Println(ctx.styles.renderError(errMsg))
		}
		ctx.allPassed = false
	} else {
		fmt.Println(ctx.styles.renderCheck("IDP certificates are valid", true))
		fmt.Println(ctx.styles.renderDetail(fmt.Sprintf("All %d certificate(s) are valid and not expiring soon", validCerts)))
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
		fmt.Println(ctx.styles.renderCheck("IDP SSO URL is reachable", false))
		fmt.Println(ctx.styles.renderError(err.Error()))
		fmt.Println(ctx.styles.renderNote("The IDP SSO endpoint may not be accessible from your network"))
		ctx.allPassed = false
	} else {
		ssoResp.Body.Close()
		// Accept any response code < 500, as auth endpoints often return 302, 401, etc.
		if ssoResp.StatusCode < 500 {
			fmt.Println(ctx.styles.renderCheck("IDP SSO URL is reachable", true))
			fmt.Println(ctx.styles.renderDetail(fmt.Sprintf("Status: HTTP %d (IDP is responding)", ssoResp.StatusCode)))
		} else {
			fmt.Println(ctx.styles.renderCheck("IDP SSO URL is reachable", false))
			fmt.Println(ctx.styles.renderError(fmt.Sprintf("HTTP %d - IDP server error", ssoResp.StatusCode)))
			ctx.allPassed = false
		}
	}
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
	// Validate that either a favorite is used or both account/alias and car are provided
	if cCtx.String("favorite") == "" && ((cCtx.String("account") == "" && cCtx.String("alias") == "") || cCtx.String("car") == "") {
		return errors.New("must specify either --fav OR --account and --car  OR --alias and --car parameters")
	}

	// Set the favorite region as the default region if a favorite is used
	favName := cCtx.String("favorite")
	_, fMap := helper.MapFavs(c.config.Favorites)
	var fav string
	if fMap[favName] != (structs.Favorite{}) {
		fav = favName
	} else {
		return errors.New("can't find favorite")
	}
	favorite := fMap[fav]
	if favorite.Region != "" {
		c.config.Kion.DefaultRegion = favorite.Region
	}

	return nil
}
