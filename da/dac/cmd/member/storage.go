package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
  Store(id string, r io.Reader) error
  Fetch(id string) (io.ReadCloser, error)
}

var ErrNotFound = errors.New("id not found")

type fileStorage struct {
  Directory string
}

func newFileStorage(dir string) fileStorage {
  return fileStorage{dir}
}

func (s fileStorage) computePath(id string) string {
  return filepath.Join(s.Directory, id)
}

func (s fileStorage) Store(id string, r io.Reader) error {
  path := s.computePath(id)
  file, err := os.Create(path)
  if err != nil {
    return err
  }

  _, err = io.Copy(file, r)
  return err
}

func (s fileStorage) Fetch(id string) (io.ReadCloser, error) {
  path := s.computePath(id)
  file, err := os.Open(path)
  if err != nil {
    if err == os.ErrNotExist {
      return nil, ErrNotFound
    }
    return nil, err
  }

  return file, nil
}
