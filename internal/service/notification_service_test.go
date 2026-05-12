package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/muhammadFittra/notification-service/internal/domain"
	"github.com/muhammadFittra/notification-service/internal/service"
	"github.com/muhammadFittra/notification-service/mocks"
)

// ─────────────────────────────────────────────────────────────
// Test Suite Setup Helper
// ─────────────────────────────────────────────────────────────

type testDeps struct {
	ctrl         *gomock.Controller
	notifRepo    *mocks.MockNotificationRepository
	templateRepo *mocks.MockTemplateRepository
	userRepo     *mocks.MockUserRepository
	emailProv    *mocks.MockEmailProvider
	pushProv     *mocks.MockPushProvider
	svc          service.NotificationService
}

func newTestDeps(t *testing.T) *testDeps {
	t.Helper()
	ctrl := gomock.NewController(t)
	notifRepo := mocks.NewMockNotificationRepository(ctrl)
	templateRepo := mocks.NewMockTemplateRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	emailProv := mocks.NewMockEmailProvider(ctrl)
	pushProv := mocks.NewMockPushProvider(ctrl)

	svc := service.NewNotificationService(
		notifRepo, emailProv, pushProv, templateRepo, userRepo,
	)
	return &testDeps{
		ctrl:         ctrl,
		notifRepo:    notifRepo,
		templateRepo: templateRepo,
		userRepo:     userRepo,
		emailProv:    emailProv,
		pushProv:     pushProv,
		svc:          svc,
	}
}

// ─────────────────────────────────────────────────────────────
// Send – Email Notification
// ─────────────────────────────────────────────────────────────

func TestSend_Email_Success(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-001",
		TemplateName: "package_status_changed",
		TriggerEvent: domain.EventPackageStatusChanged,
		TemplateData: map[string]string{"status": "In Transit", "tracking_no": "JNE-123"},
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	template := &domain.MessageTemplate{
		Name:    "package_status_changed",
		Subject: "Status Paket Anda Berubah",
		Body:    "Halo {{.Name}}, paket Anda kini berstatus: {{.status}}",
	}

	userContact := &domain.UserContact{
		UserID:      "user-001",
		Email:       "user@example.com",
		DeviceToken: "fcm-token-abc",
		Name:        "Budi Santoso",
	}

	// Define expected mock interactions
	d.templateRepo.EXPECT().
		FindByName(ctx, "package_status_changed").
		Return(template, nil).
		Times(1)

	d.userRepo.EXPECT().
		FindContactByUserID(ctx, "user-001").
		Return(userContact, nil).
		Times(1)

	d.notifRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(nil).
		Times(1)

	d.emailProv.EXPECT().
		Send(ctx, "user@example.com", gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	d.notifRepo.EXPECT().
		UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusSent, "").
		Return(nil).
		Times(1)

	results, err := d.svc.Send(ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, domain.DeliveryStatusSent, results[0].Status)
	assert.Equal(t, domain.NotificationTypeEmail, results[0].Channel)
	assert.Equal(t, "user-001", results[0].UserID)
}

