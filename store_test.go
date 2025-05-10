package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestCASPathTransformFunc(t *testing.T) {
	key := "niyati"
	pathname := CASPathTransformFunc(key, DefaultRootPath)
	fmt.Println(pathname)
}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	k := "niyati"
	s := NewStore(opts)
	if err := s.Delete(k); err != nil {
		t.Error(err)
	}

	if s.Has(k) {
		t.Errorf("expected %s to be deleted", k)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	k := "niyati"
	data := []byte("some jpg bytes")
	if _, err := s.writeStream(k, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	_, r, err := s.Read(k)

	if err != nil {
		t.Error(err)
	}
	buf, _ := io.ReadAll(r)
	if string(buf) != string(data) {
		t.Errorf("expected %s, got %s", data, buf)
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func tearDown(t *testing.T, s *Store) {
	// Clean up the test directory
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
