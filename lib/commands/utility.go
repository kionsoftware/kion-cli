package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
)

// FlushCache clears the Kion CLI cache.
func (c *Cmd) FlushCache(cCtx *cli.Context) error {
	return c.cache.FlushCache()
}

// PushFavorites pushes the local favorites to a target instance of Kion.
func (c *Cmd) PushFavorites(cCtx *cli.Context) error {

	if cCtx.App.Metadata["useFavoritesAPI"].(bool) {

		if len(c.config.Favorites) == 0 {
			color.Yellow("No local favorites found for the current profile. Nothing to push.")
			return nil
		}

		// track errors during the push process
		// this will be used to determine if we should delete local favorites after the push
		var errors bool

		// get the combined list of favorites from the CLI config and the Kion API
		apiFavorites, _, err := kion.GetAPIFavorites(c.config.Kion.Url, c.config.Kion.ApiKey)
		if err != nil {
			fmt.Printf("Error retrieving favorites from API: %v\n", err)
			return err
		}
		result, err := helper.CombineFavorites(c.config.Favorites, apiFavorites, c.config.Kion.DefaultRegion)
		if err != nil {
			fmt.Printf("Error combining favorites: %v\n", err)
			return err
		}

		// check if there's anything to push
		noChanges := len(result.LocalOnly) == 0 && len(result.Conflicts) == 0
		inSync := len(c.config.Favorites) == len(apiFavorites) || len(c.config.Favorites) == len(result.Exact)
		if noChanges {
			if inSync {
				color.Green("All local favorites are already in sync with Kion.\n")
				return c.DeleteLocalFavorites(cCtx)
			} else {
				// No changes, but counts don't match exactly — do nothing
				color.Yellow("No new or conflicting favorites to push, but local and API counts differ. Manual review may be needed.\n")
				return nil
			}
		}

		if len(result.LocalOnly) > 0 {
			fmt.Printf("\nLocal favorites to push to API:\n")
			for _, f := range result.LocalOnly {
				fmt.Printf(" - %s\n", f.Name)
			}
		}

		if len(result.Conflicts) > 0 {
			fmt.Printf("\nName conflicts with API favorites:\n")
			for _, f := range result.Conflicts {
				fmt.Printf(" - %s\n", f.Name)
			}
			color.Red("\nThese are favorites that exist in both the CLI config and the API with the same name, but have different settings.")
			color.Red("Pushing these will overwrite the API favorites with the local settings.\n")
		}

		selection, _ := helper.PromptSelect("\nDo you want to push your local favorites to the Kion API?", []string{"no", "yes"})
		if selection == "no" {
			fmt.Println("\nAborting push of favorites.")
			return nil
		}

		if len(result.Conflicts) > 0 {
			confirm, _ := helper.PromptSelect("You have some name conflicts with API favorites. Are you sure you want to continue pushing your local favorites?", []string{"no", "yes"})
			if confirm == "no" {
				fmt.Println("\nAborting push of favorites due to conflicts.")
				return nil
			}
		}

		for _, f := range result.LocalOnly {
			if len(f.Name) > 50 {
				color.Yellow("Trimming favorite %s because its name exceeds 50 characters.\n", f.Name)
				f.Name = f.Name[:50]
			}

			f.AccessType = kion.ConvertAccessType(f.AccessType)

			_, status, err := kion.CreateFavorite(c.config.Kion.Url, c.config.Kion.ApiKey, f)
			if err != nil {
				color.Red("Error creating favorite %s: %v\n", f.Name, err)
				errors = true
				continue
			}
			if status == 201 || status == 200 {
				color.Green("Successfully pushed %s favorite to Kion\n", f.Name)
			} else {
				color.Red("Failed to push favorite %s, status code: %d\n", f.Name, status)
				errors = true
			}
		}

		for _, f := range result.Conflicts {

			f.AccessType = kion.ConvertAccessType(f.AccessType)

			_, err := kion.DeleteFavorite(c.config.Kion.Url, c.config.Kion.ApiKey, f.Name)
			if err != nil {
				color.Red("Error deleting favorite %s: %v\n", f.Name, err)
				errors = true
				continue
			}
			color.Green("Successfully deleted conflicting favorite: %s\n", f.Name)

			_, _, err = kion.CreateFavorite(c.config.Kion.Url, c.config.Kion.ApiKey, f)
			if err != nil {
				color.Red("Error creating favorite %s: %v", f.Name, err)
				errors = true
				continue
			}
			color.Green("Successfully created favorite: %s", f.Name)
		}

		// send to the DeleteLocalFavorites function to remove local favorites after successful push
		if !errors {
			return c.DeleteLocalFavorites(cCtx)
		}
	} else {
		color.Yellow("Favorites API is not enabled. This requires Kion version 3.13.0 or higher.")
	}
	return nil
}

func (c *Cmd) DeleteLocalFavorites(cCtx *cli.Context) error {
	confirmDelete, err := helper.PromptSelect("Do you want to delete the local favorites?", []string{"no", "yes"})
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
			profile = "default"
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