func TestSend_Email_TemplateNotFound(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-001",
		TemplateName: "non_existent_template",
		TriggerEvent: domain.EventPackageStatusChanged,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	d.templateRepo.EXPECT().
		FindByName(ctx, "non_existent_template").
		Return(nil, &domain.ErrNotFound{Resource: "template", ID: "non_existent_template"}).
		Times(1)

	results, err := d.svc.Send(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, results)

	var notFound *domain.ErrNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestSend_Email_UserContactNotFound(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-999",
		TemplateName: "package_status_changed",
		TriggerEvent: domain.EventPackageStatusChanged,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	template := &domain.MessageTemplate{
		Name:    "package_status_changed",
		Subject: "Status Paket Anda Berubah",
		Body:    "Paket Anda kini dalam pengiriman",
	}

	d.templateRepo.EXPECT().
		FindByName(ctx, "package_status_changed").
		Return(template, nil)

	d.userRepo.EXPECT().
		FindContactByUserID(ctx, "user-999").
		Return(nil, &domain.ErrNotFound{Resource: "user", ID: "user-999"})

	results, err := d.svc.Send(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestSend_Email_ProviderFails_StatusMarkedFailed(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-002",
		TemplateName: "payment_confirmed",
		TriggerEvent: domain.EventPaymentStatusChanged,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	template := &domain.MessageTemplate{
		Name:    "payment_confirmed",
		Subject: "Pembayaran Dikonfirmasi",
		Body:    "Pembayaran Anda telah berhasil diproses.",
	}

	userContact := &domain.UserContact{
		UserID: "user-002",
		Email:  "user2@example.com",
		Name:   "Siti Rahayu",
	}

	d.templateRepo.EXPECT().
		FindByName(ctx, "payment_confirmed").
		Return(template, nil)

	d.userRepo.EXPECT().
		FindContactByUserID(ctx, "user-002").
		Return(userContact, nil)

	d.notifRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(nil)

	d.emailProv.EXPECT().
		Send(ctx, "user2@example.com", gomock.Any(), gomock.Any()).
		Return(errors.New("smtp connection refused"))

	// Status harus di-update ke Failed beserta error message
	d.notifRepo.EXPECT().
		UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusFailed, "smtp connection refused").
		Return(nil)

	results, err := d.svc.Send(ctx, req)

	// Send tetap mengembalikan result dengan status Failed, bukan error
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, domain.DeliveryStatusFailed, results[0].Status)
}

// ─────────────────────────────────────────────────────────────
// Send – Push Notification
// ─────────────────────────────────────────────────────────────

func TestSend_Push_Success(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-003",
		TemplateName: "delivery_confirmed",
		TriggerEvent: domain.EventDeliveryConfirmed,
		Channels:     []domain.NotificationType{domain.NotificationTypePush},
	}

	template := &domain.MessageTemplate{
		Name:    "delivery_confirmed",
		Subject: "Paket Telah Diterima",
		Body:    "Paket Anda telah berhasil diterima. Terima kasih!",
	}

	userContact := &domain.UserContact{
		UserID:      "user-003",
		Email:       "user3@example.com",
		DeviceToken: "fcm-device-token-xyz",
		Name:        "Ahmad Fauzi",
	}

	d.templateRepo.EXPECT().FindByName(ctx, "delivery_confirmed").Return(template, nil)
	d.userRepo.EXPECT().FindContactByUserID(ctx, "user-003").Return(userContact, nil)
	d.notifRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
	d.pushProv.EXPECT().
		Send(ctx, "fcm-device-token-xyz", gomock.Any(), gomock.Any()).
		Return(nil)
	d.notifRepo.EXPECT().
		UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusSent, "").
		Return(nil)

	results, err := d.svc.Send(ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, domain.DeliveryStatusSent, results[0].Status)
	assert.Equal(t, domain.NotificationTypePush, results[0].Channel)
}

func TestSend_Push_EmptyDeviceToken(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-004",
		TemplateName: "delivery_confirmed",
		TriggerEvent: domain.EventDeliveryConfirmed,
		Channels:     []domain.NotificationType{domain.NotificationTypePush},
	}

	template := &domain.MessageTemplate{
		Name: "delivery_confirmed",
		Body: "Paket Anda telah berhasil diterima.",
	}

	userContact := &domain.UserContact{
		UserID:      "user-004",
		DeviceToken: "", // Tidak ada device token
		Name:        "Dewi Lestari",
	}

	d.templateRepo.EXPECT().FindByName(ctx, "delivery_confirmed").Return(template, nil)
	d.userRepo.EXPECT().FindContactByUserID(ctx, "user-004").Return(userContact, nil)
	d.notifRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
	// Push provider TIDAK dipanggil karena tidak ada device token
	d.notifRepo.EXPECT().
		UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusFailed, gomock.Any()).
		Return(nil)

	results, err := d.svc.Send(ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, domain.DeliveryStatusFailed, results[0].Status)
}

// ─────────────────────────────────────────────────────────────
// Send – Multi-channel (Email + Push)
// ─────────────────────────────────────────────────────────────

func TestSend_MultiChannel_BothSuccess(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-005",
		TemplateName: "order_cancelled",
		TriggerEvent: domain.EventOrderCancelled,
		Channels: []domain.NotificationType{
			domain.NotificationTypeEmail,
			domain.NotificationTypePush,
		},
	}

	template := &domain.MessageTemplate{
		Name:    "order_cancelled",
		Subject: "Pesanan Dibatalkan",
		Body:    "Maaf, pesanan Anda telah dibatalkan.",
	}

	userContact := &domain.UserContact{
		UserID:      "user-005",
		Email:       "user5@example.com",
		DeviceToken: "fcm-token-user-005",
		Name:        "Rizky Pratama",
	}

	d.templateRepo.EXPECT().FindByName(ctx, "order_cancelled").Return(template, nil).Times(1)
	d.userRepo.EXPECT().FindContactByUserID(ctx, "user-005").Return(userContact, nil).Times(1)
	d.notifRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil).Times(2)
	d.emailProv.EXPECT().Send(ctx, "user5@example.com", gomock.Any(), gomock.Any()).Return(nil)
	d.pushProv.EXPECT().Send(ctx, "fcm-token-user-005", gomock.Any(), gomock.Any()).Return(nil)
	d.notifRepo.EXPECT().UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusSent, "").Return(nil).Times(2)

	results, err := d.svc.Send(ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 2)
	for _, r := range results {
		assert.Equal(t, domain.DeliveryStatusSent, r.Status)
	}
}

