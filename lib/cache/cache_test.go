package cache

import (
	"testing"

	"github.com/kionsoftware/kion-cli/lib/kion"
)

func TestNewNullCache(t *testing.T) {
	cache := NewNullCache()
	if cache == nil {
		t.Error("NewNullCache() returned nil")
	}
}

func TestNullCacheImplementsInterface(t *testing.T) {
	// Verify NullCache satisfies the Cache interface at compile time.
	// This is a compile-time check - if NullCache doesn't implement Cache,
	// this file won't compile.
	var _ Cache = (*NullCache)(nil)
}

func TestNullCache_Stak(t *testing.T) {
	tests := []struct {
		name     string
		carName  string
		accNum   string
		accAlias string
		stak     kion.STAK
	}{
		{
			name:     "typical values",
			carName:  "AdminRole",
			accNum:   "123456789012",
			accAlias: "my-account",
			stak: kion.STAK{
				AccessKey:       "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				SessionToken:    "token123",
			},
		},
		{
			name:     "empty strings",
			carName:  "",
			accNum:   "",
			accAlias: "",
			stak:     kion.STAK{},
		},
		{
			name:     "only car name",
			carName:  "ReadOnly",
			accNum:   "",
			accAlias: "",
			stak:     kion.STAK{AccessKey: "AKIA123"},
		},
		{
			name:     "special characters",
			carName:  "role/with/slashes",
			accNum:   "000000000000",
			accAlias: "alias-with-dashes_and_underscores",
			stak:     kion.STAK{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := NewNullCache()

			// SetStak should always succeed
			err := cache.SetStak(test.carName, test.accNum, test.accAlias, test.stak)
			if err != nil {
				t.Errorf("SetStak() returned error: %v", err)
			}

			// GetStak should always return empty, false, nil
			stak, found, err := cache.GetStak(test.carName, test.accNum, test.accAlias)
			if err != nil {
				t.Errorf("GetStak() returned error: %v", err)
			}
			if found {
				t.Error("GetStak() returned found=true, want false")
			}
			if stak != (kion.STAK{}) {
				t.Errorf("GetStak() returned non-empty STAK: %+v", stak)
			}
		})
	}
}

func TestNullCache_Session(t *testing.T) {
	tests := []struct {
		name    string
		session kion.Session
	}{
		{
			name: "typical session",
			session: kion.Session{
				IDMSID:   1,
				UserName: "user@example.com",
			},
		},
		{
			name:    "empty session",
			session: kion.Session{},
		},
		{
			name: "zero IDMS ID",
			session: kion.Session{
				IDMSID:   0,
				UserName: "user@example.com",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := NewNullCache()

			// SetSession should always succeed
			err := cache.SetSession(test.session)
			if err != nil {
				t.Errorf("SetSession() returned error: %v", err)
			}

			// GetSession should always return empty, false, nil
			session, found, err := cache.GetSession()
			if err != nil {
				t.Errorf("GetSession() returned error: %v", err)
			}
			if found {
				t.Error("GetSession() returned found=true, want false")
			}
			if session != (kion.Session{}) {
				t.Errorf("GetSession() returned non-empty Session: %+v", session)
			}
		})
	}
}

func TestNullCache_Password(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		idmsID   uint
		username string
		password string
	}{
		{
			name:     "typical values",
			host:     "https://kion.example.com",
			idmsID:   1,
			username: "user@example.com",
			password: "secretpassword123",
		},
		{
			name:     "empty strings",
			host:     "",
			idmsID:   0,
			username: "",
			password: "",
		},
		{
			name:     "zero IDMS ID",
			host:     "https://kion.example.com",
			idmsID:   0,
			username: "user@example.com",
			password: "password",
		},
		{
			name:     "special characters in password",
			host:     "https://kion.example.com",
			idmsID:   1,
			username: "user@example.com",
			password: "p@$$w0rd!#$%^&*()",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := NewNullCache()

			// SetPassword should always succeed
			err := cache.SetPassword(test.host, test.idmsID, test.username, test.password)
			if err != nil {
				t.Errorf("SetPassword() returned error: %v", err)
			}

			// GetPassword should always return empty, false, nil
			password, found, err := cache.GetPassword(test.host, test.idmsID, test.username)
			if err != nil {
				t.Errorf("GetPassword() returned error: %v", err)
			}
			if found {
				t.Error("GetPassword() returned found=true, want false")
			}
			if password != "" {
				t.Errorf("GetPassword() returned non-empty password: %q", password)
			}
		})
	}
}

func TestNullCache_FlushCache(t *testing.T) {
	cache := NewNullCache()
	err := cache.FlushCache()
	if err != nil {
		t.Errorf("FlushCache() returned error: %v", err)
	}

	// Calling multiple times should still succeed
	err = cache.FlushCache()
	if err != nil {
		t.Errorf("FlushCache() second call returned error: %v", err)
	}
}

func TestNullCache_MultipleInstances(t *testing.T) {
	// Verify that multiple NullCache instances don't interfere with each other
	// (they shouldn't, since they don't store anything)
	cache1 := NewNullCache()
	cache2 := NewNullCache()

	// Set data on cache1
	_ = cache1.SetStak("role1", "111111111111", "alias1", kion.STAK{AccessKey: "KEY1"})
	_ = cache1.SetSession(kion.Session{IDMSID: 1, UserName: "user1"})
	_ = cache1.SetPassword("host1", 1, "user1", "pass1")

	// Set different data on cache2
	_ = cache2.SetStak("role2", "222222222222", "alias2", kion.STAK{AccessKey: "KEY2"})
	_ = cache2.SetSession(kion.Session{IDMSID: 2, UserName: "user2"})
	_ = cache2.SetPassword("host2", 2, "user2", "pass2")

	// Both should return empty for any query
	stak1, found1, _ := cache1.GetStak("role1", "111111111111", "alias1")
	stak2, found2, _ := cache2.GetStak("role2", "222222222222", "alias2")

	if found1 || found2 {
		t.Error("NullCache should never return found=true")
	}
	if stak1 != (kion.STAK{}) || stak2 != (kion.STAK{}) {
		t.Error("NullCache should always return empty STAK")
	}
}

func TestRealCacheImplementsInterface(t *testing.T) {
	// Verify RealCache satisfies the Cache interface at compile time.
	var _ Cache = (*RealCache)(nil)
}
