package cache

import (
	"encoding/json"

	"github.com/99designs/keyring"
)

// flushCache clears the Kion CLI cache.
func flushCache(k keyring.Keyring) error {
	// marshal an empty cache to json
	var cacheData CacheData
	data, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	// build the keyring item
	cacheName := "Kion-CLI Cache"
	cache := keyring.Item{
		Key:         cacheName,
		Data:        data,
		Label:       cacheName,
		Description: "Cache data for the Kion-CLI.",
	}

	// store the cache
	err = k.Set(cache)
	if err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Real Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// FLushCache implements the FlushCache interface for RealCache.
func (c *RealCache) FlushCache() error {
	return flushCache(c.keyring)
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Null Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// FLushCache implements the FlushCache interface for NullCache.
// This is just to satisfy the interface and should never be called.
func (c *NullCache) FlushCache() error {
	return nil
}
