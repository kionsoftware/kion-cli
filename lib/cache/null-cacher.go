package cache

import "github.com/kionsoftware/kion-cli/lib/kion"

// NullCache implements the Cache interface and does nothing.
type NullCache struct{}

// NewNullCache creates a new NullCache.
func NewNullCache() *NullCache {
	return &NullCache{}
}

// SetStak does nothing.
func (c *NullCache) SetStak(key string, value kion.STAK) error {
	return nil
}

// GetStak returns an empty STAK, false, and a nil error.
func (c *NullCache) GetStak(key string) (kion.STAK, bool, error) {
	return kion.STAK{}, false, nil
}
