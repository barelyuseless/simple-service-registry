package internal

import (
	"net/url"
	"time"
)

type Service struct {
	ServiceURL     string    `json:"service_url" yaml:"service_url"`
	SpecURL        string    `json:"spec_url" yaml:"spec_url"`
	Tags           []string  `json:"tags" yaml:"tags"`
	Description    string    `json:"description" yaml:"description"`
	HealthcheckURL string    `json:"healthcheck_url" yaml:"healthcheck_url"`
	Available      bool      `json:"health_confirmed" yaml:"health_confirmed"`
	LastChecked    time.Time `json:"last_checked,omitempty" yaml:"last_checked,omitempty"`
	LastAvailable  time.Time `json:"last_available,omitempty" yaml:"last_available,omitempty"`
}

func (s Service) validate() bool {
	if s.ServiceURL == "" || s.SpecURL == "" || s.Description == "" {
		return false
	}

	_, err := url.ParseRequestURI(s.ServiceURL)
	if err != nil {
		return false
	}

	_, err = url.ParseRequestURI(s.SpecURL)
	if err != nil {
		return false
	}

	if s.HealthcheckURL != "" {
		_, err = url.ParseRequestURI(s.HealthcheckURL)
		if err != nil {
			return false
		}
	}

	return true
}
