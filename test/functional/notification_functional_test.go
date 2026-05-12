// +build functional

package functional

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/muhammadFittra/notification-service/internal/domain"
	"github.com/muhammadFittra/notification-service/internal/service"
)

// ─────────────────────────────────────────────────────────────
// FT-01: Kirim Email Notification – End-to-End via DB
// ─────────────────────────────────────────────────────────────

func TestFunctional_SendEmailNotification_Persisted(t *testing.T) {
	env := setupTestEnv(t)
	defer teardownTestEnv(t, env)

	// Seed template dan user contact di DB
	seedTemplate(t, env.db, "ft_package_status", "Status Paket", "Paket Anda diperbarui")
	seedUserContact(t, env.db, "ft-user-001", "functional.test@example.com", "", "Test User")

	// Buat service dengan dependency nyata (repository PostgreSQL)
	notifRepo := buildNotificationRepository(t, env.db)
	templateRepo := buildTemplateRepository(t, env.db)
	userRepo := buildUserRepository(t, env.db)
	emailProv := buildFakeEmailProvider(t)  // SMTP tes / mailtrap
	pushProv := buildFakePushProvider(t)    // FCM tes

	svc := service.NewNotificationService(notifRepo, emailProv, pushProv, templateRepo, userRepo)

	req := &domain.NotificationRequest{
		UserID:       "ft-user-001",
		TemplateName: "ft_package_status",
		TriggerEvent: domain.EventPackageStatusChanged,
		TemplateData: map[string]string{"status": "Dalam Pengiriman"},
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	results, err := svc.Send(env.ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, domain.DeliveryStatusSent, results[0].Status)

	// Verifikasi record tersimpan di DB
	notifID := results[0].NotificationID
	persisted, err := notifRepo.FindByID(env.ctx, notifID)
	require.NoError(t, err)
	assert.Equal(t, domain.DeliveryStatusSent, persisted.Status)
	assert.Equal(t, "ft-user-001", persisted.UserID)
	assert.NotNil(t, persisted.SentAt)
}

// ─────────────────────────────────────────────────────────────
// FT-02: Kirim Push Notification – End-to-End via DB
// ─────────────────────────────────────────────────────────────

func TestFunctional_SendPushNotification_Persisted(t *testing.T) {
	env := setupTestEnv(t)
	defer teardownTestEnv(t, env)

	seedTemplate(t, env.db, "ft_delivery_confirmed", "Paket Tiba", "Paket Anda telah diterima!")
	seedUserContact(t, env.db, "ft-user-002", "", "device-token-ft-002", "Test User Push")

	notifRepo := buildNotificationRepository(t, env.db)
	templateRepo := buildTemplateRepository(t, env.db)
	userRepo := buildUserRepository(t, env.db)
	emailProv := buildFakeEmailProvider(t)
	pushProv := buildFakePushProvider(t)

	svc := service.NewNotificationService(notifRepo, emailProv, pushProv, templateRepo, userRepo)

	req := &domain.NotificationRequest{
		UserID:       "ft-user-002",
		TemplateName: "ft_delivery_confirmed",
		TriggerEvent: domain.EventDeliveryConfirmed,
		Channels:     []domain.NotificationType{domain.NotificationTypePush},
	}

	results, err := svc.Send(env.ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, domain.DeliveryStatusSent, results[0].Status)

	notifID := results[0].NotificationID
	persisted, err := notifRepo.FindByID(env.ctx, notifID)
	require.NoError(t, err)
	assert.Equal(t, domain.NotificationTypePush, persisted.Type)
	assert.Equal(t, domain.DeliveryStatusSent, persisted.Status)
}

// ─────────────────────────────────────────────────────────────
// FT-03: ProcessEvent dari Queue – Full Flow
// ─────────────────────────────────────────────────────────────

func TestFunctional_ProcessEvent_PackageStatusChanged(t *testing.T) {
	env := setupTestEnv(t)
	defer teardownTestEnv(t, env)

	seedTemplate(t, env.db, "ft_package_status", "Status Paket", "Paket Anda: {{.status}}")
	seedUserContact(t, env.db, "ft-user-003", "ft-user3@example.com", "fcm-ft-003", "Test User Event")

	notifRepo := buildNotificationRepository(t, env.db)
	templateRepo := buildTemplateRepository(t, env.db)
	userRepo := buildUserRepository(t, env.db)
	emailProv := buildFakeEmailProvider(t)
	pushProv := buildFakePushProvider(t)

	svc := service.NewNotificationService(notifRepo, emailProv, pushProv, templateRepo, userRepo)

	event := &domain.EventMessage{
		EventID:   "ft-evt-001",
		EventType: domain.EventPackageStatusChanged,
		UserID:    "ft-user-003",
		Payload: map[string]string{
			"tracking_no": "FT-TRACK-001",
			"status":      "Out for Delivery",
		},
		Timestamp: time.Now(),
	}

	err := svc.ProcessEvent(env.ctx, event)
	require.NoError(t, err)

	// Verifikasi minimal satu notification tersimpan untuk user ini
	notifications, err := notifRepo.FindByUserID(env.ctx, "ft-user-003", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(notifications), 1)

	for _, n := range notifications {
		assert.Equal(t, "ft-user-003", n.UserID)
		assert.NotEmpty(t, n.Status)
	}
}

// ─────────────────────────────────────────────────────────────
// FT-04: GetNotificationStatus – Record yang Ada
// ─────────────────────────────────────────────────────────────

func TestFunctional_GetNotificationStatus_ExistingRecord(t *testing.T) {
	env := setupTestEnv(t)
	defer teardownTestEnv(t, env)

	notifRepo := buildNotificationRepository(t, env.db)
	templateRepo := buildTemplateRepository(t, env.db)
	userRepo := buildUserRepository(t, env.db)
	emailProv := buildFakeEmailProvider(t)
	pushProv := buildFakePushProvider(t)

	// Seed notification langsung ke DB
	notifID := "ft-notif-status-001"
	seedNotification(t, env.db, notifID, "ft-user-004", domain.DeliveryStatusSent)

	svc := service.NewNotificationService(notifRepo, emailProv, pushProv, templateRepo, userRepo)

	result, err := svc.GetNotificationStatus(env.ctx, notifID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, notifID, result.NotificationID)
	assert.Equal(t, domain.DeliveryStatusSent, result.Status)
}

// ─────────────────────────────────────────────────────────────
// FT-05: Delivery Failed – Record Status Failed Tersimpan
// ─────────────────────────────────────────────────────────────

func TestFunctional_Send_ProviderFails_FailedStatusPersisted(t *testing.T) {
	env := setupTestEnv(t)
	defer teardownTestEnv(t, env)

	seedTemplate(t, env.db, "ft_payment_confirmed", "Pembayaran Berhasil", "Terima kasih atas pembayaran Anda.")
	seedUserContact(t, env.db, "ft-user-005", "ft-user5@example.com", "", "Test Fail User")

	notifRepo := buildNotificationRepository(t, env.db)
	templateRepo := buildTemplateRepository(t, env.db)
	userRepo := buildUserRepository(t, env.db)
	emailProv := buildAlwaysFailEmailProvider(t) // Selalu gagal
	pushProv := buildFakePushProvider(t)

	svc := service.NewNotificationService(notifRepo, emailProv, pushProv, templateRepo, userRepo)

	req := &domain.NotificationRequest{
		UserID:       "ft-user-005",
		TemplateName: "ft_payment_confirmed",
		TriggerEvent: domain.EventPaymentStatusChanged,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail},
	}

	results, err := svc.Send(env.ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, domain.DeliveryStatusFailed, results[0].Status)

	// Verifikasi status Failed tersimpan di DB
	notifID := results[0].NotificationID
	persisted, err := notifRepo.FindByID(env.ctx, notifID)
	require.NoError(t, err)
	assert.Equal(t, domain.DeliveryStatusFailed, persisted.Status)
	assert.NotEmpty(t, persisted.ErrorMessage)
}

