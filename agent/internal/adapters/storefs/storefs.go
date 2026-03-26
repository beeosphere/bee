package storefs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

// type Store interface {
// 	ReadMeta(key string, v any) error
// 	ReadData(key string) ([]byte, error)
// 	Write(key string, data []byte, metadata any) error
// 	Delete(key string) error
// 	DeleteAll() error
// }

type FileSystemStore struct {
	path string
	log  models.Logger
}

func NewFileSystemStore(session *core.Session, basePath string) *FileSystemStore {
	// Create a path that is unique per bee/pubKey
	uniquePath := filepath.Join(basePath, session.Bee, session.PublicKey)
	return &FileSystemStore{
		path: uniquePath,
		log:  session.Log,
	}
}

// BasePath returns the base path of the store
func (fs *FileSystemStore) BasePath() string {
	return fs.path
}

// ReadMeta retrieves metadata for the given key and deserializes into the provided struct pointer
// If no metadata file exists, v is left unchanged and no error is returned
func (fs *FileSystemStore) ReadMeta(key string, v any) error {
	if v == nil {
		return fmt.Errorf("cannot deserialize into nil pointer")
	}

	metaPath := filepath.Join(fs.path, key+".meta.json")
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No metadata is fine
		}
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	if err = json.Unmarshal(metaBytes, v); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	fs.log.Tracef("read metadata for key: %s", key)
	return nil
}

// ReadData retrieves data for the given key
func (fs *FileSystemStore) ReadData(key string) ([]byte, error) {
	dataPath := filepath.Join(fs.path, key+".data")
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}
	fs.log.Tracef("read data for key: %s", key)
	return data, nil
}

// Write stores data and metadata for the given key
// Metadata can be any struct that will be serialized to JSON
func (fs *FileSystemStore) Write(key string, data []byte, metadata any) error {
	// Ensure base path exists
	if err := os.MkdirAll(fs.path, 0755); err != nil {
		return fmt.Errorf("failed to create base path: %w", err)
	}

	// Write data file
	dataPath := filepath.Join(fs.path, key+".data")
	if err := os.WriteFile(dataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}

	// Write metadata file only if metadata is provided
	if metadata != nil {
		metaBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		metaPath := filepath.Join(fs.path, key+".meta.json")
		if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
			return fmt.Errorf("failed to write metadata file: %w", err)
		}
	}
	fs.log.Tracef("stored key: %s", key)
	return nil
}

// Delete removes both data and metadata files for the given key
func (fs *FileSystemStore) Delete(key string) error {
	dataPath := filepath.Join(fs.path, key+".data")
	metaPath := filepath.Join(fs.path, key+".meta.json")

	// Delete data file
	if err := os.Remove(dataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete data file: %w", err)
	}

	// Delete metadata file
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata file: %w", err)
	}

	fs.log.Tracef("deleted key: %s", key)
	return nil
}

// DeleteAll removes all files in the base path
func (fs *FileSystemStore) DeleteAll() error {
	// Read directory contents
	entries, err := os.ReadDir(fs.path)
	if err != nil {
		if os.IsNotExist(err) {
			fs.log.Tracef("Nothing to delete in path: %s", fs.path)
			return nil // Nothing to delete
		}
		return fmt.Errorf("failed to read directory: %w", err)
	}
	// Delete each file
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(fs.path, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", entry.Name(), err)
			}
		}
	}
	fs.log.Tracef("deleted all keys in path: %s", fs.path)
	return nil
}
