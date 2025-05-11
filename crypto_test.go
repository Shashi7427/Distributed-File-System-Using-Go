package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypy(t *testing.T) {
	payload := "hello world"
	key := newEncryptionKey()
	src := bytes.NewBufferString(payload)
	dst := &bytes.Buffer{}
	// Encrypt
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	fmt.Printf("Encrypted data: %x\n", dst.Bytes())
	// Decrypt
	decrypted := &bytes.Buffer{}
	nw, err := copyDecrypt(key, dst, decrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	fmt.Printf("Decrypted data: %s\n", decrypted.String())

	if nw != 16+len(payload) {
		t.Fatalf("expected %d bytes, got %d", len(payload), nw)
	}

	if decrypted.String() != payload {
		t.Errorf("expected 'hello world', got '%s'", decrypted.String())
	}
}
