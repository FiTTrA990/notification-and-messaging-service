package provider

import (
	"context"
)

//go:generate mockgen -source=push_provider.go -destination=../../mocks/mock_push_provider.go -package=mocks

// FCMPushProvider sends Firebase Cloud Messaging push notifications (stub)
type FCMPushProvider struct {
	serverKey string
	fcmURL    string
}

// NewFCMPushProvider creates an FCM push provider
func NewFCMPushProvider(serverKey string) *FCMPushProvider {
	return &FCMPushProvider{
		serverKey: serverKey,
		fcmURL:    "https://fcm.googleapis.com/fcm/send",
	}
}

// Send sends a push notification via FCM – STUB
func (p *FCMPushProvider) Send(ctx context.Context, deviceToken, title, body string) error {
	// TODO: implement FCM HTTP v1 push
	panic("FCMPushProvider.Send: not implemented")
}
