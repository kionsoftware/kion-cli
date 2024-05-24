package cache

import (
	"github.com/99designs/keyring"
	"github.com/kionsoftware/kion-cli/lib/kion"
)

// Cache is an interface for storing and receiving data.
type Cache interface {
	SetStak(key string, value kion.STAK) error
	GetStak(key string) (kion.STAK, bool, error)
}

// RealCache is our cache object for passing the keychain to receiver methods.
type RealCache struct {
	keyring keyring.Keyring
}

// CacheData is a nested structure for storing kion-cli data.
type CacheData struct {
	STAK map[string]kion.STAK
	// TODO: extend for kion session data
}

// NewCache creates a new RealCache.
func NewCache(keyring keyring.Keyring) *RealCache {
	return &RealCache{
		keyring: keyring,
	}
}
