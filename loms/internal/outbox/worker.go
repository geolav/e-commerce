package outbox

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

//go:generate mockgen -source=worker.go -destination=mocks/worker_mock.go -package=mocks
type outboxRepository interface {
	FetchBatch(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]entity.OutboxEvent, error)
	Delete(ctx context.Context, id int64) error
}

// kafkaWriter — интерфейс для отправки сообщений в Kafka.
// Позволяет подменять реальный writer на мок в тестах.
type kafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

const (
	maxAttempts  = 3
	timeDuration = 100
)

type Config struct {
	Workers       int
	BatchSize     int
	FetchPeriod   time.Duration
	InProgressTTL time.Duration
	KafkaBrokers  string
	KafkaTopic    string
}

type Worker struct {
	repo   outboxRepository
	cfg    Config
	logger *zap.Logger
}

func NewWorker(repo outboxRepository, cfg Config, logger *zap.Logger) *Worker {
	return &Worker{repo: repo, cfg: cfg, logger: logger}
}

func (w *Worker) Run(ctx context.Context) {
	w.logger.Info("starting outbox worker",
		zap.String("brokers", w.cfg.KafkaBrokers),
		zap.String("topic", w.cfg.KafkaTopic))

	for i := 0; i < w.cfg.Workers; i++ {
		go w.runWorker(ctx)
	}
}

type notificationPayload struct {
	UserID  int64  `json:"user_id"`
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

func (w *Worker) runWorker(ctx context.Context) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(w.cfg.KafkaBrokers),
		Topic:        w.cfg.KafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  maxAttempts,
	}
	defer writer.Close()

	ticker := time.NewTicker(w.cfg.FetchPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx, writer)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context, writer kafkaWriter) {
	events, err := w.repo.FetchBatch(ctx, w.cfg.BatchSize, w.cfg.InProgressTTL)
	if err != nil {
		w.logger.Error("fetch outbox batch", zap.Error(err))
		return
	}

	for _, event := range events {
		orderID, _ := event.Payload["order_id"].(float64)
		userID, _ := event.Payload["user_id"].(float64)
		eventStatus, _ := event.Payload["status"].(string)

		payload := notificationPayload{
			UserID:  int64(userID),
			OrderID: int64(orderID),
			Status:  eventStatus,
		}

		value, err := json.Marshal(payload)
		if err != nil {
			w.logger.Error("marshal payload", zap.Error(err))
			continue
		}

		msg := kafka.Message{
			Key:   []byte(strconv.FormatInt(int64(orderID), 10)),
			Value: value,
		}

		w.logger.Info("writing kafka message",
			zap.Int64("order_id", int64(orderID)),
			zap.Int64("user_id", int64(userID)),
			zap.String("status", eventStatus),
			zap.String("key", string(msg.Key)),
		)

		var writeErr error
		for attempt := 0; attempt < 5; attempt++ {
			writeErr = writer.WriteMessages(ctx, msg)
			if writeErr == nil {
				break
			}
			w.logger.Warn("failed to write kafka message, retrying",
				zap.Error(writeErr),
				zap.Int("attempt", attempt+1),
				zap.Int64("order_id", int64(orderID)))
			time.Sleep(time.Duration(timeDuration*(1<<attempt)) * time.Millisecond)
		}

		if writeErr != nil {
			w.logger.Error("failed to write kafka message after retries",
				zap.Error(writeErr),
				zap.Int64("order_id", int64(orderID)))
			continue
		}

		if err := w.repo.Delete(ctx, event.ID); err != nil {
			w.logger.Error("delete outbox event", zap.Error(err))
		} else {
			w.logger.Info("outbox event processed and deleted",
				zap.Int64("id", event.ID),
				zap.Int64("order_id", int64(orderID)))
		}
	}
}
