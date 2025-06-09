package cache

import (
	"github.com/99designs/keyring"
	"github.com/kionsoftware/kion-cli/lib/kion"
)

// Cache is an interface for storing and receiving data.
type Cache interface {
	SetStak(carName string, accNum string, accAlias string, value kion.STAK) error
	GetStak(carName string, accNum string, accAlias string) (kion.STAK, bool, error)
	SetSession(value kion.Session) error
	GetSession() (kion.Session, bool, error)
	SetPassword(host string, idmsID uint, un string, pw string) error
	GetPassword(host string, idmsID uint, un string) (string, bool, error)
	FlushCache() error
}

// NullKeyRing implements the keyring.Keyring interface but does nothing.
// It is used when no keyring is available or when we want to disable caching.
type NullKeyRing struct{}

func (n NullKeyRing) Get(_ string) (keyring.Item, error) {
	return keyring.Item{}, keyring.ErrKeyNotFound
}
func (n NullKeyRing) Set(_ keyring.Item) error {
	return nil
}
func (n NullKeyRing) Remove(_ string) error {
	return nil
}
func (n NullKeyRing) Keys() ([]string, error) {
	return nil, nil
}
func (n NullKeyRing) GetMetadata(_ string) (keyring.Metadata, error) {
	return keyring.Metadata{}, nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Real Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// RealCache is our cache object for passing the keychain to receiver methods.
type RealCache struct {
	keyring keyring.Keyring
}

// CacheData is a nested structure for storing kion-cli data.
type CacheData struct {
	STAK     map[string]kion.STAK
	SESSION  kion.Session
	PASSWORD map[string]string
}

// NewCache creates a new RealCache.
func NewCache(keyring keyring.Keyring) *RealCache {
	return &RealCache{
		keyring: keyring,
	}
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Null Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// NullCache implements the Cache interface and does nothing.
type NullCache struct {
	keyring keyring.Keyring
}

// NewNullCache creates a new NullCache.
func NewNullCache(keyring keyring.Keyring) *NullCache {
	return &NullCache{
		keyring: keyring,
	}
}
