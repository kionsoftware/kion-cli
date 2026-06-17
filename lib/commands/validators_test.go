package commands

import (
	"flag"
	"testing"

	"github.com/kionsoftware/kion-cli/lib/structs"
	"github.com/urfave/cli/v2"
)

// newTestContext creates a cli.Context for testing with the given flags.
func newTestContext(t *testing.T, flags map[string]string) *cli.Context {
	t.Helper()
	app := &cli.App{}
	set := flag.NewFlagSet("test", flag.ContinueOnError)

	// Register all flags we might need
	for name := range flags {
		set.String(name, "", "")
	}

	// Set flag values
	for name, value := range flags {
		if err := set.Set(name, value); err != nil {
			t.Fatalf("failed to set flag %q to %q: %v", name, value, err)
		}
	}

	return cli.NewContext(app, set, nil)
}

func TestValidateCmdStak(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no flags - valid",
			flags:   map[string]string{},
			wantErr: false,
		},
		{
			name: "account and car - valid",
			flags: map[string]string{
				"account": "123456789012",
				"car":     "AdminRole",
			},
			wantErr: false,
		},
		{
			name: "alias and car - valid",
			flags: map[string]string{
				"alias": "my-account",
				"car":   "AdminRole",
			},
			wantErr: false,
		},
		{
			name: "account without car - invalid",
			flags: map[string]string{
				"account": "123456789012",
			},
			wantErr: true,
			errMsg:  "must specify --car parameter when using --account or --alias",
		},
		{
			name: "alias without car - invalid",
			flags: map[string]string{
				"alias": "my-account",
			},
			wantErr: true,
			errMsg:  "must specify --car parameter when using --account or --alias",
		},
		{
			name: "car without account or alias - invalid",
			flags: map[string]string{
				"car": "AdminRole",
			},
			wantErr: true,
			errMsg:  "must specify --account OR --alias parameter when using --car",
		},
		{
			name: "all three flags - valid",
			flags: map[string]string{
				"account": "123456789012",
				"alias":   "my-account",
				"car":     "AdminRole",
			},
			wantErr: false,
		},
	}

	cmd := &Cmd{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newTestContext(t, test.flags)
			err := cmd.ValidateCmdStak(ctx)

			if test.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if err.Error() != test.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), test.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateCmdConsole(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no flags - valid",
			flags:   map[string]string{},
			wantErr: false,
		},
		{
			name: "account and car - valid",
			flags: map[string]string{
				"account": "123456789012",
				"car":     "AdminRole",
			},
			wantErr: false,
		},
		{
			name: "alias and car - valid",
			flags: map[string]string{
				"alias": "my-account",
				"car":   "AdminRole",
			},
			wantErr: false,
		},
		{
			name: "car without account or alias - invalid",
			flags: map[string]string{
				"car": "AdminRole",
			},
			wantErr: true,
			errMsg:  "must specify --account or --alias parameter when using --car",
		},
		{
			name: "account without car - invalid",
			flags: map[string]string{
				"account": "123456789012",
			},
			wantErr: true,
			errMsg:  "must specify --car parameter when using --account or --alias",
		},
		{
			name: "alias without car - invalid",
			flags: map[string]string{
				"alias": "my-account",
			},
			wantErr: true,
			errMsg:  "must specify --car parameter when using --account or --alias",
		},
	}

	cmd := &Cmd{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newTestContext(t, test.flags)
			err := cmd.ValidateCmdConsole(ctx)

			if test.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if err.Error() != test.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), test.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateCmdRun(t *testing.T) {
	tests := []struct {
		name      string
		flags     map[string]string
		favorites []structs.Favorite
		wantErr   bool
		errMsg    string
	}{
		// Favorite-based tests
		{
			name: "valid favorite",
			flags: map[string]string{
				"favorite": "my-fav",
			},
			favorites: []structs.Favorite{
				{Name: "my-fav", Account: "123456789012", CAR: "AdminRole"},
			},
			wantErr: false,
		},
		{
			name: "favorite not found",
			flags: map[string]string{
				"favorite": "nonexistent",
			},
			favorites: []structs.Favorite{
				{Name: "my-fav", Account: "123456789012", CAR: "AdminRole"},
			},
			wantErr: true,
			errMsg:  "can't find favorite",
		},
		{
			name: "favorite not found - empty favorites list",
			flags: map[string]string{
				"favorite": "my-fav",
			},
			favorites: []structs.Favorite{},
			wantErr:   true,
			errMsg:    "can't find favorite",
		},

		// Account/alias + car tests (without favorite)
		{
			name: "account and car without favorite - valid",
			flags: map[string]string{
				"account": "123456789012",
				"car":     "AdminRole",
			},
			favorites: []structs.Favorite{},
			wantErr:   false,
		},
		{
			name: "alias and car without favorite - valid",
			flags: map[string]string{
				"alias": "my-account",
				"car":   "AdminRole",
			},
			favorites: []structs.Favorite{},
			wantErr:   false,
		},
		{
			name: "account, alias and car without favorite - valid",
			flags: map[string]string{
				"account": "123456789012",
				"alias":   "my-account",
				"car":     "AdminRole",
			},
			favorites: []structs.Favorite{},
			wantErr:   false,
		},

		// Invalid combinations
		{
			name:      "no flags at all - invalid",
			flags:     map[string]string{},
			favorites: []structs.Favorite{},
			wantErr:   true,
			errMsg:    "must specify either --fav OR --account and --car  OR --alias and --car parameters",
		},
		{
			name: "account without car - invalid",
			flags: map[string]string{
				"account": "123456789012",
			},
			favorites: []structs.Favorite{},
			wantErr:   true,
			errMsg:    "must specify either --fav OR --account and --car  OR --alias and --car parameters",
		},
		{
			name: "alias without car - invalid",
			flags: map[string]string{
				"alias": "my-account",
			},
			favorites: []structs.Favorite{},
			wantErr:   true,
			errMsg:    "must specify either --fav OR --account and --car  OR --alias and --car parameters",
		},
		{
			name: "car without account or alias - invalid",
			flags: map[string]string{
				"car": "AdminRole",
			},
			favorites: []structs.Favorite{},
			wantErr:   true,
			errMsg:    "must specify either --fav OR --account and --car  OR --alias and --car parameters",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &Cmd{
				config: &structs.Configuration{
					Favorites: test.favorites,
				},
			}
			ctx := newTestContext(t, test.flags)
			err := cmd.ValidateCmdRun(ctx)

			if test.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if err.Error() != test.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), test.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateCmdRun_SetsDefaultRegion(t *testing.T) {
	t.Run("sets region from favorite", func(t *testing.T) {
		cmd := &Cmd{
			config: &structs.Configuration{
				Favorites: []structs.Favorite{
					{Name: "my-fav", Account: "123456789012", CAR: "AdminRole", Region: "us-west-2"},
				},
				Kion: structs.Kion{
					DefaultRegion: "",
				},
			},
		}

		ctx := newTestContext(t, map[string]string{"favorite": "my-fav"})
		err := cmd.ValidateCmdRun(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cmd.config.Kion.DefaultRegion != "us-west-2" {
			t.Errorf("DefaultRegion = %q, want %q", cmd.config.Kion.DefaultRegion, "us-west-2")
		}
	})

	t.Run("does not change region when favorite has no region", func(t *testing.T) {
		cmd := &Cmd{
			config: &structs.Configuration{
				Favorites: []structs.Favorite{
					{Name: "my-fav", Account: "123456789012", CAR: "AdminRole", Region: ""},
				},
				Kion: structs.Kion{
					DefaultRegion: "us-east-1",
				},
			},
		}

		ctx := newTestContext(t, map[string]string{"favorite": "my-fav"})
		err := cmd.ValidateCmdRun(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cmd.config.Kion.DefaultRegion != "us-east-1" {
			t.Errorf("DefaultRegion = %q, want %q (should be unchanged)", cmd.config.Kion.DefaultRegion, "us-east-1")
		}
	})

	t.Run("does not set region when using account+car", func(t *testing.T) {
		cmd := &Cmd{
			config: &structs.Configuration{
				Favorites: []structs.Favorite{},
				Kion: structs.Kion{
					DefaultRegion: "us-east-1",
				},
			},
		}

		ctx := newTestContext(t, map[string]string{
			"account": "123456789012",
			"car":     "AdminRole",
		})
		err := cmd.ValidateCmdRun(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cmd.config.Kion.DefaultRegion != "us-east-1" {
			t.Errorf("DefaultRegion = %q, want %q (should be unchanged)", cmd.config.Kion.DefaultRegion, "us-east-1")
		}
	})
}

func TestValidateCmdRun_NilFavorites(t *testing.T) {
	t.Run("nil favorites with account+car succeeds", func(t *testing.T) {
		cmd := &Cmd{
			config: &structs.Configuration{
				Favorites: nil,
			},
		}

		ctx := newTestContext(t, map[string]string{
			"account": "123456789012",
			"car":     "AdminRole",
		})
		err := cmd.ValidateCmdRun(ctx)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil favorites with favorite flag fails gracefully", func(t *testing.T) {
		cmd := &Cmd{
			config: &structs.Configuration{
				Favorites: nil,
			},
		}

		ctx := newTestContext(t, map[string]string{
			"favorite": "my-fav",
		})
		err := cmd.ValidateCmdRun(ctx)

		if err == nil {
			t.Error("expected error but got nil")
		} else if err.Error() != "can't find favorite" {
			t.Errorf("error = %q, want %q", err.Error(), "can't find favorite")
		}
	})
}
