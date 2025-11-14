package commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

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
	fmt.Println(ctx.styles.renderMainHeader("SAML Configuration Validation"))
	fmt.Println(ctx.styles.renderSeparator())
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
	fmt.Println(ctx.styles.renderSeparator())
	if ctx.allPassed {
		var summary strings.Builder
		summary.WriteString("✓ All validation checks passed!\n\n")
		summary.WriteString("Your SAML configuration appears to be correct.\n")
		summary.WriteString("Try running SAML authentication to complete the flow.")

		successBox := ctx.styles.summaryBox.BorderForeground(ctx.styles.checkMark.GetForeground())
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

	failBox := ctx.styles.summaryBox.BorderForeground(ctx.styles.xMark.GetForeground())
	fmt.Println(failBox.Render(summary.String()))
	return fmt.Errorf("SAML validation failed")
}

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
