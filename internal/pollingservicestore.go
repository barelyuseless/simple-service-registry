package internal

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type PollingServiceStore struct {
	serviceStore ServiceStore

	pollingMap         map[string]bool
	pollingStatusMutex sync.Mutex

	availabilityMap   map[string]bool
	lastCheckedMap    map[string]time.Time
	lastAvailableMap  map[string]time.Time
	availabilityMutex sync.RWMutex

	pollingInterval time.Duration
	httpClient      *http.Client
}

func NewPollingServiceStore(s ServiceStore, pollingInterval time.Duration) (*PollingServiceStore, error) {
	ps := &PollingServiceStore{
		serviceStore:       s,
		pollingMap:         make(map[string]bool),
		pollingStatusMutex: sync.Mutex{},
		availabilityMap:    make(map[string]bool),
		lastCheckedMap:     make(map[string]time.Time),
		lastAvailableMap:   make(map[string]time.Time),
		availabilityMutex:  sync.RWMutex{},
		pollingInterval:    pollingInterval,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
		},
	}

	existingServiceList, err := ps.serviceStore.getServices()
	if err != nil {
		return nil, err
	}

	for _, service := range existingServiceList {
		if service.HealthcheckURL != "" {
			go ps.startServicePolling(service)
		}
	}

	return ps, nil
}

func (s *PollingServiceStore) addService(service *Service) error {
	if service.HealthcheckURL != "" {
		s.pollingStatusMutex.Lock()
		defer s.pollingStatusMutex.Unlock()

		_, ok := s.pollingMap[service.ServiceURL]
		if !ok {
			go s.startServicePolling(service)
			s.pollingMap[service.ServiceURL] = true
		}
	}

	return s.serviceStore.addService(service)
}

func (s *PollingServiceStore) getServices() ([]*Service, error) {
	serviceSlice, err := s.serviceStore.getServices()
	if err != nil {
		return serviceSlice, err
	}

	s.availabilityMutex.RLock()
	defer s.availabilityMutex.RUnlock()

	for _, service := range serviceSlice {
		status, ok := s.availabilityMap[service.ServiceURL]
		if ok {
			service.Available = status
			service.LastChecked = s.lastCheckedMap[service.ServiceURL]
			service.LastAvailable = s.lastAvailableMap[service.ServiceURL]
		}
	}

	return serviceSlice, nil
}

func (s *PollingServiceStore) startServicePolling(service *Service) {
	s.availabilityMutex.Lock()
	s.lastAvailableMap[service.ServiceURL] = time.Time{}
	s.availabilityMutex.Unlock()

	s.pollService(service)

	for range time.Tick(s.pollingInterval) {
		s.pollService(service)
	}
}

func (s *PollingServiceStore) pollService(service *Service) {
	pollSuccess := false
	resp, err := s.httpClient.Get(service.HealthcheckURL)
	if err == nil {
		if resp.StatusCode == http.StatusOK {
			pollSuccess = true
		}
		resp.Body.Close()
	}

	s.updatePollStatus(service, pollSuccess)
}

func (s *PollingServiceStore) updatePollStatus(service *Service, status bool) {
	s.availabilityMutex.Lock()
	defer s.availabilityMutex.Unlock()

	now := time.Now()

	s.availabilityMap[service.ServiceURL] = status
	s.lastCheckedMap[service.ServiceURL] = now
	if status == true {
		s.lastAvailableMap[service.ServiceURL] = now
	}
}

func (s *PollingServiceStore) Close() error {
	return s.serviceStore.Close()
}

func (s *PollingServiceStore) String() string {
	return fmt.Sprintf("polling (%v) around %s", s.pollingInterval, s.serviceStore)
}
