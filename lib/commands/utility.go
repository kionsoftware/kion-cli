package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// deleteUpstreamFavorites deletes favorites in the Kion API. It assumes you are
// passing upstream defined favorites only as we don't want to delete local
// only favorites.
func (c *Cmd) deleteUpstreamFavorites(favorites []structs.Favorite) error {
	hasErrors := false
	for _, f := range favorites {
		fmt.Printf(" removing upstream favorite %s: ", f.Name)
		_, err := kion.DeleteFavorite(c.config.Kion.URL, c.config.Kion.APIKey, f.Name)
		if err != nil {
			color.Red("x %s\n", err)
			hasErrors = true
			continue
		}
		color.Green("✓")
	}
	if hasErrors {
		return errors.New("one or more favorites failed to delete")
	}
	return nil
}

// createUpstreamFavorite creates favorites in the Kion API. It assumes you are
// passing locally defined favorites only as we convert the access type from
// cli to api format.
func (c *Cmd) createUpstreamFavorite(favorites []structs.Favorite) error {
	hasErrors := false
	for _, f := range favorites {
		fmt.Printf(" creating favorite %s: ", f.Name)
		f.AccessType = kion.ConvertAccessType(f.AccessType)
		_, _, err := kion.CreateFavorite(c.config.Kion.URL, c.config.Kion.APIKey, f)
		if err != nil {
			color.Red("x %s\n", err)
			hasErrors = true
			continue
		}
		color.Green("✓")
	}
	if hasErrors {
		return errors.New("one or more errors occurred during the creation process")
	}
	return nil
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
	// Exit if not using a compatible Kion version.
	if !cCtx.App.Metadata["useFavoritesAPI"].(bool) {
		err := errors.New("favorites API is not enabled. This requires Kion version 3.13.5, 3.14.1 or higher")
		return err
	}

	// Exit if no local favorites are defined.
	if len(c.config.Favorites) == 0 {
		color.Yellow("No local favorites found for the current profile. Nothing to push.")
		return nil
	}

	// Track errors during the push process. This will be used to determine if
	// we should delete local favorites after the push.
	var hasErrors bool

	// Authenticate before making API calls
	err := c.setAuthToken(cCtx)
	if err != nil {
		return err
	}

	// Get the combined list of favorites from the CLI config and the Kion API.
	apiFavorites, _, err := kion.GetAPIFavorites(c.config.Kion.URL, c.config.Kion.APIKey)
	if err != nil {
		fmt.Printf("Error retrieving favorites from Kion API: %v\n", err)
		return err
	}
	_, favorites, err := helper.CombineFavorites(c.config.Favorites, apiFavorites)
	if err != nil {
		fmt.Printf("Error combining favorites: %v\n", err)
		return err
	}

	// Check if there's anything to push.
	changes := len(favorites.LocalOnly) + len(favorites.ConflictsLocal) + len(favorites.UnaliasedLocal)
	if changes == 0 {
		color.Green("All local favorites are already uploaded to Kion.\n")
		return nil
	}

	// Build the prompt message.
	prompt := fmt.Sprintf("\nThe following local favorites will be pushed to Kion (%v):\n\n", c.config.Kion.URL)
	for _, f := range favorites.LocalOnly {
		prompt += fmt.Sprintf(" - %s %s\n", f.Name, color.GreenString("(new)"))
	}
	for _, f := range favorites.ConflictsLocal {
		prompt += fmt.Sprintf(" - %s %s\n", f.Name, color.RedString("(upstream conflict)"))
	}
	for _, f := range favorites.UnaliasedLocal {
		prompt += fmt.Sprintf(" - %s %s\n", f.Name, color.YellowString("(will update alias on existing favorite)"))
	}
	if len(favorites.ConflictsLocal) > 0 {
		prompt += fmt.Sprintf("%s\n", color.RedString("\nPushing local favorites with conflicts will overwrite upstream Kion favorites!"))
	}
	prompt += "\nDo you want to continue?"

	// Confirm the push.
	selection, err := helper.PromptSelect(prompt, []string{"no", "yes"})
	if selection == "no" || err != nil {
		fmt.Println("\nAborting push of favorites.")
		return err
	}
	if len(favorites.ConflictsLocal) > 0 {
		confirm, err := helper.PromptSelect(
			"\nConflicting favorites in Kion will be overwritten, are you sure you want to continue?",
			[]string{"no", "yes"},
		)
		if confirm == "no" || err != nil {
			fmt.Println("\nAborting push of favorites due to conflicts.")
			return err
		}
	}

	// Push new local-only favorites.
	err = c.createUpstreamFavorite(favorites.LocalOnly)
	if err != nil {
		hasErrors = true
	}

	// Handle conflicts by deleting and recreating.
	err = c.deleteUpstreamFavorites(favorites.ConflictsUpstream)
	if err != nil {
		hasErrors = true
	}
	err = c.createUpstreamFavorite(favorites.ConflictsLocal)
	if err != nil {
		hasErrors = true
	}

	// Handle unaliased favorites (create will overwrite / update).
	err = c.createUpstreamFavorite(favorites.UnaliasedLocal)
	if err != nil {
		hasErrors = true
	}

	// Remove local favorites after successful push.
	if !hasErrors {
		return c.DeleteLocalFavorites(cCtx)
	} else {
		return errors.New("one or more errors occurred, local favorites have not been deleted")
	}
}

func (c *Cmd) DeleteLocalFavorites(cCtx *cli.Context) error {
	confirmDelete, err := helper.PromptSelect("\nDo you want to delete the local favorites?", []string{"no", "yes"})
	if err != nil {
		color.Red("Error prompting for deletion confirmation: %v\n", err)
		return err
	}
	if confirmDelete == "yes" {

		configPath := cCtx.App.Metadata["configPath"].(string)

		// load the full config file
		var config structs.Configuration
		err := helper.LoadConfig(configPath, &config)
		if err != nil {
			color.Red("Error loading config: %v\n", err)
			return err
		}

		// if using a profile, delete favorites from that profile
		// otherwise delete favorites from the default profile
		profile := cCtx.String("profile")
		if profile == "" {
			config.Favorites = []structs.Favorite{}
		} else {
			profileConfig := config.Profiles[profile]
			profileConfig.Favorites = []structs.Favorite{}
			config.Profiles[profile] = profileConfig
		}

		// Save the updated config back to the file
		err = helper.SaveConfig(configPath, config)
		if err != nil {
			color.Red("Error saving updated config: %v\n", err)
			return err
		}
		color.Green("\nLocal favorites deleted after successful push to Kion API.\n")
	} else {
		color.Green("\nKeeping local favorites.\n")
	}

	return nil
}