func TestSend_MultiChannel_OneFailsOneSucceeds(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-006",
		TemplateName: "order_created",
		TriggerEvent: domain.EventOrderCreated,
		Channels: []domain.NotificationType{
			domain.NotificationTypeEmail,
			domain.NotificationTypePush,
		},
	}

	template := &domain.MessageTemplate{
		Name:    "order_created",
		Subject: "Pesanan Dibuat",
		Body:    "Pesanan Anda telah berhasil dibuat.",
	}

	userContact := &domain.UserContact{
		UserID:      "user-006",
		Email:       "user6@example.com",
		DeviceToken: "fcm-token-user-006",
		Name:        "Hendra Kusuma",
	}

	d.templateRepo.EXPECT().FindByName(ctx, "order_created").Return(template, nil)
	d.userRepo.EXPECT().FindContactByUserID(ctx, "user-006").Return(userContact, nil)
	d.notifRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil).Times(2)

	// Email berhasil
	d.emailProv.EXPECT().Send(ctx, "user6@example.com", gomock.Any(), gomock.Any()).Return(nil)
	d.notifRepo.EXPECT().UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusSent, "").Return(nil)

	// Push gagal
	d.pushProv.EXPECT().Send(ctx, "fcm-token-user-006", gomock.Any(), gomock.Any()).
		Return(errors.New("FCM quota exceeded"))
	d.notifRepo.EXPECT().
		UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusFailed, "FCM quota exceeded").
		Return(nil)

	results, err := d.svc.Send(ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 2)

	var sentCount, failedCount int
	for _, r := range results {
		if r.Status == domain.DeliveryStatusSent {
			sentCount++
		} else {
			failedCount++
		}
	}
	assert.Equal(t, 1, sentCount)
	assert.Equal(t, 1, failedCount)
}

// ─────────────────────────────────────────────────────────────
// Send – Validation
// ─────────────────────────────────────────────────────────────

func TestSend_EmptyUserID_ReturnsValidationError(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "", // Invalid
		TemplateName: "package_status_changed",
		TriggerEvent: domain.EventPackageStatusChanged,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	results, err := d.svc.Send(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, results)

	var invalidInput *domain.ErrInvalidInput
	assert.True(t, errors.As(err, &invalidInput))
	assert.Equal(t, "user_id", invalidInput.Field)
}

