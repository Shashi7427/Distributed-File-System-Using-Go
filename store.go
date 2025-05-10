package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var DefaultRootPath = "shashiFiles"

type PathTransformFunc func(string, string) PathKey

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

func CASPathTransformFunc(key string, root string) PathKey {
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
		PathName: root + "/" + strings.Join(path, "/"),
		FileName: hashStr,
	}
}

var DefaultPathTransformFunc = func(key string, root string) PathKey {
	return PathKey{
		PathName: root + "/" + key,
		FileName: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if len(opts.RootPath) == 0 {
		opts.RootPath = DefaultRootPath
	} else {
		DefaultRootPath = opts.RootPath
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
	pathName := s.PathTransformFunc(key, s.RootPath)
	_, err := os.Stat(pathName.FullPath())
	return !os.IsNotExist(err)

}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.RootPath)
}

func (s *Store) Delete(key string) error {
	pathName := s.PathTransformFunc(key, s.RootPath)

	defer func() {
		fmt.Printf("deleted the file %s\n", pathName.FullPath())
	}()

	return os.RemoveAll(FirstPathName(pathName.FullPath()))
	// return os.RemoveAll(pathName.PathName)
}

// Delete removes the file at the given key path
// >>> ?? done : instead of copying the file directly to a reader
// first copy this into a buffer and then return the file from the readstream
func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
	// if err != nil {
	// 	return n, nil, err
	// }
	// defer f.Close()
	// buf := new(bytes.Buffer)
	// _, err = io.Copy(buf, f)
	// return n, buf, err
}
func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathName := s.PathTransformFunc(key, s.RootPath)

	file, err := os.Open(pathName.FullPath())
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	n := fi.Size()
	return n, file, nil
}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	pathName := s.PathTransformFunc(key, s.RootPath)
	if err := os.MkdirAll(pathName.PathName, 0755); err != nil {
		return 0, err
	}
	// log.Printf("pathName: %s\n", pathName)
	f, err := os.Create(pathName.FullPath())
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, nil
	}
	log.Printf("written %d bytes to %s\n", n, pathName.FullPath())
	return n, nil
}
