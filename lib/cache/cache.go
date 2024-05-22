package cache

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"

	"github.com/kionsoftware/kion-cli/lib/kion"
)

type Cache struct {
	data map[string]kion.STAK
	key  []byte
}

func NewCache(key []byte) *Cache {
	return &Cache{
		data: make(map[string]kion.STAK),
		key:  key,
	}
}

// Set stores new STAK entries into the cache.
func (c *Cache) Set(key string, value kion.STAK) {
	c.data[key] = value
}

// Get retrieves existing STAK entries from the cache.
func (c *Cache) Get(key string) (kion.STAK, bool) {
	value, found := c.data[key]
	return value, found
}

// SaveToFile stores the current cache into the specified file.
func (c *Cache) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encryptedData, err := c.encrypt()
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(file)
	return encoder.Encode(encryptedData)
}

// LoadFromFile retrieves the full cache from the specified file. It handles
// creation of the path and file as needed.
func (c *Cache) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// if the file does not exist, create all necessary directories
			err = os.MkdirAll(filepath.Dir(filename), 0755)
			if err != nil {
				return err
			}

			// create an empty file.
			file, err = os.Create(filename)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var encryptedData []byte
	err = decoder.Decode(&encryptedData)
	if err != nil {
		if err == io.EOF {
			// if the file is empty, return nil
			return nil
		}
		return err
	}

	return c.decrypt(encryptedData)
}

// encrypt handles the encryption of the cache.
func (c *Cache) encrypt() ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	data, err := gobEncode(c.data)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decrypt handles the decryption of the cache.
func (c *Cache) decrypt(encryptedData []byte) error {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return err
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	data, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	return gobDecode(data, &c.data)
}

// gobEncode encodes an arbitrary value into a byte slice array.
func gobEncode(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// gobDecode decodes a byte slice array into a value.
func gobDecode(data []byte, value interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	return dec.Decode(value)
}
