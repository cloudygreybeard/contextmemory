package providers

import (
	"fmt"
)

// TODO: Implement when cloud storage is needed

// S3Provider skeleton for AWS S3 storage
type S3Provider struct {
	config ProviderConfig
}

func NewS3Provider(config ProviderConfig) (StorageProvider, error) {
	return nil, fmt.Errorf("S3 provider not yet implemented - coming in future release")
}

// GCSProvider skeleton for Google Cloud Storage
type GCSProvider struct {
	config ProviderConfig
}

func NewGCSProvider(config ProviderConfig) (StorageProvider, error) {
	return nil, fmt.Errorf("GCS provider not yet implemented - coming in future release")
}

// RemoteProvider skeleton for remote HTTP API storage
type RemoteProvider struct {
	config ProviderConfig
}

func NewRemoteProvider(config ProviderConfig) (StorageProvider, error) {
	return nil, fmt.Errorf("Remote provider not yet implemented - coming in future release")
}

// Example of what S3Provider would implement when ready:

/*
func (s *S3Provider) Create(req storage.CreateMemoryRequest) (*storage.Memory, error) {
	// Implementation would:
	// 1. Generate memory with ID
	// 2. Upload to S3 bucket with key: keyPrefix + memoryID + ".json"
	// 3. Update index object in S3
	// 4. Return memory
	return nil, fmt.Errorf("not implemented")
}

func (s *S3Provider) Get(id string) (*storage.Memory, error) {
	// Implementation would:
	// 1. Download from S3 with key: keyPrefix + id + ".json"
	// 2. Parse JSON to Memory struct
	// 3. Return memory
	return nil, fmt.Errorf("not implemented")
}

func (s *S3Provider) GetProviderType() ProviderType {
	return S3Provider
}

func (s *S3Provider) GetProviderInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":      "s3",
		"bucket":    s.config.Bucket,
		"region":    s.config.Region,
		"keyPrefix": s.config.KeyPrefix,
		"provider":  "aws-s3",
	}
}

func (s *S3Provider) ValidateConfig() error {
	if s.config.Bucket == "" {
		return NewProviderConfigError(S3Provider, "bucket", "bucket name is required")
	}
	if s.config.Region == "" {
		return NewProviderConfigError(S3Provider, "region", "region is required")
	}

	// Test S3 connectivity
	// return s.testS3Connection()
	return nil
}
*/
