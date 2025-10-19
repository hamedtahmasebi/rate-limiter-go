package limiter

import (
	"errors"
	"log"
)

var ErrServiceNotFound error = errors.New("service not found")
var ErrClientNotFound error = errors.New("client not found")

type Service struct {
	ID                 string
	UsagePriceInTokens uint64
}

type CreateServiceReqBody struct {
	ID                 string
	UsagePriceInTokens uint64
}

type UpdateServiceReqBody struct {
	UsagePriceInTokens uint64
}

type ServiceRegistry interface {
	CreateService(body CreateServiceReqBody) (Service, error)
	UpdateService(id string, body UpdateServiceReqBody) (Service, error)
	GetService(id string) (Service, error)
}

type ServiceRegistryImpl struct {
	servicesMap map[string]*Service
}

func (sr *ServiceRegistryImpl) CreateService(body CreateServiceReqBody) (Service, error) {
	log.Printf("action=create_service id=%q usage_price_in_tokens=%d", body.ID, body.UsagePriceInTokens)
	sr.servicesMap[body.ID] = &Service{
		ID:                 body.ID,
		UsagePriceInTokens: body.UsagePriceInTokens,
	}
	val := sr.servicesMap[body.ID]
	return *val, nil
}

func (sr *ServiceRegistryImpl) UpdateService(id string, body UpdateServiceReqBody) (Service, error) {
	log.Printf("action=update_service id=%q usage_price_in_tokens=%d", id, body.UsagePriceInTokens)
	s, exists := sr.servicesMap[id]
	if !exists {
		log.Printf("action=update_service id=%q error=%q", id, ErrServiceNotFound)
		return Service{}, ErrServiceNotFound
	}
	s = &Service{
		ID:                 id,
		UsagePriceInTokens: body.UsagePriceInTokens,
	}
	return *s, nil
}

func (sr *ServiceRegistryImpl) GetService(id string) (Service, error) {
	log.Printf("action=get_service id=%q", id)
	s, exists := sr.servicesMap[id]
	if !exists {
		log.Printf("action=get_service id=%q error=%q", id, ErrServiceNotFound)
		return Service{}, ErrServiceNotFound
	}
	return *s, nil
}

func NewServiceRegistry() ServiceRegistry {
	return &ServiceRegistryImpl{
		servicesMap: make(map[string]*Service),
	}
}
