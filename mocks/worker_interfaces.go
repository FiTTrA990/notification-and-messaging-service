// Package mocks – additional interface aliases to support queue worker tests
package mocks

import (
	"context"

	"github.com/FiTTrA990/notification-and-messaging-service/internal/domain"
)

// MockMessageConsumerInterface is used as a parameter type in worker tests.
// It is satisfied by *MockMessageConsumer.
type MockMessageConsumerInterface interface {
	Consume(ctx context.Context, handler func(context.Context, *domain.EventMessage) error) error
	Close() error
}

// MockNotificationServiceInterface is used as a parameter type in worker tests.
// It is satisfied by *MockNotificationService.
type MockNotificationServiceInterface interface {
	ProcessEvent(ctx context.Context, event *domain.EventMessage) error
}