// ─────────────────────────────────────────────────────────────
// FT-06: Multi-channel – Keduanya Tersimpan di DB
// ─────────────────────────────────────────────────────────────

func TestFunctional_Send_MultiChannel_BothPersisted(t *testing.T) {
	env := setupTestEnv(t)
	defer teardownTestEnv(t, env)

	seedTemplate(t, env.db, "ft_order_created", "Pesanan Dibuat", "Pesanan Anda berhasil.")
	seedUserContact(t, env.db, "ft-user-006", "ft-user6@example.com", "fcm-ft-006", "Test Multi")

	notifRepo := buildNotificationRepository(t, env.db)
	templateRepo := buildTemplateRepository(t, env.db)
	userRepo := buildUserRepository(t, env.db)
	emailProv := buildFakeEmailProvider(t)
	pushProv := buildFakePushProvider(t)

	svc := service.NewNotificationService(notifRepo, emailProv, pushProv, templateRepo, userRepo)

	req := &domain.NotificationRequest{
		UserID:       "ft-user-006",
		TemplateName: "ft_order_created",
		TriggerEvent: domain.EventOrderCreated,
		Channels:     []domain.NotificationType{domain.NotificationTypeEmail, domain.NotificationTypePush},
	}

	results, err := svc.Send(env.ctx, req)

	require.NoError(t, err)
	require.Len(t, results, 2)

	// Verifikasi dua record di DB untuk user ini
	notifications, err := notifRepo.FindByUserID(env.ctx, "ft-user-006", 10, 0)
	require.NoError(t, err)
	assert.Len(t, notifications, 2)

	channels := make(map[domain.NotificationType]bool)
	for _, n := range notifications {
		channels[n.Type] = true
	}
	assert.True(t, channels[domain.NotificationTypeEmail])
	assert.True(t, channels[domain.NotificationTypePush])
}

