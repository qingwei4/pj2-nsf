package nsfContext

import (
	"encoding/json"
	"errors"
	"fmt"
	"nsf/internal/db"
	"nsf/logger"
	"sync"
)

// ErrServiceConflict is returned when a service is already registered by another NF.
var ErrServiceConflict = errors.New("service already registered by another NF")

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
		exist, err := s.DbIf.Exist(service)
		if err != nil {
			return fmt.Errorf("failed to check service %s: %w", service, err)
		}

		if exist {
			stored, err := s.DbIf.Load(service)
			if err != nil {
				return fmt.Errorf("failed to load service %s: %w", service, err)
			}

			var info serviceInfo
			if err := json.Unmarshal([]byte(stored), &info); err != nil {
				return fmt.Errorf("failed to unmarshal service %s: %w", service, err)
			}

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

// TODO: Finish the implementation of the deregisterServices method with an error return type
func (s *serviceContext) deregisterServices(nfName string, services []string) error {
	s.NsfLogger.SvcLog.Tracef("De-registering %s services", nfName)

	s.NsfLogger.SvcLog.Warnln("deregisterServices: not implemented yet")
	return fmt.Errorf("deregisterServices: not implemented yet")
}

// TODO: Finish the implementation of the discoverService method with these return values:
// (string, string, int, error) representing the NF name, SBI address, SBI port, and an error if any
func (s *serviceContext) discoverService(service string) (string, string, int, error) {
	s.NsfLogger.SvcLog.Tracef("Discovering service %s", service)

	s.NsfLogger.SvcLog.Warnln("discoverService: not implemented yet")
	return "", "", 0, fmt.Errorf("discoverService: not implemented yet")
}
