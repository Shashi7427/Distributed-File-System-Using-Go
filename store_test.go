package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestCASPathTransformFunc(t *testing.T) {
	key := "niyati"
	pathname := CASPathTransformFunc(key)
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
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	k := "niyati"
	s := NewStore(opts)
	data := []byte("some jpg bytes")
	if err := s.writeStream(k, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	r, err := s.Read(k)

	if err != nil {
		t.Error(err)
	}
	buf, _ := io.ReadAll(r)
	if string(buf) != string(data) {
		t.Errorf("expected %s, got %s", data, buf)
	}
}
