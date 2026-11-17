package cryptobox

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestEncryption(t *testing.T) {
	b, err := New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("TOTPSECRET")
	a, err := b.Seal(plain, "totp:user-1")
	if err != nil {
		t.Fatal(err)
	}
	other, err := b.Seal(plain, "totp:user-1")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a, other) || bytes.Contains(a, plain) {
		t.Fatal("ciphertext must use a fresh nonce")
	}
	got, err := b.Open(a, "totp:user-1")
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("roundtrip: %s %v", got, err)
	}
	if _, err = b.Open(a, "exchange:user-1"); err == nil {
		t.Fatal("wrong purpose accepted")
	}
	a[len(a)-1] ^= 1
	if _, err = b.Open(a, "totp:user-1"); err == nil {
		t.Fatal("tampering accepted")
	}
	for _, value := range [][]byte{nil, {1}, {2, 3, 4}} {
		if _, err = b.Open(value, "totp:user-1"); err == nil {
			t.Fatal("malformed ciphertext accepted")
		}
	}
	for _, key := range []string{"", "not-base64", base64.StdEncoding.EncodeToString(make([]byte, 16))} {
		if _, err = New(key); err == nil {
			t.Fatal("invalid key accepted")
		}
	}
}
