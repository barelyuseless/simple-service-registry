package internal

import "sync"

type ServiceStore interface {
	addService(*Service) error
	getServices() ([]*Service, error)
	Close() error
	String() string
}

type MapServiceStore struct {
	mutex    sync.RWMutex
	services map[string]*Service
}

func NewMapServiceStore() *MapServiceStore {
	return &MapServiceStore{
		mutex:    sync.RWMutex{},
		services: make(map[string]*Service),
	}
}

func (s *MapServiceStore) addService(service *Service) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.services[service.ServiceURL] = service

	return nil
}

func (s *MapServiceStore) getServices() ([]*Service, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	serviceSlice := make([]*Service, len(s.services))

	i := 0
	for _, v := range s.services {
		serviceSlice[i] = v
		i++
	}

	return serviceSlice, nil
}

func (s *MapServiceStore) Close() error {
	return nil
}

func (s *MapServiceStore) String() string {
	return "map"
}