func TestSend_EmptyChannels_ReturnsValidationError(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-007",
		TemplateName: "package_status_changed",
		TriggerEvent: domain.EventPackageStatusChanged,
		Channels:     []domain.NotificationType{}, // No channels
	}

	results, err := d.svc.Send(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestSend_EmptyTemplateName_ReturnsValidationError(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-008",
		TemplateName: "",
		TriggerEvent: domain.EventPackageStatusChanged,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	results, err := d.svc.Send(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, results)
}

// ─────────────────────────────────────────────────────────────
// ProcessEvent
// ─────────────────────────────────────────────────────────────

func TestProcessEvent_PackageStatusChanged_TriggersNotification(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	event := &domain.EventMessage{
		EventID:   "evt-001",
		EventType: domain.EventPackageStatusChanged,
		UserID:    "user-010",
		Payload: map[string]string{
			"tracking_no": "JNE-456",
			"status":      "Delivered",
		},
		Timestamp: time.Now(),
	}

	template := &domain.MessageTemplate{
		Name:    "package_status_changed",
		Subject: "Status Paket Berubah",
		Body:    "Paket Anda telah berhasil diterima.",
	}

	userContact := &domain.UserContact{
		UserID:      "user-010",
		Email:       "user10@example.com",
		DeviceToken: "fcm-token-010",
		Name:        "Agus Setiawan",
	}

	d.templateRepo.EXPECT().FindByName(ctx, gomock.Any()).Return(template, nil).AnyTimes()
	d.userRepo.EXPECT().FindContactByUserID(ctx, "user-010").Return(userContact, nil)
	d.notifRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil).AnyTimes()
	d.emailProv.EXPECT().Send(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	d.pushProv.EXPECT().Send(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	d.notifRepo.EXPECT().UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusSent, "").Return(nil).AnyTimes()

	err := d.svc.ProcessEvent(ctx, event)

	assert.NoError(t, err)
}

func TestProcessEvent_PaymentStatusChanged_TriggersNotification(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	event := &domain.EventMessage{
		EventID:   "evt-002",
		EventType: domain.EventPaymentStatusChanged,
		UserID:    "user-011",
		Payload: map[string]string{
			"payment_id": "PAY-789",
			"amount":     "250000",
			"status":     "SUCCESS",
		},
		Timestamp: time.Now(),
	}

	template := &domain.MessageTemplate{
		Name:    "payment_confirmed",
		Subject: "Pembayaran Berhasil",
		Body:    "Pembayaran Anda sebesar Rp{{.amount}} telah dikonfirmasi.",
	}

	userContact := &domain.UserContact{
		UserID:      "user-011",
		Email:       "user11@example.com",
		DeviceToken: "fcm-token-011",
		Name:        "Nindya Permata",
	}

	d.templateRepo.EXPECT().FindByName(ctx, gomock.Any()).Return(template, nil).AnyTimes()
	d.userRepo.EXPECT().FindContactByUserID(ctx, "user-011").Return(userContact, nil)
	d.notifRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil).AnyTimes()
	d.emailProv.EXPECT().Send(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	d.pushProv.EXPECT().Send(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	d.notifRepo.EXPECT().UpdateStatus(ctx, gomock.Any(), domain.DeliveryStatusSent, "").Return(nil).AnyTimes()

	err := d.svc.ProcessEvent(ctx, event)

	assert.NoError(t, err)
}

func TestProcessEvent_UnknownEventType_ReturnsError(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	event := &domain.EventMessage{
		EventID:   "evt-003",
		EventType: "UNKNOWN_EVENT_TYPE",
		UserID:    "user-012",
		Timestamp: time.Now(),
	}

	err := d.svc.ProcessEvent(ctx, event)

	assert.Error(t, err)
}

func TestProcessEvent_NilEvent_ReturnsError(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()

	err := d.svc.ProcessEvent(ctx, nil)

	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────
// GetNotificationStatus
// ─────────────────────────────────────────────────────────────

func TestGetNotificationStatus_Found(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	notifID := "notif-uuid-001"
	now := time.Now()

	storedNotif := &domain.Notification{
		ID:           notifID,
		UserID:       "user-020",
		Type:         domain.NotificationTypeEmail,
		TemplateName: "package_status_changed",
		Status:       domain.DeliveryStatusSent,
		TriggerEvent: domain.EventPackageStatusChanged,
		CreatedAt:    now.Add(-10 * time.Minute),
		SentAt:       &now,
	}

	d.notifRepo.EXPECT().
		FindByID(ctx, notifID).
		Return(storedNotif, nil)

	result, err := d.svc.GetNotificationStatus(ctx, notifID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, notifID, result.NotificationID)
	assert.Equal(t, domain.DeliveryStatusSent, result.Status)
	assert.Equal(t, "user-020", result.UserID)
	assert.NotNil(t, result.SentAt)
}

func TestGetNotificationStatus_NotFound(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	notifID := "notif-non-existent"

	d.notifRepo.EXPECT().
		FindByID(ctx, notifID).
		Return(nil, &domain.ErrNotFound{Resource: "notification", ID: notifID})

	result, err := d.svc.GetNotificationStatus(ctx, notifID)

	assert.Error(t, err)
	assert.Nil(t, result)

	var notFound *domain.ErrNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestGetNotificationStatus_EmptyID_ReturnsError(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()

	result, err := d.svc.GetNotificationStatus(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ─────────────────────────────────────────────────────────────
// Repository Failure Handling
// ─────────────────────────────────────────────────────────────

func TestSend_RepositorySaveFails_ReturnsError(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx := context.Background()
	req := &domain.NotificationRequest{
		UserID:       "user-030",
		TemplateName: "package_status_changed",
		TriggerEvent: domain.EventPackageStatusChanged,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	template := &domain.MessageTemplate{
		Name:    "package_status_changed",
		Subject: "Status Paket",
		Body:    "Paket Anda diperbarui",
	}

	userContact := &domain.UserContact{
		UserID: "user-030",
		Email:  "user30@example.com",
		Name:   "Tono Hardjono",
	}

	d.templateRepo.EXPECT().FindByName(ctx, "package_status_changed").Return(template, nil)
	d.userRepo.EXPECT().FindContactByUserID(ctx, "user-030").Return(userContact, nil)
	d.notifRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(errors.New("database connection lost"))

	results, err := d.svc.Send(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, results)
}

// ─────────────────────────────────────────────────────────────
// Context Cancellation
// ─────────────────────────────────────────────────────────────

func TestSend_CancelledContext_ReturnsError(t *testing.T) {
	d := newTestDeps(t)
	defer d.ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &domain.NotificationRequest{
		UserID:       "user-040",
		TemplateName: "package_status_changed",
		TriggerEvent: domain.EventPackageStatusChanged,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	results, err := d.svc.Send(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.ErrorIs(t, err, context.Canceled)
}
