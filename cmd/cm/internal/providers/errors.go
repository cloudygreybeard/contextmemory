package providers

import "fmt"

// UnsupportedProviderError represents an error for unsupported providers
type UnsupportedProviderError struct {
	ProviderType ProviderType
}

func (e *UnsupportedProviderError) Error() string {
	return fmt.Sprintf("unsupported provider type: %s", e.ProviderType)
}

// NewUnsupportedProviderError creates a new unsupported provider error
func NewUnsupportedProviderError(providerType ProviderType) *UnsupportedProviderError {
	return &UnsupportedProviderError{ProviderType: providerType}
}

// ProviderConfigError represents a configuration error
type ProviderConfigError struct {
	ProviderType ProviderType
	Field        string
	Message      string
}

func (e *ProviderConfigError) Error() string {
	return fmt.Sprintf("provider %s config error in field '%s': %s", e.ProviderType, e.Field, e.Message)
}

// NewProviderConfigError creates a new provider configuration error
func NewProviderConfigError(providerType ProviderType, field, message string) *ProviderConfigError {
	return &ProviderConfigError{
		ProviderType: providerType,
		Field:        field,
		Message:      message,
	}
}
