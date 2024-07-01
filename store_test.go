package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestPathTranformFunc(t *testing.T) {
	key := "naturecoolpicture"
	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "75260fd17a51c8a6c68ea682a194e876bf9db965"
	expectedPathName := "75260/fd17a/51c8a/6c68e/a682a/194e8/76bf9/db965"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.Filename !=  expectedPathName {
		t.Errorf("have %s want %s", pathKey.Filename, expectedOriginalKey)
	}
	// fmt.Println(pathname)
}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "myspecialpicture"
	data := []byte("Some jpg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "myspecialpicture"
	data := []byte("Some jpg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := ioutil.ReadAll(r)

	if string(b) != string(data) {
		t.Errorf("want %s have %s", data, b)
	}

	s.Delete(key)
}