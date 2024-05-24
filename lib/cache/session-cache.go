package cache

import (
	"encoding/json"

	"github.com/99designs/keyring"
	"github.com/kionsoftware/kion-cli/lib/kion"
)

// SetSession is a common func for all Cache implementations and stores a
// Session in the cache.
func SetSession(k keyring.Keyring, session kion.Session) error {
	// pull our stak cache
	cacheName := "Kion-CLI Cache"
	cache, err := k.Get(cacheName)
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

	// store the session
	cacheData.SESSION = session

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
	err = k.Set(cache)
	if err != nil {
		return err
	}

	return nil

}

// GetSession is a common func for all Cache implementations and retrieves a
// Session in the cache.
func GetSession(k keyring.Keyring) (kion.Session, bool, error) {
	// pull our stak cache
	cache, err := k.Get("Kion-CLI Cache")
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return kion.Session{}, false, nil
		}
		return kion.Session{}, false, err
	}

	// unmarshal the json data
	var cacheData CacheData
	if len(cache.Data) > 0 {
		err = json.Unmarshal(cache.Data, &cacheData)
		if err != nil {
			return kion.Session{}, false, err
		}
	}

	// return the stak if found
	session := cacheData.SESSION
	if session != (kion.Session{}) {
		return session, true, nil
	}

	// return empty stak if not found
	return kion.Session{}, false, nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Real Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// SetSession implements the Cache interface for RealCache and wraps a common
// function for storing session data.
func (c *RealCache) SetSession(session kion.Session) error {
	return SetSession(c.keyring, session)
}

// GetSession implements the Cache interface for RealCache and wraps a common
// function for retrieving session data.
func (c *RealCache) GetSession() (kion.Session, bool, error) {
	return GetSession(c.keyring)
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Null Cacher                                                               //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// SetSession implements the Cache interface for NullCache and wraps a common
// function for storing session data.
func (c *NullCache) SetSession(session kion.Session) error {
	return SetSession(c.keyring, session)
}

// GetSession implements the Cache interface for NullCache and wraps a common
// function for retrieving session data.
func (c *NullCache) GetSession() (kion.Session, bool, error) {
	return GetSession(c.keyring)
}
