package nsfContext

import (
	"encoding/json"
	"errors"
	"fmt"
	"nsf/internal/db"
	"nsf/logger"
	"sync"
)

var (
	// ErrServiceConflict is returned when a service is already registered by another NF.
	ErrServiceConflict = errors.New("service already registered by another NF")
	// ErrServiceNotFound is returned when a service is not registered.
	ErrServiceNotFound = errors.New("service not registered")
	// ErrServiceForbidden is returned when an NF tries to deregister another NF's service.
	ErrServiceForbidden = errors.New("cannot deregister service owned by another NF")
)

// serviceInfo is the db value of a service. The db key is the service name.
type serviceInfo struct {
	NfName  string `json:"nf_name"`
	SbiAddr string `json:"sbi_addr"`
	SbiPort int    `json:"sbi_port"`
}

// TODO: You may need to add some fields
type serviceContextIE struct {
	db.DbIf

	*logger.NsfLogger
}

// TODO: You may need to add some fields
type serviceContext struct {
	db.DbIf

	// mu guards every access to the service data in db.
	mu sync.RWMutex

	*logger.NsfLogger
}

// TODO: Finish the initialization of the serviceContext instance
func newServiceContext(serviceContextIE *serviceContextIE) *serviceContext {

	return &serviceContext{
		DbIf: serviceContextIE.DbIf,

		NsfLogger: serviceContextIE.NsfLogger,
	}
}

// TODO: Finish the release of the serviceContext instance
func (s *serviceContext) release() {
	s.NsfLogger.SvcLog.Infoln("Release serviceContext...")

	s.NsfLogger.SvcLog.Infoln("serviceContext released")
}

// registerServices stores each service as key, with nfName/sbiAddr/sbiPort as value.
// A service already owned by the same NF is updated. If any service is owned by
// another NF, it returns ErrServiceConflict and nothing in the request is saved.
func (s *serviceContext) registerServices(nfName, sbiAddr string, sbiPort int, services []string) error {
	s.NsfLogger.SvcLog.Tracef("Registering %s services", nfName)

	value, err := json.Marshal(serviceInfo{
		NfName:  nfName,
		SbiAddr: sbiAddr,
		SbiPort: sbiPort,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check all services before writing, so a conflict leaves the db unchanged.
	data := make(map[string]string, len(services))
	for _, service := range services {
		info, exist, err := s.loadServiceInfo(service)
		if err != nil {
			return err
		}

		if exist {
			if info.NfName != nfName {
				return fmt.Errorf("%w: service %s is owned by %s", ErrServiceConflict, service, info.NfName)
			}
			s.NsfLogger.SvcLog.Debugf("Service %s already registered by %s, updating", service, nfName)
		}

		data[service] = string(value)
	}

	// SaveAll writes every service in one db transaction.
	if err := s.DbIf.SaveAll(data); err != nil {
		return fmt.Errorf("failed to save services: %w", err)
	}

	s.NsfLogger.SvcLog.Debugf("Registered %d services for %s", len(data), nfName)
	return nil
}

// loadServiceInfo reads a service from db. The bool is false if the service
// is not registered. The caller must hold s.mu (read or write lock).
func (s *serviceContext) loadServiceInfo(service string) (serviceInfo, bool, error) {
	var info serviceInfo

	exist, err := s.DbIf.Exist(service)
	if err != nil {
		return info, false, fmt.Errorf("failed to check service %s: %w", service, err)
	}
	if !exist {
		return info, false, nil
	}

	stored, err := s.DbIf.Load(service)
	if err != nil {
		return info, false, fmt.Errorf("failed to load service %s: %w", service, err)
	}

	if err := json.Unmarshal([]byte(stored), &info); err != nil {
		return info, false, fmt.Errorf("failed to unmarshal service %s: %w", service, err)
	}

	return info, true, nil
}

// deregisterServices removes services owned by nfName from db. If any service is
// not registered, it returns ErrServiceNotFound. If any service is owned by another
// NF, it returns ErrServiceForbidden. In both cases nothing in the request is removed.
func (s *serviceContext) deregisterServices(nfName string, services []string) error {
	s.NsfLogger.SvcLog.Tracef("De-registering %s services", nfName)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check all services before removing, so a bad request leaves the db unchanged.
	for _, service := range services {
		info, exist, err := s.loadServiceInfo(service)
		if err != nil {
			return err
		}

		if !exist {
			return fmt.Errorf("%w: service %s", ErrServiceNotFound, service)
		}
		if info.NfName != nfName {
			return fmt.Errorf("%w: service %s is owned by %s", ErrServiceForbidden, service, info.NfName)
		}
	}

	for _, service := range services {
		if err := s.DbIf.Remove(service); err != nil {
			return fmt.Errorf("failed to remove service %s: %w", service, err)
		}
	}

	s.NsfLogger.SvcLog.Debugf("De-registered %d services for %s", len(services), nfName)
	return nil
}

// discoverService returns the NF name, SBI address and SBI port of a service.
// It returns ErrServiceNotFound if the service is not registered.
func (s *serviceContext) discoverService(service string) (string, string, int, error) {
	s.NsfLogger.SvcLog.Tracef("Discovering service %s", service)

	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exist, err := s.loadServiceInfo(service)
	if err != nil {
		return "", "", 0, err
	}
	if !exist {
		return "", "", 0, fmt.Errorf("%w: service %s", ErrServiceNotFound, service)
	}

	return info.NfName, info.SbiAddr, info.SbiPort, nil
}
