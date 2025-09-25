// Package cache provides secure storage and retrieval of authentication data
// and temporary credentials for the Kion CLI using the system keyring. It
// offers both real and null cache implementations to support caching of SAML
// sessions, STAK (Short-Term Access Keys), and user passwords with automatic
// expiration handling and secure storage across macOS Keychain, Windows
// Credential Manager, and Linux Secret Service.
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
}

// NewNullCache creates a new NullCache.
func NewNullCache() *NullCache {
	return &NullCache{}
}
