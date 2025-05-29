package kion

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kionsoftware/kion-cli/lib/structs"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Favorites                                                                 //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// FavoritesResponse maps to the Kion API response.
type FavoritesResponse struct {
	Status    int                `json:"status"`
	Favorites []structs.Favorite `json:"data"`
}

// GetAPIFavorites returns a list of a user's Favorites associated with a given Kion from the API
func GetAPIFavorites(host string, token string) ([]structs.Favorite, int, error) {

	url := fmt.Sprintf("%v/api/v3/user-cloud-access-role-alias", host)
	query := map[string]string{}
	var data interface{}
	resp, statusCode, err := runQuery("GET", url, token, query, data)
	if err != nil {
		return nil, statusCode, err
	}

	// unmarshal response body
	favResp := FavoritesResponse{}
	jsonErr := json.Unmarshal(resp, &favResp)
	if jsonErr != nil {
		return nil, 0, err
	}

	var apiFavorites []structs.Favorite
	for _, apiFav := range favResp.Favorites {
		// normalize the access type to match what the CLI uses
		if apiFav.AccessType == "console_access" {
			apiFav.AccessType = "web"
		} else if apiFav.AccessType == "short_term_key_access" {
			apiFav.AccessType = "cli"
		}

		// if the upstream favorite has no name (alias), attempt to look up the account name
		// and create an alias out of the account name, car, access type, and region
		if apiFav.Name == "" {
			account, status, err := GetAccount(host, token, apiFav.Account)
			if err != nil {
				fmt.Printf("failed looking up account name for account number %s: %v\n", apiFav.Account, err)
				apiFav.Name = "unset"
			}

			if account != nil && status == 200 {
				newName := fmt.Sprintf("%s_%s_%s_%s", strings.Replace(account.Name, " ", "", -1), apiFav.CAR, apiFav.AccessType, apiFav.Region)
				if len(newName) > 50 {
					newName = fmt.Sprintf("%s...", newName[:47])
				}
				apiFav.Name = newName
			} else {
				apiFav.Name = "unset"
			}
		}
		apiFavorites = append(apiFavorites, apiFav)
	}

	return apiFavorites, favResp.Status, nil
}

func CreateFavorite(host string, token string, favorite structs.Favorite) (structs.Favorite, int, error) {
	url := fmt.Sprintf("%v/api/v3/user-cloud-access-role-alias", host)
	query := map[string]string{}
	var data interface{}
	data = map[string]string{
		"alias_name":             favorite.Name,
		"account_number":         favorite.Account,
		"cloud_access_role_name": favorite.CAR,
		"access_type":            favorite.AccessType,
	}
	resp, statusCode, err := runQuery("POST", url, token, query, data)
	if err != nil {
		return structs.Favorite{}, statusCode, err
	}

	// unmarshal response body
	var createdFav structs.Favorite
	jsonErr := json.Unmarshal(resp, &createdFav)
	if jsonErr != nil {
		return structs.Favorite{}, 0, jsonErr
	}

	return createdFav, statusCode, nil
}

func DeleteFavorite(host string, token string, favoriteName string) (int, error) {
	url := fmt.Sprintf("%v/api/v3/user-cloud-access-role-alias", host)
	query := map[string]string{}
	var data interface{}
	data = map[string]string{"alias_name": favoriteName}
	resp, statusCode, err := runQuery("DELETE", url, token, query, data)
	if err != nil {
		return statusCode, err
	}

	// check if the response is successful
	if statusCode != 200 {
		return statusCode, fmt.Errorf("failed to delete favorite with name %s: %s", favoriteName, string(resp))
	}

	return statusCode, nil
}
