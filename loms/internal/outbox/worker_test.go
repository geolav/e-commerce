package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/outbox/mocks"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// ── mock kafkaWriter ──────────────────────────────────────────────────────────

type mockKafkaWriter struct {
	writeFunc func(ctx context.Context, msgs ...kafka.Message) error
	closed    bool
}

func (m *mockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if m.writeFunc != nil {
		return m.writeFunc(ctx, msgs...)
	}
	return nil
}

func (m *mockKafkaWriter) Close() error {
	m.closed = true
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestWorker(t *testing.T, repo *mocks.MockoutboxRepository) *Worker {
	t.Helper()
	return NewWorker(repo, Config{
		BatchSize:     10,
		InProgressTTL: time.Second,
		FetchPeriod:   time.Millisecond,
		Workers:       1,
		KafkaBrokers:  "localhost:9092",
		KafkaTopic:    "test-topic",
	}, zap.NewNop())
}

func validEvent(id int64) entity.OutboxEvent {
	return entity.OutboxEvent{
		ID: id,
		Payload: map[string]interface{}{
			"order_id": float64(id),
			"user_id":  float64(42),
			"status":   "awaiting_payment",
		},
	}
}

// ── NewWorker ─────────────────────────────────────────────────────────────────

func TestNewWorker_NotNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	w := newTestWorker(t, repo)

	require.NotNil(t, w)
	assert.Equal(t, repo, w.repo)
	assert.Equal(t, 10, w.cfg.BatchSize)
}

// ── Run ───────────────────────────────────────────────────────────────────────

func TestWorker_Run_ZeroWorkers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	w := NewWorker(repo, Config{Workers: 0, FetchPeriod: time.Millisecond}, zap.NewNop())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w.Run(ctx)
}

func TestWorker_Run_WithWorkers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]entity.OutboxEvent{}, nil).
		AnyTimes()

	w := NewWorker(repo, Config{
		Workers:       2,
		BatchSize:     5,
		FetchPeriod:   5 * time.Millisecond,
		InProgressTTL: time.Second,
	}, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	w.Run(ctx)
	<-ctx.Done()
}

// ── processBatch — fetch ──────────────────────────────────────────────────────

func TestProcessBatch_FetchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), 10, time.Second).
		Return(nil, errors.New("db error"))

	w := newTestWorker(t, repo)
	w.processBatch(context.Background(), nil) // writer не нужен при ошибке fetch
}

func TestProcessBatch_EmptyBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), 10, time.Second).
		Return([]entity.OutboxEvent{}, nil)

	w := newTestWorker(t, repo)
	w.processBatch(context.Background(), nil)
}

func TestProcessBatch_ContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), 10, time.Second).
		Return(nil, context.Canceled)

	w := newTestWorker(t, repo)
	w.processBatch(ctx, nil)
}

// ── processBatch — write success + delete ────────────────────────────────────

func TestProcessBatch_WriteSuccess_DeleteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), 10, time.Second).
		Return([]entity.OutboxEvent{validEvent(1)}, nil)
	repo.EXPECT().
		Delete(gomock.Any(), int64(1)).
		Return(nil)

	writer := &mockKafkaWriter{writeFunc: func(_ context.Context, _ ...kafka.Message) error {
		return nil
	}}

	w := newTestWorker(t, repo)
	w.processBatch(context.Background(), writer)
}

func TestProcessBatch_WriteSuccess_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), 10, time.Second).
		Return([]entity.OutboxEvent{validEvent(2)}, nil)
	repo.EXPECT().
		Delete(gomock.Any(), int64(2)).
		Return(errors.New("delete failed"))

	writer := &mockKafkaWriter{writeFunc: func(_ context.Context, _ ...kafka.Message) error {
		return nil
	}}

	w := newTestWorker(t, repo)
	w.processBatch(context.Background(), writer) // не должно паниковать
}

// ── processBatch — write error (все попытки) ─────────────────────────────────

func TestProcessBatch_WriteAlwaysFails_NoDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), 10, time.Second).
		Return([]entity.OutboxEvent{validEvent(3)}, nil)
	// Delete НЕ должен вызываться если write провалился

	callCount := 0
	writer := &mockKafkaWriter{writeFunc: func(_ context.Context, _ ...kafka.Message) error {
		callCount++
		return errors.New("kafka unavailable")
	}}

	w := newTestWorker(t, repo)
	w.processBatch(context.Background(), writer)

	assert.Equal(t, 5, callCount, "должно быть ровно 5 попыток")
}

// ── processBatch — несколько событий ─────────────────────────────────────────

func TestProcessBatch_MultipleEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	events := []entity.OutboxEvent{validEvent(10), validEvent(11), validEvent(12)}

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), 10, time.Second).
		Return(events, nil)
	repo.EXPECT().Delete(gomock.Any(), int64(10)).Return(nil)
	repo.EXPECT().Delete(gomock.Any(), int64(11)).Return(nil)
	repo.EXPECT().Delete(gomock.Any(), int64(12)).Return(nil)

	writer := &mockKafkaWriter{writeFunc: func(_ context.Context, _ ...kafka.Message) error {
		return nil
	}}

	w := newTestWorker(t, repo)
	w.processBatch(context.Background(), writer)
}

// ── processBatch — write fail первые N попыток, потом успех ──────────────────

func TestProcessBatch_WriteRetry_ThenSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	repo.EXPECT().
		FetchBatch(gomock.Any(), 10, time.Second).
		Return([]entity.OutboxEvent{validEvent(5)}, nil)
	repo.EXPECT().
		Delete(gomock.Any(), int64(5)).
		Return(nil)

	callCount := 0
	writer := &mockKafkaWriter{writeFunc: func(_ context.Context, _ ...kafka.Message) error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}}

	w := newTestWorker(t, repo)
	w.processBatch(context.Background(), writer)

	assert.Equal(t, 3, callCount)
}

// ── Config / notificationPayload ──────────────────────────────────────────────

func TestConfig_Fields(t *testing.T) {
	cfg := Config{
		Workers:       3,
		BatchSize:     20,
		FetchPeriod:   200 * time.Millisecond,
		InProgressTTL: time.Minute,
		KafkaBrokers:  "broker1:9092",
		KafkaTopic:    "orders",
	}
	assert.Equal(t, 3, cfg.Workers)
	assert.Equal(t, 20, cfg.BatchSize)
	assert.Equal(t, "broker1:9092", cfg.KafkaBrokers)
}

func TestNotificationPayload_Fields(t *testing.T) {
	p := notificationPayload{UserID: 1, OrderID: 2, Status: "paid"}
	assert.Equal(t, int64(1), p.UserID)
	assert.Equal(t, int64(2), p.OrderID)
	assert.Equal(t, "paid", p.Status)
}
