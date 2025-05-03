package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var DefaultRootPath = "shashiFiles"

type PathTransformFunc func(string) PathKey

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
	RootPath          string
}

type PathKey struct {
	PathName string
	FileName string
}

func FirstPathName(key string) string {
	firstPath := strings.Split(key, "/")
	if len(firstPath) == 0 {
		return ""
	}
	return firstPath[0] + "/" + firstPath[1]
}

func (p *PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	path := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		path[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: DefaultRootPath + "/" + strings.Join(path, "/"),
		FileName: hashStr,
	}
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: DefaultRootPath + "/" + key,
		FileName: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if len(opts.RootPath) == 0 {
		opts.RootPath = DefaultRootPath
	}
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	return &Store{
		StoreOpts: opts,
	}
}

// Store is a simple file store that uses the file system to store files

func (s *Store) Has(key string) bool {
	pathName := s.PathTransformFunc(key)
	_, err := os.Stat(pathName.FullPath())
	return !os.IsNotExist(err)

}

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.RootPath)
}

func (s *Store) Delete(key string) error {
	pathName := s.PathTransformFunc(key)

	defer func() {
		fmt.Printf("deleted the file %s\n", pathName.FullPath())
	}()

	return os.RemoveAll(FirstPathName(pathName.FullPath()))
	// return os.RemoveAll(pathName.PathName)
}

// Delete removes the file at the given key path

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathName := s.PathTransformFunc(key)
	return os.Open(pathName.FullPath())
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)
	if err := os.MkdirAll(pathName.PathName, 0755); err != nil {
		return err
	}
	// log.Printf("pathName: %s\n", pathName)
	f, err := os.Create(pathName.FullPath())
	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return nil
	}
	log.Printf("written %d bytes to %s\n", n, pathName.FullPath())
	return nil
}
