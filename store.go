package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

const defaultRootFolderName = "ggnetwork"

var mu sync.Mutex

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	blockSize := 5
	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathKey struct {
	PathName string
	Filename string
}

type PathTransformFunc func(string) PathKey

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

type StoreOpts struct {
	Root              string // Root is the folder name of root, containing all the folders/files of the system.
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(id string, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	mu.Lock()
	defer mu.Unlock()
	log.Printf("Checking existence of file: %s", fullPathWithRoot)
	_, err := os.Stat(fullPathWithRoot)

	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	mu.Lock()
	defer mu.Unlock()
	log.Printf("Clearing store at root: %s", s.Root)
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(id string, key string) error {
	pathKey := s.PathTransformFunc(key)
	firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FirstPathName())

	mu.Lock()
	defer mu.Unlock()
	log.Printf("Deleting file: %s", firstPathNameWithRoot)
	defer log.Printf("Deleted [%s] from disk", pathKey.Filename)
	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Write(id string, key string, r io.Reader) (int64, error) {
	log.Printf("Writing key: %s", key)
	return s.writeStream(id, key, r)
}

func (s Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	log.Printf("Decrypting and writing key: %s", key)
	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

func (s *Store) openFileForWriting(id string, key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)

	mu.Lock()
	defer mu.Unlock()
	log.Printf("Creating directories for path: %s", pathNameWithRoot)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	log.Printf("Creating file for path: %s", fullPathWithRoot)
	return os.Create(fullPathWithRoot)
}

func (s *Store) writeStream(id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	log.Printf("Writing data to file for key: %s", key)
	return io.Copy(f, r)
}

func (s *Store) Read(id string, key string) (int64, io.Reader, error) {
	log.Printf("Reading key: %s", key)
	return s.readStream(id, key)
}

func (s *Store) readStream(id string, key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	mu.Lock()
	defer mu.Unlock()
	log.Printf("Opening file for path: %s", fullPathWithRoot)
	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		file.Close()
		return 0, nil, err
	}

	return fi.Size(), file, nil
}
