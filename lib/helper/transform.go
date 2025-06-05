package helper

import (
	"fmt"
	"slices"
	"sort"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Transform                                                                 //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// FavoritesComparison holds the results of comparing local favorites with API
// favorites. It includes all favorites, exact matches, non-matches, conflicts,
// and local-only favorites. It's returned by the CombineFavorites function.
type FavoritesComparison struct {
	All       []structs.Favorite // Combined local + API, deduplicated and deconflicted
	Exact     []structs.Favorite // Exact matches (local + API)
	APIOnly   []structs.Favorite // API-only favorites
	Conflicts []structs.Favorite // Name conflicts (same name, different settings)
	LocalOnly []structs.Favorite // Local-only favorites (not matched in API)
}

// MapProjects transforms a slice of Projects into a slice of their names and a
// map indexed by their names.
func MapProjects(projects []kion.Project) ([]string, map[string]kion.Project) {
	var pNames []string
	pMap := make(map[string]kion.Project)
	for _, project := range projects {
		name := fmt.Sprintf("%v (%v)", project.Name, project.ID)
		pNames = append(pNames, name)
		pMap[name] = project
	}
	sort.Strings(pNames)

	return pNames, pMap
}

// MapAccounts transforms a slice of Accounts into a slice of their names and a
// map indexed by their names.
func MapAccounts(accounts []kion.Account) ([]string, map[string]kion.Account) {
	var aNames []string
	aMap := make(map[string]kion.Account)
	for _, account := range accounts {
		var name string
		if account.Alias != "" {
			name = fmt.Sprintf("%v [%v] (%v)", account.Name, account.Alias, account.Number)
		} else {
			name = fmt.Sprintf("%v (%v)", account.Name, account.Number)
		}
		aNames = append(aNames, name)
		aMap[name] = account
	}
	sort.Strings(aNames)

	return aNames, aMap
}

// MapAccountsFromCARS transforms a slice of CARs into a slice of account names
// and a map of account numbers indexed by their names. If a project ID is
// passed it will only return accounts in the given project. Note that some
// versions of Kion will not populate account metadata in CAR objects so use
// carefully (see useUpdatedCloudAccessRoleAPI bool).
func MapAccountsFromCARS(cars []kion.CAR, pid uint) ([]string, map[string]string) {
	var aNames []string
	aMap := make(map[string]string)
	for _, car := range cars {
		if pid == 0 || car.ProjectID == pid {
			var name string
			if car.AccountAlias != "" {
				name = fmt.Sprintf("%v [%v] (%v)", car.AccountName, car.AccountAlias, car.AccountNumber)
			} else {
				name = fmt.Sprintf("%v (%v)", car.AccountName, car.AccountNumber)
			}
			if slices.Contains(aNames, name) {
				continue
			}
			aNames = append(aNames, name)
			aMap[name] = car.AccountNumber
		}
	}
	sort.Strings(aNames)

	return aNames, aMap
}

// MapCAR transforms a slice of CARs into a slice of their names and a map
// indexed by their names.
func MapCAR(cars []kion.CAR) ([]string, map[string]kion.CAR) {
	var cNames []string
	cMap := make(map[string]kion.CAR)
	for _, car := range cars {
		name := fmt.Sprintf("%v (%v)", car.Name, car.ID)
		cNames = append(cNames, name)
		cMap[name] = car
	}
	sort.Strings(cNames)

	return cNames, cMap
}

// MapIDMSs transforms a slice of IDMSs into a slice of their names and a map
// indexed by their names.
func MapIDMSs(idmss []kion.IDMS) ([]string, map[string]kion.IDMS) {
	var iNames []string
	iMap := make(map[string]kion.IDMS)
	for _, idms := range idmss {
		iNames = append(iNames, idms.Name)
		iMap[idms.Name] = idms
	}
	sort.Strings(iNames)

	return iNames, iMap
}

// MapFavs transforms a slice of Favorites into a slice of their names and a
// map indexed by their names.
func MapFavs(favs []structs.Favorite) ([]string, map[string]structs.Favorite) {
	var fNames []string
	fMap := make(map[string]structs.Favorite)
	for _, fav := range favs {
		fNames = append(fNames, fav.Name)
		fMap[fav.Name] = fav
	}
	sort.Strings(fNames)

	return fNames, fMap
}

// FindCARByName returns a CAR identified by its name.
func FindCARByName(cars []kion.CAR, carName string) (*kion.CAR, error) {
	for _, c := range cars {
		if c.Name == carName {
			return &c, nil
		}
	}
	return &kion.CAR{}, fmt.Errorf("cannot find cloud access role with name %v", carName)
}

// CombineFavorites combines local favorites with API favorites, ensuring that
// local favorites are prioritized and that there are no duplicates or name conflicts.
// It returns a slice of Favorites that contains all unique favorites, with local
// favorites appearing first. If an API favorite has no name, it attempts to
// generate a name based on the account name, car, access type, and region.
func CombineFavorites(localFavs []structs.Favorite, apiFavs []structs.Favorite, defaultRegion string) (FavoritesComparison, error) {

	result := FavoritesComparison{}

	// Track all exact matches by a unique composite key
	exactMap := make(map[string]bool)

	// Ensure all local favorites have a region set
	for i := range localFavs {
		if localFavs[i].Region == "" {
			if defaultRegion != "" {
				localFavs[i].Region = defaultRegion
			} else {
				localFavs[i].Region = "us-east-1"
			}
		}
	}

	// Start with localFavs in the final set
	result.All = append(result.All, localFavs...)

	for _, apiFav := range apiFavs {

		if apiFav.Region == "" {
			if defaultRegion != "" {
				apiFav.Region = defaultRegion
			} else {
				apiFav.Region = "us-east-1"
			}
		}
		apiKey := fmt.Sprintf("%s|%s|%s|%s|%s", apiFav.Name, apiFav.Account, apiFav.CAR, apiFav.AccessType, apiFav.Region)
		apiFav.AccessType = kion.NormalizeAccessType(apiFav.AccessType)
		foundMatch := false

		for _, localFav := range localFavs {

			localKey := fmt.Sprintf("%s|%s|%s|%s|%s", localFav.Name, localFav.Account, localFav.CAR, localFav.AccessType, localFav.Region)

			// Exact match
			if apiKey == localKey {
				foundMatch = true
				exactMap[localKey] = true
				result.Exact = append(result.Exact, localFav)
				break
			}

			// Name conflict
			if apiFav.Name == localFav.Name {
				apiFav.Name = fmt.Sprintf("%s (conflict)", apiFav.Name)
				result.All = append(result.All, apiFav)
				result.Conflicts = append(result.Conflicts, localFav)
				foundMatch = true
				break
			}
		}

		if !foundMatch {
			result.All = append(result.All, apiFav)
			result.APIOnly = append(result.APIOnly, apiFav)
		}
	}

	// Determine which localFavs were not part of exact matches
	for _, localFav := range localFavs {
		key := fmt.Sprintf("%s|%s|%s|%s|%s", localFav.Name, localFav.Account, localFav.CAR, localFav.AccessType, localFav.Region)
		if !exactMap[key] {
			result.LocalOnly = append(result.LocalOnly, localFav)
		}
	}

	return result, nil
}
