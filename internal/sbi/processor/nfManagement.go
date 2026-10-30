package processor

import (
	"errors"
	"fmt"
	"net/http"
	"nsf/internal/nsfContext"
	"nsf/model"
	"time"

	"github.com/free-ran-ue/util"
	"github.com/google/uuid"
)

func (p *Processor) ProcessServicesRegistration(request model.NnsfNfmPostRequest) (*model.NnsfNfmPostResponse, *model.HttpError) {
	p.NsfLogger.ProcLog.Debugln("Processing services registration")

	if err := p.NsfContext.RegisterServices(request.NfName, request.SbiAddress, request.SbiPort, request.Services); err != nil {
		p.NsfLogger.ProcLog.Errorf("Failed to register services: %v", err)
		statusCode := http.StatusInternalServerError
		if errors.Is(err, nsfContext.ErrServiceConflict) {
			statusCode = http.StatusConflict
		}
		return nil, &model.HttpError{
			StatusCode:  statusCode,
			ErrorDetail: fmt.Errorf("failed to register services: %v", err),
		}
	}

	nfUuid := uuid.New()
	token, err := util.CreateJWT("nsf", request.NfName, 24*time.Hour, nil)
	if err != nil {
		p.NsfLogger.ProcLog.Errorf("Failed to create token: %v", err)
		return nil, &model.HttpError{
			StatusCode:  http.StatusInternalServerError,
			ErrorDetail: fmt.Errorf("failed to create token: %v", err),
		}
	}

	response := &model.NnsfNfmPostResponse{
		ServiceToken: token,
		ServiceUuid:  nfUuid.String(),
	}

	p.NsfLogger.ProcLog.Debugln("Services registration processed successfully")
	return response, nil
}

func (p *Processor) ProcessServicesDeregistration(request model.NnsfNfmDeleteRequest) *model.HttpError {
	p.NsfLogger.ProcLog.Debugln("Processing services deregistration")

	if err := p.NsfContext.DeregisterServices(request.NfName, request.Services); err != nil {
		p.NsfLogger.ProcLog.Errorf("Failed to deregister services: %v", err)
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, nsfContext.ErrServiceNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, nsfContext.ErrServiceForbidden):
			statusCode = http.StatusForbidden
		}
		return &model.HttpError{
			StatusCode:  statusCode,
			ErrorDetail: fmt.Errorf("failed to deregister services: %v", err),
		}
	}

	p.NsfLogger.ProcLog.Debugln("Services deregistration processed successfully")
	return nil
}
