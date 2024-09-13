package helper

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kionsoftware/kion-cli/lib/kion"
	"github.com/kionsoftware/kion-cli/lib/structs"
)

func TestMapProjects(t *testing.T) {
	tests := []struct {
		name     string
		projects []kion.Project
		wantOne  []string
		wantTwo  map[string]kion.Project
	}{
		{
			"Basic",
			kionTestProjects,
			[]string{
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[4], kionTestProjects[4].ID),
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[3], kionTestProjects[3].ID),
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[0], kionTestProjects[0].ID),
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[5], kionTestProjects[5].ID),
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[2], kionTestProjects[2].ID),
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[1], kionTestProjects[1].ID),
			},
			map[string]kion.Project{
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[0], kionTestProjects[0].ID): kionTestProjects[0],
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[1], kionTestProjects[1].ID): kionTestProjects[1],
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[2], kionTestProjects[2].ID): kionTestProjects[2],
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[3], kionTestProjects[3].ID): kionTestProjects[3],
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[4], kionTestProjects[4].ID): kionTestProjects[4],
				fmt.Sprintf("%v (%v)", kionTestProjectsNames[5], kionTestProjects[5].ID): kionTestProjects[5],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			one, two := MapProjects(test.projects)
			if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", one, two, test.wantOne, test.wantTwo)
			}
		})
	}
}

func TestMapAccounts(t *testing.T) {
	tests := []struct {
		name     string
		accounts []kion.Account
		wantOne  []string
		wantTwo  map[string]kion.Account
	}{
		{
			"Basic",
			kionTestAccounts,
			[]string{
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[4], kionTestAccounts[4].Alias, kionTestAccounts[4].Number),
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[3], kionTestAccounts[3].Alias, kionTestAccounts[3].Number),
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[0], kionTestAccounts[0].Alias, kionTestAccounts[0].Number),
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[5], kionTestAccounts[5].Alias, kionTestAccounts[5].Number),
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[2], kionTestAccounts[2].Alias, kionTestAccounts[2].Number),
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[1], kionTestAccounts[1].Alias, kionTestAccounts[1].Number),
			},
			map[string]kion.Account{
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[0], kionTestAccounts[0].Alias, kionTestAccounts[0].Number): kionTestAccounts[0],
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[1], kionTestAccounts[1].Alias, kionTestAccounts[1].Number): kionTestAccounts[1],
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[2], kionTestAccounts[2].Alias, kionTestAccounts[2].Number): kionTestAccounts[2],
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[3], kionTestAccounts[3].Alias, kionTestAccounts[3].Number): kionTestAccounts[3],
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[4], kionTestAccounts[4].Alias, kionTestAccounts[4].Number): kionTestAccounts[4],
				fmt.Sprintf("%v [%v] (%v)", kionTestAccountsNames[5], kionTestAccounts[5].Alias, kionTestAccounts[5].Number): kionTestAccounts[5],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			one, two := MapAccounts(test.accounts)
			if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", one, two, test.wantOne, test.wantTwo)
			}
		})
	}
}
func TestMapCAR(t *testing.T) {
	tests := []struct {
		name    string
		cars    []kion.CAR
		wantOne []string
		wantTwo map[string]kion.CAR
	}{
		{
			"Basic",
			kionTestCARs,
			[]string{
				fmt.Sprintf("%v (%v)", kionTestCARsNames[4], kionTestCARs[4].ID),
				fmt.Sprintf("%v (%v)", kionTestCARsNames[3], kionTestCARs[3].ID),
				fmt.Sprintf("%v (%v)", kionTestCARsNames[0], kionTestCARs[0].ID),
				fmt.Sprintf("%v (%v)", kionTestCARsNames[5], kionTestCARs[5].ID),
				fmt.Sprintf("%v (%v)", kionTestCARsNames[2], kionTestCARs[2].ID),
				fmt.Sprintf("%v (%v)", kionTestCARsNames[1], kionTestCARs[1].ID),
			},
			map[string]kion.CAR{
				fmt.Sprintf("%v (%v)", kionTestCARsNames[0], kionTestCARs[0].ID): kionTestCARs[0],
				fmt.Sprintf("%v (%v)", kionTestCARsNames[1], kionTestCARs[1].ID): kionTestCARs[1],
				fmt.Sprintf("%v (%v)", kionTestCARsNames[2], kionTestCARs[2].ID): kionTestCARs[2],
				fmt.Sprintf("%v (%v)", kionTestCARsNames[3], kionTestCARs[3].ID): kionTestCARs[3],
				fmt.Sprintf("%v (%v)", kionTestCARsNames[4], kionTestCARs[4].ID): kionTestCARs[4],
				fmt.Sprintf("%v (%v)", kionTestCARsNames[5], kionTestCARs[5].ID): kionTestCARs[5],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			one, two := MapCAR(test.cars)
			if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", one, two, test.wantOne, test.wantTwo)
			}
		})
	}
}

