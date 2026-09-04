package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
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
		token, err := c.freshAPIKey()
		if err != nil {
			color.Red("x %s\n", err)
			hasErrors = true
			continue
		}
		_, err = kion.DeleteFavorite(c.config.Kion.URL, token, f.Name)
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
		token, err := c.freshAPIKey()
		if err != nil {
			color.Red("x %s\n", err)
			hasErrors = true
			continue
		}
		_, _, err = kion.CreateFavorite(c.config.Kion.URL, token, f)
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

// FlushCache clears the Kion CLI cache.
func (c *Cmd) FlushCache(cCtx *cli.Context) error {
	return c.cache.FlushCache()
}

// AuthStatus prints the state of the cached session — access/refresh token
// presence, expiry, and (with --force-refresh) exercises the refresh flow
// against the Kion API so you can verify SAML/UNPW refresh end-to-end.
func (c *Cmd) AuthStatus(cCtx *cli.Context) error {
	timeFormat := "2006-01-02T15:04:05-0700"

	session, found, err := c.cache.GetSession()
	if err != nil {
		return fmt.Errorf("failed to read session cache: %w", err)
	}
	if !found {
		color.Yellow("No cached session. Run any auth'd command to create one.")
		return nil
	}

	fmt.Printf("Kion URL:    %s\n", c.config.Kion.URL)
	if session.UserName != "" {
		fmt.Printf("Username:    %s\n", session.UserName)
	}
	if session.IDMSID != 0 {
		fmt.Printf("IDMS ID:     %d\n", session.IDMSID)
	}

	// access token
	fmt.Println()
	color.Cyan("Access token")
	if session.Access.Token == "" {
		color.Red("  (none)")
	} else {
		exp, err := time.Parse(timeFormat, session.Access.Expiry)
		if err != nil {
			color.Red("  unparseable expiry %q: %v", session.Access.Expiry, err)
		} else {
			fmt.Printf("  expiry:     %s\n", exp.Local().Format(time.RFC1123))
			remaining := time.Until(exp).Round(time.Second)
			if remaining > 0 {
				color.Green("  remaining:  %s (valid)", remaining)
			} else {
				color.Yellow("  remaining:  expired %s ago", (-remaining).Round(time.Second))
			}
		}
	}

	// refresh token
	fmt.Println()
	color.Cyan("Refresh token")
	if session.Refresh.Token == "" {
		color.Yellow("  (not cached — this session can't be refreshed)")
		if session.Access.Token != "" {
			fmt.Println("  Likely cause: SAML on a Kion that didn't return ct_auth, or")
			fmt.Println("  the session was cached by an older version of this CLI.")
		}
	} else {
		exp, err := time.Parse(timeFormat, session.Refresh.Expiry)
		if err != nil {
			color.Red("  unparseable expiry %q: %v", session.Refresh.Expiry, err)
		} else {
			fmt.Printf("  expiry:     %s\n", exp.Local().Format(time.RFC1123))
			remaining := time.Until(exp).Round(time.Second)
			if remaining > 0 {
				color.Green("  remaining:  %s (valid)", remaining)
			} else {
				color.Yellow("  remaining:  expired %s ago", (-remaining).Round(time.Second))
			}
		}
	}

	// optional: actually exercise the refresh endpoint
	if cCtx.Bool("force-refresh") {
		fmt.Println()
		color.Cyan("Force refresh")
		if session.Refresh.Token == "" {
			color.Red("  No refresh token to exchange — skipping.")
			return nil
		}
		refreshed, err := kion.RefreshSession(c.config.Kion.URL, session.Refresh.Token)
		if err != nil {
			color.Red("  Refresh failed: %v", err)
			return err
		}
		if refreshed.Access.Token == "" {
			color.Red("  Refresh succeeded but no access token in response.")
			return errors.New("empty access token from refresh endpoint")
		}
		// RefreshSession normalizes to timeFormat already; surface a parse
		// error if for some reason normalization didn't take.
		newExp, err := time.Parse(timeFormat, refreshed.Access.Expiry)
		if err != nil {
			color.Red("  Refresh OK, but unparseable expiry %q: %v", refreshed.Access.Expiry, err)
			return nil
		}
		color.Green("  Refresh OK. New access token expiry: %s (%s from now)",
			newExp.Local().Format(time.RFC1123),
			time.Until(newExp).Round(time.Second),
		)
		fmt.Println("  (cache not updated — this is a dry run)")
	}

	return nil
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
	token, err := c.freshAPIKey()
	if err != nil {
		return err
	}
	apiFavorites, _, err := kion.GetAPIFavorites(c.config.Kion.URL, token)
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
	selection, err := helper.PromptSelect(prompt, "", []string{"no", "yes"})
	if selection == "no" || err != nil {
		fmt.Println("\nAborting push of favorites.")
		return err
	}
	if len(favorites.ConflictsLocal) > 0 {
		confirm, err := helper.PromptSelect(
			"Conflicting favorites in Kion will be overwritten, are you sure you want to continue?",
			"",
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

// DeleteLocalFavorites prompts for confirmation and deletes local favorites.
func (c *Cmd) DeleteLocalFavorites(cCtx *cli.Context) error {
	confirmDelete, err := helper.PromptSelect("Do you want to delete the local favorites?", "", []string{"no", "yes"})
	if err != nil {
		color.Red("Error prompting for deletion confirmation: %v\n", err)
		return err
	}
	if confirmDelete == "yes" {

		configPath := cCtx.App.Metadata["configPath"].(string)

		// if using a profile, delete favorites from that profile
		// otherwise delete favorites from the default profile
		profile := cCtx.String("profile")
		if profile == "" {
			err = helper.DeleteConfigField(configPath, []string{"favorites"})
			if err != nil {
				return fmt.Errorf("error deleting favorites from default profile: %v", err)
			}
		} else {
			err = helper.DeleteConfigField(configPath, []string{"profiles", profile, "favorites"})
			if err != nil {
				return fmt.Errorf("error deleting favorites from profile %s: %v", profile, err)
			}
		}
		color.Green("\nLocal favorites deleted after successful push to Kion API.\n")
	} else {
		color.Green("\nKeeping local favorites.\n")
	}

	return nil
}

// RotateAPIKey rotates the Kion App API Key for the default or specified profile in the configuration file.
func (c *Cmd) RotateAPIKey(cCtx *cli.Context) error {
	configPath := cCtx.App.Metadata["configPath"].(string)

	var config structs.Configuration
	err := helper.LoadConfigStruct(configPath, &config)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	profile := cCtx.String("profile")
	label := "default profile"
	url := config.Kion.URL
	storedKey := config.Kion.APIKey
	fieldPath := []string{"kion", "api_key"}

	if profile != "" {
		profileStruct, found := config.Profiles[profile]
		if !found {
			return fmt.Errorf("profile not found: %s", profile)
		}
		label = fmt.Sprintf("%s profile", profile)
		url = profileStruct.Kion.URL
		storedKey = profileStruct.Kion.APIKey
		fieldPath = []string{"profiles", profile, "kion", "api_key"}
	}

	if storedKey != c.config.Kion.APIKey {
		fmt.Printf("Active key doesn't match the %s's stored key (likely --token or KION_API_KEY).\n", label)
		fmt.Println("Rotating it and printing the new key — config file will not be updated.")

		newAPIKey, err := kion.RotateAPIKey(url, c.config.Kion.APIKey)
		if err != nil {
			return fmt.Errorf("error rotating Kion App API Key: %v", err)
		}
		return printKeyFallback(newAPIKey)
	}

	fmt.Printf("Rotating Kion App API Key (%s)\n", label)
	newAPIKey, err := kion.RotateAPIKey(url, storedKey)
	if err != nil {
		return fmt.Errorf("error rotating Kion App API Key (%s): %v", label, err)
	}

	err = helper.UpdateConfigField(configPath, fieldPath, newAPIKey)
	if err != nil {
		color.Red("Error updating Kion App API Key in configuration file (%s): %v", label, err)
		return printKeyFallback(newAPIKey)
	}

	color.Green("Kion App API Key rotated successfully.\n")
	return nil
}

func printKeyFallback(s string) error {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("non-interactive terminal detected, unable to display new Kion App API Key. Please manually regenerate a key and update your configuration")
	}
	timeout := 10 * time.Second
	color.Green("\nDo you wish to print the key? [y/N] (auto-declines in %s): ", timeout)
	answers := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			answers <- scanner.Text()
		}
		close(answers)
	}()

	select {
	case answer, ok := <-answers:
		if ok && strings.EqualFold(strings.TrimSpace(answer), "y") {
			fmt.Println(s)
			color.Yellow("Key rotated but not saved — copy it from above before it's gone.\n")
		}
	case <-time.After(timeout):
		fmt.Println("\nNo response, not printing the key.")
	}
	return nil
}
