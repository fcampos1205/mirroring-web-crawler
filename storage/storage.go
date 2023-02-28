package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Storage is an interface for storing and retrieving files.
type Storage interface {
	Store(filePath string, data []byte) error
	Retrieve(filename string) ([]byte, error)
}

// DiskStorage stores files on disk.
type DiskStorage struct {
	path string
}

// NewDiskStorage creates a new instance of the DiskStorage with the given path.
func NewDiskStorage(path string) Storage {
	return &DiskStorage{
		path: path,
	}
}

// Store stores a file on disk.
func (s *DiskStorage) Store(filePath string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory: %s", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing file: %s", err)
	}

	return nil
}

// Retrieve retrieves a file from disk.
func (s *DiskStorage) Retrieve(filePath string) ([]byte, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		return nil, fmt.Errorf("error reading file: %s", err)
	}

	return data, nil
}
