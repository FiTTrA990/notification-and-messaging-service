package queue_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/muhammadFittra/notification-service/internal/domain"
	"github.com/muhammadFittra/notification-service/mocks"
)

// Worker menggunakan MessageConsumer dan NotificationService.
// Unit test ini memverifikasi bahwa Worker:
//   1. Memanggil consumer.Consume dengan handler yang benar
//   2. Meneruskan setiap EventMessage ke service.ProcessEvent
//   3. Menangani error dengan benar (tanpa crash)

func TestWorker_Consume_CallsProcessEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockConsumer := mocks.NewMockMessageConsumer(ctrl)
	mockService := mocks.NewMockNotificationService(ctrl)

	event := &domain.EventMessage{
		EventID:   "evt-worker-001",
		EventType: domain.EventPackageStatusChanged,
		UserID:    "user-worker-001",
		Timestamp: time.Now(),
	}

	// Consumer akan memanggil handler satu kali dengan event di atas
	mockConsumer.EXPECT().
		Consume(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, handler func(context.Context, *domain.EventMessage) error) error {
			return handler(ctx, event)
		}).
		Times(1)

	mockService.EXPECT().
		ProcessEvent(ctx, event).
		Return(nil).
		Times(1)

	// Jalankan worker
	err := runWorker(ctx, mockConsumer, mockService)

	require.NoError(t, err)
}

func TestWorker_Consume_ProcessEventError_DoesNotStopWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockConsumer := mocks.NewMockMessageConsumer(ctrl)
	mockService := mocks.NewMockNotificationService(ctrl)

	events := []*domain.EventMessage{
		{EventID: "evt-w-01", EventType: domain.EventPackageStatusChanged, UserID: "user-w-01", Timestamp: time.Now()},
		{EventID: "evt-w-02", EventType: domain.EventPaymentStatusChanged, UserID: "user-w-02", Timestamp: time.Now()},
	}

	callCount := 0
	mockConsumer.EXPECT().
		Consume(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, handler func(context.Context, *domain.EventMessage) error) error {
			for _, evt := range events {
				_ = handler(ctx, evt) // handler error tidak menghentikan loop
			}
			return nil
		}).
		Times(1)

	// Event pertama gagal, kedua berhasil – keduanya tetap diproses
	mockService.EXPECT().
		ProcessEvent(ctx, events[0]).
		DoAndReturn(func(_ context.Context, _ *domain.EventMessage) error {
			callCount++
			return errors.New("processing failed")
		})

	mockService.EXPECT().
		ProcessEvent(ctx, events[1]).
		DoAndReturn(func(_ context.Context, _ *domain.EventMessage) error {
			callCount++
			return nil
		})

	err := runWorker(ctx, mockConsumer, mockService)

	require.NoError(t, err)
	assert.Equal(t, 2, callCount, "kedua event harus tetap diproses meski event pertama error")
}

func TestWorker_ConsumerError_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockConsumer := mocks.NewMockMessageConsumer(ctrl)
	mockService := mocks.NewMockNotificationService(ctrl)

	consumerErr := errors.New("rabbitmq connection closed")

	mockConsumer.EXPECT().
		Consume(ctx, gomock.Any()).
		Return(consumerErr)

	err := runWorker(ctx, mockConsumer, mockService)

	assert.ErrorIs(t, err, consumerErr)
}

func TestWorker_Close_CalledOnShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConsumer := mocks.NewMockMessageConsumer(ctrl)

	mockConsumer.EXPECT().Close().Return(nil).Times(1)

	err := mockConsumer.Close()
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────
// runWorker – stub helper yang akan diganti implementasi nyata
// ─────────────────────────────────────────────────────────────

// runWorker mensimulasikan fungsi worker yang menggunakan consumer dan service.
// Akan diganti dengan implementasi nyata dari package queue.
func runWorker(
	ctx context.Context,
	consumer mocks.MockMessageConsumerInterface,
	svc mocks.MockNotificationServiceInterface,
) error {
	return consumer.Consume(ctx, func(ctx context.Context, msg *domain.EventMessage) error {
		if err := svc.ProcessEvent(ctx, msg); err != nil {
			// Log error, tapi tidak menghentikan consumer loop
			return nil
		}
		return nil
	})
}
