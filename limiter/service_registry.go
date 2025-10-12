package limiter

import "errors"

var ErrServiceNotFound error = errors.New("Service not found")

type Service struct {
	ID                 string
	UsagePriceInTokens uint64
}

type CreateServiceReqBody struct {
	ID                 string
	UsagePriceInTokens uint64
}

type UpdateServiceReqBody struct {
	usagePriceInTokens uint64
}

type ServiceRegistry interface {
	CreateService(body CreateServiceReqBody) (Service, error)
	UpdateService(id string, body UpdateServiceReqBody) (Service, error)
	GetService(id string) (Service, error)
}

type ServiceRegistryImpl struct {
	servicesMap map[string]Service
}

func (sr *ServiceRegistryImpl) CreateService(body CreateServiceReqBody) (Service, error) {
	sr.servicesMap[body.ID] = Service{
		ID:                 body.ID,
		UsagePriceInTokens: body.UsagePriceInTokens,
	}
	val := sr.servicesMap[body.ID]
	return val, nil
}

func (sr *ServiceRegistryImpl) UpdateService(id string, body UpdateServiceReqBody) (Service, error) {
	s, exists := sr.servicesMap[id]
	if !exists {
		return Service{}, ErrServiceNotFound
	}
	s = Service{
		ID:                 id,
		UsagePriceInTokens: body.usagePriceInTokens,
	}
	return s, nil
}

func (sr *ServiceRegistryImpl) GetService(id string) (Service, error) {
	s, exists := sr.servicesMap[id]
	if !exists {
		return Service{}, ErrServiceNotFound
	}
	return s, nil
}

func NewServiceRegistry() ServiceRegistry {
	return &ServiceRegistryImpl{
		servicesMap: make(map[string]Service),
	}
}
