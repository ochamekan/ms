package discovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type Registry interface {
	// Register creates a service instance record in the registry.
	Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error
	// Deregister removes a service instance record from the registry.
	Deregister(ctx context.Context, instanceID string, serviceName string) error
	// ServiceAddresses return the list of addresses of active instances of the given service.
	ServiceAddresses(ctx context.Context, serviceID string) ([]string, error)
	// ReportHealthyState is a push mechanism for reporting healthy state to the registry.
	ReportHealthyState(instanceID string, serviceName string) error
}

var ErrNotFound = errors.New("no service addresses found")

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
