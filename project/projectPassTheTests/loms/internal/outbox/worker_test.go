package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/outbox/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNewWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	logger := zap.NewNop()
	cfg := Config{
		Workers:           1,
		BatchSize:         10,
		FetchPeriod:       time.Second,
		InProgressTTL:     time.Second,
		NotificationsAddr: "localhost:50053",
	}

	w := NewWorker(repo, cfg, logger)
	require.NotNil(t, w)
}

func TestWorker_Run_NoWorkers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	logger := zap.NewNop()
	cfg := Config{
		Workers:           0,
		BatchSize:         10,
		FetchPeriod:       time.Millisecond,
		InProgressTTL:     time.Second,
		NotificationsAddr: "",
	}

	w := NewWorker(repo, cfg, logger)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w.Run(ctx) // не должно падать
}

func TestProcessBatch_EmptyEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	logger := zap.NewNop()
	w := NewWorker(repo, Config{BatchSize: 10, InProgressTTL: time.Second}, logger)

	repo.EXPECT().FetchBatch(gomock.Any(), 10, time.Second).Return([]entity.OutboxEvent{}, nil)

	w.processBatch(context.Background(), nil)
}

func TestProcessBatch_FetchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	logger := zap.NewNop()
	w := NewWorker(repo, Config{BatchSize: 10, InProgressTTL: time.Second}, logger)

	repo.EXPECT().FetchBatch(gomock.Any(), 10, time.Second).Return(nil, errors.New("db error"))

	w.processBatch(context.Background(), nil)
}
