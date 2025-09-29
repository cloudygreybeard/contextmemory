package providers

import (
	"fmt"
)

// TODO: Cloud storage providers - planned for future releases
// These will provide distributed storage options for team collaboration
// and backup scenarios. Current implementation focuses on local file storage.

// Cloud provider constructors (not yet implemented)
func NewS3Provider(config ProviderConfig) (interface{}, error) {
	return nil, fmt.Errorf("S3 provider not yet implemented - planned for v1.1.0")
}

func NewGCSProvider(config ProviderConfig) (interface{}, error) {
	return nil, fmt.Errorf("GCS provider not yet implemented - planned for v1.1.0")
}

func NewRemoteProvider(config ProviderConfig) (interface{}, error) {
	return nil, fmt.Errorf("remote API provider not yet implemented - planned for v1.2.0")
}