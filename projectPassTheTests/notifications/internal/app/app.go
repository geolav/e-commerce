////package app
////
////import (
////	"net"
////
////	"go.uber.org/zap"
////	"google.golang.org/grpc"
////	"google.golang.org/grpc/reflection"
////
////	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
////	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/service"
////	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
////)
////
////func Run(logger *zap.Logger, cfg *config.Config) {
////	notificationsServer := service.NewNotificationsServer(cfg.Clients.CallbackAddr, logger)
////
////	grpcPort := ":" + cfg.GRPC.Port
////	lis, err := net.Listen("tcp", grpcPort)
////	if err != nil {
////		logger.Fatal("failed to listen", zap.Error(err))
////	}
////
////	grpcServer := grpc.NewServer()
////	notificationspb.RegisterNotificationsServer(grpcServer, notificationsServer)
////	reflection.Register(grpcServer)
////
////	logger.Info("notifications gRPC server started", zap.String("port", grpcPort))
////	if err := grpcServer.Serve(lis); err != nil {
////		logger.Fatal("failed to serve", zap.Error(err))
////	}
////}
//
//package app
//
//import (
//	"context"
//	"encoding/json"
//	"strings"
//
//	"github.com/segmentio/kafka-go"
//	"go.uber.org/zap"
//
//	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
//	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/service"
//)
//
//func Run(logger *zap.Logger, cfg *config.Config) {
//	notifier := service.NewNotifier(cfg.Clients.CallbackAddr, logger)
//
//	reader := kafka.NewReader(kafka.ReaderConfig{
//		Brokers: strings.Split(cfg.Kafka.Brokers, ","),
//		Topic:   cfg.Kafka.Topic,
//		GroupID: cfg.Kafka.ConsumerGroup,
//	})
//	defer reader.Close()
//
//	logger.Info("notifications kafka consumer started")
//
//	ctx := context.Background()
//	for {
//		logger.Info("waiting for kafka messages...")
//
//		msg, err := reader.FetchMessage(ctx)
//		if err != nil {
//			logger.Error("fetch kafka message", zap.Error(err))
//			continue
//		}
//
//		logger.Info("received kafka message",
//			zap.String("topic", msg.Topic),
//			zap.Int("partition", msg.Partition),
//			zap.Int64("offset", msg.Offset),
//			zap.ByteString("key", msg.Key),
//			zap.ByteString("value", msg.Value))
//
//		var payload service.NotificationPayload
//		if err := json.Unmarshal(msg.Value, &payload); err != nil {
//			logger.Error("unmarshal message", zap.Error(err))
//			// невалидное сообщение — коммитим чтобы не застрять
//			_ = reader.CommitMessages(ctx, msg)
//			continue
//		}
//
//		logger.Info("unmarshaled payload",
//			zap.Int64("order_id", payload.OrderID),
//			zap.Int64("user_id", payload.UserID),
//			zap.String("status", payload.Status))
//
//		if err := notifier.Notify(ctx, payload); err != nil {
//			// не коммитим — at least once гарантия
//			logger.Warn("failed to notify", zap.Error(err))
//			continue
//		}
//
//		if err := reader.CommitMessages(ctx, msg); err != nil {
//			logger.Error("commit kafka message", zap.Error(err))
//		}
//	}
//}

package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/service"
)

func checkKafkaConnection(brokers []string, topic string) error {
	conn, err := kafka.DialLeader(context.Background(), "tcp", brokers[0], topic, 0)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ReadPartitions(topic)
	return err
}

