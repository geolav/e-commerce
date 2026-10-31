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

func newTestWorker(t *testing.T, repo *mocks.MockoutboxRepository) *Worker {
	t.Helper()
	return NewWorker(repo, Config{
		Workers:       1,
		BatchSize:     10,
		FetchPeriod:   time.Millisecond,
		InProgressTTL: time.Second,
	}, zap.NewNop())
}

func TestNewWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	w := newTestWorker(t, repo)
	require.NotNil(t, w)
}

func TestWorker_Run_NoWorkers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockoutboxRepository(ctrl)
	w := NewWorker(repo, Config{Workers: 0, FetchPeriod: time.Millisecond}, zap.NewNop())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w.Run(ctx) // не должно падать
}

func TestProcessBatch(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockoutboxRepository)
	}{
		{
			name: "empty events",
			mockSetup: func(repo *mocks.MockoutboxRepository) {
				repo.EXPECT().
					FetchBatch(gomock.Any(), 10, time.Second).
					Return([]entity.OutboxEvent{}, nil)
			},
		},
		{
			name: "fetch error",
			mockSetup: func(repo *mocks.MockoutboxRepository) {
				repo.EXPECT().
					FetchBatch(gomock.Any(), 10, time.Second).
					Return(nil, errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockoutboxRepository(ctrl)
			tt.mockSetup(repo)

			w := NewWorker(repo, Config{BatchSize: 10, InProgressTTL: time.Second}, zap.NewNop())
			w.processBatch(context.Background(), nil) // не должно падать
		})
	}
}
