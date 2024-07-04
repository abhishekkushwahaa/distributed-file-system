package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"testing"
)

func TestPathTranformFunc(t *testing.T) {
	key := "naturecoolpicture"
	pathKey := CASPathTransformFunc(key)
	expectedFilename := "75260fd17a51c8a6c68ea682a194e876bf9db965"
	expectedPathName := "75260/fd17a/51c8a/6c68e/a682a/194e8/76bf9/db965"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.Filename != expectedFilename {
		t.Errorf("have %s want %s", pathKey.Filename, expectedFilename)
	}
}

func TestStore(t *testing.T) {
	log.Println("Starting TestStore")
	s := newStore()
	id := generateID()

	defer func() {
		log.Println("Teardown")
		teardown(t, s)
	}()

	for i := 0; i < 5; i++ { // Reduce the number of iterations to 10
		log.Printf("Iteration: %d", i)
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("Some jpg bytes")

		log.Printf("Writing key: %s", key)
		if _, err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
			t.Errorf("Error writing key %s: %v", key, err)
			return
		}

		log.Printf("Checking existence of key: %s", key)
		if ok := s.Has(id, key); !ok {
			t.Errorf("Expected to have key %s", key)
			return
		}

		log.Printf("Reading key: %s", key)
		_, r, err := s.Read(id, key)
		if err != nil {
			t.Errorf("Error reading key %s: %v", key, err)
			return
		}

		log.Printf("Reading data for key: %s", key)
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Errorf("Error reading data for key %s: %v", key, err)
			return
		}

		log.Printf("Data read for key %s: %s", key, b)
		if string(b) != string(data) {
			t.Errorf("Expected data %s, got %s", data, b)
			return
		}

		if rc, ok := r.(io.ReadCloser); ok {
			log.Printf("Closing reader for key: %s", key)
			rc.Close()
		}

		log.Printf("Deleting key: %s", key)
		if err := s.Delete(id, key); err != nil {
			t.Errorf("Error deleting key %s: %v", key, err)
			return
		}

		log.Printf("Checking non-existence of key: %s", key)
		if ok := s.Has(id, key); ok {
			t.Errorf("Expected to not have key %s", key)
			return
		}
	}
	log.Println("TestStore completed")
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
