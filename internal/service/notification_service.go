package service

import (
	"context"

	"github.com/muhammadFittra/notification-service/internal/domain"
)

//go:generate mockgen -source=notification_service.go -destination=../../mocks/mock_notification_service.go -package=mocks

// NotificationService defines the core business logic contract
type NotificationService interface {
	// Send dispatches a notification based on a request
	Send(ctx context.Context, req *domain.NotificationRequest) ([]*domain.NotificationResult, error)

	// ProcessEvent handles an inbound event message from the queue
	// and triggers appropriate notifications
	ProcessEvent(ctx context.Context, event *domain.EventMessage) error

	// GetNotificationStatus retrieves the delivery status of a notification
	GetNotificationStatus(ctx context.Context, notificationID string) (*domain.NotificationResult, error)
}

// notificationServiceImpl is the concrete implementation (stub – not yet complete)
type notificationServiceImpl struct {
	repo          NotificationRepository
	emailProvider EmailProvider
	pushProvider  PushProvider
	templateRepo  TemplateRepository
	userRepo      UserRepository
}

// NewNotificationService constructs a new NotificationService
func NewNotificationService(
	repo NotificationRepository,
	emailProvider EmailProvider,
	pushProvider PushProvider,
	templateRepo TemplateRepository,
	userRepo UserRepository,
) NotificationService {
	return &notificationServiceImpl{
		repo:          repo,
		emailProvider: emailProvider,
		pushProvider:  pushProvider,
		templateRepo:  templateRepo,
		userRepo:      userRepo,
	}
}

// Send – STUB: implementation not yet complete
func (s *notificationServiceImpl) Send(ctx context.Context, req *domain.NotificationRequest) ([]*domain.NotificationResult, error) {
	// TODO: implement full send logic
	panic("Send: not implemented")
}

// ProcessEvent – STUB: implementation not yet complete
func (s *notificationServiceImpl) ProcessEvent(ctx context.Context, event *domain.EventMessage) error {
	// TODO: implement event routing logic
	panic("ProcessEvent: not implemented")
}

// GetNotificationStatus – STUB: implementation not yet complete
func (s *notificationServiceImpl) GetNotificationStatus(ctx context.Context, notificationID string) (*domain.NotificationResult, error) {
	// TODO: implement status lookup
	panic("GetNotificationStatus: not implemented")
}
