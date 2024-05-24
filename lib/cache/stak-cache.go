package cache

import (
	"encoding/json"
	"time"

	"github.com/99designs/keyring"
	"github.com/kionsoftware/kion-cli/lib/kion"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Real Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// SetStak stores a STAK in the cache.
func (c *RealCache) SetStak(key string, value kion.STAK) error {
	// pull our stak cache
	cacheName := "Kion-CLI Cache"
	cache, err := c.keyring.Get(cacheName)
	if err != nil && err != keyring.ErrKeyNotFound {
		return err
	}

	// unmarshal the json data
	var cacheData CacheData
	if len(cache.Data) > 0 {
		err = json.Unmarshal(cache.Data, &cacheData)
		if err != nil {
			return err
		}
	}

	// initialize the map if it is still nil
	if cacheData.STAK == nil {
		cacheData.STAK = make(map[string]kion.STAK)
	}

	// clean expired entries
	now := time.Now()
	for key, stak := range cacheData.STAK {
		if stak.Expiration.Before(now) {
			delete(cacheData.STAK, key)
		}
	}

	// create our entry
	cacheData.STAK[key] = value

	// marshal the stack cache to json
	data, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	// build the keyring item
	cache = keyring.Item{
		Key:         cacheName,
		Data:        data,
		Label:       cacheName,
		Description: "Cache data for the Kion-CLI.",
	}

	// store the cache
	err = c.keyring.Set(cache)
	if err != nil {
		return err
	}

	return nil
}

// GetStak retrieves a STAK from the cache.
func (c *RealCache) GetStak(key string) (kion.STAK, bool, error) {
	// pull our stak cache
	cache, err := c.keyring.Get("Kion-CLI Cache")
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return kion.STAK{}, false, nil
		}
		return kion.STAK{}, false, err
	}

	// unmarshal the json data
	var cacheData CacheData
	if len(cache.Data) > 0 {
		err = json.Unmarshal(cache.Data, &cacheData)
		if err != nil {
			return kion.STAK{}, false, err
		}
	}

	// return the stak if found
	stak, found := cacheData.STAK[key]
	if found {
		return stak, true, nil
	}

	// return empty stak if not found
	return kion.STAK{}, false, nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Null Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// SetStak does nothing.
func (c *NullCache) SetStak(key string, value kion.STAK) error {
	return nil
}

// GetStak returns an empty STAK, false, and a nil error.
func (c *NullCache) GetStak(key string) (kion.STAK, bool, error) {
	return kion.STAK{}, false, nil
}
