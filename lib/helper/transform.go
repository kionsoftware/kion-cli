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
// passed it will only return accounts in the given project.
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
