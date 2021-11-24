package storage

import (
	"io/ioutil"
	"path"
)

type Storage interface {
	Read(fileName string) ([]byte, error)
}

type FileStorage struct {
	Directory string
}

func (f *FileStorage) Read(fileName string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(f.Directory, fileName))
}
