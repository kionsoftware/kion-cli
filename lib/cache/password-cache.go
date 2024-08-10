package cache

import (
	"encoding/json"
	"fmt"

	"github.com/99designs/keyring"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Real Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// SetPassword stores a Password in the cache (or removes it if the password is nil).
func (c *RealCache) SetPassword(host string, idmsID uint, un string, pw string) error {
	// set the key based on what was passed
	key := fmt.Sprintf("%s-%d-%s", host, idmsID, un)

	// pull our cache
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
	if cacheData.PASSWORD == nil {
		cacheData.PASSWORD = make(map[string]string)
	}

	if pw != "" {
		// create/update our entry
		cacheData.PASSWORD[key] = pw
	} else {
		// Delete the entry
		delete(cacheData.PASSWORD, key)
	}

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

// GetPassword retrieves a password from the cache.
func (c *RealCache) GetPassword(host string, idmsID uint, un string) (string, bool, error) {
	// set the key based on what was passed
	key := fmt.Sprintf("%s-%d-%s", host, idmsID, un)

	// pull our cache
	cache, err := c.keyring.Get("Kion-CLI Cache")
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", false, nil
		}
		return "", false, err
	}

	// unmarshal the json data
	var cacheData CacheData
	if len(cache.Data) > 0 {
		err = json.Unmarshal(cache.Data, &cacheData)
		if err != nil {
			return "", false, err
		}
	}

	// return the password if found
	password, found := cacheData.PASSWORD[key]
	if found {
		return password, true, nil
	}

	// return empty password if not found
	return "", false, nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Null Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// SetPassword does nothing.
func (c *NullCache) SetPassword(host string, idmsID uint, un string, pw string) error {
	return nil
}

// GetPassword returns an empty password, false, and a nil error.
func (c *NullCache) GetPassword(host string, idmsID uint, un string) (string, bool, error) {
	return "", false, nil
}
