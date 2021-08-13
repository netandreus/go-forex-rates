package provider

import (
	"errors"
)

// Registry is providers container
type Registry struct {
	providers map[string]RatesProvider
}

// BuildRegistry constructor
func BuildRegistry() (*Registry, error) {
	var registry = &Registry{
		providers: make(map[string]RatesProvider),
	}
	return registry, nil
}

// AddProvider add provider instance to Registry
func (r *Registry) AddProvider(provider RatesProvider) {
	code := provider.GetCode()
	r.providers[code] = provider
}

// GetProvider returns provider's instance from Registry by code
func (r *Registry) GetProvider(code string) (RatesProvider, error) {
	for providerCode, provider := range r.providers {
		if providerCode == code {
			return provider, nil
		}
	}
	return nil, errors.New("provider with code " + code + "does not registered")
}
