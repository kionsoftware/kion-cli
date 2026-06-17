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
				"fav five     [local] (151515151515 car five web)",
				"fav four     [local] (141414141414 car four web)",
				"fav one      [local] (111111111111 car one web)",
				"fav six      [local] (161616161616 car six web)",
				"fav three    [local] (131313131313 car three web)",
				"fav two      [local] (121212121212 car two web)",
			},
			map[string]structs.Favorite{
				// Indexed by DescriptiveName
				"fav one      [local] (111111111111 car one web)":   kionTestFavorites[0],
				"fav two      [local] (121212121212 car two web)":   kionTestFavorites[1],
				"fav three    [local] (131313131313 car three web)": kionTestFavorites[2],
				"fav four     [local] (141414141414 car four web)":  kionTestFavorites[3],
				"fav five     [local] (151515151515 car five web)":  kionTestFavorites[4],
				"fav six      [local] (161616161616 car six web)":   kionTestFavorites[5],
				// Also indexed by plain Name for CLI argument lookup
				"fav one":   kionTestFavorites[0],
				"fav two":   kionTestFavorites[1],
				"fav three": kionTestFavorites[2],
				"fav four":  kionTestFavorites[3],
				"fav five":  kionTestFavorites[4],
				"fav six":   kionTestFavorites[5],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			one, two := MapFavs(test.favorites)
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

func TestCombineFavorites_EmptyInputs(t *testing.T) {
	all, comparison, err := CombineFavorites(nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty All, got %d items", len(all))
	}
	if len(comparison.LocalOnly) != 0 {
		t.Errorf("expected empty LocalOnly, got %d items", len(comparison.LocalOnly))
	}
	if len(comparison.ConflictsLocal) != 0 {
		t.Errorf("expected empty ConflictsLocal, got %d items", len(comparison.ConflictsLocal))
	}
}

func TestCombineFavorites_OnlyUpstream(t *testing.T) {
	upstream := []structs.Favorite{
		{Name: "upstream1", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
		{Name: "upstream2", Account: "222222222222", CAR: "Role2", AccessType: "web"},
	}

	all, comparison, err := CombineFavorites(nil, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 in All, got %d", len(all))
	}
	if len(comparison.LocalOnly) != 0 {
		t.Errorf("expected empty LocalOnly, got %d items", len(comparison.LocalOnly))
	}
	for _, fav := range all {
		if fav.DescriptiveName == "" {
			t.Errorf("expected DescriptiveName to be set for %s", fav.Name)
		}
	}
}

func TestCombineFavorites_OnlyLocal(t *testing.T) {
	local := []structs.Favorite{
		{Name: "local1", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
		{Name: "local2", Account: "222222222222", CAR: "Role2", AccessType: "web"},
	}

	all, comparison, err := CombineFavorites(local, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 in All, got %d", len(all))
	}
	if len(comparison.LocalOnly) != 2 {
		t.Errorf("expected 2 in LocalOnly, got %d", len(comparison.LocalOnly))
	}
	if comparison.LocalOnly[0].Name != "local1" || comparison.LocalOnly[1].Name != "local2" {
		t.Errorf("LocalOnly contains wrong favorites")
	}
}

func TestCombineFavorites_ExactMatch(t *testing.T) {
	local := []structs.Favorite{
		{Name: "shared", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
	}
	upstream := []structs.Favorite{
		{Name: "shared", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 in All (upstream only), got %d", len(all))
	}
	if len(comparison.LocalOnly) != 0 {
		t.Errorf("expected empty LocalOnly for exact match, got %d", len(comparison.LocalOnly))
	}
	if len(comparison.ConflictsLocal) != 0 {
		t.Errorf("expected empty ConflictsLocal for exact match, got %d", len(comparison.ConflictsLocal))
	}
}

func TestCombineFavorites_NameConflict(t *testing.T) {
	local := []structs.Favorite{
		{Name: "myfav", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
	}
	upstream := []structs.Favorite{
		{Name: "myfav", Account: "222222222222", CAR: "Role2", AccessType: "web", Unaliased: false},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 in All, got %d", len(all))
	}
	if len(comparison.ConflictsLocal) != 1 {
		t.Errorf("expected 1 in ConflictsLocal, got %d", len(comparison.ConflictsLocal))
	}
	if len(comparison.ConflictsUpstream) != 1 {
		t.Errorf("expected 1 in ConflictsUpstream, got %d", len(comparison.ConflictsUpstream))
	}
	if comparison.ConflictsLocal[0].Name != "myfav" {
		t.Errorf("expected conflict to be 'myfav', got %s", comparison.ConflictsLocal[0].Name)
	}
}

func TestCombineFavorites_Duplicate(t *testing.T) {
	local := []structs.Favorite{
		{Name: "local-name", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
	}
	upstream := []structs.Favorite{
		{Name: "upstream-name", Account: "111111111111", CAR: "Role1", AccessType: "cli", Unaliased: false},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 in All, got %d", len(all))
	}
	if len(comparison.ConflictsLocal) != 1 {
		t.Errorf("expected 1 in ConflictsLocal, got %d", len(comparison.ConflictsLocal))
	}
	if len(comparison.ConflictsUpstream) != 1 {
		t.Errorf("expected 1 in ConflictsUpstream, got %d", len(comparison.ConflictsUpstream))
	}
}

func TestCombineFavorites_UnaliasedMatch(t *testing.T) {
	local := []structs.Favorite{
		{Name: "my-local-name", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
	}
	upstream := []structs.Favorite{
		{Name: "", Account: "111111111111", CAR: "Role1", AccessType: "cli", Unaliased: true},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 in All, got %d", len(all))
	}
	if all[0].Name != "my-local-name" {
		t.Errorf("expected local favorite in All, got %s", all[0].Name)
	}
	if len(comparison.UnaliasedLocal) != 1 {
		t.Errorf("expected 1 in UnaliasedLocal, got %d", len(comparison.UnaliasedLocal))
	}
	if len(comparison.UnaliasedUpstream) != 1 {
		t.Errorf("expected 1 in UnaliasedUpstream, got %d", len(comparison.UnaliasedUpstream))
	}
	if len(comparison.LocalOnly) != 0 {
		t.Errorf("expected empty LocalOnly, got %d", len(comparison.LocalOnly))
	}
}

func TestCombineFavorites_MultipleConflictsWithSameUpstream(t *testing.T) {
	local := []structs.Favorite{
		{Name: "shared", Account: "111111111111", CAR: "RoleA", AccessType: "cli"},
		{Name: "shared", Account: "222222222222", CAR: "RoleB", AccessType: "web"},
	}
	upstream := []structs.Favorite{
		{Name: "shared", Account: "333333333333", CAR: "RoleC", AccessType: "cli", Unaliased: false},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 in All, got %d", len(all))
	}
	if len(comparison.ConflictsLocal) != 2 {
		t.Errorf("expected 2 in ConflictsLocal, got %d", len(comparison.ConflictsLocal))
	}
	if len(comparison.ConflictsUpstream) != 1 {
		t.Errorf("expected 1 in ConflictsUpstream (deduped), got %d", len(comparison.ConflictsUpstream))
	}
}

func TestCombineFavorites_MixedScenarios(t *testing.T) {
	local := []structs.Favorite{
		{Name: "exact", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
		{Name: "local-only", Account: "999999999999", CAR: "UniqueRole", AccessType: "web"},
		{Name: "conflict", Account: "333333333333", CAR: "RoleX", AccessType: "cli"},
	}
	upstream := []structs.Favorite{
		{Name: "exact", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
		{Name: "conflict", Account: "444444444444", CAR: "RoleY", AccessType: "web", Unaliased: false},
		{Name: "upstream-only", Account: "555555555555", CAR: "RoleZ", AccessType: "cli"},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("expected 5 in All, got %d", len(all))
	}
	if len(comparison.LocalOnly) != 1 {
		t.Errorf("expected 1 in LocalOnly, got %d", len(comparison.LocalOnly))
	}
	if comparison.LocalOnly[0].Name != "local-only" {
		t.Errorf("expected 'local-only' in LocalOnly, got %s", comparison.LocalOnly[0].Name)
	}
	if len(comparison.ConflictsLocal) != 1 {
		t.Errorf("expected 1 in ConflictsLocal, got %d", len(comparison.ConflictsLocal))
	}
	if comparison.ConflictsLocal[0].Name != "conflict" {
		t.Errorf("expected 'conflict' in ConflictsLocal, got %s", comparison.ConflictsLocal[0].Name)
	}
}

func TestCombineFavorites_NilVsEmptySlice(t *testing.T) {
	all1, comp1, err1 := CombineFavorites(nil, nil)
	all2, comp2, err2 := CombineFavorites([]structs.Favorite{}, []structs.Favorite{})

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if len(all1) != len(all2) {
		t.Errorf("nil vs empty slice produced different All lengths: %d vs %d", len(all1), len(all2))
	}
	if len(comp1.LocalOnly) != len(comp2.LocalOnly) {
		t.Errorf("nil vs empty slice produced different LocalOnly lengths")
	}
}

func TestCombineFavorites_UnaliasedUpstreamWithSameName(t *testing.T) {
	// When upstream has Unaliased=true but all fields match (including name),
	// it's still treated as an exact match (exact match check comes first)
	local := []structs.Favorite{
		{Name: "samename", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
	}
	upstream := []structs.Favorite{
		{Name: "samename", Account: "111111111111", CAR: "Role1", AccessType: "cli", Unaliased: true},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be exact match (not unaliased match) because all fields including name match
	if len(comparison.ConflictsLocal) != 0 {
		t.Errorf("expected no conflicts, got %d", len(comparison.ConflictsLocal))
	}
	if len(comparison.UnaliasedLocal) != 0 {
		t.Errorf("expected no unaliased (exact match takes priority), got %d", len(comparison.UnaliasedLocal))
	}
	// Only upstream in All (local is exact match, not added)
	if len(all) != 1 {
		t.Errorf("expected 1 in All, got %d", len(all))
	}
}

func TestCombineFavorites_UnaliasedUpstreamDifferentName(t *testing.T) {
	// When upstream has Unaliased=true and different name but same account/CAR/AccessType,
	// it should be treated as unaliased match (local provides the name)
	local := []structs.Favorite{
		{Name: "my-local-name", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
	}
	upstream := []structs.Favorite{
		{Name: "different-name", Account: "111111111111", CAR: "Role1", AccessType: "cli", Unaliased: true},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be unaliased match
	if len(comparison.UnaliasedLocal) != 1 {
		t.Errorf("expected 1 in UnaliasedLocal, got %d", len(comparison.UnaliasedLocal))
	}
	if len(comparison.UnaliasedUpstream) != 1 {
		t.Errorf("expected 1 in UnaliasedUpstream, got %d", len(comparison.UnaliasedUpstream))
	}
	// Local replaces upstream in All
	if len(all) != 1 {
		t.Errorf("expected 1 in All, got %d", len(all))
	}
	if all[0].Name != "my-local-name" {
		t.Errorf("expected local name in All, got %s", all[0].Name)
	}
}

func TestCombineFavorites_ExactMatchTakesPriority(t *testing.T) {
	// If local matches first upstream exactly, it shouldn't conflict with second
	local := []structs.Favorite{
		{Name: "myfav", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
	}
	upstream := []structs.Favorite{
		// First upstream is exact match
		{Name: "myfav", Account: "111111111111", CAR: "Role1", AccessType: "cli"},
		// Second upstream has same name but different settings (would be conflict)
		{Name: "myfav", Account: "222222222222", CAR: "Role2", AccessType: "web"},
	}

	all, comparison, err := CombineFavorites(local, upstream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be exact match with first upstream, no conflict
	if len(comparison.ConflictsLocal) != 0 {
		t.Errorf("expected no conflicts (exact match takes priority), got %d", len(comparison.ConflictsLocal))
	}
	if len(comparison.LocalOnly) != 0 {
		t.Errorf("expected no LocalOnly, got %d", len(comparison.LocalOnly))
	}
	// All should have both upstreams (local not added due to exact match)
	if len(all) != 2 {
		t.Errorf("expected 2 in All, got %d", len(all))
	}
}

func TestCombineFavorites_ErrorAlwaysNil(t *testing.T) {
	// Verify function never returns an error (current implementation)
	testCases := []struct {
		local    []structs.Favorite
		upstream []structs.Favorite
	}{
		{nil, nil},
		{[]structs.Favorite{}, []structs.Favorite{}},
		{[]structs.Favorite{{Name: "a"}}, nil},
		{nil, []structs.Favorite{{Name: "b"}}},
	}

	for _, tc := range testCases {
		_, _, err := CombineFavorites(tc.local, tc.upstream)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	}
}

func TestMapAccountsFromCARS_Empty(t *testing.T) {
	names, aMap := MapAccountsFromCARS(nil, 0)

	if len(names) != 0 {
		t.Errorf("expected empty names, got %d", len(names))
	}
	if len(aMap) != 0 {
		t.Errorf("expected empty map, got %d", len(aMap))
	}
}

func TestMapAccountsFromCARS_EmptySlice(t *testing.T) {
	names, aMap := MapAccountsFromCARS([]kion.CAR{}, 0)

	if len(names) != 0 {
		t.Errorf("expected empty names, got %d", len(names))
	}
	if len(aMap) != 0 {
		t.Errorf("expected empty map, got %d", len(aMap))
	}
}

func TestMapAccountsFromCARS_WithAliases(t *testing.T) {
	cars := []kion.CAR{
		{AccountName: "account one", AccountAlias: "alias-one", AccountNumber: "111111111111", ProjectID: 100},
		{AccountName: "account two", AccountAlias: "alias-two", AccountNumber: "222222222222", ProjectID: 100},
	}

	names, aMap := MapAccountsFromCARS(cars, 0)

	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}

	expectedName1 := "account one [alias-one] (111111111111)"
	expectedName2 := "account two [alias-two] (222222222222)"

	// Check map has correct entries
	if aMap[expectedName1] != "111111111111" {
		t.Errorf("expected account number 111111111111 for %q, got %q", expectedName1, aMap[expectedName1])
	}
	if aMap[expectedName2] != "222222222222" {
		t.Errorf("expected account number 222222222222 for %q, got %q", expectedName2, aMap[expectedName2])
	}
}

func TestMapAccountsFromCARS_WithoutAliases(t *testing.T) {
	cars := []kion.CAR{
		{AccountName: "account one", AccountAlias: "", AccountNumber: "111111111111", ProjectID: 100},
		{AccountName: "account two", AccountAlias: "", AccountNumber: "222222222222", ProjectID: 100},
	}

	names, aMap := MapAccountsFromCARS(cars, 0)

	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}

	expectedName1 := "account one (111111111111)"
	expectedName2 := "account two (222222222222)"

	if aMap[expectedName1] != "111111111111" {
		t.Errorf("expected account number 111111111111 for %q, got %q", expectedName1, aMap[expectedName1])
	}
	if aMap[expectedName2] != "222222222222" {
		t.Errorf("expected account number 222222222222 for %q, got %q", expectedName2, aMap[expectedName2])
	}

	// Verify aliases are not included in format
	for _, name := range names {
		if contains := len(name) > 0 && name[len(name)-1] == ']'; contains {
			// Check for bracket pattern indicating alias
			for i := len(name) - 2; i >= 0; i-- {
				if name[i] == '[' {
					t.Errorf("expected no alias brackets in name %q", name)
					break
				}
			}
		}
	}
}

func TestMapAccountsFromCARS_ProjectFiltering(t *testing.T) {
	cars := []kion.CAR{
		{AccountName: "project100 account", AccountNumber: "111111111111", ProjectID: 100},
		{AccountName: "project200 account", AccountNumber: "222222222222", ProjectID: 200},
		{AccountName: "project100 account2", AccountNumber: "333333333333", ProjectID: 100},
	}

	// pid=0 returns all
	names, aMap := MapAccountsFromCARS(cars, 0)
	if len(names) != 3 {
		t.Errorf("pid=0: expected 3 names, got %d", len(names))
	}
	if len(aMap) != 3 {
		t.Errorf("pid=0: expected 3 map entries, got %d", len(aMap))
	}

	// pid=100 returns only project 100
	names, aMap = MapAccountsFromCARS(cars, 100)
	if len(names) != 2 {
		t.Errorf("pid=100: expected 2 names, got %d", len(names))
	}
	if len(aMap) != 2 {
		t.Errorf("pid=100: expected 2 map entries, got %d", len(aMap))
	}

	// pid=200 returns only project 200
	names, aMap = MapAccountsFromCARS(cars, 200)
	if len(names) != 1 {
		t.Errorf("pid=200: expected 1 name, got %d", len(names))
	}
	if aMap["project200 account (222222222222)"] != "222222222222" {
		t.Errorf("pid=200: expected project200 account in map")
	}

	// pid=999 returns empty (no matching project)
	names, aMap = MapAccountsFromCARS(cars, 999)
	if len(names) != 0 {
		t.Errorf("pid=999: expected 0 names, got %d", len(names))
	}
	if len(aMap) != 0 {
		t.Errorf("pid=999: expected 0 map entries, got %d", len(aMap))
	}
}

func TestMapAccountsFromCARS_Deduplication(t *testing.T) {
	// Same account appears in multiple CARs (different roles for same account)
	cars := []kion.CAR{
		{AccountName: "shared account", AccountAlias: "shared", AccountNumber: "111111111111", ProjectID: 100, Name: "AdminRole"},
		{AccountName: "shared account", AccountAlias: "shared", AccountNumber: "111111111111", ProjectID: 100, Name: "ReadOnlyRole"},
		{AccountName: "shared account", AccountAlias: "shared", AccountNumber: "111111111111", ProjectID: 100, Name: "DevRole"},
	}

	names, aMap := MapAccountsFromCARS(cars, 0)

	// Should only have 1 entry (deduplicated)
	if len(names) != 1 {
		t.Errorf("expected 1 deduplicated name, got %d", len(names))
	}
	if len(aMap) != 1 {
		t.Errorf("expected 1 deduplicated map entry, got %d", len(aMap))
	}

	expectedName := "shared account [shared] (111111111111)"
	if names[0] != expectedName {
		t.Errorf("expected %q, got %q", expectedName, names[0])
	}
}

func TestMapAccountsFromCARS_Sorting(t *testing.T) {
	cars := []kion.CAR{
		{AccountName: "zebra account", AccountNumber: "333333333333", ProjectID: 100},
		{AccountName: "alpha account", AccountNumber: "111111111111", ProjectID: 100},
		{AccountName: "beta account", AccountNumber: "222222222222", ProjectID: 100},
	}

	names, _ := MapAccountsFromCARS(cars, 0)

	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}

	// Should be sorted alphabetically
	expected := []string{
		"alpha account (111111111111)",
		"beta account (222222222222)",
		"zebra account (333333333333)",
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("position %d: expected %q, got %q", i, expected[i], name)
		}
	}
}

func TestMapAccountsFromCARS_MixedAliases(t *testing.T) {
	// Some accounts have aliases, some don't
	cars := []kion.CAR{
		{AccountName: "account with alias", AccountAlias: "my-alias", AccountNumber: "111111111111", ProjectID: 100},
		{AccountName: "account without alias", AccountAlias: "", AccountNumber: "222222222222", ProjectID: 100},
	}

	names, aMap := MapAccountsFromCARS(cars, 0)

	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}

	withAlias := "account with alias [my-alias] (111111111111)"
	withoutAlias := "account without alias (222222222222)"

	if aMap[withAlias] != "111111111111" {
		t.Errorf("expected account with alias in map")
	}
	if aMap[withoutAlias] != "222222222222" {
		t.Errorf("expected account without alias in map")
	}
}
