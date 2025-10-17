package helper

import (
	"fmt"
	"slices"
	"sort"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"

	"github.com/fatih/color"
)

func padName(name string) string {
	nameLen := len(name)
	padding := fmt.Sprintf("%*s", max(12-nameLen, 0), "")
	return fmt.Sprintf("%s%s", name, padding)
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Transform                                                                 //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

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
		fNames = append(fNames, fav.DescriptiveName)
		fMap[fav.DescriptiveName] = fav
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

// CombineFavorites combines local favorites with API favorites, identifying
// exact matches, unaliased matches, conflicts, and local-only favorites. It
// returns a combined slice of all favorites and a detailed comparison struct.
func CombineFavorites(localFavs []structs.Favorite, upstreamFavs []structs.Favorite) ([]structs.Favorite, *structs.FavoritesComparison, error) {

	result := structs.FavoritesComparison{}
	upstreamConflictsSeen := make(map[string]bool)

	for _, fav := range upstreamFavs {
		// Include metadata with name
		fav.DescriptiveName = fmt.Sprintf("%s %s %s",
			padName(fav.Name),
			color.GreenString("[Kion] "),
			color.New(color.Faint).Sprintf("(%s %s %s)", fav.Account, fav.CAR, fav.AccessType),
		)

		result.All = append(result.All, fav)
	}

	for _, localFav := range localFavs {
		localKey := fmt.Sprintf("%s|%s|%s|%s", localFav.Name, localFav.Account, localFav.CAR, localFav.AccessType)
		localUnaliased := fmt.Sprintf("%s|%s|%s", localFav.Account, localFav.CAR, localFav.AccessType)
		foundMatch := false

		for _, upstreamFav := range upstreamFavs {
			// upstreamFav.AccessType = kion.ConvertAccessType(upstreamFav.AccessType)
			upstreamKey := fmt.Sprintf("%s|%s|%s|%s", upstreamFav.Name, upstreamFav.Account, upstreamFav.CAR, upstreamFav.AccessType)
			upstreamUnaliased := fmt.Sprintf("%s|%s|%s", upstreamFav.Account, upstreamFav.CAR, upstreamFav.AccessType)

			// Exact match
			if upstreamKey == localKey {
				foundMatch = true
				break
			}

			// Name conflict (two with same name but different
			// account/CAR/AccessType)
			if upstreamFav.Name == localFav.Name && !upstreamFav.Unaliased {
				result.ConflictsLocal = append(result.ConflictsLocal, localFav)
				if !upstreamConflictsSeen[upstreamKey] {
					result.ConflictsUpstream = append(result.ConflictsUpstream, upstreamFav)
					upstreamConflictsSeen[upstreamKey] = true
				}
				localFav.DescriptiveName = fmt.Sprintf("%s %s %s %s",
					padName(localFav.Name),
					color.New(color.FgBlue).Sprintf("[local]"),
					color.New(color.Faint).Sprintf("(%s %s %s)", localFav.Account, localFav.CAR, localFav.AccessType),
					color.New(color.FgRed).Sprintf("conflicts w/ %s", upstreamFav.Name),
				)
				result.All = append(result.All, localFav)
				foundMatch = true
				break
			}

			// Name conflict (different name but same
			// account/CAR/AccessType)
			if upstreamUnaliased == localUnaliased && !upstreamFav.Unaliased {
				result.ConflictsLocal = append(result.ConflictsLocal, localFav)
				if !upstreamConflictsSeen[upstreamKey] {
					result.ConflictsUpstream = append(result.ConflictsUpstream, upstreamFav)
					upstreamConflictsSeen[upstreamKey] = true
				}
				localFav.DescriptiveName = fmt.Sprintf("%s %s %s %s",
					padName(localFav.Name),
					color.New(color.FgBlue).Sprintf("[local]"),
					color.New(color.Faint).Sprintf("(%s %s %s)", localFav.Account, localFav.CAR, localFav.AccessType),
					color.New(color.FgYellow).Sprintf("duplicate of %s", upstreamFav.Name),
				)
				result.All = append(result.All, localFav)
				foundMatch = true
				break
			}

			// Account + CAR + AccessType match (no name)
			if upstreamUnaliased == localUnaliased && upstreamFav.Unaliased {
				result.UnaliasedLocal = append(result.UnaliasedLocal, localFav)
				result.UnaliasedUpstream = append(result.UnaliasedUpstream, upstreamFav)
				// Remove upstreamFav from result.All
				for i, fav := range result.All {
					if fav.Account == upstreamFav.Account && fav.CAR == upstreamFav.CAR &&
						fav.AccessType == upstreamFav.AccessType && fav.Name == upstreamFav.Name {
						result.All = append(result.All[:i], result.All[i+1:]...)
						break
					}
				}
				localFav.DescriptiveName = fmt.Sprintf("%s %s %s %s",
					padName(localFav.Name),
					color.New(color.FgBlue).Sprintf("[local]"),
					color.New(color.Faint).Sprintf("(%s %s %s)", localFav.Account, localFav.CAR, localFav.AccessType),
					color.New(color.FgYellow).Sprintf("duplicate of unaliased"),
				)
				result.All = append(result.All, localFav)
				foundMatch = true
				break
			}
		}

		if !foundMatch {
			result.LocalOnly = append(result.LocalOnly, localFav)
			localFav.DescriptiveName = fmt.Sprintf("%s %s %s",
				padName(localFav.Name),
				color.New(color.FgBlue).Sprintf("[local]"),
				color.New(color.Faint).Sprintf("(%s %s %s)", localFav.Account, localFav.CAR, localFav.AccessType),
			)
			result.All = append(result.All, localFav)
		}
	}

	return result.All, &result, nil
}
