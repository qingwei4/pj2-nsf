package processor

import (
	"fmt"
	"net/http"
	"nsf/model"
)

func (p *Processor) ProcessDiscoveryService(request model.NnsfDiscGetRequest) (*model.NnsfDiscGetResponse, *model.HttpError) {
	p.NsfLogger.ProcLog.Debugln("Processing discovery service")

	nfName, sbiAddress, sbiPort, err := p.NsfContext.DiscoverService(request.Service)
	if err != nil {
		p.NsfLogger.ProcLog.Errorf("Failed to discover service: %v", err)
		return nil, &model.HttpError{
			StatusCode:  http.StatusNotFound,
			ErrorDetail: fmt.Errorf("failed to discover service: %v", err),
		}
	}

	p.NsfLogger.ProcLog.Debugln("Discovery service processed successfully")
	return &model.NnsfDiscGetResponse{
		Service:    request.Service,
		NfName:     nfName,
		SbiAddress: sbiAddress,
		SbiPort:    sbiPort,
	}, nil
}

func (p *Processor) ProcessDiscoveryServices(request model.NnsfDiscGroupGetRequest) (*model.NnsfDiscGroupGetResponse, *model.HttpError) {
	p.NsfLogger.ProcLog.Debugln("Processing discovery services")

	services := make([]model.NnsfDiscGetResponse, 0, len(request.Services))
	for _, service := range request.Services {
		nfName, sbiAddress, sbiPort, err := p.NsfContext.DiscoverService(service)
		if err != nil {
			p.NsfLogger.ProcLog.Errorf("Failed to discover service: %v", err)
			continue
		}

		services = append(services, model.NnsfDiscGetResponse{
			Service:    service,
			NfName:     nfName,
			SbiAddress: sbiAddress,
			SbiPort:    sbiPort,
		})
	}

	if len(services) == 0 {
		p.NsfLogger.ProcLog.Errorf("Failed to discover any service in %v", request.Services)
		return nil, &model.HttpError{
			StatusCode:  http.StatusNotFound,
			ErrorDetail: fmt.Errorf("failed to discover any service in %v", request.Services),
		}
	}

	p.NsfLogger.ProcLog.Debugln("Discovery services processed successfully")
	return &model.NnsfDiscGroupGetResponse{
		Services: services,
	}, nil
}