// ─────────────────────────────────────────────────────────────
// Seed helpers – hanya untuk functional test
// ─────────────────────────────────────────────────────────────

func seedTemplate(t *testing.T, db interface{ Exec(sql string, values ...interface{}) }, name, subject, body string) {
	t.Helper()
	// TODO: insert ke tabel message_templates
	_ = fmt.Sprintf("INSERT INTO message_templates (name, subject, body) VALUES ('%s','%s','%s') ON CONFLICT DO NOTHING", name, subject, body)
}

func seedUserContact(t *testing.T, db interface{ Exec(sql string, values ...interface{}) }, userID, email, deviceToken, name string) {
	t.Helper()
	// TODO: insert ke tabel user_contacts
}

func seedNotification(t *testing.T, db interface{ Exec(sql string, values ...interface{}) }, id, userID string, status domain.DeliveryStatus) {
	t.Helper()
	// TODO: insert ke tabel notifications
}

// ─────────────────────────────────────────────────────────────
// Fake provider builders – lightweight, tidak perlu mock library
// ─────────────────────────────────────────────────────────────

type fakeEmailProvider struct{ fail bool }

func (f *fakeEmailProvider) Send(_ interface{}, to, subject, body string) error {
	if f.fail {
		return fmt.Errorf("fake email provider: forced failure")
	}
	return nil
}

type fakePushProvider struct{}

func (f *fakePushProvider) Send(_ interface{}, deviceToken, title, body string) error {
	return nil
}

func buildFakeEmailProvider(t *testing.T) service.EmailProvider {
	t.Helper()
	// TODO: return real SMTP or mailtrap provider from config
	panic("buildFakeEmailProvider: not implemented")
}

func buildAlwaysFailEmailProvider(t *testing.T) service.EmailProvider {
	t.Helper()
	panic("buildAlwaysFailEmailProvider: not implemented")
}

func buildFakePushProvider(t *testing.T) service.PushProvider {
	t.Helper()
	panic("buildFakePushProvider: not implemented")
}

func buildNotificationRepository(t *testing.T, db *gorm.DB) service.NotificationRepository {
	t.Helper()
	// TODO: return real repository.NewPostgresNotificationRepository(db)
	panic("buildNotificationRepository: not implemented")
}

func buildTemplateRepository(t *testing.T, db *gorm.DB) service.TemplateRepository {
	t.Helper()
	panic("buildTemplateRepository: not implemented")
}

func buildUserRepository(t *testing.T, db *gorm.DB) service.UserRepository {
	t.Helper()
	panic("buildUserRepository: not implemented")
}
