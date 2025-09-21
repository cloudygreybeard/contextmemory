package providers

import (
	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/internal/storage"
)

// ProviderType represents different storage backend types
type ProviderType string

const (
	FileProvider   ProviderType = "file"
	S3Provider     ProviderType = "s3"
	GCSProvider    ProviderType = "gcs"
	RemoteProvider ProviderType = "remote"
)

// ProviderConfig contains configuration for storage providers
type ProviderConfig struct {
	Type ProviderType `yaml:"type" json:"type"`

	// File provider config
	StorageDir string `yaml:"storageDir,omitempty" json:"storageDir,omitempty"`

	// Cloud provider config
	Bucket    string `yaml:"bucket,omitempty" json:"bucket,omitempty"`
	Region    string `yaml:"region,omitempty" json:"region,omitempty"`
	KeyPrefix string `yaml:"keyPrefix,omitempty" json:"keyPrefix,omitempty"`

	// Remote provider config
	Endpoint string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	APIKey   string            `yaml:"apiKey,omitempty" json:"apiKey,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// Common config
	Timeout    int  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	RetryCount int  `yaml:"retryCount,omitempty" json:"retryCount,omitempty"`
	EnableTLS  bool `yaml:"enableTLS,omitempty" json:"enableTLS,omitempty"`
}

// StorageProvider interface that all storage backends must implement
type StorageProvider interface {
	storage.FileStorage // Embed the existing interface for now

	// Provider-specific methods
	GetProviderType() ProviderType
	GetProviderInfo() map[string]interface{}
	ValidateConfig() error
}

// ProviderFactory creates storage providers based on configuration
type ProviderFactory struct {
	providers map[ProviderType]func(ProviderConfig) (StorageProvider, error)
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	factory := &ProviderFactory{
		providers: make(map[ProviderType]func(ProviderConfig) (StorageProvider, error)),
	}

	// Register built-in providers
	factory.RegisterProvider(FileProvider, NewFileProvider)
	// TODO: Register cloud providers when implemented
	// factory.RegisterProvider(S3Provider, NewS3Provider)
	// factory.RegisterProvider(GCSProvider, NewGCSProvider)
	// factory.RegisterProvider(RemoteProvider, NewRemoteProvider)

	return factory
}

// RegisterProvider registers a new storage provider
func (f *ProviderFactory) RegisterProvider(providerType ProviderType, constructor func(ProviderConfig) (StorageProvider, error)) {
	f.providers[providerType] = constructor
}

// CreateProvider creates a storage provider instance
func (f *ProviderFactory) CreateProvider(config ProviderConfig) (StorageProvider, error) {
	constructor, exists := f.providers[config.Type]
	if !exists {
		return nil, NewUnsupportedProviderError(config.Type)
	}

	provider, err := constructor(config)
	if err != nil {
		return nil, err
	}

	// Validate configuration
	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// GetSupportedProviders returns list of supported provider types
func (f *ProviderFactory) GetSupportedProviders() []ProviderType {
	var providers []ProviderType
	for providerType := range f.providers {
		providers = append(providers, providerType)
	}
	return providers
}

// GetProviderDefaults returns default configuration for a provider type
func GetProviderDefaults(providerType ProviderType) ProviderConfig {
	switch providerType {
	case FileProvider:
		return ProviderConfig{
			Type:       FileProvider,
			StorageDir: "", // Will default to ~/.contextmemory
			Timeout:    30,
		}
	case S3Provider:
		return ProviderConfig{
			Type:       S3Provider,
			Region:     "us-east-1",
			KeyPrefix:  "contextmemory/",
			Timeout:    30,
			RetryCount: 3,
			EnableTLS:  true,
		}
	case GCSProvider:
		return ProviderConfig{
			Type:       GCSProvider,
			KeyPrefix:  "contextmemory/",
			Timeout:    30,
			RetryCount: 3,
			EnableTLS:  true,
		}
	case RemoteProvider:
		return ProviderConfig{
			Type:       RemoteProvider,
			Timeout:    30,
			RetryCount: 3,
			EnableTLS:  true,
		}
	default:
		return ProviderConfig{Type: FileProvider}
	}
}