func TestMapIDMSs(t *testing.T) {
	tests := []struct {
		name    string
		idmss   []kion.IDMS
		wantOne []string
		wantTwo map[string]kion.IDMS
	}{
		{
			"Basic",
			kionTestIDMSs,
			[]string{
				kionTestIDMSsNames[4],
				kionTestIDMSsNames[3],
				kionTestIDMSsNames[0],
				kionTestIDMSsNames[5],
				kionTestIDMSsNames[2],
				kionTestIDMSsNames[1],
			},
			map[string]kion.IDMS{
				kionTestIDMSsNames[0]: kionTestIDMSs[0],
				kionTestIDMSsNames[1]: kionTestIDMSs[1],
				kionTestIDMSsNames[2]: kionTestIDMSs[2],
				kionTestIDMSsNames[3]: kionTestIDMSs[3],
				kionTestIDMSsNames[4]: kionTestIDMSs[4],
				kionTestIDMSsNames[5]: kionTestIDMSs[5],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			one, two := MapIDMSs(test.idmss)
			// if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
			if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", one, two, test.wantOne, test.wantTwo)
			}
		})
	}
}

func TestMapFavs(t *testing.T) {
	tests := []struct {
		name      string
		favorites []structs.Favorite
		wantOne   []string
		wantTwo   map[string]structs.Favorite
	}{
		{
			"Basic",
			kionTestFavorites,
			[]string{
				kionTestFavoritesNames[4],
				kionTestFavoritesNames[3],
				kionTestFavoritesNames[0],
				kionTestFavoritesNames[5],
				kionTestFavoritesNames[2],
				kionTestFavoritesNames[1],
			},
			map[string]structs.Favorite{
				kionTestFavoritesNames[0]: kionTestFavorites[0],
				kionTestFavoritesNames[1]: kionTestFavorites[1],
				kionTestFavoritesNames[2]: kionTestFavorites[2],
				kionTestFavoritesNames[3]: kionTestFavorites[3],
				kionTestFavoritesNames[4]: kionTestFavorites[4],
				kionTestFavoritesNames[5]: kionTestFavorites[5],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			one, two := MapFavs(test.favorites)
			// if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
			if !reflect.DeepEqual(test.wantOne, one) || !reflect.DeepEqual(test.wantTwo, two) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", one, two, test.wantOne, test.wantTwo)
			}
		})
	}
}

func TestFindCARByName(t *testing.T) {
	tests := []struct {
		name    string
		find    string
		cars    []kion.CAR
		wantCAR kion.CAR
		wantErr error
	}{
		{
			"Find Match",
			"car one",
			kionTestCARs,
			kionTestCARs[0],
			nil,
		},
		{
			"Find No Match",
			"fake car",
			kionTestCARs,
			kion.CAR{},
			fmt.Errorf("cannot find cloud access role with name %v", "fake car"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			car, err := FindCARByName(test.cars, test.find)
			// if !reflect.DeepEqual(&test.wantCAR, car) || test.wantErr != err {
			if !reflect.DeepEqual(&test.wantCAR, car) || !reflect.DeepEqual(test.wantErr, err) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", car, err, &test.wantCAR, test.wantErr)
			}
		})
	}
}

// New tests for alias and account ID

func TestFindCARByNameAndAccountNumber(t *testing.T) {
	tests := []struct {
		name          string
		carName       string
		accountNumber string
		cars          []kion.CAR
		wantCAR       *kion.CAR
		wantErr       error
	}{
		{
			"Find Match",
			"car one",
			"111111111111",
			kionTestCARs,
			&kionTestCARs[0],
			nil,
		},
		{
			"Find No Match - Wrong Account Number",
			"car one",
			"999999999999",
			kionTestCARs,
			nil,
			fmt.Errorf("cannot find cloud access role with name %v and account number %v", "car one", "999999999999"),
		},
		{
			"Find No Match - Wrong CAR Name",
			"fake car",
			"111111111111",
			kionTestCARs,
			nil,
			fmt.Errorf("cannot find cloud access role with name %v and account number %v", "fake car", "111111111111"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			car, err := FindCARByNameAndAccountNumber(test.cars, test.carName, test.accountNumber)
			if !reflect.DeepEqual(test.wantCAR, car) || (err != nil && err.Error() != test.wantErr.Error()) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", car, err, test.wantCAR, test.wantErr)
			}
		})
	}
}

func TestFindCARByNameAndAlias(t *testing.T) {
	tests := []struct {
		name         string
		carName      string
		accountAlias string
		cars         []kion.CAR
		wantCAR      *kion.CAR
		wantErr      error
	}{
		{
			"Find Match",
			"car two",
			"acct-two-alias",
			kionTestCARs,
			&kionTestCARs[1],
			nil,
		},
		{
			"Find No Match - Wrong Account Alias",
			"car two",
			"non-existent-alias",
			kionTestCARs,
			nil,
			fmt.Errorf("cannot find cloud access role with name %v and account alias %v", "car two", "non-existent-alias"),
		},
		{
			"Find No Match - Wrong CAR Name",
			"fake car",
			"acct-two-alias",
			kionTestCARs,
			nil,
			fmt.Errorf("cannot find cloud access role with name %v and account alias %v", "fake car", "acct-two-alias"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			car, err := FindCARByNameAndAlias(test.cars, test.carName, test.accountAlias)
			if !reflect.DeepEqual(test.wantCAR, car) || (err != nil && err.Error() != test.wantErr.Error()) {
				t.Errorf("\ngot:\n  %v\n  %v\nwanted:\n  %v\n  %v", car, err, test.wantCAR, test.wantErr)
			}
		})
	}
}
