package domain

import "time"

// NotificationType defines the channel of notification
type NotificationType string

const (
	NotificationTypePush  NotificationType = "PUSH"
	NotificationTypeEmail NotificationType = "EMAIL"
)

// DeliveryStatus represents the result of a notification send attempt
type DeliveryStatus string

const (
	DeliveryStatusSent   DeliveryStatus = "Sent"
	DeliveryStatusFailed DeliveryStatus = "Failed"
	DeliveryStatusPending DeliveryStatus = "Pending"
)

// TriggerEvent represents events from other services that trigger notifications
type TriggerEvent string

const (
	EventPackageStatusChanged TriggerEvent = "PACKAGE_STATUS_CHANGED"
	EventPaymentStatusChanged TriggerEvent = "PAYMENT_STATUS_CHANGED"
	EventOrderCreated         TriggerEvent = "ORDER_CREATED"
	EventOrderCancelled       TriggerEvent = "ORDER_CANCELLED"
	EventDeliveryConfirmed    TriggerEvent = "DELIVERY_CONFIRMED"
)

// Notification is the core domain entity
type Notification struct {
	ID           string
	UserID       string
	Type         NotificationType
	TemplateName string
	Message      string
	TriggerEvent TriggerEvent
	Status       DeliveryStatus
	ErrorMessage string
	CreatedAt    time.Time
	SentAt       *time.Time
	UpdatedAt    time.Time
}

// NotificationRequest is the inbound request payload
type NotificationRequest struct {
	UserID       string            `json:"user_id" validate:"required"`
	TemplateName string            `json:"template_name" validate:"required"`
	TriggerEvent TriggerEvent      `json:"trigger_event" validate:"required"`
	TemplateData map[string]string `json:"template_data"`
	Channels     []NotificationType `json:"channels" validate:"required,min=1"`
}

// NotificationResult is the outbound response
type NotificationResult struct {
	NotificationID string         `json:"notification_id"`
	UserID         string         `json:"user_id"`
	Status         DeliveryStatus `json:"status"`
	Channel        NotificationType `json:"channel"`
	Message        string         `json:"message,omitempty"`
	SentAt         *time.Time     `json:"sent_at,omitempty"`
}

// EventMessage represents a message consumed from the queue (RabbitMQ/Kafka)
type EventMessage struct {
	EventID   string            `json:"event_id"`
	EventType TriggerEvent      `json:"event_type"`
	UserID    string            `json:"user_id"`
	Payload   map[string]string `json:"payload"`
	Timestamp time.Time         `json:"timestamp"`
}

// MessageTemplate holds the template definition for a notification
type MessageTemplate struct {
	Name    string
	Subject string
	Body    string
}

// UserContact holds the contact information for a user
type UserContact struct {
	UserID      string
	Email       string
	DeviceToken string // FCM / APNs token for push notifications
	Name        string
}

// ErrNotFound is returned when a resource is not found
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e *ErrNotFound) Error() string {
	return "resource " + e.Resource + " with id " + e.ID + " not found"
}

// ErrInvalidInput is returned for invalid input data
type ErrInvalidInput struct {
	Field   string
	Message string
}

func (e *ErrInvalidInput) Error() string {
	return "invalid input for field " + e.Field + ": " + e.Message
}
