package providers

import (
	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/storage"
)

// FileStorageProvider implements file-based storage
type FileStorageProvider struct {
	*storage.FileStorage
	config ProviderConfig
}

// NewFileProvider creates a new file storage provider
func NewFileProvider(config ProviderConfig) (StorageProvider, error) {
	fileStorage, err := storage.NewFileStorage(config.StorageDir)
	if err != nil {
		return nil, err
	}

	return &FileStorageProvider{
		FileStorage: fileStorage,
		config:      config,
	}, nil
}

// GetProviderType returns the provider type
func (f *FileStorageProvider) GetProviderType() ProviderType {
	return FileProvider
}

// GetProviderInfo returns provider-specific information
func (f *FileStorageProvider) GetProviderInfo() map[string]interface{} {
	info, _ := f.FileStorage.GetStorageInfo()
	return map[string]interface{}{
		"type":          "file",
		"storageDir":    info.StorageDir,
		"memoriesCount": info.MemoriesCount,
		"totalSize":     info.TotalSize,
		"provider":      "local-filesystem",
	}
}

// ValidateConfig validates the file provider configuration
func (f *FileStorageProvider) ValidateConfig() error {
	// File provider validation - basic health check
	return f.FileStorage.Health()
}