func Run(logger *zap.Logger, cfg *config.Config) {
	notifier := service.NewNotifier(cfg.Clients.CallbackAddr, logger)

	logger.Info("connecting to kafka",
		zap.String("brokers", cfg.Kafka.Brokers),
		zap.String("topic", cfg.Kafka.Topic),
		zap.String("group", cfg.Kafka.ConsumerGroup))

	var reader *kafka.Reader
	var err error

	// Retry подключения к Kafka с проверкой через DialLeader
	for i := 0; i < 30; i++ {
		// Проверяем подключение к Kafka
		err = checkKafkaConnection(strings.Split(cfg.Kafka.Brokers, ","), cfg.Kafka.Topic)
		if err == nil {
			logger.Info("connected to kafka", zap.Int("attempt", i+1))

			// Создаём reader только после успешного подключения
			reader = kafka.NewReader(kafka.ReaderConfig{
				Brokers:        strings.Split(cfg.Kafka.Brokers, ","),
				Topic:          cfg.Kafka.Topic,
				GroupID:        cfg.Kafka.ConsumerGroup,
				StartOffset:    kafka.LastOffset,
				MinBytes:       1,
				MaxBytes:       10e6,
				MaxWait:        1 * time.Second,
				CommitInterval: 0, // Ручной коммит
			})
			break
		}

		logger.Warn("failed to connect to kafka, retrying",
			zap.Error(err),
			zap.Int("attempt", i+1))
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logger.Fatal("failed to connect to kafka after 30 retries", zap.Error(err))
	}
	defer reader.Close()

	logger.Info("notifications kafka consumer started")

	ctx := context.Background()
	for {
		// Используем таймаут для FetchMessage
		fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		msg, err := reader.FetchMessage(fetchCtx)
		cancel()

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logger.Debug("no messages in kafka, waiting...")
				continue
			}
			logger.Error("fetch kafka message", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		logger.Info("✅ RECEIVED KAFKA MESSAGE",
			zap.Int64("offset", msg.Offset),
			zap.Int("partition", msg.Partition),
			zap.ByteString("key", msg.Key),
			zap.ByteString("value", msg.Value))

		var payload service.NotificationPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			logger.Error("unmarshal message", zap.Error(err))
			// невалидное сообщение — коммитим чтобы не застрять
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		logger.Info("unmarshaled payload",
			zap.Int64("order_id", payload.OrderID),
			zap.Int64("user_id", payload.UserID),
			zap.String("status", payload.Status))

		// Retry с экспоненциальной задержкой
		//for attempt := 0; attempt < 5; attempt++ {
		//	err := notifier.Notify(ctx, payload)
		//	if err == nil {
		//		break
		//	}
		//	logger.Warn("failed to notify, retrying",
		//		zap.Error(err),
		//		zap.Int("attempt", attempt+1),
		//		zap.Int64("order_id", payload.OrderID))
		//	time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
		//}
		//
		//if err := reader.CommitMessages(ctx, msg); err != nil {
		//	logger.Error("failed to commit kafka message", zap.Error(err))
		//	continue
		//}

		const maxAttempts = 10
		const maxDelay = 30 * time.Second

		notified := false
		for attempt := 0; attempt < maxAttempts; attempt++ {
			notifyErr := notifier.Notify(ctx, payload)
			if notifyErr == nil {
				notified = true
				break
			}
			logger.Warn("failed to notify, retrying",
				zap.Error(notifyErr),
				zap.Int("attempt", attempt+1),
				zap.Int64("order_id", payload.OrderID))

			delay := time.Duration(100*(1<<attempt)) * time.Millisecond
			if delay > maxDelay {
				delay = maxDelay
			}
			time.Sleep(delay)
		}

		if !notified {
			// Исчерпали попытки — НЕ коммитим offset.
			// Сообщение будет переполучено после rebalance / перезапуска.
			// Делаем паузу и продолжаем цикл без коммита.
			logger.Warn("exhausted all notify attempts, will retry message later",
				zap.Int64("order_id", payload.OrderID))
			time.Sleep(2 * time.Second)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			logger.Error("failed to commit kafka message", zap.Error(err))
			continue
		}

		logger.Info("message committed",
			zap.Int64("offset", msg.Offset),
			zap.Int("partition", msg.Partition))
	}
}
