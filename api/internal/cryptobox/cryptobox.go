// Package cryptobox encrypts application secrets with AES-256-GCM.
package cryptobox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

type Box struct{ aead cipher.AEAD }

func New(encodedKey string) (*Box, error) {
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil || len(key) != 32 {
		return nil, errors.New("APP_ENCRYPTION_KEY must be 32 bytes encoded as base64")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Box{aead: aead}, nil
}

// Seal prefixes a version and random nonce. Purpose binds ciphertext to its
// intended use (e.g. a user's TOTP secret versus an exchange API key).
func (b *Box) Seal(plain []byte, purpose string) ([]byte, error) {
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	out := append([]byte{1}, nonce...)
	return b.aead.Seal(out, nonce, plain, []byte(purpose)), nil
}

func (b *Box) Open(encrypted []byte, purpose string) ([]byte, error) {
	n := b.aead.NonceSize()
	if len(encrypted) < 1+n+b.aead.Overhead() || encrypted[0] != 1 {
		return nil, errors.New("invalid encrypted secret")
	}
	return b.aead.Open(nil, encrypted[1:1+n], encrypted[1+n:], []byte(purpose))
}
