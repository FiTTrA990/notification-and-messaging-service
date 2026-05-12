package service

import (
	"context"

	"github.com/FiTTrA990/notification-and-messaging-service/internal/domain"
)

//go:generate mockgen -source=interfaces.go -destination=../../mocks/mock_interfaces.go -package=mocks

// NotificationRepository handles persistence of notification records
type NotificationRepository interface {
	Save(ctx context.Context, n *domain.Notification) error
	FindByID(ctx context.Context, id string) (*domain.Notification, error)
	UpdateStatus(ctx context.Context, id string, status domain.DeliveryStatus, errMsg string) error
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Notification, error)
}

// TemplateRepository retrieves message templates
type TemplateRepository interface {
	FindByName(ctx context.Context, name string) (*domain.MessageTemplate, error)
}

// UserRepository retrieves user contact information
type UserRepository interface {
	FindContactByUserID(ctx context.Context, userID string) (*domain.UserContact, error)
}

// EmailProvider sends email notifications
type EmailProvider interface {
	Send(ctx context.Context, to, subject, body string) error
}

// PushProvider sends push notifications
type PushProvider interface {
	Send(ctx context.Context, deviceToken, title, body string) error
}
